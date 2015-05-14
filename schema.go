package web2

import (
	"reflect"

	"gitlab.doit9.com/backend/web2/swagger"
)

const K_DOC = swagger.DocTag
const K_DEFAULT = swagger.DefaultTag
const K_MIN = swagger.MinTag
const K_MAX = swagger.MaxTag
const K_MAXLEN = swagger.MaxLenTag
const K_MINLEN = swagger.MinLenTag
const K_HIDDEN = swagger.HiddenTag
const K_REQUIRED = swagger.RequiredTag

// recrusively describe a struct's field using our custom struct tags.
// This is recursive to allow embedding
func describeStructFields(T reflect.Type) (ret []swagger.Param) {

	ret = make([]swagger.Param, 0, T.NumField())

	for i := 0; i < T.NumField(); i++ {

		field := T.FieldByIndex([]int{i})
		if field.Name == "_" {
			continue
		}

		// a struct means this is an embedded request object
		if field.Type.Kind() == reflect.Struct {
			ret = append(describeStructFields(field.Type), ret...)
		} else {

			ret = append(ret, swagger.ParamFromField(field))

		}

	}

	return

}

//describe the API for a request
func (r Route) toSwagger() swagger.Method {

	ret := swagger.Method{
		Description: r.Description,
		Parameters:  describeStructFields(reflect.TypeOf(r.Handler)),
		Responses: map[string]swagger.Response{
			"default": {"", swagger.Schema{}},
		},
	}

	return ret
}

// Return info on all the request
func (a API) describe() *swagger.API {

	ret := swagger.NewAPI(a.Host, a.Title, a.Version, a.Doc, a.abspath(""))

	for path, route := range a.Routes {
		p := ret.AddPath(path)
		method := route.toSwagger()

		if route.Methods&POST == POST {
			p["post"] = method
		}
		if route.Methods&GET == GET {
			p["get"] = method
		}
	}

	return ret

}
