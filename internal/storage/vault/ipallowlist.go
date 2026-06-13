package vault

import (
	"net"
	"strings"
)

// CIDRList wraps a slice of CIDR strings and provides a Contains helper.
type CIDRList []string

// Contains reports whether the given IP is matched by any entry in the list.
// An empty / nil list matches nothing. The "no restriction" semantics are
// handled by IsAllowed, not here.
func (l CIDRList) Contains(ip string) bool {
	if len(l) == 0 {
		return false
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	for _, cidr := range l {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			// Allow bare IPs (e.g. "10.0.0.1") as a single-host match.
			if bare := net.ParseIP(cidr); bare != nil {
				if bare.Equal(parsed) {
					return true
				}
			}
			continue
		}
		if network.Contains(parsed) {
			return true
		}
	}
	return false
}

// IsAllowed returns true when the IP is in allowedIPs and not in deniedIPs.
// An empty allowedIPs list means "no restriction" — the IP is allowed
// unless explicitly denied.
func IsAllowed(ip string, allowedIPs, deniedIPs []string) bool {
	if CIDRList(deniedIPs).Contains(ip) {
		return false
	}
	if len(allowedIPs) == 0 {
		return true
	}
	return CIDRList(allowedIPs).Contains(ip)
}
