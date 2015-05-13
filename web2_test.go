package web2

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/dvirsky/go-pylog/logging"
)

type TestHandler struct {
	Bar string `schema:"bar" api.default:"foo" api.required:"false"`
}

func (h TestHandler) Handle(w http.ResponseWriter, r *http.Request) (interface{}, error) {

	return fmt.Sprintf("Your bar is %s", h.Bar), nil
}

var loggingMW = MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next Middleware) (interface{}, error) {
	logging.Info("Logging request %s", r.URL.String())
	return "WAT", nil
})

func TestAPI(t *testing.T) {

	a := API{
		Name:    "testung",
		Version: "1.0",
		//Middleware: []Middleware{loggingMW},
		Routes: RouteMap{
			"/foo": {
				Description: "Get Bar By Foo",
				Handler:     TestHandler{},
				Methods:     POST,
			},
			"/bar": {
				Description: "Get Bar By Foo",
				Handler:     TestHandler{},
				Methods:     GET | POST,
			},
		},
	}

	if err := a.Run(":9947"); err != nil {
		t.Fatal(err)
	}

}
