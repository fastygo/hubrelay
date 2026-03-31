package egress

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	healthLevelUnknown   = "unknown"
	healthLevelDisabled  = "disabled"
	healthLevelWG        = "wg"
	healthLevelTransport = "transport"
	healthLevelHealthy   = "healthy"
)

type Checker struct {
	interfaceByName  func(string) (*net.Interface, error)
	interfaceIPs     func(*net.Interface) ([]net.IP, error)
	lookupIPs        func(context.Context, string) ([]net.IP, error)
	localIPForIP     func(context.Context, net.IP) (net.IP, error)
	transportProbe   func(context.Context, Gateway) error
	businessProbe    func(context.Context, Gateway) error
	now              func() time.Time
	transportTimeout time.Duration
	businessTimeout  time.Duration
}

func NewChecker() *Checker {
	return &Checker{
		interfaceByName:  net.InterfaceByName,
		interfaceIPs:     collectInterfaceIPs,
		lookupIPs:        lookupHostIPs,
		localIPForIP:     localIPForTarget,
		now:              func() time.Time { return time.Now().UTC() },
		transportTimeout: 5 * time.Second,
		businessTimeout:  10 * time.Second,
	}
}

func (c *Checker) CheckGateway(ctx context.Context, gateway Gateway) GatewayStatus {
	status := GatewayStatus{
		Name:        gateway.Name,
		Interface:   gateway.Interface,
		Priority:    gateway.Priority,
		Enabled:     gateway.Enabled,
		HealthLevel: healthLevelUnknown,
	}

	if !gateway.Enabled {
		status.HealthLevel = healthLevelDisabled
		return status
	}

	wgErr := c.checkWG(ctx, gateway)
	status.Levels.WG = c.levelStatus(wgErr)
	if wgErr != nil {
		status.LastError = wgErr.Error()
		status.HealthLevel = healthLevelUnknown
		return status
	}
	status.HealthLevel = healthLevelWG

	transportErr := c.checkTransport(ctx, gateway)
	status.Levels.Transport = c.levelStatus(transportErr)
	if transportErr != nil {
		status.LastError = transportErr.Error()
		return status
	}
	status.HealthLevel = healthLevelTransport

	businessErr := c.checkBusiness(ctx, gateway)
	status.Levels.Business = c.levelStatus(businessErr)
	if businessErr != nil {
		status.LastError = businessErr.Error()
		return status
	}

	status.HealthLevel = healthLevelHealthy
	status.Healthy = true
	return status
}

func (c *Checker) levelStatus(err error) LevelStatus {
	status := LevelStatus{
		CheckedAt: c.now(),
	}
	if err == nil {
		status.OK = true
		return status
	}
	status.Error = err.Error()
	return status
}

func (c *Checker) checkWG(ctx context.Context, gateway Gateway) error {
	iface, err := c.interfaceByName(strings.TrimSpace(gateway.Interface))
	if err != nil {
		return err
	}
	if iface.Flags&net.FlagUp == 0 {
		return fmt.Errorf("interface %s is down", gateway.Interface)
	}
	if strings.TrimSpace(gateway.TestHost) == "" {
		return nil
	}

	targets, err := c.lookupIPs(ctx, gateway.TestHost)
	if err != nil {
		return err
	}
	interfaceIPs, err := c.interfaceIPs(iface)
	if err != nil {
		return err
	}
	for _, target := range targets {
		localIP, localErr := c.localIPForIP(ctx, target)
		if localErr != nil {
			continue
		}
		if containsIP(interfaceIPs, localIP) {
			return nil
		}
	}
	return fmt.Errorf("test host %s does not route through %s", gateway.TestHost, gateway.Interface)
}

func (c *Checker) checkTransport(ctx context.Context, gateway Gateway) error {
	if c.transportProbe != nil {
		return c.transportProbe(ctx, gateway)
	}
	probeURL, err := parseProbeURL(gateway.ProbeURL)
	if err != nil {
		return err
	}
	address, err := resolvedAddress(probeURL)
	if err != nil {
		return err
	}

	dialer, err := c.boundDialer(ctx, gateway.Interface, c.transportTimeout)
	if err != nil {
		return err
	}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return err
	}
	defer conn.Close()

	if probeURL.Scheme != "https" {
		return nil
	}

	serverName := probeURL.Hostname()
	if serverName == "" {
		return fmt.Errorf("probe URL hostname is required")
	}
	tlsConn := tls.Client(conn, &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: serverName,
	})
	defer tlsConn.Close()

	transportCtx, cancel := context.WithTimeout(ctx, c.transportTimeout)
	defer cancel()
	if err := tlsConn.HandshakeContext(transportCtx); err != nil {
		return err
	}
	return nil
}

