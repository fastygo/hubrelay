package outbound

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

var (
	ErrPrivateEgressInterfaceRequired = errors.New("private egress interface is required")
	ErrPrivateEgressInterfaceDown     = errors.New("private egress interface is down")
	ErrPrivateEgressRouteMismatch     = errors.New("private egress test host does not route through the private interface")
)

type PrivateEgressChecker interface {
	Check(context.Context, string, string) error
}

type InterfaceEgressChecker struct {
	interfaceByName func(string) (*net.Interface, error)
	lookupIPs       func(context.Context, string) ([]net.IP, error)
	localIPForIP    func(context.Context, net.IP) (net.IP, error)
}

func NewInterfaceEgressChecker() *InterfaceEgressChecker {
	return &InterfaceEgressChecker{
		interfaceByName: net.InterfaceByName,
		lookupIPs:       lookupHostIPs,
		localIPForIP:    localIPForTarget,
	}
}

func (c *InterfaceEgressChecker) Check(ctx context.Context, interfaceName, testHost string) error {
	name := strings.TrimSpace(interfaceName)
	if name == "" {
		return ErrPrivateEgressInterfaceRequired
	}
	if c.interfaceByName == nil || c.lookupIPs == nil || c.localIPForIP == nil {
		return errors.New("private egress checker is not fully configured")
	}

	iface, err := c.interfaceByName(name)
	if err != nil {
		return err
	}
	if iface.Flags&net.FlagUp == 0 {
		return fmt.Errorf("%w: %s", ErrPrivateEgressInterfaceDown, name)
	}
	if strings.TrimSpace(testHost) == "" {
		return nil
	}

	targets, err := c.lookupIPs(ctx, testHost)
	if err != nil {
		return err
	}
	ifaceIPs, err := interfaceIPs(iface)
	if err != nil {
		return err
	}
	for _, target := range targets {
		localIP, localErr := c.localIPForIP(ctx, target)
		if localErr != nil {
			continue
		}
		if containsIP(ifaceIPs, localIP) {
			return nil
		}
	}
	return fmt.Errorf("%w: host=%s interface=%s", ErrPrivateEgressRouteMismatch, testHost, name)
}

var lookupHostIPs = func(ctx context.Context, host string) ([]net.IP, error) {
	trimmed := strings.TrimSpace(host)
	if ip := net.ParseIP(trimmed); ip != nil {
		return []net.IP{ip}, nil
	}
	return net.DefaultResolver.LookupIP(ctx, "ip", trimmed)
}

var localIPForTarget = func(ctx context.Context, target net.IP) (net.IP, error) {
	dialer := net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(ctx, "udp", net.JoinHostPort(target.String(), "9"))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	udpAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok || udpAddr.IP == nil {
		return nil, errors.New("private egress local address is not UDP")
	}
	return udpAddr.IP, nil
}

var interfaceIPs = func(iface *net.Interface) ([]net.IP, error) {
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
