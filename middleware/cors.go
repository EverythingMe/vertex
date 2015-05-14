package middleware

import (
	"net/http"

	"gitlab.doit9.com/backend/web2"
)

//Access-Control-Allow-Origin

var CORS = web2.MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next web2.HandlerFunc) (interface{}, error) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	return next(w, r)
})
