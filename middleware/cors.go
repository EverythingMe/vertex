package middleware

import (
	"net/http"
	"strings"

	"github.com/EverythingMe/vertex"
)

type CORS struct {
	AllowOrigin      string
	exposeHeaders    []string
	allowHeaders     []string
	allowMethods     []string
	allowCredentials bool
}

//Access-Control-Allow-Origin

// CORS is a middleware that injects Access-Control-Allow-Origin headers
func (c *CORS) Handle(w http.ResponseWriter, r *vertex.Request, next vertex.HandlerFunc) (interface{}, error) {
	if c.AllowOrigin != "" {
		w.Header().Set("Access-Control-Allow-Origin", c.AllowOrigin)
	}

	if c.allowHeaders != nil && len(c.allowHeaders) > 0 {
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(c.allowHeaders, ","))
	}

	if c.exposeHeaders != nil && len(c.exposeHeaders) > 0 {
		w.Header().Set("Access-Control-Expose-Headers", strings.Join(c.exposeHeaders, ","))
	}

	if c.allowMethods != nil && len(c.allowMethods) > 0 {
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(c.allowMethods, ","))
	}

	if c.allowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	return next(w, r)
}

func NewCORS() *CORS {
	return &CORS{
		AllowOrigin: "*",
	}
}

// Default initializes a generic CORS configuration
func (c *CORS) Default() *CORS {
	c.ExposeHeaders("WWW-Authenticate", "Authorization")
	c.AllowHeaders(c.exposeHeaders...)
	c.allowCredentials = true
	c.AllowMethods("GET", "POST", "OPTIONS", "PUT")
	return c
}

func (c *CORS) ExposeHeaders(headers ...string) *CORS {
	c.exposeHeaders = headers
	return c
}

func (c *CORS) AllowHeaders(headers ...string) *CORS {
	c.allowHeaders = headers
	return c
}

func (c *CORS) AllowCredentials(allow bool) *CORS {
	c.allowCredentials = allow
	return c
}

func (c *CORS) AllowMethods(methods ...string) *CORS {
	c.allowMethods = methods
	return c
}
