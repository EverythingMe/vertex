package middleware

import (
	"net/http"

	"gitlab.doit9.com/server/vertex"
)

//Access-Control-Allow-Origin
// CORS is a middleware that injects Access-Control-Allow-Origin headers
var CORS = vertex.MiddlewareFunc(func(w http.ResponseWriter, r *vertex.Request, next vertex.HandlerFunc) (interface{}, error) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	return next(w, r)
})
