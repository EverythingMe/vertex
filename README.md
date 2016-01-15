# vertex
--
    import "github.com/EverythingMe/vertex"

Vertex is a friendly, fast and flexible RESTful API building framework


### What Vertex Includes

1. An API definition framework

2. Request handlers as structs with automatic data mapping

3. Automatic Data Validation

4. Automatic generation of Swagger from API definitions, for easy documentation 

5. An integrated testing framework for your API

6. A middleware framework similar (but not compatible) to negroni

7. Batteries included: JSON rendering, Auto Recover, Static File Serving,
Request Logging, and more


### Request Handlers

The basic idea of Vertex revolves around friendly, pre-validated request
handlers, that leave the developer with the need to write as little boilerplate
code as possible. Routes in the API are mapped to the RequestHandler interface:

      type RequestHandler interface {
    	   Handle(w http.ResponseWriter, r *http.Request) (interface{}, error)
      }

RequestHandlers have a few interesting characteristics:

1. Fields in structs implementing RequestHandler get automtically filled by
request data.

2. Field values are automatically validated and sanitized

3. They do not *(need to)* write to the response writer, they just need to
return a response object.

You create structs that have all the parameters you need to handle the requests,
define validations for these parameters, and Vertex does the rest for you - just
return a response object and you're done.

Here is an example super simple RequestHandler:

    type UserHandler struct {
    	Id string `schema:"id" required:"true" doc:"The Id Of the user" maxlen:"30"`
    }

    func (h UserHandler) Handle(w http.ResponseWriter, r *http.Request) (interface{}, error) {

    	// load the user from the database
    	user, err := db.Load(h.Id)

    	// return it to the response. No need to write anything directly to the writer
    	return user, err
    }

As you can see, the "id" parameter that is received as a post/get/path parameter
is automatically parsed into the struct when the handler is invoked. If it is
missing or invalid, the handler won't even be invoked, but an error will be
generated to the client.


### Handler Field Tags List

These are the allowed tags for fields in RequestHandler structs:

    - schema - the parameter name in the request
    - doc - a short documentation string for the field
    - default - the default value for the parameter in case it's missing
    - min - the minimum allowed value for numeric fields (inclusive)
    - max - the maximum allowed value for numeric fields (inclusive)
    - maxlen - the maximal allowed length for strings
    - minlen - the minimal allowed length for strings
    - required [true/false] - if set to "true", forces the request to have this parameter set
    - allowEmpty [true/false] - do we allow empty values?
    - pattern - a regular expression that a string must match if this tag is set
    - in [query/body/path] - optional for non path params. mainly for documentation needs

    TODO: Support min/max length for string lists

