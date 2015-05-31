package middleware

import (
	"net/http"

	"github.com/dvirsky/go-pylog/logging"

	"gitlab.doit9.com/server/vertex"
)

type BasicAuth struct {
	User     string
	Password string
	Realm    string
}

func (b BasicAuth) requireAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="`+b.Realm+`"`)
	w.WriteHeader(401)
	w.Write([]byte("401 Unauthorized\n"))
}

func (b BasicAuth) Handle(w http.ResponseWriter, r *vertex.Request, next vertex.HandlerFunc) (interface{}, error) {
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

	return next(w, r)
}
