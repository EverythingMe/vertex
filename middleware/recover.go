package middleware

import (
	"net/http"

	"github.com/dvirsky/go-pylog/logging"

	"gitlab.doit9.com/server/vertex"
)

var AutoRecover = vertex.MiddlewareFunc(func(w http.ResponseWriter, r *vertex.Request, next vertex.HandlerFunc) (ret interface{}, err error) {

	defer func() {

		e := recover()
		if e != nil {
			logging.Critical("Caught panic: %v", e)

			err = vertex.NewErrorf("PANIC handling %s: %s", r.URL.Path, e)
			return
		}
	}()

	return next(w, r)

})
