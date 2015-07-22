package middleware

import "gitlab.doit9.com/server/vertex"

var DefaultMiddleware = []vertex.Middleware{
	AutoRecover,
	RequestLogger,
	CORS,
	&InstrumentationMiddleware{},
}
