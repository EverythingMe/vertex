package middleware

import (
	"net/http"

	"gitlab.doit9.com/backend/vertex"
)

//Access-Control-Allow-Origin

var CORS = vertex.MiddlewareFunc(func(w http.ResponseWriter, r *vertex.Request, next vertex.HandlerFunc) (interface{}, error) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	return next(w, r)
})
