package middleware

import (
	"net/http"

	"github.com/dvirsky/go-pylog/logging"
	"gitlab.doit9.com/backend/web2"
)

var RequestLogger = web2.MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next web2.HandlerFunc) (interface{}, error) {

	logging.Info("Handling %s %s", r.Method, r.URL.String())

	ret, err := next(w, r)

	logging.Info("Return value was %v %v", ret, err)
	return ret, err
})
