package vertex

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dvirsky/go-pylog/logging"
)

// Renderer is an interface for response renderers. A renderer gets the response object after the entire
// middleware chain processed it, and renders it directly to the client
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

// JSONRenderer renders a response as a JSON object
type JSONRenderer struct{}

func (JSONRenderer) Render(res *response, w http.ResponseWriter, r *http.Request) error {

	if err := writeResponse(w, res); err != nil {
		writeError(w, "Error sending response")
	}

	return nil
}

func (JSONRenderer) ContentTypes() []string {
	return []string{"text/json"}
}

//serialize an error string inside an object
func writeError(w http.ResponseWriter, message string) {

	//WriteError is called from recovery, so it must be panic proof
	defer func() {
		e := recover()
		if e != nil {
			logging.Error("Could not write error response! %s", e)
		}
	}()

	http.Error(w, message, http.StatusInternalServerError)

}

//serialize a response object to JSON
func writeResponse(w http.ResponseWriter, resp *response) (err error) {

	buf, err := json.Marshal(resp.ResponseObject)
	if err == nil {

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set(HeaderProcessingTime, fmt.Sprintf("%v", resp.ProcessingTime))
		w.Header().Set(HeaderRequestId, resp.RequestId)
		if code := httpCode(resp.ErrorCode); code != http.StatusOK {
			http.Error(w, errorString(resp.ErrorCode), code)
			return
		}

		if resp.callback != "" {
			if _, err = fmt.Fprintf(w, "%s(", resp.callback); err != nil {
				return
			}
		}

		if _, err = w.Write(buf); err != nil {
			return
		}

		if resp.callback != "" {
			if _, err = fmt.Fprintln(w, ");"); err != nil {
				return
			}
		}

	}

	return
}
