package vertex

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dvirsky/go-pylog/logging"
)

// Renderer is an interface for response renderers
type Renderer interface {
	Render(*response, http.ResponseWriter, *http.Request) error
	ContentTypes() []string
}

type funcRenderer struct {
	f            func(*response, http.ResponseWriter, *http.Request) error
	contentTypes []string
}

// Wrap a rendering function as an renderer
func RenderFunc(f func(*response, http.ResponseWriter, *http.Request) error, contentTypes ...string) Renderer {
	return funcRenderer{
		f:            f,
		contentTypes: contentTypes,
	}
}

func (f funcRenderer) Render(res *response, w http.ResponseWriter, r *http.Request) error {
	return f.f(res, w, r)
}
func (f funcRenderer) ContentTypes() []string {
	return f.contentTypes
}

type JSONRenderer struct{}

func (JSONRenderer) Render(res *response, w http.ResponseWriter, r *http.Request) error {

	if err := writeResponse(w, res, formValueDefault(r, "callback", "")); err != nil {
		writeError(w, "Error sending response", formValueDefault(r, "callback", ""))
	}

	return nil
}

func (JSONRenderer) ContentTypes() []string {
	return []string{"text/json"}
}

//serialize an error string inside an object
func writeError(w http.ResponseWriter, message string, callback string) {

	//WriteError is called from recovery, so it must be panic proof
	defer func() {
		e := recover()
		if e != nil {
			logging.Error("Could not write error response! %s", e)
		}
	}()

	response := response{
		ErrorCode:   -1,
		ErrorString: message,
	}

	w.WriteHeader(http.StatusInternalServerError)

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
func writeResponse(w http.ResponseWriter, resp *response, callback string) error {

	buf, e := json.Marshal(resp)
	if e == nil {

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if code := httpCode(resp.ErrorCode); code != http.StatusOK {
			fmt.Println("Writing error code", resp.ErrorCode, code)
			w.WriteHeader(code)
		}

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
