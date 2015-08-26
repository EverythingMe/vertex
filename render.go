package vertex

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/dvirsky/go-pylog/logging"
)

// Renderer is an interface for response renderers. A renderer gets the response object after the entire
// middleware chain processed it, and renders it directly to the client
type Renderer interface {
	Render(interface{}, error, http.ResponseWriter, *Request) error
	ContentTypes() []string
}

type funcRenderer struct {
	f            func(interface{}, error, http.ResponseWriter, *Request) error
	contentTypes []string
}

// Wrap a rendering function as an renderer
func RenderFunc(f func(interface{}, error, http.ResponseWriter, *Request) error, contentTypes ...string) Renderer {
	return funcRenderer{
		f:            f,
		contentTypes: contentTypes,
	}
}

func (f funcRenderer) Render(v interface{}, e error, w http.ResponseWriter, r *Request) error {
	return f.f(v, e, w, r)
}
func (f funcRenderer) ContentTypes() []string {
	return f.contentTypes
}

// JSONRenderer renders a response as a JSON object
type JSONRenderer struct{}

func (JSONRenderer) Render(v interface{}, e error, w http.ResponseWriter, r *Request) error {

	if err := writeResponse(w, r, v, e); err != nil {
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
func writeResponse(w http.ResponseWriter, r *Request, response interface{}, e error) (err error) {

	// Dump meta-data headers
	w.Header().Set(HeaderProcessingTime, fmt.Sprintf("%.03f", time.Since(r.StartTime).Seconds()*1000))
	w.Header().Set(HeaderRequestId, r.RequestId)

	// Dump Error if the request failed
	if e != nil {
		code, message := httpError(e)
		http.Error(w, message, code)
		return
	}

	var buf []byte
	buf, err = json.Marshal(response)
	if err == nil {

		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		if r.Callback != "" {
			if _, err = fmt.Fprintf(w, "%s(", r.Callback); err != nil {
				return
			}
		}

		if _, err = w.Write(buf); err != nil {
			return
		}

		if r.Callback != "" {
			if _, err = fmt.Fprintln(w, ");"); err != nil {
				return
			}
		}

	}

	return
}

type HTMLRenderer struct {
	template *template.Template
}

func NewHTMLRendererFiles(funcMap map[string]interface{}, fileNames ...string) *HTMLRenderer {

	if funcMap == nil {
		funcMap = template.FuncMap{}
	}

	tpl, err := template.New("html").Funcs(funcMap).ParseFiles(fileNames...)
	if err != nil {
		panic(err)
	}

	logging.Info("Created template from files %s (%#v)", fileNames, tpl)
	tpl.ExecuteTemplate(os.Stderr, "html", nil)
	return &HTMLRenderer{
		template: tpl,
	}
}

func NewHTMLRenderer(src string, funcMap template.FuncMap) *HTMLRenderer {

	if funcMap == nil {
		funcMap = template.FuncMap{}
	}

	tpl, err := template.New("html").Funcs(funcMap).Parse(src)
	if err != nil {
		panic(err)
	}

	return &HTMLRenderer{
		template: tpl,
	}

}

func (h *HTMLRenderer) Render(v interface{}, e error, w http.ResponseWriter, r *Request) error {

	// Dump meta-data headers
	w.Header().Set(HeaderProcessingTime, fmt.Sprintf("%.03f", time.Since(r.StartTime).Seconds()*1000))
	w.Header().Set(HeaderRequestId, r.RequestId)

	// Dump Error if the request failed
	if e != nil {
		code, message := httpError(e)
		http.Error(w, message, code)
		return nil
	}

	if err := h.template.ExecuteTemplate(w, "html", v); err != nil {
		http.Error(w, "Could not render html template: "+err.Error(), http.StatusInternalServerError)
	}
	return nil
}

func (h *HTMLRenderer) ContentTypes() []string {
	return []string{"text/html"}
}
