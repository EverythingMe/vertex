package vertex

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	gorilla "github.com/gorilla/schema"

	"github.com/dvirsky/go-pylog/logging"
)

// Headers for responses
const (

	// The POST/GET param we pass if we want a JSONP callback response
	CallbackParam = "callback"

	HeaderProcessingTime = "X-Vertex-ProcessingTime"
	HeaderRequestId      = "X-Vertex-RequestId"
	HeaderHost           = "X-Vertex-Host"
	HeaderServerVersion  = "X-Vertex-Version"
)

// RequestHandler is the interface that request handler structs should implement.
//
// The idea is that you define your request parameters as struct fields, and they get mapped automatically
// and validated, leaving you with just pure logic work.
//
// An example Request handler:
//
//	type UserHandler struct {
//		Id   string `schema:"id" required:"true" doc:"The Id Of the user" maxlen:"20" in:"path"`
//		Name string `schema:"name" maxlen:"100" required:"true" doc:"The Name Of the user"`
//		Admin bool `schema:"bool" default:"true" required:"false" doc:"Is this user an admin"`
//	}
//
//	func (h UserHandler) Handle(w http.ResponseWriter, r *http.Request) (interface{}, error) {
//		return fmt.Sprintf("Your name is %s and id is %s", h.Name, h.Id), nil
//	}
//
// Supported types for automatic param mapping: string, int(32/64), float(32/64), bool, []string
type RequestHandler interface {
	Handle(w http.ResponseWriter, r *Request) (interface{}, error)
}

// HandlerFunc is an adapter that allows you to register normal functions as handlers. It is used mainly by middleware
// and should not be used in an application context
type HandlerFunc func(http.ResponseWriter, *Request) (interface{}, error)

// Handle calls the underlying function
func (h HandlerFunc) Handle(w http.ResponseWriter, r *Request) (interface{}, error) {
	return h(w, r)
}

// Unmarshaler is an interface for types who are interested in automatic decoding.
// The unmarshaler should return a new instance of itself with the value set correctly.
//
// Example: a type that takes a string and splits in two
//	type Banana struct {
//		Foo string
//		Bar string
//	}
//
//	func (b Banana) UnmarshalRequestData(data string) interface{} {
//		parts := strings.Split(data, ",")
//		if len(parts) == 2 {
//			return Banana{parts[0], parts[1]}
//		}
//		return Banana{}
//	}
type Unmarshaler interface {
	UnmarshalRequestData(data string) interface{}
}

// Params are a string map for path formatting
type Params map[string]string

// FormatPath takes a path template and formats it according to the given path params
//
// e.g.
//	FormatPath("/foo/{id}", Params{"id":"bar"})
//  // Output: "/foo/bar"
func FormatPath(path string, params Params) string {

	if params != nil {
		for k, v := range params {
			path = strings.Replace(path, fmt.Sprintf("{%s}", k), v, -1)
		}
	}
	return path
}

// StaticHandler is a batteries-included handler for serving static files inside a directory.
//
// root is the path the root path for this static handler, and will get stripped.
//
// NOTE: root should be the full path to the API root. so if your handler path is "/static/*filepath",
// root should be something like "/myapi/1.0/static".
// Because the handler is created before the API object is configured, we do not know the root on creation
func StaticHandler(root string, dir http.Dir) RequestHandler {

	h := http.StripPrefix(root, http.FileServer(dir))

	return HandlerFunc(func(w http.ResponseWriter, r *Request) (interface{}, error) {

		h.ServeHTTP(w, r.Request)
		return nil, Hijacked

	})
}

// VoidHandler is a batteries-included handler that does nothing, useful for testing, or when
// a middleware takes over the request completely
type VoidHandler struct{}

// Handle does nothing :)
func (VoidHandler) Handle(w http.ResponseWriter, r *Request) (interface{}, error) {
	return nil, nil
}

// SecurityScheme is a special interface that validates a request and is outside the middleware chain.
// An API has a default security scheme, and each route can override it
type SecurityScheme interface {
	Validate(r *Request) error
}

// MethodFlag is used for const flags for method handling on API declaration
type MethodFlag int

// Method flag definitions
const (
	GET  MethodFlag = 0x01
	POST MethodFlag = 0x02
	PUT  MethodFlag = 0x03
)

var schemaDecoder = gorilla.NewDecoder()

// Parse the user input into a request handler struct, with input validation
func parseInput(r *http.Request, input interface{}, validator *RequestValidator) error {

	schemaDecoder.IgnoreUnknownKeys(true)

	if err := r.ParseForm(); err != nil {
		return InvalidRequestError("Error parsing request data: %s", err)
	}

	// We do not map and validate input to non-struct handlers
	if reflect.TypeOf(input).Kind() != reflect.Func {

		if err := schemaDecoder.Decode(input, r.Form); err != nil {
			return InvalidRequestError("Error decoding schema: %s", err)
		}

		// Validate the input based on the API spec
		if err := validator.Validate(input, r); err != nil {
			logging.Error("Error validating http.Request!: %s", err)
			return NewError(err)

		}

	}

	return nil

}

// FormValueDefault returns the value from  a form param, with an optional default argument if the value was not set
func formValueDefault(r *Request, key, def string) string {
	ret := r.FormValue(key)
	if ret == "" {
		return def
	}
	return strings.Replace(ret, "\"", "", -1)
}
