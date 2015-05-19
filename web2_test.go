package web2_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"gitlab.doit9.com/backend/web2"
	"gitlab.doit9.com/backend/web2/middleware"
	"gitlab.doit9.com/backend/web2/swagger"

	"github.com/dvirsky/go-pylog/logging"
)

type DateTime time.Time

func (d *DateTime) UnmarshalParam(v string) error {
	return json.Unmarshal([]byte(v), d)
}

type UserHandler struct {
	Id   string `schema:"id" required:"true" doc:"The Id Of the user" in:"path"`
	Name string `schema:"name" maxlen:"100" required:"true" doc:"The Name Of the user"`
}

func (h UserHandler) Handle(w http.ResponseWriter, r *http.Request) (interface{}, error) {

	return fmt.Sprintf("Your name is %s and id is %s", h.Name, h.Id), nil
}

var loggingMW = web2.MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next web2.HandlerFunc) (interface{}, error) {
	logging.Info("Logging request %s", r.URL.String())
	return next(w, r)
})

func testUserHandler(t *web2.TestContext) error {

	req, err := t.NewRequest("GET", "/user/foo?name=bar", nil)
	if err != nil {
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	res.Body.Close()

	return nil

}

func TestMiddleware(t *testing.T) {

	//	mw1 := web2.MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next web2.HandlerFunc) (interface{}, error) {
	//		fmt.Println("mw1")
	//		return next(w, r)
	//	})

	//	mw2 := web2.MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next web2.HandlerFunc) (interface{}, error) {
	//		fmt.Println("mw2")
	//		if next != nil {
	//			return next(w, r)
	//		}
	//		return nil, nil

	//	})

	//chain := web2.buildChain([]Middleware{mw1, mw2})

	///chain.Handle(nil, nil)

}

func TestAPI(t *testing.T) {

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
		},
		Tests: []web2.Tester{
			web2.TestFunc("TestUserHandler", "critical", testUserHandler),
		},
	}

	srv := web2.NewServer(":9947")
	srv.AddAPI(a)

	//	if err := srv.Run(); err != nil {
	//		t.Fatal(err)
	//	}

	s := httptest.NewServer(srv.Handler())
	defer s.Close()

	u := fmt.Sprintf("http://%s%s", s.Listener.Addr().String(), a.FullPath("/swagger"))
	t.Log(u)

	res, err := http.Get(u)
	if err != nil {
		t.Errorf("Could not get swagger data")
	}

	defer res.Body.Close()
	//	b, err := ioutil.ReadAll(res.Body)
	//	fmt.Println(string(b))
	var sw swagger.API
	dec := json.NewDecoder(res.Body)
	if err = dec.Decode(&sw); err != nil {
		t.Errorf("Could not decode swagger def: %s", err)
	}

	swexp := a.ToSwagger()

	if !reflect.DeepEqual(sw, *swexp) {
		t.Errorf("Unmatching api descs:\n%#v\n%#v", sw, *swexp)
	}
	//fmt.Println(sw)

}
