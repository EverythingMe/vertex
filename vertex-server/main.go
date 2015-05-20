package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"gitlab.doit9.com/backend/web2"
	"gitlab.doit9.com/backend/web2/middleware"
)

type BaseHandler struct {
	Context Context
}

type UserHandler struct {
	BaseHandler
	Id   string `schema:"id" required:"true" doc:"The Id Of the user" in:"path"`
	Name string `schema:"name" maxlen:"100" required:"true" doc:"The Name Of the user"`
}

func (h UserHandler) Handle(w http.ResponseWriter, r *http.Request) (interface{}, error) {

	return fmt.Sprintf("Your name is %s and id is %s", h.Name, h.Id), nil
}

func testUserHandler(t *web2.TestContext) error {

	vals := url.Values{}
	//vals.Set("name", "foofi")
	params := web2.Params{"id": "foo"}

	req, err := t.NewRequest("POST", vals, params)
	if err != nil {
		return err
	}

	resp := &web2.Response{}
	if r, err := t.JsonRequest(req, resp); err != nil {
		b, _ := ioutil.ReadAll(r.Body)
		t.Log("Got response: %v", string(b))
		return err
	}

	return nil
}

func main() {

	//t.SkipNow()

	root := "/testung/1.0"
	a := &web2.API{
		Host:          "localhost:9947",
		Name:          "testung",
		Version:       "1.0",
		Root:          root,
		Doc:           "Our fancy testung API",
		Title:         "Testung API!",
		Middleware:    middleware.DefaultMiddleware,
		Renderer:      web2.RenderJSON,
		AllowInsecure: true,
		Routes: web2.RouteMap{
			"/User/byId/{id}": {
				Description: "Get User Info by id or name",
				Handler:     UserHandler{},
				Methods:     web2.GET | web2.POST,
				Test:        web2.WarningTest(testUserHandler),
			},
			"/static/*filepath": {
				Description: "Static",
				Handler:     web2.StaticHanlder(path.Join(root, "static"), http.Dir("/tmp")),
				Methods:     web2.GET,
				Test:        web2.DummyTest,
			},
		},
	}

	srv := web2.NewServer(":9947")
	srv.AddAPI(a)

	if err := srv.Run(); err != nil {
		panic(err)
	}

}
