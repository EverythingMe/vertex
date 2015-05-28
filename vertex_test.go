package vertex

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"gitlab.doit9.com/backend/vertex/schema"

	"github.com/stretchr/testify/assert"
)

type MockHandler struct {
	Foo string `schema:"foo" required:"true"`
	Bar string `schema:"bar" required:"true"`
}

func (h MockHandler) Handle(w http.ResponseWriter, r *Request) (interface{}, error) {

	return map[string]string{"foo": h.Foo, "bar": h.Bar}, nil
}

const middlewareHeader = "X-Middleware-Message"

func makeMockMW(header string) MiddlewareFunc {

	return MiddlewareFunc(func(w http.ResponseWriter, r *Request, next HandlerFunc) (interface{}, error) {
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
	//t.SkipNow()
	mw1 := MiddlewareFunc(func(w http.ResponseWriter, r *Request, next HandlerFunc) (interface{}, error) {
		fmt.Fprint(w, "mw1,")
		return next(w, r)
	})

	mw2 := MiddlewareFunc(func(w http.ResponseWriter, r *Request, next HandlerFunc) (interface{}, error) {
		fmt.Fprint(w, "mw2,")
		if next != nil {
			return next(w, r)
		}
		return nil, nil

	})

	mw3 := MiddlewareFunc(func(w http.ResponseWriter, r *Request, next HandlerFunc) (interface{}, error) {
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
			Test: WarningTest(func(t *TestContext) {

				req, err := t.NewRequest("GET", url.Values{"foo": []string{"bar"}, "bar": []string{"baz"}}, nil)
				if err != nil {
					t.Fail("Failed creating request %v", err)
				}

				m := map[string]interface{}{}
				_, err = t.GetJSON(req, &m)
				if err != nil {
					t.Fail("Failed running request %v", err)
				}
				if len(m) == 0 {
					t.Fail("VAlue not serialized")
				}

			}),
		},
		{
			Path:        "/test2",
			Description: "test2",
			Handler: HandlerFunc(func(w http.ResponseWriter, r *Request) (interface{}, error) {
				return map[string]string{"YO": "YO"}, nil
			}),
			Methods:    GET,
			Middleware: []Middleware{makeMockMW("Private middleware")},
			Test:       CriticalTest(func(t *TestContext) {}),
		},
		{
			Path:        "/testvoid",
			Description: "testvoid",
			Handler:     VoidHandler{},
			Methods:     GET,
			Middleware:  []Middleware{makeMockMW("Private middleware")},
			Test:        CriticalTest(func(t *TestContext) {}),
		},
	},
}

func TestRegistration(t *testing.T) {

	//t.SkipNow()
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
	////t.SkipNow()
	srv := NewServer(":9947")
	srv.AddAPI(mockAPI)

	s := httptest.NewServer(srv.Handler())
	defer s.Close()

	checkRequest := func(u string, tests ...func(r interface{}, hr *http.Response)) {

		res, err := http.Get(u)
		if err != nil {
			t.Errorf("Could not get response data")
		}

		resp := map[string]interface{}{}
		if res.StatusCode == http.StatusOK {
			dec := json.NewDecoder(res.Body)
			if err = dec.Decode(&resp); err != nil {
				t.Errorf("Could not decode json response def: %s. Response code:%v", err, res.Status)
			}

		} else {
			b, _ := ioutil.ReadAll(res.Body)
			t.Log("Response body: %s", string(b))
		}

		res.Body.Close()

		if h, found := res.Header[middlewareHeader]; !found || len(h) == 0 {
			t.Error("Could not fine middleware headers")
		} else {
			if h[0] != "Global middleware" || h[1] != "Private middleware" {
				t.Errorf("Bad m2 injected headers: %s", h)
			}
		}

		for _, f := range tests {
			f(resp, res)
		}

	}

	basicTest := func(resp interface{}, hr *http.Response) {
		if hr.StatusCode != http.StatusOK {
			t.Errorf("bad response code: %d", hr.StatusCode)
		}

		if hr.Header.Get(HeaderProcessingTime) == "" {
			t.Errorf("Bad processing time")
		}

	}

	u := fmt.Sprintf("http://%s%s", s.Listener.Addr().String(), mockAPI.FullPath("/test"))
	u += "?foo=f&bar=b"
	t.Log(u)

	checkRequest(u, basicTest, func(resp interface{}, hr *http.Response) {
		if m, ok := resp.(map[string]interface{}); !ok {
			t.Errorf("Bad response type: %s", reflect.TypeOf(resp))
		} else {

			if m["foo"] != "f" || m["bar"] != "b" {
				t.Errorf("Bad response map: %v", m)
			}

		}
	})

	u = fmt.Sprintf("http://%s%s", s.Listener.Addr().String(), mockAPI.FullPath("/test"))
	checkRequest(u, func(resp interface{}, hr *http.Response) {
		assert.Equal(t, http.StatusBadRequest, hr.StatusCode)
	})

	u = fmt.Sprintf("http://%s%s", s.Listener.Addr().String(), mockAPI.FullPath("/test2"))

	checkRequest(u, basicTest, func(resp interface{}, hr *http.Response) {
		if m, ok := resp.(map[string]interface{}); !ok {
			t.Errorf("Bad response type: %s", reflect.TypeOf(resp))
		} else {

			if m["YO"] != "YO" {
				t.Errorf("Bad response: %v", m)
			}

		}
	})

	u = fmt.Sprintf("http://%s%s", s.Listener.Addr().String(), mockAPI.FullPath("/testvoid"))
	checkRequest(u, basicTest, func(resp interface{}, hr *http.Response) {
		if m, ok := resp.(map[string]interface{}); !ok {
			t.Errorf("Bad response type: %s", reflect.TypeOf(resp))
		} else {
			if len(m) != 0 {
				t.Error("Expected empty map, got", m)
			}
		}
	})
	// Test integration tests

	u = fmt.Sprintf("http://%s/test/%s/warning", s.Listener.Addr().String(), mockAPI.root())

	res, err := http.Get(u)
	if err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(b), "[PASS]") {
		t.Errorf("API tests did not pass")
		t.Log(string(b))
	}
	if !strings.Contains(string(b), "warning") {
		t.Errorf("API tests did not contain critical")
	}
	res.Body.Close()

	u = fmt.Sprintf("http://%s/test/%s/critical", s.Listener.Addr().String(), mockAPI.root())

	if res, err = http.Get(u); err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if b, err = ioutil.ReadAll(res.Body); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(b), "[PASS]") {
		t.Errorf("API tests did not pass")
	}
	if !strings.Contains(string(b), "critical") {
		t.Errorf("API tests did not contain critical")
	}

}