func (c *Checker) checkBusiness(ctx context.Context, gateway Gateway) error {
	if c.businessProbe != nil {
		return c.businessProbe(ctx, gateway)
	}
	probeURL, err := parseProbeURL(gateway.ProbeURL)
	if err != nil {
		return err
	}

	dialer, err := c.boundDialer(ctx, gateway.Interface, c.businessTimeout)
	if err != nil {
		return err
	}
	transport := &http.Transport{
		Proxy: nil,
		DialContext: func(dialCtx context.Context, network, address string) (net.Conn, error) {
			return dialer.DialContext(dialCtx, network, address)
		},
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			ServerName: probeURL.Hostname(),
		},
	}
	defer transport.CloseIdleConnections()

	client := &http.Client{
		Timeout:   c.businessTimeout,
		Transport: transport,
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodHead, probeURL.String(), nil)
	if err != nil {
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 1024))
	if response.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("business probe returned status %d", response.StatusCode)
	}
	return nil
}

func (c *Checker) boundDialer(ctx context.Context, interfaceName string, timeout time.Duration) (*net.Dialer, error) {
	iface, err := c.interfaceByName(strings.TrimSpace(interfaceName))
	if err != nil {
		return nil, err
	}
	ips, err := c.interfaceIPs(iface)
	if err != nil {
		return nil, err
	}
	localIP := firstUsableIP(ips)
	if localIP == nil {
		return nil, fmt.Errorf("interface %s has no usable IP", interfaceName)
	}
	return &net.Dialer{
		Timeout:   timeout,
		LocalAddr: &net.TCPAddr{IP: localIP},
	}, nil
}

func parseProbeURL(raw string) (*url.URL, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, fmt.Errorf("probe URL is required")
	}
	probeURL, err := url.Parse(trimmed)
	if err != nil {
		return nil, err
	}
	if probeURL.Scheme != "https" && probeURL.Scheme != "http" {
		return nil, fmt.Errorf("probe URL must use http or https")
	}
	if probeURL.Hostname() == "" {
		return nil, fmt.Errorf("probe URL host is required")
	}
	return probeURL, nil
}

func resolvedAddress(probeURL *url.URL) (string, error) {
	port := probeURL.Port()
	if port == "" {
		switch probeURL.Scheme {
		case "https":
			port = "443"
		case "http":
			port = "80"
		default:
			return "", fmt.Errorf("probe URL port is required")
		}
	}
	return net.JoinHostPort(probeURL.Hostname(), port), nil
}

func lookupHostIPs(ctx context.Context, host string) ([]net.IP, error) {
	trimmed := strings.TrimSpace(host)
	if ip := net.ParseIP(trimmed); ip != nil {
		return []net.IP{ip}, nil
	}
	return net.DefaultResolver.LookupIP(ctx, "ip", trimmed)
}

func localIPForTarget(ctx context.Context, target net.IP) (net.IP, error) {
	dialer := net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(ctx, "udp", net.JoinHostPort(target.String(), "9"))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	udpAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok || udpAddr.IP == nil {
		return nil, fmt.Errorf("local address is not UDP")
	}
	return udpAddr.IP, nil
}

func collectInterfaceIPs(iface *net.Interface) ([]net.IP, error) {
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}
	ips := make([]net.IP, 0, len(addrs))
	for _, addr := range addrs {
		switch value := addr.(type) {
		case *net.IPNet:
			if value.IP != nil {
				ips = append(ips, value.IP)
			}
		case *net.IPAddr:
			if value.IP != nil {
				ips = append(ips, value.IP)
			}
		}
	}
	return ips, nil
}

func containsIP(candidates []net.IP, target net.IP) bool {
	for _, candidate := range candidates {
		if candidate.Equal(target) {
			return true
		}
	}
	return false
}

func firstUsableIP(candidates []net.IP) net.IP {
	for _, candidate := range candidates {
		if candidate == nil || candidate.IsLoopback() || candidate.IsUnspecified() {
			continue
		}
		return candidate
	}
	return nil
}
