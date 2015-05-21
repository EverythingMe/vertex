// Vertex is a friendly, fast and flexible RESTful API building framework
//
// What Vertex Includes
//
// 1. An API definition framework
//
// 2. Request handlers as structs with automatic data mapping
//
// 3. Automatic Data Validation
//
// 4. An integrated testing framework for your API
//
// 5. A middleware framework similar (but not compliant) to negroni
//
// 6. Batteries included: JSON rendering, Auto Recover, Static File Serving, Request Logging, and more
//
//
// Request Handlers
//
// The basic idea of Vertex revolves around friendly, pre-validated request handlers, that leave the developer with the need to write
// as little boilerplate code as possible. Routes in the API are mapped to the RequestHandler interface:
//
//   type RequestHandler interface {
//	   Handle(w http.ResponseWriter, r *http.Request) (interface{}, error)
//   }
//
// RequestHandlers have a few interesting characteristics:
//
// 1. Fields in structs implementing RequestHandler get automtically filled by request data.
//
// 2. Field values are automatically validated and sanitized
//
// 3. They do not *(need to)* write to the response writer, they just need to return a response object.
//
// You create structs that have all the parameters you need to handle the requests, define validations for these parameters,
// and Vertex does the rest for you - just return a response object and you're done.
//
// Here is an example super simple RequestHandler:
//
//	type UserHandler struct {
//		Id string `schema:"id" required:"true" doc:"The Id Of the user" maxlen:"30"`
//	}
//
//	func (h UserHandler) Handle(w http.ResponseWriter, r *http.Request) (interface{}, error) {
//
//		// load the user from the database
//		user, err := db.Load(h.Id)
//
//		// return it to the response. No need to write anything directly to the writer
//		return user, err
//	}
//
// As you can see, the "id" parameter that is received as a post/get/path parameter is automatically parsed into the struct when the handler
// is invoked. If it is missing or invalid, the handler won't even be invoked, but an error will be generated to the client.
//
// Handler Field Tags List
//
// These are the allowed tags for fields in RequestHandler structs:
//
//  - schema - the parameter name in the request
//  - doc - a short documentation string for the field
//  - default - the default value for the parameter in case it's missing
//  - min - the minimum allowed value for numeric fields (inclusive)
//  - max - the maximum allowed value for numeric fields (inclusive)
//  - maxlen - the maximal allowed length for strings
//  - minlen - the minimal allowed length for strings
//  - required [true/false] - if set to "true", forces the request to have this parameter set
//  - allowEmpty [true/false] - do we allow empty values?
//  - pattern - a regular expression that a string must match if this tag is set
//  - in [query/body/path] - optional for non path params. mainly for documentation needs
//
//  TODO: Support min/max length for string lists
//
// Supported types for struct fields are (see :
//	- bool
//	- float variants (float32, float64)
//	- int variants (int, int8, int16, int32, int64)
//	- string
//	- uint variants (uint, uint8, uint16, uint32, uint64)
//	- struct - only if it implements Unmarshaler (see below)
//	- a pointer to one of the above types
//	- a slice or a pointer to a slice of one of the above types
//
// Custom Unmarshalers
//
// If a field has a custom type that needs automatic deserialization (e.g. a binary Thrift or Protobuf object),
// we can define a custom Unmarshal method to the type, letting it automatically deserialize parameters. (See the Unmarshaler interface)
//
// The unmarshaler should return a new instance of itself with the value set correctly.
//
// Example: a type that takes a string and splits in two
//
//	type Banana struct {
//		Foo string
//		Bar string
//	}
//
//	func (b Banana) UnmarshalRequestData(data string) interface{} {
//		parts := strings.Split(data, ",")
//		if len(parts) == 2 {
//			return Banana{parts[0], parts[1]}
//		}
//		return Banana{}
//	}
//
// Defining An API
//
// APIs are defined in a declarative way, preferably separately from defining the the actual handler logic.
//
// An API has a few major parts:
//
//  1. High level definitions - like name, version, documentation, etc.
//  2. Routes - defining routing paths and mapping them to handlers and tests
//  3. Middleware - defining a middleware chain to pre/post-process requests
//  4. SecurityScheme - defining the default way requests are validated
//
// Here is an example simple API definition:
//	var myAPI = &vertex.API{
//
//		// The API's name, optionally used in the path
//		Name:          "testung",
//
//		// The API's version, optionally used in the path
//		Version:       "1.0",
//
//		// Optional root path. If not set, the root is /<name>/<version>
//		Root:          "/testung/1.0",
//
//		// Some documentation
//		Doc:           "This is our Test API. It is used to demonstrate declaring an API",
//
//		// Friendly API title for documentation
//		Title:         "Test API!",
//
//		// A middleware chain. The default chain includes panic recovery and request logging
//		Middleware:    middleware.DefaultMiddleware,
//
//		// Response renderer. The default is of course a JSON renderer
//		Renderer:      vertex.JSONRenderer{},
//
//		// A SecurityScheme. Each route can have an alternative scheme if needed
//		DefaultSecurityScheme: APIKeyValidator,
//
//		// Unless explicitly set, we only allow https traffic
//		AllowInsecure: false,
//
//		// The routes of the API
//		Routes: vertex.RouteMap{
//
//			// Path parameters are defined as {param}
//			"/user/byId/{id}": {
//
//				// Short request description
//				Description: "Get User Info by id",
//
//				// An instance of the handler. We use reflection to create a new instance per request
//				Handler:     UserHandler{},
//
//				// a flag mask of supported requests
//				Methods:     vertex.GET | vertex.POST,
//
//				// An integration test for the request. Each request must have a test.
//				// Tests can be "warning" tests or "critical" tests
//				Test:        vertex.WarningTest(testUserHandler),
//
//				// Optional object returned by the request, that will be automatically added to the documentation
//				Returns:     User{},
//			},
//		},
//	}
//
// Security Schemes
//
// Security Schemes are used to validate requests. The scheme simply receives the request, and returns an error if it is not valid.
// It can be used to authenticate the user, validate the API key, etc.
//
// Middleware
//
// TODO
//
// Renderers
//
// TODO
//
// Running The Server
//
// TODO
//
// Integration Tests
//
// TODO
//
// API Console
//
// TODO
package vertex
