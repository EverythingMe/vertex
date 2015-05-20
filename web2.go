package web2

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	gorilla "github.com/gorilla/schema"
	"gitlab.doit9.com/backend/web2/schema"

	"github.com/dvirsky/go-pylog/logging"
)

// This wraps every response object, with an error code and processing time info.
// It is compatible with a gondor response object.
type Response struct {
	ErrorString    string      `json:"errorString"`
	ErrorCode      int         `json:"errorCode"`
	ProcessingTime float64     `json:"processingTime"`
	RequestId      string      `json"requestId"`
	ResponseObject interface{} `json:"response,omitempty"`
}

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
//      Admin bool `schema:"bool" default:"true" required:"false" doc:"Is this user an admin"`
//	}
//
//	func (h UserHandler) Handle(w http.ResponseWriter, r *http.Request) (interface{}, error) {
//		return fmt.Sprintf("Your name is %s and id is %s", h.Name, h.Id), nil
//	}
//
// Supported types for automatic param mapping: string, int(32/64), float(32/64), bool, []string
type RequestHandler interface {
	Handle(w http.ResponseWriter, r *http.Request) (interface{}, error)
}

// HandlerFunc is an adapter that allows you to register normal functions as handlers. It is used mainly by middleware
// and should not be used in an application context
type HandlerFunc func(http.ResponseWriter, *http.Request) (interface{}, error)

// Handle calls the underlying function
func (h HandlerFunc) Handle(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return h(w, r)
}

type Params map[string]string

func FormatPath(path string, params Params) string {

	if params != nil {
		for k, v := range params {
			path = strings.Replace(path, fmt.Sprintf("{%s}", k), v, -1)
		}
	}
	return path
}

func StaticHanlder(root string, dir http.Dir) RequestHandler {

	h := http.StripPrefix(root, http.FileServer(dir))

	return HandlerFunc(func(w http.ResponseWriter, r *http.Request) (interface{}, error) {

		h.ServeHTTP(w, r)
		return nil, ErrHijacked

	})
}

var VoidHandler = HandlerFunc(func(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return nil, nil
})

func StaticText(message string) RequestHandler {

	return HandlerFunc(func(w http.ResponseWriter, r *http.Request) (interface{}, error) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, message)
		return nil, ErrHijacked

	})
}

// SecurityScheme is a special interface that validates a request and is outside the middleware chain.
// An API has a default security scheme, and each route can override it
type SecurityScheme interface {
	Validate(r *http.Request) error
}

// Flags for method handling on API declaration
type MethodFlag int

// Method flag definitions
const (
	GET  MethodFlag = 0x01
	POST MethodFlag = 0x02
	PUT  MethodFlag = 0x03
)

var schemaDecoder = gorilla.NewDecoder()

// Parse the user input into a request handler struct, with input validation
func parseInput(r *http.Request, input interface{}, validator *schema.RequestValidator) error {

	schemaDecoder.IgnoreUnknownKeys(true)
	err := r.ParseForm()

	if err != nil {
		return NewErrorCode(err.Error(), InvalidRequest)
	}

	// We do not map and validate input to non-struct handlers
	if reflect.TypeOf(input).Kind() == reflect.Struct {

		err = schemaDecoder.Decode(input, r.Form)

		if err != nil {
			logging.Error("Error decoding input: %s", err)
			return NewErrorCode(err.Error(), InvalidRequest)
		}

		// Validate the input based on the API spec
		if err := validator.Validate(input, r); err != nil {
			logging.Error("Error validating http.Request!: %s", err)
			return NewErrorCode(err.Error(), InvalidRequest)

		}

	}

	return nil

}

// Get the value from  a form param, with an optional default argument if the value was not set
func FormValueDefault(r *http.Request, key, def string) string {
	ret := r.FormValue(key)
	if ret == "" {
		return def
	}
	return strings.Replace(ret, "\"", "", -1)
}
