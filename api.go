package web2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"

	"gitlab.doit9.com/backend/web2/schema"
	"gitlab.doit9.com/backend/web2/swagger"

	"github.com/dvirsky/go-pylog/logging"
	"github.com/julienschmidt/httprouter"
)

// API represents the definition of a single, versioned API and all its routes, middleware and handlers
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
	Tests                 []Tester
	AllowInsecure         bool
}

// return an httprouter compliant handler function for a route
func (a *API) handler(route Route) func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	// extract the handler type to create a reflect based factory for it
	T := reflect.TypeOf(route.Handler)
	if T.Kind() == reflect.Ptr {
		T = T.Elem()
	}

	validator := schema.NewRequestValidator(route.requestInfo)

	security := route.Security
	if security == nil {
		security = a.DefaultSecurityScheme
	}

	// Build the middleware chain for the API middleware and the rout middleware.
	// The route middleware comes after the API middleware
	chain := buildChain(append(a.Middleware, route.Middleware...))

	// add the handler itself as the final middleware
	handlerMW := MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next HandlerFunc) (interface{}, error) {

		// create a new request handler instance
		reqHandler := reflect.New(T).Interface().(RequestHandler)

		//read params
		if err := parseInput(r, reqHandler); err != nil {
			logging.Error("Error reading input: %s", err)
			return nil, NewErrorCode(err.Error(), InvalidRequest)
		}

		// Validate the input based on the API spec
		if err := validator.Validate(reqHandler, r); err != nil {
			logging.Error("Error validating http.Request!: %s", err)
			return nil, NewErrorCode(err.Error(), InvalidRequest)

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

		r.ParseForm()

		// Copy values from the router params to the request params
		for _, v := range p {
			r.Form.Set(v.Key, v.Value)
		}

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

var routeRe = regexp.MustCompile("\\{([a-zA-Z_\\.0-9]+)\\}")

// FullPath returns the calculated full versioned path inside the API of a request.
//
// e.g. if my API name is "myapi" and the version is 1.0, FullPath("/foo") returns "/myapi/1.0/foo"
func (a *API) FullPath(relpath string) string {

	relpath = routeRe.ReplaceAllString(relpath, ":$1")

	ret := strings.TrimRight(fmt.Sprintf("/%s/%s/%s", a.Name, a.Version, strings.TrimLeft(relpath, "/")), "/")
	logging.Debug("FullPath for %s => %s", relpath, ret)
	return ret
}

// Run runs a single API server

func (a *API) Run(addr string) error {
	router := a.configure(nil)
	// Server the console swagger UI
	router.ServeFiles("/console/*filepath", http.Dir("./console"))

	// Add a listener for integration tests
	router.Handle("GET", fmt.Sprintf("/test/%s/%s/:category", a.Name, a.Version), a.testHandler(addr))
	return http.ListenAndServe(addr, router)
}

func (a *API) configure(router *httprouter.Router) *httprouter.Router {
	if router == nil {
		router = httprouter.New()
	}

	for path, route := range a.Routes {
		route.parseInfo(path)
		a.Routes[path] = route
		h := a.handler(route)

		path = a.FullPath(path)

		if route.Methods&GET == GET {
			logging.Info("Registering GET handler %v to path %s", h, path)
			router.Handle("GET", path, h)
		}
		if route.Methods&POST == POST {
			logging.Info("Registering POST handler %v to path %s", h, path)
			router.Handle("POST", path, h)

		}

	}

	// Server the API documentation swagger
	router.GET(a.FullPath("/swagger"), a.docsHandler())

	// Redirect /$api/$version/console => /console?url=/$api/$version/swagger
	uiPath := fmt.Sprintf("/console?url=%s", url.QueryEscape(a.FullPath("/swagger")))
	router.Handler("GET", a.FullPath("/console"), http.RedirectHandler(uiPath, 301))

	return router

}

func (a *API) docsHandler() func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	apiDesc := a.ToSwagger()

	// A hander that generates html documentation of the API. Bind it to a url explicitly
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

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

func (a *API) testHandler(addr string) func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		w.Header().Set("Content-Type", "text/plain")
		category := p.ByName("category")

		runner := newTestRunner(w, a, addr, category)

		err := runner.Run(a.Tests)
		if err != nil {
			w.Write([]byte("TESTS FAILED\r\n"))
		}

	}
}

// Return info on all the request
func (a API) ToSwagger() *swagger.API {

	schemes := []string{"https"}

	// http first is important for the swagger ui
	if a.AllowInsecure {
		schemes = []string{"http", "https"}
	}
	ret := swagger.NewAPI(a.Host, a.Title, a.Version, a.Doc, a.FullPath(""), schemes)
	ret.Consumes = []string{"text/json"}
	ret.Produces = a.Renderer.ContentTypes()
	for path, route := range a.Routes {

		ri := route.requestInfo

		p := ret.AddPath(path)
		method := ri.ToSwagger()
		fmt.Printf("%#v\n", ri)

		if route.Methods&POST == POST {
			p["post"] = method
		}
		if route.Methods&GET == GET {
			p["get"] = method
		}
	}

	return ret
}
