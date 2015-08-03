package middleware

import (
	"net/http"

	"gitlab.doit9.com/server/vertex"
)

type IPAddressFilter struct {
	allowed map[string]struct{}
	denied  map[string]struct{}
}

func NewIPAddressFilter(allowed ...string) *IPAddressFilter {
	ret := &IPAddressFilter{
		allowed: map[string]struct{}{},
	}

	for _, addr := range allowed {
		ret.allowed[addr] = struct{}{}
	}
	return ret
}

func (f *IPAddressFilter) Allow(addrs ...string) {
	f.allowed = map[string]struct{}{}

	for _, addr := range addrs {
		f.allowed[addr] = struct{}{}
	}
}

func (f *IPAddressFilter) Deny(addrs ...string) {
	f.denied = map[string]struct{}{}

	for _, addr := range addrs {
		f.denied[addr] = struct{}{}
	}
}

//Access-Control-Allow-Origin
// CORS is a middleware that injects Access-Control-Allow-Origin headers
func (f *IPAddressFilter) Handle(w http.ResponseWriter, r *vertex.Request, next vertex.HandlerFunc) (interface{}, error) {
	if f.denied != nil {
		if _, found := f.denied[r.RemoteIP]; found {
			return nil, vertex.UnauthorizedError("IP Address %s blocked", r.RemoteIP)
		}
	}

	if _, found := f.allowed[r.RemoteIP]; !found {
		return nil, vertex.UnauthorizedError("IP Address %s blocked", r.RemoteIP)
	}
	return next(w, r)
}