func TestFormatPath(t *testing.T) {
	//t.SkipNow()
	data := []struct {
		path     string
		params   Params
		expected string
	}{
		{"/foo/{bar}", Params{"bar": "baz", "baz": "foo"}, "/foo/baz"},
		{"/foo/{bar}", nil, "/foo/{bar}"},
		{"/foo/{biz}", Params{"bar": "baz", "baz": "foo"}, "/foo/{biz}"},
	}

	for _, x := range data {
		out := FormatPath(x.path, x.params)
		if out != x.expected {
			t.Errorf("Bad formatting. expected '%s', got '%s'", x.expected, out)
		}
	}

}

func TestIsHijacked(t *testing.T) {
	//t.SkipNow()
	assert.False(t, IsHijacked(nil))
	assert.False(t, IsHijacked(errors.New("Foo")))
	assert.True(t, IsHijacked(newErrorCode(ErrHijacked, "hijacked")))
	assert.True(t, IsHijacked(Hijacked))
}

func TestRenderer(t *testing.T) {
	//t.SkipNow()
	r := RenderFunc(func(v interface{}, err error, w http.ResponseWriter, r *Request) error {
		fmt.Fprintln(w, "testung")
		return nil
	}, "text/plain")

	assert.EqualValues(t, r.ContentTypes(), []string{"text/plain"})
	out := httptest.NewRecorder()

	err := r.Render(nil, nil, out, nil)
	if err != nil {
		t.Error(err)
	}
	out.Flush()

	assert.Equal(t, "testung\n", out.Body.String())

	// test json renderer

	jr := JSONRenderer{}
	out = httptest.NewRecorder()
	hr, _ := http.NewRequest("GET", "http://foo.bar?callback=foo", nil)
	req := newRequest(hr)
	req.ParseForm()

	assert.NoError(t, jr.Render("ello", nil, out, req))
	assert.Equal(t, "foo(\"ello\");\n", out.Body.String())

	out = httptest.NewRecorder()
	writeError(out, "watwat")
	assert.Equal(t, "watwat\n", out.Body.String())

}

