package middleware

import (
	"net/http"

	"github.com/dvirsky/go-pylog/logging"

	"github.com/EverythingMe/vertex"
)

// BasicAuth is a middleware that forces basic auth user/pass authentication on requests.
//
// When creating the auth middleware, give it a user/pass/realm config, and this is what it will validate
type BasicAuth struct {
	User           string
	Password       string
	Realm          string
	BypassForLocal bool
}

func (b BasicAuth) requireAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="`+b.Realm+`"`)
	w.WriteHeader(401)
	w.Write([]byte("401 Unauthorized\n"))
}

func (b BasicAuth) Handle(w http.ResponseWriter, r *vertex.Request, next vertex.HandlerFunc) (interface{}, error) {

	if !r.IsLocal() || !b.BypassForLocal {
		user, pass, ok := r.BasicAuth()
		if !ok {
			logging.Debug("No auth header, denying")
			b.requireAuth(w)
			return nil, vertex.Hijacked
		}

		if user != b.User || pass != b.Password {
			logging.Warning("Unmatching auth: %s/%s", user, pass)
			b.requireAuth(w)
			return nil, vertex.Hijacked
		}
	}

	return next(w, r)
}
