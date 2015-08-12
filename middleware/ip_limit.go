package middleware

import (
	"net"
	"net/http"

	"github.com/dvirsky/go-pylog/logging"

	"gitlab.doit9.com/server/vertex"
)

// IPRangeFilter allows or denies ips based on a given set of IP ranges (CIDRs)
type IPRangeFilter struct {
	allowed []*net.IPNet
	denied  []*net.IPNet
}

// NewIPRangeFilter creates a new filter with the given allowed CIDRs (e.g. 127.0.0.0/8 for local addresses)
func NewIPRangeFilter(allowed ...string) *IPRangeFilter {
	ret := &IPRangeFilter{
		allowed: make([]*net.IPNet, 0, len(allowed)),
	}

	ret.Allow(allowed...)
	return ret
}

// AlloPrivate allows IP ranges from all private ranges according to RFC 1918
func (f *IPRangeFilter) AlloPrivate() *IPRangeFilter {
	f.Allow("10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "127.0.0.0/8")
	return f
}

// Allow allows traffic from the given allowed CIDRs
func (f *IPRangeFilter) Allow(cidrs ...string) *IPRangeFilter {
	f.allowed = make([]*net.IPNet, 0, len(cidrs))
	for _, addr := range cidrs {

		_, ipnet, err := net.ParseCIDR(addr)
		if err != nil {
			logging.Error("Error parsing CIDR: %s", err)
			continue
		}
		logging.Info("Allowing traffic from %s", ipnet)
		f.allowed = append(f.allowed, ipnet)

	}

	return f
}

// Deny denies traffic from the given CIDRs (e.g. 127.0.0.0/8 for local addresses)
func (f *IPRangeFilter) Deny(cidrs ...string) *IPRangeFilter {
	f.denied = make([]*net.IPNet, 0, len(cidrs))

	for _, addr := range cidrs {

		_, ipnet, err := net.ParseCIDR(addr)
		if err != nil {
			logging.Error("Error parsing CIDR: %s", err)
			continue
		}
		f.denied = append(f.denied, ipnet)

	}
	return f
}

// Handle checks the current requests IP against the allowed and blocked IP ranges in the filter
func (f *IPRangeFilter) Handle(w http.ResponseWriter, r *vertex.Request, next vertex.HandlerFunc) (interface{}, error) {
	ip := net.ParseIP(r.RemoteIP)

	if f.denied != nil {
		for _, ipnet := range f.denied {
			if ipnet.Contains(ip) {
				return nil, vertex.UnauthorizedError("IP Address %s blocked", r.RemoteIP)
			}
		}

	}

	for _, ipnet := range f.allowed {
		if ipnet.Contains(ip) {
			logging.Info("IP Address %s allowed", r.RemoteIP)
			return next(w, r)
		}

	}
	return nil, vertex.UnauthorizedError("IP Address %s not allowed", r.RemoteIP)
}
