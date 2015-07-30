package middleware

import (
	"net/http"

	"gitlab.doit9.com/server/vertex"
)

type ForceSecure struct {
	AllowLocalInsecure bool
}

// ForceSecure validates that a request is sent over SSL regardless of the global API config
func (f ForceSecure) Handle(w http.ResponseWriter, r *vertex.Request, next vertex.HandlerFunc) (interface{}, error) {

	if !r.Secure {

		if !r.IsLocal() || !f.AllowLocalInsecure {

			return nil, vertex.UnauthorizedError("Insecure Access Forbidden")
		}
	}

	return next(w, r)
}
