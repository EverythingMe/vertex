package schema

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"gitlab.doit9.com/backend/web2/swagger"

	"github.com/dvirsky/go-pylog/logging"
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

// ParamInfo represents metadata about a requests parameter
type ParamInfo struct {
	// the struct name of the param
	StructKey string

	// The request name of the parameter, case sensitive
	Name string

	// Documentation description
	Description string

	// Is this param required or optional
	Required bool

	// The param's type as a reflect.Kind. We allow string,int,float,bool,slice
	Type reflect.Kind

	// extra format info for swagger compliance. see https://github.com/swagger-api/swagger-spec/blob/master/versions/1.2.md#431-primitives
	Format string

	// Default value, parsed from string based on the param type
	Default interface{}

	// The unparsed, raw value of the default
	RawDefault string

	// did we have a default value? the default may legitimately be nil or empty or 0
	HasDefault bool

	// Max for numbers
	Max float64
	// Did we have a max definition
	HasMax bool

	// Min for numbers
	Min float64
	//did we have a min definition?
	HasMin bool

	// Maxlength for strings. irrelevant if 0
	MaxLength int

	// Minlength for strings. irrelevant if 0
	MinLength int

	// Regex pattern match. TODO: add to the validator logic
	Pattern string

	// One-of string selection
	Options []string

	// Where is the param in. empty is query/body. should be set only to "path" in case of path params
	In string
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

func newParamInfo(field reflect.StructField) ParamInfo {

	ret := ParamInfo{Name: field.Name, StructKey: field.Name}

	//allow schema overrides of fields
	schemaName := field.Tag.Get("schema")
	if schemaName != "" {
		ret.Name = schemaName
	}

	ret.In = getTag(field, InTag, "query")
	ret.Description = field.Tag.Get(DocTag)

	ret.Type = field.Type.Kind()
	ret.Required = boolTag(field, RequiredTag, false)
	ret.Pattern = field.Tag.Get(PatternTag)

	ret.Min, ret.HasMin = floatTag(field, MinTag, 0)
	ret.Max, ret.HasMax = floatTag(field, MaxTag, 0)
	ret.MaxLength, _ = intTag(field, MaxLenTag, 0)
	ret.MinLength, _ = intTag(field, MinLenTag, 0)

	ret.RawDefault = getTag(field, DefaultTag, "")
	ret.Default, ret.HasDefault = parseDefault(getTag(field, DefaultTag, ""), field.Type.Kind())

	return ret
}

// RequestInfo represents a single request's descriptor
type RequestInfo struct {
	Path        string
	Description string
	Params      []ParamInfo
}

func (r RequestInfo) ToSwagger() swagger.Method {
	ret := swagger.Method{
		Description: r.Description,
		Responses: map[string]swagger.Response{
			"default": {"", swagger.Schema{}},
		},
		Parameters: make([]swagger.Param, 0),
	}

	for _, p := range r.Params {
		ret.Parameters = append(ret.Parameters, p.ToSwagger())
	}
	return ret
}

// recrusively describe a struct's field using our custom struct tags.
// This is recursive to allow embedding
func extractParams(T reflect.Type) (ret []ParamInfo) {

	ret = make([]ParamInfo, 0, T.NumField())

	for i := 0; i < T.NumField(); i++ {

		field := T.FieldByIndex([]int{i})
		if field.Name == "_" {
			continue
		}

		// a struct means this is an embedded request object
		if field.Type.Kind() == reflect.Struct {
			ret = append(extractParams(field.Type), ret...)
		} else {

			ret = append(ret, newParamInfo(field))

		}

	}

	return

}

// NewRequestInfo Builds a requestInfo from a requestHandler struct using reflection
func NewRequestInfo(T reflect.Type, path string, description string) (RequestInfo, error) {

	if T.Kind() == reflect.Ptr {
		T = T.Elem()
	}

	if T.Kind() != reflect.Struct {
		return RequestInfo{}, fmt.Errorf("Could not extract request info from non struct type")
	}

	ret := RequestInfo{Path: path, Description: description, Params: make([]ParamInfo, 0)}

	for i := 0; i < T.NumField(); i++ {

		field := T.FieldByIndex([]int{i})
		if field.Name == "_" {
			continue
		}

		// a struct means this is an embedded request object
		if field.Type.Kind() == reflect.Struct {
			ret.Params = append(extractParams(field.Type), ret.Params...)
		} else {

			ret.Params = append(ret.Params, newParamInfo(field))

		}

	}

	return ret, nil

}

// ToSwagger converts the paramInfo into a swagger Param - they are almost the same, but kept separate
// for decoupling purposes
func (p ParamInfo) ToSwagger() swagger.Param {
	return swagger.Param{
		Name:        p.Name,
		Description: p.Description,
		Type:        swagger.TypeOf(p.Type),
		Required:    p.Required,
		Format:      p.Format,
		Default:     p.RawDefault,
		Max:         p.Max,
		Min:         p.Min,
		MaxLength:   p.MaxLength,
		MinLength:   p.MinLength,
		Pattern:     p.Pattern,
		Enum:        p.Options,
		In:          p.In,
	}
}