const mockConfs = `
server:
  listen: :8686
apis:
  testung:
     foo: baz
`

func TestConfigs(t *testing.T) {
	//t.SkipNow()
	var apiConf = struct {
		Foo string `yaml:"foo"`
	}{"not baz"}

	registerAPIConfig("testung", &apiConf)

	confile := "/tmp/apiconf.test.yaml"
	fp, err := os.Create(confile)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = fp.WriteString(mockConfs); err != nil {
		t.Fatal(err)
	}
	fp.Close()

	flag.Set("conf", confile)
	assert.NoError(t, ReadConfigs())
	assert.Equal(t, Config.Server.ListenAddr, ":8686")
	assert.Equal(t, "baz", apiConf.Foo)
}

func TestErrors(t *testing.T) {
	//t.SkipNow()
	err := NewError(errors.New("wat"))
	if e, ok := err.(*internalError); !ok {
		t.Fatal("returned not an internal error")
	} else {
		assert.Equal(t, e.Code, ErrGeneralFailure)
		assert.Equal(t, e.Message, "wat")
	}

	err = UnauthorizedError("word")
	if e, ok := err.(*internalError); !ok {
		t.Fatal("returned not an internal error")
	} else {
		assert.Equal(t, e.Code, ErrUnauthorized)
		assert.Equal(t, e.Message, "word")

		code, msg := httpError(e)
		assert.Equal(t, http.StatusUnauthorized, code)
		fmt.Println(msg)
	}

	err = NewErrorf("word %s", "dawg")
	if e, ok := err.(*internalError); !ok {
		t.Fatal("returned not an internal error")
	} else {
		assert.Equal(t, e.Code, ErrGeneralFailure)
		assert.Equal(t, e.Message, "word dawg")

		code, msg := httpError(e)
		assert.Equal(t, http.StatusInternalServerError, code)
		fmt.Println(msg)
	}

	testErr := func(err error, code int, httpCode int) {

		if e, ok := err.(*internalError); !ok {
			t.Error("returned not an internal error")
		} else {
			assert.Equal(t, e.Code, code)
			hcode, _ := httpError(e)
			assert.Equal(t, httpCode, hcode)
		}

	}

	testErr(MissingParamError("foo"), ErrMissingParam, http.StatusBadRequest)
	testErr(InvalidParamError("foo"), ErrInvalidParam, http.StatusBadRequest)
	testErr(InvalidRequestError("sdfsd"), ErrInvalidRequest, http.StatusBadRequest)
	testErr(UnauthorizedError("sdfsd"), ErrUnauthorized, http.StatusUnauthorized)
	testErr(InsecureAccessDenied("sdfsd"), ErrInsecureAccessDenied, http.StatusForbidden)
	testErr(ResourceUnavailableError("sdfsd"), ErrResourceUnavailable, http.StatusServiceUnavailable)
	testErr(BackOffError(0), ErrBackOff, http.StatusServiceUnavailable)

}

