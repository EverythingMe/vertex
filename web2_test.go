package web2_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"gitlab.doit9.com/backend/web2"
	"gitlab.doit9.com/backend/web2/middleware"

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

	return fmt.Sprintf("Your name is %s", h.Name), nil
}

var loggingMW = web2.MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next web2.HandlerFunc) (interface{}, error) {
	logging.Info("Logging request %s", r.URL.String())
	return next(w, r)
})

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

	a := web2.API{
		Host:       "localhost:9947",
		Name:       "testung",
		Version:    "1.0",
		Doc:        "Our fancy testung API",
		Title:      "Testung API!",
		Middleware: middleware.DefaultMiddleware,
		Renderer:   web2.RenderJSON,
		Routes: web2.RouteMap{
			"/user/{id}": {
				Description: "Get User Info by id or name",
				Handler:     UserHandler{},
				Methods:     web2.GET,
			},
		},
	}

	if err := a.Run(":9947"); err != nil {
		t.Fatal(err)
	}

}
