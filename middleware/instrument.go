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
	defer instrument.SampleTime(p)()

	v, err = next(w, r)
	if err != nil {
		p = p + ".failure"
	} else {
		p = p + ".success"
	}

	instrument.Hit(p)
	return v, err

}
