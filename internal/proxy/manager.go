package proxy

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	xproxy "golang.org/x/net/proxy"
)

const (
	StatusPending = "pending"
	StatusHealthy = "healthy"
	StatusFailed  = "failed"
	StatusBanned  = "banned"
)

var (
	ErrSessionNotFound = errors.New("proxy session not found")
	ErrNoHealthyProxy  = errors.New("no healthy proxy available")
)

type ProbeFunc func(context.Context, string) (time.Duration, error)

type Candidate struct {
	Address       string    `json:"address"`
	Status        string    `json:"status"`
	LatencyMS     int64     `json:"latency_ms"`
	LastError     string    `json:"last_error,omitempty"`
	LastCheckedAt time.Time `json:"last_checked_at,omitempty"`
	FailedCount   int       `json:"failed_count"`
}

type Lease struct {
	SessionID string    `json:"session_id"`
	Address   string    `json:"address"`
	LatencyMS int64     `json:"latency_ms"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Session struct {
	ID            string      `json:"id"`
	PrincipalID   string      `json:"principal_id,omitempty"`
	SelectedProxy string      `json:"selected_proxy,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	Candidates    []Candidate `json:"candidates"`
}

type Manager struct {
	mu            sync.Mutex
	sessions      map[string]*Session
	prober        ProbeFunc
	probeTimeout  time.Duration
	sessionTTL    time.Duration
	banCooldown   time.Duration
	reapFrequency time.Duration
}

func NewManager(prober ProbeFunc) *Manager {
	return &Manager{
		sessions:      make(map[string]*Session),
		prober:        prober,
		probeTimeout:  12 * time.Second,
		sessionTTL:    12 * time.Hour,
		banCooldown:   10 * time.Minute,
		reapFrequency: time.Hour,
	}
}

func (m *Manager) Start(ctx context.Context) {
	ticker := time.NewTicker(m.reapFrequency)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.reapExpired()
		}
	}
}

func (m *Manager) CreateSession(principalID string, rawLines []string) (Session, error) {
	candidates, err := parseCandidates(rawLines)
	if err != nil {
		return Session{}, err
	}
	now := time.Now().UTC()
	session := Session{
		ID:          fmt.Sprintf("proxy-%d", now.UnixNano()),
		PrincipalID: strings.TrimSpace(principalID),
		CreatedAt:   now,
		UpdatedAt:   now,
		Candidates:  candidates,
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.ID] = &session
	return cloneSession(session), nil
}

func (m *Manager) GetSession(id string) (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	session, ok := m.sessions[id]
	if !ok {
		return Session{}, ErrSessionNotFound
	}
	return cloneSession(*session), nil
}

