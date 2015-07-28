package middleware

import (
	"net/http"
	"strings"

	"gitlab.doit9.com/backend/instrument"
	"gitlab.doit9.com/server/vertex"
)

// InstrumentationMiddleware is a middleware that instruments request times and success/failure rate
type InstrumentationMiddleware struct {
}

func (i *InstrumentationMiddleware) Handle(w http.ResponseWriter, r *vertex.Request, next vertex.HandlerFunc) (interface{}, error) {

	p := strings.Join(strings.Split(strings.Trim(r.Request.URL.Path, "/"), "/"), ".")

	var v interface{}
	var err error

	instrument.Profile(p, func() error {
		v, err = next(w, r)
		return err
	})

	return v, err

}
