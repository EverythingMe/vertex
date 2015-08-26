package vertex

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"regexp"
	"strings"
	"time"

	"gitlab.doit9.com/server/vertex/swagger"

	"github.com/alecthomas/jsonschema"
	"github.com/dvirsky/go-pylog/logging"
	"github.com/julienschmidt/httprouter"
)

// API represents the definition of a single, versioned API and all its routes, middleware and handlers
type API struct {
	Name                  string
	Title                 string
	Version               string
	Root                  string
	Doc                   string
	DefaultSecurityScheme SecurityScheme
	Renderer              Renderer
	Routes                Routes
	Middleware            []Middleware
	TestMiddleware        []Middleware
	SwaggerMiddleware     []Middleware
	AllowInsecure         bool
}

// return an httprouter compliant handler function for a route
func (a *API) handler(route Route) func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	// extract the handler type to create a reflect based factory for it
	T := reflect.TypeOf(route.Handler)
	if T.Kind() == reflect.Ptr {
		T = T.Elem()
	}

	validator := NewRequestValidator(route.requestInfo)

	security := route.Security
	if security == nil {
		security = a.DefaultSecurityScheme
	}

	// Build the middleware chain for the API middleware and the rout middleware.
	// The route middleware comes after the API middleware
	chain := buildChain(append(a.Middleware, route.Middleware...)...)

	// add the handler itself as the final middleware
	handlerMW := MiddlewareFunc(func(w http.ResponseWriter, r *Request, next HandlerFunc) (interface{}, error) {

		var reqHandler RequestHandler
		if T.Kind() == reflect.Struct {
			// create a new request handler instance
			reqHandler = reflect.New(T).Interface().(RequestHandler)
		} else {
			reqHandler = route.Handler
		}

		//read params
		if err := parseInput(r.Request, reqHandler, validator); err != nil {
			logging.Error("Error reading input: %s", err)
			return nil, NewError(err)
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

	return a.middlewareHandler(chain, security, route.Renderer)
}

func (a *API) middlewareHandler(chain *step, security SecurityScheme, renderer Renderer) func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	// allow overriding the API's default renderer with a per-route one
	if renderer == nil {
		renderer = a.Renderer
	}

	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

		req := NewRequest(r)

		if !a.AllowInsecure && !req.Secure {
			// local requests bypass security
			if req.RemoteIP != "127.0.0.1" {
				http.Error(w, insecureAccessMessage, http.StatusForbidden)
				return
			}
		}

		r.ParseForm()
		// Copy values from the router params to the request params
		for _, v := range p {
			r.Form.Set(v.Key, v.Value)
		}

		var ret interface{}
		var err error

		if security != nil {
			if err = security.Validate(req); err != nil {
				logging.Warning("Error validating security scheme: %s", err)

				if e, ok := err.(*internalError); ok {
					e.Code = ErrUnauthorized
					err = e
				}
			}
		}
		if err == nil {
			ret, err = chain.handle(w, req)
		}

		if err != Hijacked {

			if err = renderer.Render(ret, err, w, req); err != nil {
				logging.Error("Error rendering response: %s", err)
			}
		} else {
			logging.Debug("Not rendering hijacked request %s", r.RequestURI)
		}

	}

}

var routeRe = regexp.MustCompile("\\{([a-zA-Z_\\.0-9]+)\\}")

func (a *API) root() string {
	if len(a.Root) == 0 {
		a.Root = strings.Join([]string{"", a.Name, a.Version}, "/")
	}

	return a.Root
}

// FullPath returns the calculated full versioned path inside the API of a request.
//
// e.g. if my API name is "myapi" and the version is 1.0, FullPath("/foo") returns "/myapi/1.0/foo"
func (a *API) FullPath(relpath string) string {

	relpath = routeRe.ReplaceAllString(relpath, ":$1")

	ret := path.Join(a.root(), relpath)
	logging.Debug("FullPath for %s => %s", relpath, ret)
	return ret
}

