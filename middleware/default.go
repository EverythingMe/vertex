package middleware

import "github.com/EverythingMe/vertex"

// DefaultMiddleware is a quick set-up of the default middleware - logger, recover, CORS
var DefaultMiddleware = []vertex.Middleware{
	AutoRecover,
	RequestLogger,
	NewCORS().Default(),
}
