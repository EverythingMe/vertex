package vertex

import (
	"reflect"

	"github.com/dvirsky/go-pylog/logging"
	gorilla "github.com/gorilla/schema"

	"gitlab.doit9.com/server/vertex/schema"
)

// A routing map for an API
type Routes []Route

// Route represents a single route (path) in the API and its handler and optional extra middleware
type Route struct {
	Path        string
	Description string
	Handler     RequestHandler
	Methods     MethodFlag
	Security    SecurityScheme
	Middleware  []Middleware
	Test        Tester
	Returns     interface{}
	requestInfo schema.RequestInfo
}

func (r *Route) parseInfo(path string) error {

	ri, err := schema.NewRequestInfo(reflect.TypeOf(r.Handler), path, r.Description, r.Returns)
	if err != nil {
		return err
	}

	// search for custom unmarshallers in the request info
	for _, param := range ri.Params {
		if param.Type.Kind() == reflect.Struct {

			logging.Debug("Checking unmarshaller for %s", param.Type)
			val := reflect.Zero(param.Type).Interface()

			if unm, ok := val.(Unmarshaler); ok {
				logging.Info("Registering unmarshaller for %#v", val)

				schemaDecoder.RegisterConverter(val, gorilla.Converter(func(s string) reflect.Value {
					return reflect.ValueOf(unm.UnmarshalRequestData(s))
				}))

			}
		}

	}

	r.requestInfo = ri
	return nil

}