func (m *Manager) CheckSession(ctx context.Context, id string) (Session, error) {
	m.mu.Lock()
	session, ok := m.sessions[id]
	if !ok {
		m.mu.Unlock()
		return Session{}, ErrSessionNotFound
	}
	candidates := append([]Candidate(nil), session.Candidates...)
	m.mu.Unlock()

	if m.prober == nil {
		return Session{}, errors.New("proxy prober is not configured")
	}

	for idx := range candidates {
		checkCtx, cancel := context.WithTimeout(ctx, m.probeTimeout)
		latency, err := m.prober(checkCtx, candidates[idx].Address)
		cancel()
		candidates[idx].LastCheckedAt = time.Now().UTC()
		candidates[idx].LatencyMS = latency.Milliseconds()
		if err != nil {
			candidates[idx].Status = StatusFailed
			candidates[idx].LastError = err.Error()
			continue
		}
		candidates[idx].Status = StatusHealthy
		candidates[idx].LastError = ""
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	session, ok = m.sessions[id]
	if !ok {
		return Session{}, ErrSessionNotFound
	}
	session.Candidates = candidates
	session.UpdatedAt = time.Now().UTC()
	if session.SelectedProxy != "" && !isAddressHealthy(session.Candidates, session.SelectedProxy) {
		session.SelectedProxy = ""
	}
	return cloneSession(*session), nil
}

func (m *Manager) SelectProxy(id, address string) (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	session, ok := m.sessions[id]
	if !ok {
		return Session{}, ErrSessionNotFound
	}

	selected := strings.TrimSpace(address)
	if strings.EqualFold(selected, "fastest") || selected == "" {
		best, err := selectFastest(session.Candidates)
		if err != nil {
			return Session{}, err
		}
		selected = best.Address
	}
	if !isAddressHealthy(session.Candidates, selected) {
		return Session{}, fmt.Errorf("proxy %q is not healthy", selected)
	}
	session.SelectedProxy = selected
	session.UpdatedAt = time.Now().UTC()
	return cloneSession(*session), nil
}

func (m *Manager) AcquireLease(id string) (Lease, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	session, ok := m.sessions[id]
	if !ok {
		return Lease{}, ErrSessionNotFound
	}
	if session.SelectedProxy != "" && isAddressHealthy(session.Candidates, session.SelectedProxy) {
		return buildLease(*session, session.SelectedProxy), nil
	}
	best, err := selectFastest(session.Candidates)
	if err != nil {
		return Lease{}, err
	}
	session.SelectedProxy = best.Address
	session.UpdatedAt = time.Now().UTC()
	return buildLease(*session, best.Address), nil
}

func (m *Manager) ReportFailure(id, address, message string) (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	session, ok := m.sessions[id]
	if !ok {
		return Session{}, ErrSessionNotFound
	}
	for idx := range session.Candidates {
		if session.Candidates[idx].Address != address {
			continue
		}
		session.Candidates[idx].Status = StatusBanned
		session.Candidates[idx].FailedCount++
		session.Candidates[idx].LastError = strings.TrimSpace(message)
		session.Candidates[idx].LastCheckedAt = time.Now().UTC()
		break
	}
	if session.SelectedProxy == address {
		session.SelectedProxy = ""
	}
	if best, err := selectFastest(session.Candidates); err == nil {
		session.SelectedProxy = best.Address
	}
	session.UpdatedAt = time.Now().UTC()
	return cloneSession(*session), nil
}

func parseCandidates(rawLines []string) ([]Candidate, error) {
	seen := make(map[string]struct{})
	candidates := make([]Candidate, 0, len(rawLines))
	for _, raw := range rawLines {
		for _, line := range strings.Split(raw, "\n") {
			address := strings.TrimSpace(line)
			if address == "" {
				continue
			}
			if err := validateAddress(address); err != nil {
				return nil, err
			}
			if _, ok := seen[address]; ok {
				continue
			}
			seen[address] = struct{}{}
			candidates = append(candidates, Candidate{
				Address: address,
				Status:  StatusPending,
			})
		}
	}
	if len(candidates) == 0 {
		return nil, errors.New("proxy list is empty")
	}
	return candidates, nil
}

func validateAddress(address string) error {
	host, port, err := net.SplitHostPort(strings.TrimSpace(address))
	if err != nil {
		return fmt.Errorf("invalid proxy address %q", address)
	}
	if strings.TrimSpace(host) == "" || strings.TrimSpace(port) == "" {
		return fmt.Errorf("invalid proxy address %q", address)
	}
	return nil
}

func selectFastest(candidates []Candidate) (Candidate, error) {
	healthy := make([]Candidate, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.Status == StatusHealthy {
			healthy = append(healthy, candidate)
		}
	}
	if len(healthy) == 0 {
		return Candidate{}, ErrNoHealthyProxy
	}
	slices.SortStableFunc(healthy, func(a, b Candidate) int {
		switch {
		case a.LatencyMS < b.LatencyMS:
			return -1
		case a.LatencyMS > b.LatencyMS:
			return 1
		default:
			return strings.Compare(a.Address, b.Address)
		}
	})
	return healthy[0], nil
}

func isAddressHealthy(candidates []Candidate, address string) bool {
	for _, candidate := range candidates {
		if candidate.Address == address {
			return candidate.Status == StatusHealthy
		}
	}
	return false
}

func buildLease(session Session, address string) Lease {
	lease := Lease{
		SessionID: session.ID,
		Address:   address,
		UpdatedAt: session.UpdatedAt,
	}
	for _, candidate := range session.Candidates {
		if candidate.Address == address {
			lease.LatencyMS = candidate.LatencyMS
			break
		}
	}
	return lease
}

func cloneSession(session Session) Session {
	session.Candidates = append([]Candidate(nil), session.Candidates...)
	return session
}

func (m *Manager) reapExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()
	cutoff := time.Now().UTC().Add(-m.sessionTTL)
	for id, session := range m.sessions {
		if session.UpdatedAt.Before(cutoff) {
			delete(m.sessions, id)
		}
	}
}

func NewHTTPClient(proxyAddress string, timeout time.Duration) (*http.Client, error) {
	transport := &http.Transport{
		TLSHandshakeTimeout:   timeout,
		ResponseHeaderTimeout: timeout,
		IdleConnTimeout:       timeout,
		DisableKeepAlives:     true,
	}
	if strings.TrimSpace(proxyAddress) != "" {
		dialer, err := xproxy.SOCKS5("tcp", proxyAddress, nil, xproxy.Direct)
		if err != nil {
			return nil, err
		}
		transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
			type contextDialer interface {
				DialContext(context.Context, string, string) (net.Conn, error)
			}
			if withContext, ok := dialer.(contextDialer); ok {
				return withContext.DialContext(ctx, network, address)
			}
			return dialer.Dial(network, address)
		}
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}, nil
}
