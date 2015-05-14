package web2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/dvirsky/go-pylog/logging"
	"github.com/julienschmidt/httprouter"
)

type API struct {
	Name                  string
	Title                 string
	Version               string
	Doc                   string
	Host                  string
	DefaultSecurityScheme SecurityScheme
	Renderer              Renderer
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

	chain := buildChain(append(a.Middleware, route.Middleware...))

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

	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		resp := &Response{
			ErrorString:    "OK",
			ErrorCode:      1,
			ResponseObject: nil,
		}

		//sample processing time
		st := time.Now()

		ret, err := chain.handle(w, r)

		et := time.Now()
		resp.ProcessingTime = float64(et.Sub(st)) / float64(time.Millisecond)
		resp.ResponseObject = ret
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

		if err = a.Renderer.Render(resp, w, r); err != nil {
			logging.Error("Error rendering response: %s", err)
		}

	}

}

func (a *API) abspath(relpath string) string {

	return strings.TrimRight(fmt.Sprintf("/%s/%s/%s", a.Name, a.Version, strings.TrimLeft(relpath, "/")), "/")
}

func (a *API) Run(addr string) error {

	router := httprouter.New()

	for path, route := range a.Routes {

		h := a.handler(route)

		path = a.abspath(path)

		if route.Methods&GET == GET {
			logging.Info("Registering GET handler %v to path %s", h, path)
			router.Handle("GET", path, h)
		}
		if route.Methods&POST == POST {
			logging.Info("Registering POST handler %v to path %s", h, path)
			router.Handle("POST", path, h)

		}

	}

	router.GET(a.abspath("/docs"), a.docsHandler())

	router.ServeFiles("/console/*filepath", http.Dir("./console"))

	return http.ListenAndServe(addr, router)

}

func (a *API) docsHandler() func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	// A hander that generates html documentation of the API. Bind it to a url explicitly
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		apiDesc := a.describe()
		apiDesc.Consumes = []string{"text/json"}
		apiDesc.Produces = apiDesc.Consumes

		b, _ := json.MarshalIndent(apiDesc, "", "  ")

		w.Header().Set("Content-Type", "text/json")
		fmt.Fprintf(w, string(b))
		fmt.Println(string(b))

		//		t, e := template.New("doc").Parse(schemaDocTemplate)
		//		if e != nil {
		//			w.Write([]byte(e.Error()))
		//			return
		//		}

		//		t.Execute(w, &apiDesc)
	}

}
