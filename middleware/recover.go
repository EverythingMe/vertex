package middleware

import (
	"fmt"
	"net/http"

	"github.com/dvirsky/go-pylog/logging"

	"gitlab.doit9.com/backend/web2"
)

var AutoRecover = web2.MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next web2.HandlerFunc) (ret interface{}, err error) {

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
