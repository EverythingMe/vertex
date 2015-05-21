package schema

import (
	"net/http"
	"reflect"
	"testing"

	"gitlab.doit9.com/backend/vertex/swagger"
)

type MockHandler struct {
	Int    int      `schema:"int" required:"true" doc:"integer field" min:"-100" max:"100" default:"4"`
	Float  float64  `schema:"float" required:"true" doc:"float field" min:"-100" max:"100" default:"3.141"`
	Bool   bool     `schema:"bool" required:"false" doc:"bool field" default:"true"`
	String string   `schema:"string" required:"false" doc:"string field" default:"WAT WAT" minlen:"1" maxlen:"4" pattern:"^[a-zA-Z]+$"`
	Lst    []string `schema:"list" required:"false" doc:"string list field" default:"  foo, bar, baz    "`
}

var expected = []ParamInfo{
	{StructKey: "Int", Name: "int", Kind: reflect.Int, Required: true, Description: "integer field",
		HasMin: true, Min: -100, HasMax: true, Max: 100, HasDefault: true, Default: int64(4), RawDefault: "4", In: "query"},
	{StructKey: "Float", Name: "float", Kind: reflect.Float64, Required: true, Description: "float field",
		HasMin: true, Min: -100, HasMax: true, Max: 100, HasDefault: true, Default: float64(3.141), RawDefault: "3.141", In: "query"},
	{StructKey: "Bool", Name: "bool", Kind: reflect.Bool, Required: false, Description: "bool field",
		HasDefault: true, Default: true, RawDefault: "true", In: "query"},
	{StructKey: "String", Name: "string", Kind: reflect.String, Required: false, Description: "string field",
		HasDefault: true, Default: "WAT WAT", RawDefault: "WAT WAT", MinLength: 1, MaxLength: 4, Pattern: "^[a-zA-Z]+$", In: "query"},
	{StructKey: "Lst", Name: "list", Kind: reflect.Slice, Required: false, Description: "string list field",
		HasDefault: true, Default: []string{"foo", "bar", "baz"}, RawDefault: "  foo, bar, baz    ", In: "query"},
}

func TestParamInfo(t *testing.T) {

	pi := newParamInfo(reflect.TypeOf(MockHandler{}).Field(0))
	if pi.Kind != reflect.Int {
		t.Errorf("Wrong reflect type. want int got %v", pi.Kind)
	}
	if pi.Name != "int" {
		t.Errorf("Wrong name, want int, got %v", pi.Name)
	}

	if pi.Required == false {
		t.Errorf("expected required")
	}

	if pi.Description != "integer field" {
		t.Errorf("Wrong doc: '%s'", pi.Description)
	}
	if !pi.HasDefault {
		t.Errorf("pi should have default")
	}

	if pi.Default != int64(4) {
		t.Errorf("Bad default: %v (%s)", pi.Default, reflect.TypeOf(pi.Default))
	}

	if !pi.HasMin {
		t.Errorf("pi should have min")
	}
	if pi.Min != -100 {
		t.Errorf("Wrong min: %v", pi.Min)
	}
	if !pi.HasMax {
		t.Errorf("pi should have max")
	}
	if pi.Max != 100 {
		t.Errorf("Wrong max: %v", pi.Max)
	}

}

func TestRequestInfo(t *testing.T) {

	path := "/foo/bar"
	desc := "this is a description yo"
	// Test failure
	if _, err := NewRequestInfo(reflect.TypeOf(35), path, desc); err == nil {
		t.Errorf("RequestInfo on non struct should fail")
	}

	ri, err := NewRequestInfo(reflect.TypeOf(MockHandler{}), path, desc)
	if err != nil {
		t.Error(err)
	}

	if ri.Group != "foo" {
		t.Errorf("Bad group '%s'", ri.Group)
	}
	if ri.Path != path {
		t.Errorf("Wrong path, expected %s, got '%s'", path, ri.Path)
	}
	if ri.Description != desc {
		t.Errorf("Wrong desc, expected %s, got '%s'", desc, ri.Description)
	}

	if len(ri.Params) != len(expected) {
		t.Errorf("Wrong number of params, expected %d, got %d", len(expected), len(ri.Params))
	}

	for i := range ri.Params {
		if !reflect.DeepEqual(ri.Params[i], expected[i]) {
			t.Errorf("Wrong param match %d: \ngot: %#v\nexp: %#v", i, ri.Params[i], expected[i])
		}
	}

	// Test conversion to swagger
	sw := ri.ToSwagger()
	if sw.Description != ri.Description {
		t.Errorf("Unmatching descriptions")
	}
	if len(sw.Parameters) != len(ri.Params) {
		t.Errorf("Unmatching parameters")
	}

	for i := range ri.Params {

		p := ri.Params[i]
		s := sw.Parameters[i]

		if p.Name != s.Name || p.RawDefault != s.Default || s.Type != swagger.TypeOf(p.Kind) {
			t.Errorf("Unmatching param %s", p.Name)
		}

	}
}

func TestValidation(t *testing.T) {

	req, err := http.NewRequest("GET", "http://example.com/foo?int=4&float=1.4&string=word&bool=true&list=foo&list=bar", nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := req.ParseForm(); err != nil {
		t.Fatal(err)
	}

	h := MockHandler{
		Int:    4,
		Float:  3.14,
		Bool:   true,
		Lst:    []string{"foo", "bar"},
		String: "word",
	}

	ri, err := NewRequestInfo(reflect.TypeOf(MockHandler{}), "/foo", "bar")
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
