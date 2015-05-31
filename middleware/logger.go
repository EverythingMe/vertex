package middleware

import (
	"net/http"

	"github.com/dvirsky/go-pylog/logging"
	"gitlab.doit9.com/server/vertex"
)

var RequestLogger = vertex.MiddlewareFunc(func(w http.ResponseWriter, r *vertex.Request, next vertex.HandlerFunc) (interface{}, error) {

	logging.Info("Handling %s %s", r.Method, r.URL.String())

	ret, err := next(w, r)

	logging.Info("Return value was %v %v", ret, err)
	return ret, err
})

//func StaticText(msg string) vertex.MiddlewareFunc {

//	//h := http.FileServer(dir)
//	return HandlerFunc(func(w http.ResponseWriter, r *http.Request) (interface{}, error) {

//		fmt.Fprintln(w, msg)
//		return nil, ErrHijacked

//	})
//}
