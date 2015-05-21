package middleware

import "gitlab.doit9.com/backend/vertex"

var DefaultMiddleware = []vertex.Middleware{
	AutoRecover,
	RequestLogger,
	CORS,
}
