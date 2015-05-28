package vertex

import (
	"gitlab.doit9.com/backend/vertex/schema"
	"net/http"
	"reflect"
	"regexp"

	"github.com/dvirsky/go-pylog/logging"
)

// Param validator interface
type validator interface {
	Validate(v reflect.Value, r *http.Request) error
	GetDefault() (interface{}, bool)
	GetKey() string
	IsOptional() bool
	GetParamName() string
}

// Base param validator
type fieldValidator struct {
	schema.ParamInfo
}

func (v *fieldValidator) Validate(field reflect.Value, r *http.Request) error {

	//validate required fields
	if v.Required {

		if _, found := r.Form[v.Name]; !found || !field.IsValid() {
			return MissingParamError("missing required param '%s'", v.Name)
		}

	}

	return nil
}

func (v *fieldValidator) GetKey() string {
	return v.StructKey
}

func (v *fieldValidator) GetParamName() string {
	return v.Name
}
func (v *fieldValidator) IsOptional() bool {
	return !v.Required
}

func (v *fieldValidator) GetDefault() (interface{}, bool) {

	if v.Required {
		return 0, false
	}

	return v.Default, v.HasDefault

}

func newFieldValidator(pi schema.ParamInfo) *fieldValidator {
	ret := &fieldValidator{
		ParamInfo: pi,
	}

	return ret
}

///////////////////////////////////////////////////////
//
// Int validator
//
///////////////////////////////////////////////////////
type intValidator struct {
	*fieldValidator
}

func (v *intValidator) Validate(field reflect.Value, r *http.Request) error {
	err := v.fieldValidator.Validate(field, r)
	if err != nil {
		return err
	}

	i := field.Int()

	if v.HasMin && i < int64(v.Min) {
		return InvalidParamError("Value too small for %s", v.GetParamName())
	}
	if v.HasMax && i > int64(v.Max) {
		return InvalidParamError("Value too large for %s", v.GetParamName())
	}

	return nil

}

func newIntValidator(pi schema.ParamInfo) *intValidator {

	return &intValidator{
		fieldValidator: newFieldValidator(pi),
	}

}

///////////////////////////////////////////////////////
//
// String validator
//
///////////////////////////////////////////////////////
type stringValidator struct {
	*fieldValidator
	re *regexp.Regexp
}

func (v *stringValidator) Validate(field reflect.Value, r *http.Request) error {
	err := v.fieldValidator.Validate(field, r)
	if err != nil {
		return err
	}

	s := field.String()

	if v.MaxLength > 0 && len(s) > v.MaxLength {
		return InvalidParamError("%s is too long", v.GetParamName())
	}

	if v.MinLength > 0 && len(s) < v.MinLength {
		return InvalidParamError("%s is too short", v.GetParamName())
	}

	if v.re != nil && !v.re.MatchString(s) {
		return InvalidParamError("%s does not match regex pattern", v.GetParamName())
	}

	return nil

}

func newStringValidator(pi schema.ParamInfo) *stringValidator {

	ret := &stringValidator{
		fieldValidator: newFieldValidator(pi),
	}

	if pi.Pattern != "" {
		re, err := regexp.Compile(pi.Pattern)
		if err != nil {
			logging.Error("Could not create regexp validator - invalid regexp: %s - %s", pi.Pattern, err)
		} else {
			ret.re = re
		}
	}

	return ret

}

//////////////////////////////////////////////////
//
// Float validator
//
//////////////////////////////////////////////////

type floatValidator struct {
	*fieldValidator
}

func (v *floatValidator) Validate(field reflect.Value, r *http.Request) error {
	err := v.fieldValidator.Validate(field, r)

	if err != nil {
		return err
	}

	f := field.Float()
	if v.HasMin && f < v.Min {
		return InvalidParamError("Value too small for %s", v.GetParamName())
	}
	if v.HasMax && f > v.Max {
		return InvalidParamError("Value too large for %s", v.GetParamName())
	}

	return nil
}

func newFloatValidator(pi schema.ParamInfo) *floatValidator {

	ret := &floatValidator{
		fieldValidator: newFieldValidator(pi),
	}
	return ret

}

//////////////////////////////////////////////////
//
// Bool validator
//
//////////////////////////////////////////////////

type boolValidator struct {
	*fieldValidator
}

func (v *boolValidator) Validate(field reflect.Value, r *http.Request) error {
	return v.fieldValidator.Validate(field, r)
}

func newBoolValidator(pi schema.ParamInfo) *boolValidator {

	return &boolValidator{
		fieldValidator: newFieldValidator(pi),
	}
}

type RequestValidator struct {
	fieldValidators []validator
}

func (rv *RequestValidator) Validate(request interface{}, r *http.Request) error {

	val := reflect.ValueOf(request)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	//go over all the validators
	for _, v := range rv.fieldValidators {

		// find the field in the struct. we assume it's there since we build the validators on start time
		field := val.FieldByName(v.GetKey())

		// if the arg is optional and not set, we set the default
		if v.IsOptional() && (!field.IsValid() || r.FormValue(v.GetParamName()) == "") {
			def, ok := v.GetDefault()
			if ok {
				logging.Info("Default value for %s: %v", v.GetKey(), def)
				field.Set(reflect.ValueOf(def))
			}
		}

		// now we validate!
		e := v.Validate(field, r)

		if e != nil {
			logging.Error("Could not validate field %s: %s", v.GetParamName(), e)
			return e
		}

	}

	return nil
}

// Create new request validator for a request handler interface.
// This function walks the struct tags of the handler's fields and extracts validation metadata.
//
// You should give it the reflect type of your request handler struct
func NewRequestValidator(ri schema.RequestInfo) *RequestValidator {

	//if the user passes a pointer we walk the actual struct

	ret := &RequestValidator{
		fieldValidators: make([]validator, 0),
	}

	//iterate over the fields and create a validator for each
	for _, pi := range ri.Params {

		var vali validator
		switch pi.Kind {
		//		case reflect.Struct:

		//			//for structs - we add validators recursively
		//			validator := NewRequestValidator(field.Type)
		//			if validator != nil && len(validator.fieldValidators) > 0 {
		//				ret.fieldValidators = append(ret.fieldValidators, validator.fieldValidators...)
		//			}
		//			continue

		case reflect.String:
			vali = newStringValidator(pi)

		case reflect.Int, reflect.Int32, reflect.Int64:
			vali = newIntValidator(pi)

		case reflect.Float32, reflect.Float64:
			vali = newFloatValidator(pi)
		case reflect.Bool:
			vali = newBoolValidator(pi)
		default:
			logging.Error("I don't know how to validate %s", pi.Kind)
			continue
		}

		if vali != nil {
			logging.Debug("Adding validator %v to request validator %v", vali, ri)
			ret.fieldValidators = append(ret.fieldValidators, vali)
		}

	}

	return ret
}
