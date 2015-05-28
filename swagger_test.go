package vertex_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"gitlab.doit9.com/backend/vertex"
	"gitlab.doit9.com/backend/vertex/middleware"
	"gitlab.doit9.com/backend/vertex/swagger"
)

type UserHandler struct {
	Id   string `schema:"id" required:"true" doc:"The Id Of the user" in:"path"`
	Name string `schema:"name" maxlen:"100" required:"true" doc:"The Name Of the user"`
}

func (h UserHandler) Handle(w http.ResponseWriter, r *vertex.Request) (interface{}, error) {

	return fmt.Sprintf("Your name is %s and id is %s", h.Name, h.Id), nil
}

func assertEqual(t *testing.T, v1, v2 interface{}, msg ...string) {

	if len(msg) == 0 {
		msg = append(msg, "Equality Assertion failed: \n%#v !=\n%#v")
	}
	switch reflect.TypeOf(v1).Kind() {
	case reflect.String, reflect.Struct, reflect.Slice, reflect.Map:
		if !reflect.DeepEqual(v1, v2) {
			t.Errorf(msg[0], v1, v2)
			return
		}
	default:
		if v1 != v2 {
			t.Errorf(msg[0], v1, v2)
			return
		}
	}

}

func TestSwagger(t *testing.T) {
	//t.SkipNow()
	a := &vertex.API{
		Name:          "testung",
		Version:       "1.0",
		Doc:           "Our fancy testung API",
		Title:         "Testung API!",
		Middleware:    middleware.DefaultMiddleware,
		Renderer:      vertex.JSONRenderer{},
		AllowInsecure: true,
		Routes: vertex.Routes{
			{
				Path:        "/user/{id}",
				Description: "Get User Info by id or name",
				Handler:     UserHandler{},
				Methods:     vertex.GET,
			},
			{
				Path:        "/func/handler",
				Description: "Test handling by a pure func",
				Methods:     vertex.POST,
				Handler: vertex.HandlerFunc(func(w http.ResponseWriter, r *vertex.Request) (interface{}, error) {
					return "WAT WAT", nil
				}),
			},
		},
	}

	srv := vertex.NewServer(":9947")
	srv.AddAPI(a)

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

	swexp := a.ToSwagger(s.Listener.Addr().String())

	assertEqual(t, sw.Basepath, swexp.Basepath)
	assertEqual(t, sw.Consumes, swexp.Consumes)
	assertEqual(t, sw.Host, swexp.Host)
	assertEqual(t, sw.Info, swexp.Info)
	assertEqual(t, sw.Produces, swexp.Produces)
	assertEqual(t, sw.Schemes, swexp.Schemes)
	assertEqual(t, sw.SwaggerVersion, swexp.SwaggerVersion)

	for k, v := range swexp.Paths {

		v2 := sw.Paths[k]
		assertEqual(t, v, v2, "Path mismatch \n%#v\n%#v")
	}

	//fmt.Println(sw)

}
