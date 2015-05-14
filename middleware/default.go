package middleware

import "gitlab.doit9.com/backend/web2"

var DefaultMiddleware = []web2.Middleware{
	AutoRecover,
	RequestLogger,
	CORS,
}