// configure registers the API's routes on a router. If the passed router is nil, we create a new one and return it.
// The nil mode is used when an API is run in stand-alone mode.
func (a *API) configure(router *httprouter.Router) *httprouter.Router {

	if router == nil {
		router = httprouter.New()
	}

	for i, route := range a.Routes {

		if err := route.parseInfo(route.Path); err != nil {
			logging.Error("Error parsing info for %s: %s", route.Path, err)
		}
		a.Routes[i] = route
		h := a.handler(route)

		pth := a.FullPath(route.Path)

		if route.Methods&GET == GET {
			logging.Info("Registering GET handler %v to path %s", h, pth)
			router.Handle("GET", pth, h)
		}
		if route.Methods&POST == POST {
			logging.Info("Registering POST handler %v to path %s", h, pth)
			router.Handle("POST", pth, h)

		}

	}

	chain := buildChain(a.SwaggerMiddleware...)
	if chain == nil {
		chain = buildChain(a.swaggerHandler())
	} else {
		chain.append(a.swaggerHandler())
	}

	// Server the API documentation swagger
	router.GET(a.FullPath("/swagger"), a.middlewareHandler(chain, nil, nil))

	chain = buildChain(a.TestMiddleware...)
	if chain == nil {
		chain = buildChain(a.testHandler())
	} else {
		chain.append(a.testHandler())
	}

	router.GET(path.Join("/test", a.root(), ":category"), a.middlewareHandler(chain, nil, nil))

	// Redirect /$api/$version/console => /console?url=/$api/$version/swagger
	uiPath := fmt.Sprintf("/console?url=%s", url.QueryEscape(a.FullPath("/swagger")))
	router.Handler("GET", a.FullPath("/console"), http.RedirectHandler(uiPath, 301))

	return router

}

// swaggerHandler handles the swagger description request for the API
func (a *API) swaggerHandler() MiddlewareFunc {
	return MiddlewareFunc(func(w http.ResponseWriter, r *Request, next HandlerFunc) (interface{}, error) {
		apiDesc := a.ToSwagger(r.Host)
		return apiDesc, nil
	})
}

// testHandler handles the running of integration tests on the API's special testing url
func (a *API) testHandler() MiddlewareFunc {

	// configure server address for testing
	var serverAddr = "127.0.0.1:9944"
	if addr, err := net.ResolveTCPAddr("tcp", Config.Server.ListenAddr); err == nil {
		serverAddr = fmt.Sprintf("127.0.0.1:%d", addr.Port)
	}

	return MiddlewareFunc(func(w http.ResponseWriter, r *Request, next HandlerFunc) (interface{}, error) {

		category := r.FormValue("category")

		format := r.FormValue("format")
		switch format {
		case TestFormatJson:
			w.Header().Set("Content-Type", "application/json")
		default:
			format = TestFormatText
			w.Header().Set("Content-Type", "text/plain")
		}

		buf := bytes.NewBuffer(nil)

		runner := newTestRunner(buf, a, fmt.Sprintf("http://%s", serverAddr), category, format)

		st := time.Now()
		success := runner.Run()

		if success {
			w.Write(buf.Bytes())
		} else {
			http.Error(w, buf.String(), http.StatusInternalServerError)
		}

		fmt.Fprintln(w, time.Since(st))
		return nil, Hijacked
	})
}

// ToSwagger Converts an API definition into a swagger API object for serialization
func (a API) ToSwagger(serverUrl string) *swagger.API {

	schemes := []string{"https"}

	// http first is important for the swagger ui
	if a.AllowInsecure {
		schemes = []string{"http", "https"}
	}
	ret := swagger.NewAPI(serverUrl, a.Title, a.Doc, a.Version, a.FullPath(""), schemes)
	ret.Consumes = []string{"text/json"}
	ret.Produces = a.Renderer.ContentTypes()
	for _, route := range a.Routes {

		ri := route.requestInfo

		p := ret.AddPath(route.Path)
		method := ri.ToSwagger()

		// copy response definitions to API definitions
		for rk, resp := range method.Responses {

			if resp.Schema != nil && resp.Schema.Definitions != nil {
				for k, v := range resp.Schema.Definitions {
					ret.Definitions[k] = swagger.Schema(&jsonschema.Schema{Type: v})

				}

				resp.Schema.Definitions = nil

				method.Responses[rk] = resp
			}
		}

		// copy global param definitions to param definitions
		for i, parm := range method.Parameters {
			if parm.Global {
				ret.Parameters[parm.Name] = parm

				method.Parameters[i] = swagger.Param{Ref: fmt.Sprintf("#/parameters/%s", parm.Name)}
			}
		}

		// register methods
		if route.Methods&POST == POST {
			p["post"] = method
		}
		if route.Methods&GET == GET {
			p["get"] = method
		}
		if route.Methods&PUT == PUT {
			p["put"] = method
		}
	}

	return ret
}