func TestServer(t *testing.T) {
	//t.SkipNow()
	s := NewServer(":9934")

	s.AddAPI(mockAPI)

	go func() {
		if err := s.Run(); err != nil {
			t.Fatal(err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	s.Stop()

}

type MockHandlerV struct {
	Int    int      `schema:"int" required:"true" doc:"integer field" min:"-100" max:"100" default:"4"`
	Float  float64  `schema:"float" required:"true" doc:"float field" min:"-100" max:"100" default:"3.141"`
	Bool   bool     `schema:"bool" required:"false" doc:"bool field" default:"true"`
	String string   `schema:"string" required:"false" doc:"string field" default:"WAT WAT" minlen:"1" maxlen:"4" pattern:"^[a-zA-Z]+$"`
	Lst    []string `schema:"list" required:"false" doc:"string list field" default:"  foo, bar, baz    "`
}

func TestValidation(t *testing.T) {
	//t.SkipNow()
	req, err := http.NewRequest("GET", "http://example.com/foo?int=4&float=1.4&string=word&bool=true&list=foo&list=bar", nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := req.ParseForm(); err != nil {
		t.Fatal(err)
	}

	h := MockHandlerV{
		Int:    4,
		Float:  3.14,
		Bool:   true,
		Lst:    []string{"foo", "bar"},
		String: "word",
	}

	ri, err := schema.NewRequestInfo(reflect.TypeOf(MockHandlerV{}), "/foo", "bar", nil)
	if err != nil {
		t.Error(err)
	}

	v := NewRequestValidator(ri)
	if v == nil {
		t.Fatal("nil request validator")
	}

	if err = v.Validate(h, req); err != nil {
		t.Errorf("Failed validation: %s", err)
	}

	// fail on missing param
	badreq, _ := http.NewRequest("GET", "http://example.com/foo?float=1.4&string=word&bool=true&list=foo&list=bar", nil)
	badreq.ParseForm()
	if err = v.Validate(h, badreq); err == nil {
		t.Errorf("We didn't fail on missing int from request: %s", err)
	}

	// fail on bad string value
	h.String = " wat" // spaces not allowed
	if err = v.Validate(h, req); err == nil || err.Error() != "string does not match regex pattern" {
		t.Errorf("We didn't fail on regex: %s", err)
	}

	h.String = "watwatwat" // spaces not allowed
	if err = v.Validate(h, req); err == nil || err.Error() != "string is too long" {
		t.Errorf("We didn't fail on maxlen: %s", err)
	}

	h.String = ""
	if err = v.Validate(h, req); err == nil || err.Error() != "string is too short" {
		t.Errorf("We didn't fail on minlen: %s", err)
	}
	h.String = "wat"

	// Fail on bad int value
	h.Int = -1000
	if err = v.Validate(h, req); err == nil || err.Error() != "Value too small for int" {
		t.Errorf("We didn't fail on min: %s", err)
	}

	h.Int = 1000
	if err = v.Validate(h, req); err == nil || err.Error() != "Value too large for int" {
		t.Errorf("We didn't fail on max: %s", err)
	}

	h.Int = 5

	// Fail on bad float value
	h.Float = -1000
	if err = v.Validate(h, req); err == nil || err.Error() != "Value too small for float" {
		t.Errorf("We didn't fail on min: %s", err)
	}

	h.Float = 1000
	if err = v.Validate(h, req); err == nil || err.Error() != "Value too large for float" {
		t.Errorf("We didn't fail on max: %s", err)
	}
}

func TestRequest(t *testing.T) {

	req, err := http.NewRequest("GET", "http://example.com?callback=foo", nil)

	assert.NoError(t, err)
	req.Header.Set("X-Forwarded-For", "1.1.1.1,2.2.2.2,,,")
	req.Header.Set("Accept-Language", "en-GB,en;q=0.8,he;q=0.6,da;q=0.4")
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set(HeaderGeoPosition, "32.4,31.8")
	req.Header.Set("User-Agent", "Vertex/test")
	r := newRequest(req)

	assert.True(t, r.Secure)
	assert.Equal(t, "2.2.2.2", r.RemoteIP)
	assert.Equal(t, "en-GB", r.Locale)
	assert.Equal(t, "foo", r.Callback)
	assert.Equal(t, 32.4, r.Location.Lat)
	assert.Equal(t, 31.8, r.Location.Long)
	assert.NotEmpty(t, r.UserAgent)
	assert.NotEmpty(t, r.RequestId)

	// Test attributes
	r.SetAttribute("Foo", 123)
	v, found := r.Attribute("Foo")
	assert.True(t, found)
	assert.Equal(t, 123, v)

	// Test Secure parsing
	req.Header.Del("X-Forwarded-Proto")
	assert.False(t, newRequest(req).Secure)
	req.Header.Set("X-Scheme", "https")
	assert.True(t, newRequest(req).Secure)
	req, _ = http.NewRequest("GET", "https://example.com?callback=foo", nil)
	req.RequestURI = req.URL.String()
	assert.True(t, newRequest(req).Secure)

	// default locale if no header set
	assert.Equal(t, "en-US", newRequest(req).Locale)

	// try fucked up ips in XFF
	req.Header.Set("X-Forwarded-For", "1.1.1.1,2.2.2.2,word up")
	assert.Empty(t, newRequest(req).RemoteIP)
	req.Header.Set("X-Real-Ip", "1.1.1.1")
	assert.Equal(t, "1.1.1.1", newRequest(req).RemoteIP)
}