Supported types for struct fields are (see :

    - bool
    - float variants (float32, float64)
    - int variants (int, int8, int16, int32, int64)
    - string
    - uint variants (uint, uint8, uint16, uint32, uint64)
    - struct - only if it implements Unmarshaler (see below)
    - a pointer to one of the above types
    - a slice or a pointer to a slice of one of the above types


### Custom Unmarshalers

If a field has a custom type that needs automatic deserialization (e.g. a binary
Thrift or Protobuf object), we can define a custom Unmarshal method to the type,
letting it automatically deserialize parameters. (See the Unmarshaler interface)

The unmarshaler should return a new instance of itself with the value set
correctly.

Example: a type that takes a string and splits in two

    type Banana struct {
    	Foo string
    	Bar string
    }

    func (b Banana) UnmarshalRequestData(data string) interface{} {
    	parts := strings.Split(data, ",")
    	if len(parts) == 2 {
    		return Banana{parts[0], parts[1]}
    	}
    	return Banana{}
    }


### Defining An API

APIs are defined in a declarative way, preferably separately from defining the
the actual handler logic.

An API has a few major parts:

    1. High level definitions - like name, version, documentation, etc.
    2. Routes - defining routing paths and mapping them to handlers and tests
    3. Middleware - defining a middleware chain to pre/post-process requests
    4. SecurityScheme - defining the default way requests are validated

Here is an example simple API definition:

    var myAPI = &vertex.API{

    	// The API's name, optionally used in the path
    	Name:          "testung",

    	// The API's version, optionally used in the path
    	Version:       "1.0",

    	// Optional root path. If not set, the root is /<name>/<version>
    	Root:          "/testung/1.0",

    	// Some documentation
    	Doc:           "This is our Test API. It is used to demonstrate declaring an API",

    	// Friendly API title for documentation
    	Title:         "Test API!",

    	// A middleware chain. The default chain includes panic recovery and request logging
    	Middleware:    middleware.DefaultMiddleware,

    	// Response renderer. The default is of course a JSON renderer
    	Renderer:      vertex.JSONRenderer{},

    	// A SecurityScheme. Each route can have an alternative scheme if needed
    	DefaultSecurityScheme: APIKeyValidator,

    	// Unless explicitly set, we only allow https traffic
    	AllowInsecure: false,

    	// The routes of the API
    	Routes: vertex.RouteMap{

    		// Path parameters are defined as {param}
    		"/user/byId/{id}": {

    			// Short request description
    			Description: "Get User Info by id",

    			// An instance of the handler. We use reflection to create a new instance per request
    			Handler:     UserHandler{},

    			// a flag mask of supported requests
    			Methods:     vertex.GET | vertex.POST,

    			// An integration test for the request. Each request must have a test.
    			// Tests can be "warning" tests or "critical" tests
    			Test:        vertex.WarningTest(testUserHandler),

    			// Optional object returned by the request, that will be automatically added to the documentation
    			Returns:     User{},
    		},
    	},
    }


### Security Schemes

Security Schemes are used to validate requests. The scheme simply receives the
request, and returns an error if it is not valid. It can be used to authenticate
the user, validate the API key, etc.


### Middleware

Vertex comes with some middleware modules included. Currently implemented
middleware include:

    - CORS configuration
    - Auto Recover from panic in handlers
    - Request Logging
    - OAuth authentication
    - IP-range filter
    - Simple API Key validation
    - HTTP Basic Auth
    - Response Caching
    - Force Secure (https) Access


### Renderers

Responses have renderers - that transform the response object to some
serialization format.

The default is of course JSON, but an HTML renderer using templates also exists.


### Running The Server

### TODO


### Integration Tests

### TODO


### API Console

### TODO

## Usage

```go
const (
	// The request succeeded
	Ok = iota

	// General failure
	ErrGeneralFailure

	// Input validation failed
	ErrInvalidRequest

	// Missing parameter
	ErrMissingParam

	// Invalid parameter value
	ErrInvalidParam

	// The request was denied for auth reasons
	ErrUnauthorized

	// Insecure access denied
	ErrInsecureAccessDenied

	// We do not want to server this request, the client should not retry
	ErrResourceUnavailable

	// Please back off
	ErrBackOff

	// Some middleware took over the request, and the renderer should not render the response
	ErrHijacked
)
```

```go
const (
	CriticalTests = "critical"
	WarningTests  = "warning"
	AllTests      = "all"
)
```
test categories

```go
const (
	TestFormatText = "text"
	TestFormatJson = "json"
)
```

```go
const (

	// The POST/GET param we pass if we want a JSONP callback response
	CallbackParam = "callback"

	HeaderProcessingTime = "X-Vertex-ProcessingTime"
	HeaderRequestId      = "X-Vertex-RequestId"
	HeaderHost           = "X-Vertex-Host"
	HeaderServerVersion  = "X-Vertex-Version"
)
```
Headers for responses

```go
const DefaultLocale = "en-US"
```

```go
const HeaderGeoPosition = "X-LatLong"
```

```go
var Config = struct {
	Server     serverConfig           `yaml:"server"`
	Auth       authConfig             `yaml:"auth"`
	APIConfigs map[string]interface{} `yaml:"apis,flow"`

	apiconfs map[string]interface{}
}{
	Server: serverConfig{
		ListenAddr:       ":9944",
		AllowInsecure:    false,
		ConsoleFilesPath: "../console",
		LoggingLevel:     "INFO",
		ClientTimeout:    60,
	},

	Auth: authConfig{
		User:     "vertext",
		Password: "xetrev",
	},

	APIConfigs: make(map[string]interface{}),

	apiconfs: make(map[string]interface{}),
}
```

```go
var Hijacked = newErrorCode(ErrHijacked, "Request Hijacked, Do not rendere response")
```
A special error that should be returned when hijacking a request, taking over
response rendering from the renderer

```go
var NopSecurity = SecuritySchemeFunc(func(r *Request) error {
	return nil
})
```

#### func  BackOffError

```go
func BackOffError(duration time.Duration) error
```
BackOff returns a back-off error with a message formatted for the given amount
of backoff time

#### func  FormatPath

```go
func FormatPath(path string, params Params) string
```
FormatPath takes a path template and formats it according to the given path
params

e.g.

    	FormatPath("/foo/{id}", Params{"id":"bar"})
     // Output: "/foo/bar"

#### func  InsecureAccessDenied

```go
func InsecureAccessDenied(msg string, args ...interface{}) error
```
InsecureAccessDenied returns an error signifying the client has no access to the
requested resource

#### func  InvalidParamError

```go
func InvalidParamError(msg string, args ...interface{}) error
```
InvalidParam returns an error signifying an invalid parameter value.

NOTE: The error string will be returned directly to the client

#### func  InvalidRequestError

```go
func InvalidRequestError(msg string, args ...interface{}) error
```
InvalidRequest returns an error signifying something went bad reading the
request data (not the validation process). This in general should not be used by
APIs

#### func  IsHijacked

```go
func IsHijacked(err error) bool
```
IsHijacked inspects an error and checks whether it represents a hijacked
response

#### func  MiddlewareChain

```go
func MiddlewareChain(mw ...Middleware) []Middleware
```
MiddlewareChain just wraps a variadic list of middlewares to make your code less
ugly :)

