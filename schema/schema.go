package schema

import (
	"reflect"
	"strconv"
	"strings"

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

type ParamInfo struct {
	Name        string
	Description string
	Required    bool
	Type        reflect.Kind
	Format      string
	Default     interface{}
	HasDefault  bool
	Max         float64
	HasMax      bool
	Min         float64
	HasMin      bool
	MaxLength   int
	MinLength   int
	Pattern     string
	Options     []string
	In          string
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

func parseDefault(val string, t reflect.Kind) (interface{}, bool) {
	return nil, false
}

func newParamInfo(field reflect.StructField) ParamInfo {

	ret := ParamInfo{Name: field.Name}

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

	ret.Default, ret.HasDefault = parseDefault(getTag(field, DefaultTag, ""), field.Type.Kind())

	return ret
}

type RequestInfo struct {
	Path   string
	Params []ParamInfo
}

// recrusively describe a struct's field using our custom struct tags.
// This is recursive to allow embedding
func describeStructFields(T reflect.Type) (ret []ParamInfo) {

	ret = make([]ParamInfo, 0, T.NumField())

	for i := 0; i < T.NumField(); i++ {

		field := T.FieldByIndex([]int{i})
		if field.Name == "_" {
			continue
		}

		// a struct means this is an embedded request object
		if field.Type.Kind() == reflect.Struct {
			ret = append(describeStructFields(field.Type), ret...)
		} else {

			ret = append(ret, newParamInfo(field))

		}

	}

	return

}
