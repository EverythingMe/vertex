package vertex

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type MockHandler struct {
	Foo string `schema:"foo" required:"true"`
	Bar string `schema:"bar" required:"true"`
}

func (h MockHandler) Handle(w http.ResponseWriter, r *http.Request) (interface{}, error) {

	return map[string]string{"foo": h.Foo, "bar": h.Bar}, nil
}

const middlewareHeader = "X-Middleware-Message"

func makeMockMW(header string) MiddlewareFunc {

	return MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next HandlerFunc) (interface{}, error) {
		w.Header().Add(middlewareHeader, header)
		return next(w, r)
	})
}

func testUserHandler(t *TestContext) error {

	req, err := t.NewRequest("GET", nil, nil)
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

	mw1 := MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next HandlerFunc) (interface{}, error) {
		fmt.Fprint(w, "mw1,")
		return next(w, r)
	})

	mw2 := MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next HandlerFunc) (interface{}, error) {
		fmt.Fprint(w, "mw2,")
		if next != nil {
			return next(w, r)
		}
		return nil, nil

	})

	mw3 := MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next HandlerFunc) (interface{}, error) {
		fmt.Fprint(w, "mw3")

		return nil, nil

	})
	recorder := httptest.NewRecorder()
	chain := buildChain([]Middleware{mw1, mw2, mw3})

	chain.handle(recorder, nil)
	expected := "mw1,mw2,mw3"
	if recorder.Body.String() != expected {
		t.Errorf("Expected response to be '%s', got '%s'", expected, recorder.Body.String())
	}

}

var mockAPI = &API{
	Root:          "/mock",
	Name:          "testung",
	Version:       "1.0",
	Doc:           "Our fancy testung API",
	Title:         "Testung API!",
	Renderer:      JSONRenderer{},
	AllowInsecure: true,
	Middleware:    []Middleware{makeMockMW("Global middleware")},
	Routes: Routes{
		{
			Path:        "/test",
			Description: "test",
			Handler:     MockHandler{},
			Methods:     GET,
			Middleware:  []Middleware{makeMockMW("Private middleware")},
		},
	},
}

func TestRegistration(t *testing.T) {

	builder := func() *API {
		return mockAPI
	}

	Register("testung", builder, nil)

	if len(apiBuilders) != 1 {
		t.Fatalf("Wrong number of registered APIs: %d", len(apiBuilders))
	}
	srv := NewServer(":9947")
	srv.InitAPIs()
	if len(srv.apis) != 1 {
		t.Errorf("API not registered in server")
	}

}

func TestIntegration(t *testing.T) {
	srv := NewServer(":9947")
	srv.AddAPI(mockAPI)

	s := httptest.NewServer(srv.Handler())
	defer s.Close()

	u := fmt.Sprintf("http://%s%s", s.Listener.Addr().String(), mockAPI.FullPath("/test"))
	u += "?foo=f&bar=b"
	t.Log(u)

	res, err := http.Get(u)
	if err != nil {
		t.Errorf("Could not get response data")
	}

	defer res.Body.Close()

	resp := response{}
	dec := json.NewDecoder(res.Body)
	if err = dec.Decode(&resp); err != nil {
		t.Errorf("Could not decode swagger def: %s", err)
	}

	if resp.ErrorCode != Ok {
		t.Errorf("bad response code: %d", resp.ErrorCode)
	}

	if resp.ErrorString != "OK" {
		t.Errorf("Bad response strin: %s", resp.ErrorString)
	}

	if resp.ProcessingTime <= 0 {
		t.Errorf("Bad processing time: %v", resp.ProcessingTime)
	}

	if resp.ResponseObject == nil {
		t.Errorf("Bad response object")
	}

	if m, ok := resp.ResponseObject.(map[string]interface{}); !ok {
		t.Errorf("Bad response type: %s", reflect.TypeOf(resp.ResponseObject))
	} else {

		if m["foo"] != "f" || m["bar"] != "b" {
			t.Errorf("Bad response map: %v", m)
		}

	}

	if h, found := res.Header[middlewareHeader]; !found || len(h) == 0 {
		t.Error("Could not fine middleware headers")
	} else {
		if h[0] != "Global middleware" || h[1] != "Private middleware" {
			t.Errorf("Bad m2 injected headers: %s", h)
		}
	}
}
