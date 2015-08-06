package java

import (
	"path"
	"strings"

	"gitlab.doit9.com/server/vertex/swagger"

	"github.com/alecthomas/jsonschema"
)

// Class represents a local class definition from a swagger return type schema
type Class struct {
	Name    string
	Members []Member
	// Optionally, a class can just extend an external class, as defined in the generator.
	Extends    string
	Implements string
}

// Member is a member in a class
type Member struct {
	Name string
	Type TypeRef
}

// TypeRef represents the type of a member or a parameter, either to a java type or to a class defined in the API
type TypeRef struct {
	Namespace string
	Type      string
	Contained []TypeRef
}

func newTypeRef(t *jsonschema.Type) TypeRef {
	ret := TypeRef{}
	switch swagger.Type(t.Type) {
	case swagger.String:
		ret.Type = "String"
	case swagger.Number:
		ret.Type = "Float"
	case swagger.Boolean:
		ret.Type = "Boolean"
	case swagger.Integer:
		ret.Type = "Integer"
	case swagger.Array:
		ret.Type = "List"
		ret.Contained = []TypeRef{newTypeRef(t.Items)}
	case swagger.Object:
		ret.Type = "Object"
	default:
		ret.Namespace = "Types"
		if t.Ref != "" {
			ret.Type = path.Base(t.Ref)
		} else {
			ret.Type = t.Type
		}
	}
	return ret
}

func newTypeRefSwagger(t swagger.Type) TypeRef {
	ret := TypeRef{}
	switch t {

	case swagger.Number:
		ret.Type = "Float"
	case swagger.Boolean:
		ret.Type = "Boolean"
	case swagger.Integer:
		ret.Type = "Integer"
		//	case Array:
		//		ret.Type = "List"
		//		ret.Contained = []JavaType{newJavaType(t.Items)}
	case swagger.Object:

		ret.Type = "Object"
	case swagger.String:
		fallthrough
	default:
		ret.Type = "String"
	}
	return ret
}

func (t TypeRef) String() string {

	ret := ""
	if t.Namespace != "" {
		ret = t.Namespace + "."
	}
	ret += t.Type
	if t.Contained != nil && len(t.Contained) > 0 {
		ret += "<"
		for _, sub := range t.Contained {
			ret += sub.String() + ","
		}

		ret = strings.TrimRight(ret, ",") + ">"
	}

	return ret
}

// Method hodls the mapping of a route to a java method and its parameters
type Method struct {
	Name     string
	Returns  TypeRef
	Params   []Param
	HttpVerb string
	Doc      string
	Path     string
}

// Param is a method parameter
type Param struct {
	Name string
	Doc  string
	Type TypeRef
	In   string
}

// API holds the java-ready structure of an API definition
type API struct {
	Name    string
	Doc     string
	Root    string
	Types   []Class
	Methods []Method
}
