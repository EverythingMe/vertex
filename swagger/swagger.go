package swagger

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/dvirsky/go-pylog/logging"
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

// Struct field definitions
const (
	DocTag        = "doc"
	DefaultTag    = "default"
	MinTag        = "min"
	MaxTag        = "max"
	MaxLenTag     = "maxlen"
	MinLenTag     = "minlen"
	HiddenTag     = "hidden"
	RequiredTag   = "required"
	AllowEmptyTag = "allowEmpty"
	PatternTag    = "pattern"
	InTag         = "in"
)

func TypeOf(t reflect.Type) Type {
	switch t.Kind() {
	case reflect.Bool:
		return Boolean
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint32, reflect.Uint16, reflect.Uint64:
		return Integer
	case reflect.Float32, reflect.Float64:
		return Number
	case reflect.String:
		return String
	case reflect.Array, reflect.Slice:
		return Array
	default:
		return Object
	}
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
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Type        Type   `json:"type"`

	Format    string   `json:"format,omitempty"`
	Default   string   `json:"default,omitempty"`
	Max       float64  `json:"maximum,omitempty"`
	HasMax    bool     `json:"-"`
	Min       float64  `json:"minimum,omitempty"`
	HasMin    bool     `json:"-"`
	MaxLength int      `json:"maxLength,omitempty"`
	MinLength int      `json:"minLength,omitempty"`
	Pattern   string   `json:"pattern,omitempty"`
	Enum      []string `json:"enum,omitempty"`
	In        string   `json:"in"`
}

func getTag(f reflect.StructField, key, def string) string {
	ret := f.Tag.Get(key)
	if ret == "" {
		return def
	}
	return ret
}

func boolTag(f reflect.StructField, tag string, deflt bool) bool {
	t := f.Tag.Get(tag)
	if t == "" {
		return deflt
	}
	return t == "1" || strings.ToLower(t) == "true"
}

func floatTag(f reflect.StructField, tag string, deflt float64) (float64, bool) {

	var err error
	var ret float64

	v := f.Tag.Get(tag)
	if v == "" {
		return deflt, false
	}

	if ret, err = strconv.ParseFloat(v, 64); err != nil {
		logging.Panic("Invalid value for float: %s", v)
	}

	return ret, true

}

func intTag(f reflect.StructField, tag string, deflt int) (int, bool) {

	var err error
	var ret int64

	v := f.Tag.Get(tag)
	if v == "" {
		return deflt, false
	}

	if ret, err = strconv.ParseInt(v, 10, 64); err != nil {
		logging.Panic("Invalid value for int: %s", v)
	}
	return int(ret), true

}

func ParamFromField(field reflect.StructField) Param {

	ret := Param{Name: field.Name}

	//allow schema overrides of fields
	schemaName := field.Tag.Get("schema")
	if schemaName != "" {
		ret.Name = schemaName
	}

	ret.In = getTag(field, InTag, "query")
	ret.Description = field.Tag.Get(DocTag)
	ret.Default = field.Tag.Get(DefaultTag)
	ret.Type = TypeOf(field.Type)
	ret.Required = boolTag(field, RequiredTag, false)
	ret.Pattern = field.Tag.Get(PatternTag)

	ret.Min, ret.HasMin = floatTag(field, MinTag, 0)
	ret.Max, ret.HasMax = floatTag(field, MaxTag, 0)
	ret.MaxLength, _ = intTag(field, MaxLenTag, 0)
	ret.MinLength, _ = intTag(field, MinLenTag, 0)

	return ret
}

// Schema is a generic jsonschema definition - TBD how we want to represent it
type Schema map[string]interface{}

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
}

func NewAPI(host, title, description, version, basePath string) *API {
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
		Schemes:        []string{"http"},
	}
}

func (a *API) AddPath(path string) Path {
	p := make(Path)
	a.Paths[path] = p
	return p
}
