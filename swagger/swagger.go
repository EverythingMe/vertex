package swagger

import (
	"github.com/alecthomas/jsonschema"

	"reflect"
)

const SwaggerVersion = "2.0"

type Type string

// Type Defintions
const (
	String  Type = "string"
	Number  Type = "number"
	Boolean Type = "boolean"
	Integer Type = "integer"
	Array   Type = "array"
	Object  Type = "object"
)

func TypeOf(t reflect.Type, defaultType Type) (tp Type, items Type) {
	switch t.Kind() {
	case reflect.Bool:
		tp = Boolean
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint32, reflect.Uint16, reflect.Uint64:
		tp = Integer
	case reflect.Float32, reflect.Float64:
		tp = Number
	case reflect.String:
		tp = String
	case reflect.Array, reflect.Slice:
		tp = Array
		items, _ = TypeOf(t.Elem(), defaultType)
	default:
		tp = defaultType
	}
	return
}

// Contact Info
type Contact struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	URL   string `json:"url,omitempty"`
}

// License info
type License struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

// Info describes the meta-info of the API
type Info struct {
	Version        string   `json:"version,omitempty"`
	Title          string   `json:"title,omitempty"`
	Description    string   `json:"description,omitempty"`
	Termsofservice string   `json:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty"`
	License        *License `json:"license,omitempty"`
}

// Param describes a single request param
type Param struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Type        Type   `json:"type,omitempty"`
	Items       Type   `json:"items,omitempty"`

	Format    string      `json:"format,omitempty"`
	Default   interface{} `json:"default,omitempty"`
	Max       float64     `json:"maximum,omitempty"`
	HasMax    bool        `json:"-"`
	Min       float64     `json:"minimum,omitempty"`
	HasMin    bool        `json:"-"`
	MaxLength int         `json:"maxLength,omitempty"`
	MinLength int         `json:"minLength,omitempty"`
	Pattern   string      `json:"pattern,omitempty"`
	Enum      []string    `json:"enum,omitempty"`
	In        string      `json:"in,omitempty"`
	Global    bool        `json:"-"`
	Ref       string      `json:"$ref,omitempty"`
}

// Schema is a generic jsonschema definition - TBD how we want to represent it
type Schema *jsonschema.Schema

// Response describes a response schema
type Response struct {
	Description string `json:"description"`
	Schema      Schema `json:"schema"`
}

// Method describes an API method
type Method struct {
	Description string              `json:"description,omitempty"`
	Operationid string              `json:"operationId,omitempty"`
	Produces    []string            `json:"produces,omitempty"`
	Parameters  []Param             `json:"parameters,omitempty"`
	Responses   map[string]Response `json:"responses"`
	Tags        []string            `json:"tags",omitempty`
}

type Path map[string]Method

// API describes the base of the API
type API struct {
	SwaggerVersion string            `json:"swagger"`
	Info           Info              `json:"info,omitempty"`
	Host           string            `json:"host"`
	Basepath       string            `json:"basePath"`
	Schemes        []string          `json:"schemes"`
	Consumes       []string          `json:"consumes"`
	Produces       []string          `json:"produces"`
	Paths          map[string]Path   `json:"paths"`
	Definitions    map[string]Schema `json:"definitions,omitempty"`
	Parameters     map[string]Param  `json:"parameters,omitempty"`
}

func NewAPI(host, title, description, version, basePath string, schemes []string) *API {
	return &API{
		Info: Info{
			Version:     version,
			Title:       title,
			Description: description,
		},
		Host:           host,
		Basepath:       basePath,
		SwaggerVersion: SwaggerVersion,
		Paths:          make(map[string]Path),
		Schemes:        schemes,
		Definitions:    make(map[string]Schema),
		Parameters:     make(map[string]Param),
	}
}

func (a *API) AddPath(path string) Path {
	p := make(Path)
	a.Paths[path] = p
	return p
}
