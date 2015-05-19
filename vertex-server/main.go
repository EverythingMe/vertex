package main

import (
	"fmt"
	"net/http"
	"time"

	"gitlab.doit9.com/backend/web2"
	"gitlab.doit9.com/backend/web2/middleware"
)

type UserHandler struct {
	Id   string `schema:"id" required:"true" doc:"The Id Of the user" in:"path"`
	Name string `schema:"name" maxlen:"100" required:"true" doc:"The Name Of the user"`
}

func (h UserHandler) Handle(w http.ResponseWriter, r *http.Request) (interface{}, error) {

	return fmt.Sprintf("Your name is %s and id is %s", h.Name, h.Id), nil
}

func testUserHandler(t *web2.TestContext) error {
	t.Log("I want banana")
	t.Fatal("WAT?")
	//return errors.New("MEGA FAIL!")
	req, err := t.NewRequest("GET", "/user/foo?name=bar", nil)
	if err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	//	b, err := ioutil.ReadAll(res.Body)
	//	if err != nil {
	//		return err
	//	}
	//t.Log("Got response: %v", string(b))

	return nil //fmt.Errorf("WATTTT....")

}

func main() {

	//t.SkipNow()

	a := &web2.API{
		Host:          "localhost:9947",
		Name:          "testung",
		Version:       "1.0",
		Doc:           "Our fancy testung API",
		Title:         "Testung API!",
		Middleware:    middleware.DefaultMiddleware,
		Renderer:      web2.RenderJSON,
		AllowInsecure: true,
		Routes: web2.RouteMap{
			"/user/{id}": {
				Description: "Get User Info by id or name",
				Handler:     UserHandler{},
				Methods:     web2.GET,
			},
			"/sometest": {
				Handler: UserHandler{}
				Methods: web2.GET,
				Middleware: []web2.Middleware {
					
				}
				
			},
		},
		Tests: []web2.Tester{
			web2.TestFunc("TestUserHandler", "critical", testUserHandler),
			web2.TestFunc("TestSomethingElse", "critical", testUserHandler),
		},
	}

	srv := web2.NewServer(":9947")
	srv.AddAPI(a)

	if err := srv.Run(); err != nil {
		panic(err)
	}

}
