package middleware

import (
	"fmt"
	"net/http"

	"github.com/dvirsky/go-pylog/logging"

	"gitlab.doit9.com/backend/vertex"
)

var AutoRecover = vertex.MiddlewareFunc(func(w http.ResponseWriter, r *vertex.Request, next vertex.HandlerFunc) (ret interface{}, err error) {

	defer func() {

		e := recover()
		if e != nil {
			logging.Critical("Caught panic: %v", e)

			err = fmt.Errorf("PANIC handling %s: %s", r.URL.Path, e)
			return
		}
	}()

	return next(w, r)

})
