package web2

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dvirsky/go-pylog/logging"
)

// Renderer is an interface for response renderers
type Renderer interface {
	Render(*Response, http.ResponseWriter, *http.Request) error
	ContentTypes() []string
}

type funcRenderer struct {
	f            func(*Response, http.ResponseWriter, *http.Request) error
	contentTypes []string
}

// Wrap a rendering function as an renderer
func RenderFunc(f func(*Response, http.ResponseWriter, *http.Request) error, contentTypes ...string) Renderer {
	return funcRenderer{
		f:            f,
		contentTypes: contentTypes,
	}
}

func (f funcRenderer) Render(res *Response, w http.ResponseWriter, r *http.Request) error {
	return f.f(res, w, r)
}
func (f funcRenderer) ContentTypes() []string {
	return f.contentTypes
}

// RenderJSON is the default JSON renderer. It dumps the response object to a JSON object
var RenderJSON = RenderFunc(func(res *Response, w http.ResponseWriter, r *http.Request) error {

	if err := writeResponse(w, res, FormValueDefault(r, "callback", "")); err != nil {

		writeError(w, "Error sending response", FormValueDefault(r, "callback", ""))
	}

	return nil
},
	"text/json")

//serialize an error string inside an object
func writeError(w http.ResponseWriter, message string, callback string) {

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

//serialize a response object to JSON
func writeResponse(w http.ResponseWriter, resp *Response, callback string) error {

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
