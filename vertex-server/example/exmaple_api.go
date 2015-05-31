package example

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"

	"gitlab.doit9.com/server/vertex"
	"gitlab.doit9.com/server/vertex/middleware"
)

type UserHandler struct {
	Id     string `schema:"id" maxlen:"100" pattern:"[a-zA-Z]+" required:"true" doc:"The Id Of the user" in:"path"`
	Name   string `schema:"name" maxlen:"100" minlen:"1" required:"true" doc:"The Name Of the user"`
	Banana Banana `schema:"banana" required:"true"`
}

func (h UserHandler) Handle(w http.ResponseWriter, r *vertex.Request) (interface{}, error) {

	fmt.Printf("%#v\n", h)
	return User{Id: h.Id, Name: h.Name, Banana: h.Banana}, nil
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

type User struct {
	Name   string `json:"name"`
	Id     string `json:"id"`
	Banana Banana `json:"banana"`
}

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

var config = struct {
	Foo string
	Bar string
}{
	Foo: "Hello",
	Bar: "Wrold",
}

func init() {

	root := "/testung/1.0"
	vertex.Register("testung", func() *vertex.API {

		return &vertex.API{
			Name:          "testung",
			Version:       "1.0",
			Root:          root,
			Doc:           "Our fancy testung API",
			Title:         "Testung API!",
			Middleware:    middleware.DefaultMiddleware,
			Renderer:      vertex.JSONRenderer{},
			AllowInsecure: true,
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
						middleware.BasicAuth{config.Foo, config.Bar, "Secureee"},
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
