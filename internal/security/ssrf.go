package security

import (
	"net"
	"net/url"
	"strings"
)

// ValidateTargetURL rejects private, loopback, and link-local targets (SSRF mitigation).
func ValidateTargetURL(raw string) bool {
	u, err := url.ParseRequestURI(strings.TrimSpace(raw))
	if err != nil || u.Host == "" {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}

	host := strings.ToLower(u.Hostname())
	if host == "" {
		return false
	}
	if blockedHostname(host) {
		return false
	}

	if ip := net.ParseIP(host); ip != nil {
		return !blockedIP(ip)
	}
	return true
}

func blockedHostname(host string) bool {
	switch host {
	case "localhost", "0.0.0.0", "::", "::1":
		return true
	}
	if strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".internal") {
		return true
	}
	if strings.HasSuffix(host, ".localhost") {
		return true
	}
	return false
}

func blockedIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsPrivate() {
		return true
	}
	if ip.IsUnspecified() {
		return true
	}
	// Unique local IPv6 fc00::/7
	if ip.To4() == nil && len(ip) == net.IPv6len && ip[0]&0xfe == 0xfc {
		return true
	}
	return false
}
