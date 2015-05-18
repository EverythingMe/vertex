package web2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"gitlab.doit9.com/backend/cards/Godeps/_workspace/src/github.com/dvirsky/schema"

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

type Request struct {
	*http.Request
	context map[string]interface{}
}

type RequestHandler interface {
	Handle(w http.ResponseWriter, r *http.Request) (interface{}, error)
}

type HandlerFunc func(http.ResponseWriter, *http.Request) (interface{}, error)

func (h HandlerFunc) Handle(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return h(w, r)
}

type SecurityScheme interface {
	Validate(r *http.Request) error
}

type Renderer interface {
	Render(*Response, http.ResponseWriter, *http.Request) error
}

type RenderFunc func(*Response, http.ResponseWriter, *http.Request) error

func (f RenderFunc) Render(res *Response, w http.ResponseWriter, r *http.Request) error {
	return f(res, w, r)
}

type MethodFlag int

const (
	GET  MethodFlag = 0x01
	POST MethodFlag = 0x02
	PUT  MethodFlag = 0x03
)

type RouteMap map[string]Route

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