#### func  MissingParamError

```go
func MissingParamError(msg string, args ...interface{}) error
```
MissingParamError Returns a formatted error stating that a parameter was
missing.

NOTE: The message will be returned to the client directly

#### func  NewError

```go
func NewError(err error) error
```
Wrap a normal error object with an internal object

#### func  NewErrorf

```go
func NewErrorf(format string, args ...interface{}) error
```
Format a new web error from message

#### func  ReadConfigs

```go
func ReadConfigs() error
```

#### func  Register

```go
func Register(name string, builder func() *API, config interface{})
```
Register lest you automatically add an API to the server from your module's
init() function.

name is a unique name for your API (doesn't have to match the API name exactly).

builder is a func that creates the API when we are ready to start the server.

Optionally, you can pass a pointer to a config struct, or nil if you don't need
to. This way, we can read the config struct's values from a unified config file
BEFORE we call the builder, so the builder can use values in the config struct.

#### func  ResourceUnavailableError

```go
func ResourceUnavailableError(msg string, args ...interface{}) error
```
ResourceUnavailable returns an error meaning we do not want to serve this
request, the client should not retry

#### func  RunCLITest

```go
func RunCLITest(apiName, serverAddr, category, format string, out io.Writer) bool
```

#### func  UnauthorizedError

```go
func UnauthorizedError(msg string, args ...interface{}) error
```
Unauthorized returns an error signifying the request was not authorized, but the
client may log-in and retry

#### type API

```go
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
```

API represents the definition of a single, versioned API and all its routes,
middleware and handlers

#### func (*API) FullPath

```go
func (a *API) FullPath(relpath string) string
```
FullPath returns the calculated full versioned path inside the API of a request.

e.g. if my API name is "myapi" and the version is 1.0, FullPath("/foo") returns
"/myapi/1.0/foo"

#### func (API) ToSwagger

```go
func (a API) ToSwagger(serverUrl string) *swagger.API
```
ToSwagger Converts an API definition into a swagger API object for serialization

#### type HTMLRenderer

```go
type HTMLRenderer struct {
}
```


#### func  NewHTMLRenderer

```go
func NewHTMLRenderer(src string, funcMap template.FuncMap) *HTMLRenderer
```

#### func  NewHTMLRendererFiles

```go
func NewHTMLRendererFiles(funcMap map[string]interface{}, fileNames ...string) *HTMLRenderer
```

#### func (*HTMLRenderer) ContentTypes

```go
func (h *HTMLRenderer) ContentTypes() []string
```

#### func (*HTMLRenderer) Render

```go
func (h *HTMLRenderer) Render(v interface{}, e error, w http.ResponseWriter, r *Request) error
```

#### type HandlerFunc

```go
type HandlerFunc func(http.ResponseWriter, *Request) (interface{}, error)
```

HandlerFunc is an adapter that allows you to register normal functions as
handlers. It is used mainly by middleware and should not be used in an
application context

#### func (HandlerFunc) Handle

```go
func (h HandlerFunc) Handle(w http.ResponseWriter, r *Request) (interface{}, error)
```
Handle calls the underlying function

#### type JSONRenderer

```go
type JSONRenderer struct{}
```

JSONRenderer renders a response as a JSON object

#### func (JSONRenderer) ContentTypes

```go
func (JSONRenderer) ContentTypes() []string
```

#### func (JSONRenderer) Render

```go
func (JSONRenderer) Render(v interface{}, e error, w http.ResponseWriter, r *Request) error
```

#### type MethodFlag

```go
type MethodFlag int
```

MethodFlag is used for const flags for method handling on API declaration

```go
const (
	GET  MethodFlag = 0x01
	POST MethodFlag = 0x02
	PUT  MethodFlag = 0x03
)
```
Method flag definitions

#### type Middleware

```go
type Middleware interface {
	Handle(w http.ResponseWriter, r *Request, next HandlerFunc) (interface{}, error)
}
```

Middleware are pre/post processors that can inspect, change, or fail the
request. e.g. authentication, logging, etc

Each middleware needs to call next(w,r) so its next-in-line middleware will
work, or return without it if it wishes to terminate the processing chain

#### type MiddlewareFunc

```go
type MiddlewareFunc func(http.ResponseWriter, *Request, HandlerFunc) (interface{}, error)
```

MiddlewareFunc is a wrapper that allows functions to act as middleware

#### func (MiddlewareFunc) Handle

```go
func (f MiddlewareFunc) Handle(w http.ResponseWriter, r *Request, next HandlerFunc) (interface{}, error)
```
Handle runs the underlying func

#### type Params

```go
type Params map[string]string
```

Params are a string map for path formatting

#### type Renderer

```go
type Renderer interface {
	Render(interface{}, error, http.ResponseWriter, *Request) error
	ContentTypes() []string
}
```

Renderer is an interface for response renderers. A renderer gets the response
object after the entire middleware chain processed it, and renders it directly
to the client

#### func  RenderFunc

```go
func RenderFunc(f func(interface{}, error, http.ResponseWriter, *Request) error, contentTypes ...string) Renderer
```
Wrap a rendering function as an renderer

#### type Request

```go
type Request struct {
	*http.Request
	StartTime time.Time
	Deadline  time.Time
	Locale    string
	UserAgent string
	RemoteIP  string
	Location  struct{ Lat, Long float64 }
	RequestId string
	Callback  string
	Secure    bool
}
```

Request wraps the standard http request object with higher level contextual data

#### func  NewRequest

```go
func NewRequest(r *http.Request) *Request
```
NewRequest wraps a new http request with a vertex request

#### func (*Request) Attribute

```go
func (r *Request) Attribute(key string) (interface{}, bool)
```

#### func (*Request) IsLocal

```go
func (r *Request) IsLocal() bool
```
IsLocal returns true if a request is coming from localhost

#### func (*Request) SetAttribute

```go
func (r *Request) SetAttribute(key string, val interface{})
```

#### func (*Request) String

```go
func (r *Request) String() string
```

#### type RequestHandler

```go
type RequestHandler interface {
	Handle(w http.ResponseWriter, r *Request) (interface{}, error)
}
```

RequestHandler is the interface that request handler structs should implement.

The idea is that you define your request parameters as struct fields, and they
get mapped automatically and validated, leaving you with just pure logic work.

An example Request handler:

    type UserHandler struct {
    	Id   string `schema:"id" required:"true" doc:"The Id Of the user" maxlen:"20" in:"path"`
    	Name string `schema:"name" maxlen:"100" required:"true" doc:"The Name Of the user"`
    	Admin bool `schema:"bool" default:"true" required:"false" doc:"Is this user an admin"`
    }

    func (h UserHandler) Handle(w http.ResponseWriter, r *http.Request) (interface{}, error) {
    	return fmt.Sprintf("Your name is %s and id is %s", h.Name, h.Id), nil
    }

Supported types for automatic param mapping: string, int(32/64), float(32/64),
bool, []string

#### func  Hijacker

```go
func Hijacker(f func(w http.ResponseWriter, r *Request)) RequestHandler
```

#### func  StaticHandler

```go
func StaticHandler(root string, dir http.Dir) RequestHandler
```
StaticHandler is a batteries-included handler for serving static files inside a
directory.

root is the path the root path for this static handler, and will get stripped.

NOTE: root should be the full path to the API root. so if your handler path is
"/static/*filepath", root should be something like "/myapi/1.0/static". Because
the handler is created before the API object is configured, we do not know the
root on creation

#### func  Wrap

```go
func Wrap(f func(w http.ResponseWriter, r *http.Request)) RequestHandler
```

#### type RequestValidator

```go
type RequestValidator struct {
}
```


#### func  NewRequestValidator

```go
func NewRequestValidator(ri schema.RequestInfo) *RequestValidator
```
Create new request validator for a request handler interface. This function
walks the struct tags of the handler's fields and extracts validation metadata.

You should give it the reflect type of your request handler struct

#### func (*RequestValidator) Validate

```go
func (rv *RequestValidator) Validate(request interface{}, r *http.Request) error
```

#### type Route

```go
type Route struct {
	Path        string
	Description string
	Handler     RequestHandler
	Methods     MethodFlag
	Security    SecurityScheme
	Middleware  []Middleware
	Test        Tester
	Returns     interface{}
	Renderer    Renderer
}
```

Route represents a single route (path) in the API and its handler and optional
extra middleware

#### type Routes

```go
type Routes []Route
```

A routing map for an API

#### type SecurityScheme

```go
type SecurityScheme interface {
	Validate(r *Request) error
}
```

SecurityScheme is a special interface that validates a request and is outside
the middleware chain. An API has a default security scheme, and each route can
override it

#### type SecuritySchemeFunc

```go
type SecuritySchemeFunc func(r *Request) error
```


#### func (SecuritySchemeFunc) Validate

```go
func (f SecuritySchemeFunc) Validate(r *Request) error
```

#### type Server

```go
type Server struct {
}
```

Server represents a multi-API http server with a single router

#### func  NewServer

```go
func NewServer(addr string) *Server
```
NewServer creates a new blank server to add APIs to

#### func (*Server) AddAPI

```go
func (s *Server) AddAPI(a *API)
```
AddAPI adds an API to the server manually. It's preferred to use Register in an
init() function

#### func (*Server) Handler

```go
func (s *Server) Handler() http.Handler
```
Handler returns the underlying router, mainly for testing

#### func (*Server) InitAPIs

```go
func (s *Server) InitAPIs()
```
InitAPIs initializes and adds all the APIs registered from API builders

#### func (*Server) Run

```go
func (s *Server) Run() (err error)
```
Run runs the server if it has any APIs registered on it

#### func (*Server) Stop

```go
func (s *Server) Stop()
```
Stop waits up to a second and closes the server

#### type TestContext

```go
type TestContext struct {
}
```

TestContext is a utility available for all testing functions, allowing them to
easily test the current route. It is inspired by Go's testing framework.

In general, a tester needs to call t.Fail(), t.Fatal() or t.Skip() to stop the
execution of the test. A test that doesn't call either of them is considered
passing

#### func (*TestContext) Fail

```go
func (t *TestContext) Fail(format string, params ...interface{})
```
Fail aborts the test with a FAIL status, that is the normal case for failing
tests

#### func (*TestContext) Fatal

```go
func (t *TestContext) Fatal(format string, params ...interface{})
```
Fatal aborts the test with a FATAL status

#### func (*TestContext) FormatUrl

```go
func (t *TestContext) FormatUrl(pathParams Params) string
```
FormatUrl returns a fully formatted URL for the context's route, with all path
params replaced by their respective values in the pathParams map

#### func (*TestContext) GetJSON

```go
func (t *TestContext) GetJSON(r *http.Request, v interface{}) (*http.Response, error)
```
GetJSON performs the given request, and tries to deserialize the response object
to v. If we received an error or decoding is impossible, we return an error. The
raw http response is also returned for inspection

#### func (*TestContext) Log

```go
func (t *TestContext) Log(format string, params ...interface{})
```
Log writes a message to be displayed alongside the test result ONLY if the test
failed

#### func (*TestContext) NewRequest

```go
func (t *TestContext) NewRequest(method string, values url.Values, pathParams Params) (*http.Request, error)
```
NewRequest creates a new http request to the route we are testing now, with
optional values for post/get, and optional path params

#### func (*TestContext) ServerUrl

```go
func (t *TestContext) ServerUrl() string
```
ServerUrl returns the URL of the vertex server we are testing

#### func (*TestContext) Skip

```go
func (t *TestContext) Skip()
```
Skip aborts the test with a SKIP status, that is considered passing

#### type Tester

```go
type Tester interface {
	Test(*TestContext)
	Category() string
}
```

Tester represents a testcase the API runs for a certain API.

Each API contains a list of integration tests that can be run to monitor it.
Each test can have a category associated with it, and we can run tests by a
specific category only.

A test should fail or succeed, and can optionally write error output

#### func  CriticalTest

```go
func CriticalTest(f func(ctx *TestContext)) Tester
```
CrititcalTest wraps testers to signify that the tester is considered critical

#### func  WarningTest

```go
func WarningTest(f func(ctx *TestContext)) Tester
```
WarningTest wraps testers to signify that the tester is a warning test

#### type Unmarshaler

```go
type Unmarshaler interface {
	UnmarshalRequestData(data string) interface{}
}
```

Unmarshaler is an interface for types who are interested in automatic decoding.
The unmarshaler should return a new instance of itself with the value set
correctly.

Example: a type that takes a string and splits in two

    type Banana struct {
    	Foo string
    	Bar string
    }

    func (b Banana) UnmarshalRequestData(data string) interface{} {
    	parts := strings.Split(data, ",")
    	if len(parts) == 2 {
    		return Banana{parts[0], parts[1]}
    	}
    	return Banana{}
    }

#### type VoidHandler

```go
type VoidHandler struct{}
```

VoidHandler is a batteries-included handler that does nothing, useful for
testing, or when a middleware takes over the request completely

#### func (VoidHandler) Handle

```go
func (VoidHandler) Handle(w http.ResponseWriter, r *Request) (interface{}, error)
```
Handle does nothing :)
