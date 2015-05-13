package web2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"gitlab.doit9.com/backend/cards/Godeps/_workspace/src/github.com/dvirsky/schema"

	"github.com/dvirsky/go-pylog/logging"
	"github.com/julienschmidt/httprouter"
)

// This wraps every response object, with an error code and processing time info.
// It is compatible with a gondor response object.
type Response struct {
	ErrorString    string      `json:"errorString"`
	ErrorCode      int         `json:"errorCode"`
	ProcessingTime float64     `json:"processingTime"`
	ResponseObject interface{} `json:"response,omitempty"`
}

type Request struct {
	*http.Request
	context map[string]interface{}
}

type RequestHandler interface {
	Handle(w http.ResponseWriter, r *http.Request) (interface{}, error)
}

type SecurityScheme interface {
	Validate(r *http.Request) error
}

type MethodFlag int

const (
	GET  MethodFlag = 0x01
	POST MethodFlag = 0x02
	PUT  MethodFlag = 0x03
)

type Route struct {
	Description string
	Handler     RequestHandler
	Methods     MethodFlag
	Security    SecurityScheme
	Middleware  []Middleware
}

type RouteMap map[string]Route

type API struct {
	Name                  string
	Version               string
	DefaultSecurityScheme SecurityScheme
	Routes                RouteMap
	Middleware            []Middleware
}

func (a *API) handler(route Route) func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	T := reflect.TypeOf(route.Handler)
	if T.Kind() == reflect.Ptr {
		T = T.Elem()
	}
	validator := NewRequestValidator(T)

	security := route.Security
	if security == nil {
		security = a.DefaultSecurityScheme
	}

	chain := buildChain(a.Middleware)

	handlerMW := MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next HandlerFunc) (interface{}, error) {
		reqHandler := reflect.New(T).Interface().(RequestHandler)

		//read params
		if err := parseInput(r, reqHandler); err != nil {
			logging.Error("Error reading input: %s", err)
			return nil, NewErrorCode(err.Error(), ErrInvalidInput)
		}

		if err := validator.Validate(reqHandler, r); err != nil {
			logging.Error("Error validating http.Request!: %s", err)
			return nil, NewErrorCode(err.Error(), ErrInvalidInput)

		}

		return reqHandler.Handle(w, r)
	})

	if chain == nil {
		chain = &step{
			mw: handlerMW,
		}
	} else {
		chain.append(handlerMW)
	}

	handlerFunc := func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		//create a response object
		resp := &Response{
			ErrorString:    "OK",
			ErrorCode:      1,
			ResponseObject: nil,
		}

		//sample processing time
		st := time.Now()
		var err error

		// if the input is valid - validete security
		if err == nil {
			resp.ResponseObject, err = chain.Handle(w, r)
		}

		et := time.Now()
		resp.ProcessingTime = float64(et.Sub(st)) / float64(time.Millisecond)

		//handle errors if needed
		if err != nil {

			switch e := err.(type) {
			//handle a "proper" internal API error
			case *Error:
				resp.ErrorCode = e.Code
				resp.ErrorString = e.Message
			default:
				resp.ErrorCode = -1
				resp.ErrorString = err.Error()
			}
		}

		if err = writeResonse(w, resp, FormValueDefault(r, "callback", "")); err != nil {

			WriteError(w, "Error sending response", FormValueDefault(r, "callback", ""))
		}

	}

	return handlerFunc
}

func (a *API) Run(addr string) error {
	root := fmt.Sprintf("/%s/%s", a.Name, a.Version)

	router := httprouter.New()

	for path, route := range a.Routes {

		h := a.handler(route)

		path = fmt.Sprintf("%s/%s", root, strings.TrimLeft(path, "/"))

		if route.Methods&GET == GET {
			logging.Info("Registering GET handler %v to path %s", h, path)
			router.Handle("GET", path, h)
		}
		if route.Methods&POST == POST {
			logging.Info("Registering POST handler %v to path %s", h, path)
			router.Handle("POST", path, h)

		}

	}

	return http.ListenAndServe(addr, router)

}

type UserHandler struct {
}

func (u UserHandler) Handle(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return nil, nil
}

var schemaDecoder = schema.NewDecoder()

// Parse the user input into a request handler struct, with input validation
func parseInput(r *http.Request, input interface{}) error {
	schemaDecoder.IgnoreUnknownKeys(true)
	err := r.ParseForm()

	if err != nil {
		return NewErrorCode(err.Error(), ErrInvalidInput)
	}

	// r.PostForm is a map of our POST form values
	err = schemaDecoder.Decode(input, r.Form)

	if err != nil {
		logging.Error("Error decoding input: %s", err)
		return NewErrorCode(err.Error(), ErrInvalidInput)
	}

	return nil

}

//serialize a response object to JSON
func writeResonse(w http.ResponseWriter, resp *Response, callback string) error {

	buf, e := json.Marshal(resp)
	if e == nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		if callback != "" {
			w.Write([]byte(fmt.Sprintf("%s(", callback)))
		}

		w.Write(buf)

		if callback != "" {
			w.Write([]byte(");"))
		}

		return nil
	}

	return e
}

// Get the value from  a form param, with an optional default argument if the value was not set
func FormValueDefault(r *http.Request, key, def string) string {
	ret := r.FormValue(key)
	if ret == "" {
		return def
	}
	return strings.Replace(ret, "\"", "", -1)
}

//serialize an error string inside an object
func WriteError(w http.ResponseWriter, message string, callback string) {

	//WriteError is called from recovery, so it must be panic proof
	defer func() {
		e := recover()
		if e != nil {
			logging.Error("Could not write error response! %s", e)
		}
	}()

	response := Response{
		ErrorCode:   -1,
		ErrorString: message,
	}

	buf, e := json.Marshal(response)
	if e == nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		if callback != "" {
			w.Write([]byte(fmt.Sprintf("%s(", callback)))
		}

		w.Write(buf)
		if callback != "" {
			w.Write([]byte(");"))
		}

	} else {
		logging.Error("Could not marshal response object: %s", e)
		w.Write([]byte("Error sending response!"))
	}

}
