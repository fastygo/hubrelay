package outbound

import (
	"errors"
	"strings"

	proxymgr "sshbot/internal/proxy"
)

var (
	ErrProxyRequired        = errors.New("proxy session is required by outbound policy")
	ErrProxyManagerRequired = errors.New("proxy session manager is required by outbound policy")
	ErrProxyLeaseAddress    = errors.New("proxy lease has no address")
)

type Policy struct {
	RequireProxy bool
}

type LeaseResolver interface {
	ResolveProxyAddress(sessionID string) (string, error)
}

type ProxyLeaseResolver struct {
	Manager *proxymgr.Manager
}

func (r ProxyLeaseResolver) ResolveProxyAddress(sessionID string) (string, error) {
	if r.Manager == nil {
		return "", ErrProxyManagerRequired
	}
	lease, err := r.Manager.AcquireLease(sessionID)
	if err != nil {
		return "", err
	}
	address := strings.TrimSpace(lease.Address)
	if address == "" {
		return "", ErrProxyLeaseAddress
	}
	return address, nil
}

func (p Policy) ResolveProxyAddress(proxySessionID string, resolver LeaseResolver) (string, error) {
	sessionID := strings.TrimSpace(proxySessionID)
	if sessionID != "" {
		if resolver == nil {
			return "", ErrProxyManagerRequired
		}
		return resolver.ResolveProxyAddress(sessionID)
	}
	if p.RequireProxy {
		if resolver == nil {
			return "", ErrProxyManagerRequired
		}
		return "", ErrProxyRequired
	}
	return "", nil
}
