package example

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/EverythingMe/vertex"
	"github.com/EverythingMe/vertex/middleware"
)

type BaseHandler struct {
	APIKey string `schema:"apiKey" maxlen:"64" required:"true" doc:"Client API Key" global:"true"`
}

type UserHandler struct {
	BaseHandler
	Id   string   `schema:"id" maxlen:"100" pattern:"[a-zA-Z]+" required:"true" doc:"The Id Of the user" in:"path"`
	Name string   `schema:"name" maxlen:"100" minlen:"1" required:"true" doc:"The Name Of the user"`
	Foo  int      `schema:"foo" default:"500"`
	Bars []string `schema:"bars" in:"query" hidden:"true" global:"true"`
}

func (h UserHandler) Handle(w http.ResponseWriter, r *vertex.Request) (interface{}, error) {
	return User{Id: h.Id, Name: h.Name, Foo: h.Foo}, nil
}

type User struct {
	Name   string `json:"name"`
	Id     string `json:"id"`
	Banana Banana `json:"banana"`
	Foo    int    `json:"foo"`
}

type Banana struct {
	Foo string
	Bar string
}

func testUserHandler(t *vertex.TestContext) {

	vals := url.Values{}
	vals.Set("name", "foofi")
	params := vertex.Params{"id": "foo"}

	req, err := t.NewRequest("POST", vals, params)
	if err != nil {
		t.Fail("Error performing request: %v", err)
	}

	resp := map[string]interface{}{}
	if r, err := t.GetJSON(req, &resp); err != nil {
		if r != nil && r.Body != nil {
			b, _ := ioutil.ReadAll(r.Body)
			t.Log("Got response: %v", string(b))
		}

		t.Fail("Error getting json: %v", err)
	}

}

func (b Banana) UnmarshalRequestData(data string) interface{} {

	parts := strings.Split(data, ",")
	if len(parts) == 2 {
		return Banana{parts[0], parts[1]}
	}
	return Banana{}
}

var config = struct {
	User   string `yaml:"user"`
	Pass   string `yaml:"password"`
	APIKey string `yaml:"api_key"`
}{
	User:   "Hello",
	Pass:   "World",
	APIKey: "01bea5da73af5b",
}

func APIKeyValidator(r *vertex.Request) error {
	if r.FormValue("apiKey") != config.APIKey {
		return vertex.UnauthorizedError("Inalid API key")
	}
	return nil
}

func init() {

	root := "/testung/1.0"
	vertex.Register("testung", func() *vertex.API {

		return &vertex.API{
			Name:          "testung",
			Version:       "1.0",
			Root:          root,
			Doc:           "Our fancy testung API",
			Title:         "TestungAPI",
			Middleware:    middleware.DefaultMiddleware,
			Renderer:      vertex.JSONRenderer{},
			AllowInsecure: vertex.Config.Server.AllowInsecure,
			SwaggerMiddleware: vertex.MiddlewareChain(
				middleware.NewCORS().Default(),
				middleware.NewIPRangeFilter().AllowPrivate(),
			),
			TestMiddleware: vertex.MiddlewareChain(middleware.BasicAuth{config.User, config.Pass, "Secure", true}),
			//DefaultSecurityScheme: vertex.SecuritySchemeFunc(APIKeyValidator),
			Routes: vertex.Routes{
				{
					Path:        "/user/byId/{id}",
					Description: "Get User Info by id ",
					Handler:     UserHandler{},
					Methods:     vertex.GET,
					Test:        vertex.WarningTest(testUserHandler),
					Returns:     User{},
				},

				{
					Path:        "/user/byName/{name}",
					Description: "Get User Info by  name",
					Handler:     UserHandler{},
					Methods:     vertex.GET,
					Test:        vertex.WarningTest(testUserHandler),
					Returns:     User{},
					Middleware: []vertex.Middleware{
						middleware.BasicAuth{config.User, config.Pass, "Secureee", true},
					},
				},

				{
					Path:        "/static/*filepath",
					Description: "Static",
					Handler:     vertex.StaticHandler(path.Join(root, "static"), http.Dir("/tmp")),
					Methods:     vertex.GET,
				},
			},
		}
	}, &config)

}
