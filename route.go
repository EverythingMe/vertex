package web2

import (
	"reflect"

	"gitlab.doit9.com/backend/web2/schema"
)

// A routing map for an API
type RouteMap map[string]Route

// Route represents a single route (path) in the API and its handler and optional extra middleware
type Route struct {
	Description string
	Handler     RequestHandler
	Methods     MethodFlag
	Security    SecurityScheme
	Middleware  []Middleware
	requestInfo schema.RequestInfo
}

func (r *Route) parseInfo(path string) error {

	ri, err := schema.NewRequestInfo(reflect.TypeOf(r.Handler), path, r.Description)
	if err != nil {
		return err
	}
	r.requestInfo = ri
	return nil

}
