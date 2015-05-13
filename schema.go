package web2

import (
	"fmt"
	"reflect"
)

type requestDescriptor struct {
	handlerType reflect.Type
	Path        string
}

var requests = []requestDescriptor{}

// Info about a request parameter from its struct tags
type ParamInfo struct {
	Name      string
	Doc       string
	Type      string
	Default   string
	Inherited bool
	Required  bool
	Visible   bool
}

//info about a request
type RequestInfo struct {
	Path   string
	Doc    string
	Params []ParamInfo
}

const K_DOC = "api.doc"
const K_DEFAULT = "api.default"
const K_MIN = "api.min"
const K_MAX = "api.max"
const K_MAXLEN = "api.maxlen"
const K_MINLEN = "api.minlen"
const K_HIDDEN = "api.hidden"
const K_REQUIRED = "api.required"

// describe a request param from its field info
func describeParam(field reflect.StructField, inherited bool) ParamInfo {

	ret := ParamInfo{Name: field.Name}

	//allow schema overrides of fields
	schemaName := field.Tag.Get("schema")
	if schemaName != "" {
		ret.Name = schemaName
	}

	ret.Doc = field.Tag.Get(K_DOC)

	ret.Default = field.Tag.Get(K_DEFAULT)
	ret.Type = field.Type.Kind().String()
	ret.Inherited = inherited
	ret.Visible = field.Tag.Get(K_HIDDEN) != "true"

	ret.Required = field.Tag.Get(K_REQUIRED) == "true" || field.Tag.Get(K_REQUIRED) == "1"
	fmt.Println(ret)
	return ret
}

// recrusively describe a struct's field using our custom struct tags.
// This is recursive to allow embedding
func describeStructFields(T reflect.Type, inherited bool) (ret []ParamInfo) {
	ret = make([]ParamInfo, 0)

	for i := 0; i < T.NumField(); i++ {

		field := T.FieldByIndex([]int{i})

		if field.Name == "_" {
			continue
		}

		if field.Type.Kind() == reflect.Struct {
			ret = append(describeStructFields(field.Type, true), ret...)
		} else {

			ret = append(ret, describeParam(field, inherited))

		}

	}

	return

}

//describe the API for a request
func describeRequest(desc requestDescriptor) RequestInfo {

	ret := RequestInfo{
		Path: desc.Path,
	}

	ret.Params = describeStructFields(desc.handlerType, false)

	j := len(ret.Params) - 1
	for i, _ := range ret.Params {

		ret.Params[i], ret.Params[j-i] = ret.Params[j-i], ret.Params[i]
		if i > j/2 {
			break

		}
	}

	field, found := desc.handlerType.FieldByName("_")
	if found {
		ret.Doc = field.Tag.Get(K_DOC)
	}

	return ret
}

// Return info on all the request
func DescribeRequests() []RequestInfo {

	ret := make([]RequestInfo, 0)
	for _, req := range requests {
		ret = append(ret, describeRequest(req))
	}

	return ret

}
