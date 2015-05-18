package schema

import (
	"fmt"
	"net/http"
	"reflect"

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
	ParamInfo
	Key string
}

func newValidationError(msg string, args ...interface{}) error {
	return fmt.Errorf(msg, args...)
}

func (v *fieldValidator) Validate(field reflect.Value, r *http.Request) error {

	//validate required fields
	if v.Required {

		if r.FormValue(v.Name) == "" || !field.IsValid() {
			return newValidationError("missing required param %s", v.Key)
		}

	}

	return nil
}

func (v *fieldValidator) GetKey() string {
	return v.Key
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

func newFieldValidator(pi ParamInfo) *fieldValidator {
	ret := &fieldValidator{
		ParamInfo: pi,
		Key:       pi.Name,
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
	Min     *int64
	Max     *int64
	Default int64
}

func (v *intValidator) Validate(field reflect.Value, r *http.Request) error {
	err := v.fieldValidator.Validate(field, r)
	if err != nil {
		return err
	}

	i := field.Int()
	if v.Min != nil && i < *v.Min {
		return newValidationError("Value too small for %s", v.GetParamName())
	}
	if v.Max != nil && i > *v.Max {
		return newValidationError("Value too large for %s", v.GetParamName())
	}

	return nil

}

func newIntValidator(pi ParamInfo) *intValidator {

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
	MaxLen  *int64
	MinLen  *int64
	Default string
	//*Regex  string
}

func (v *stringValidator) Validate(field reflect.Value, r *http.Request) error {
	err := v.fieldValidator.Validate(field, r)
	if err != nil {
		return err
	}

	s := field.String()

	if v.MaxLen != nil && len(s) > int(*v.MaxLen) {
		return newValidationError("%s is too long", v.GetParamName())
	}

	if v.MinLen != nil && len(s) < int(*v.MinLen) {
		return newValidationError("%s is too short", v.GetParamName())
	}

	return nil

}

func newStringValidator(pi ParamInfo) *stringValidator {

	ret := &stringValidator{
		fieldValidator: newFieldValidator(pi),
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
	Min        *float64
	Max        *float64
	Default    float64
	hasDefault bool
}

func (v *floatValidator) Validate(field reflect.Value, r *http.Request) error {
	err := v.fieldValidator.Validate(field, r)

	if err != nil {
		return err
	}

	f := field.Float()
	if v.Min != nil && f < *v.Min {
		return newValidationError("Value too small for %s", v.GetParamName())
	}
	if v.Max != nil && f > *v.Max {
		return newValidationError("Value too large for %s", v.GetParamName())
	}

	return nil
}

func (v *floatValidator) GetDefault() (interface{}, bool) {
	return v.Default, v.hasDefault
}

func newFloatValidator(pi ParamInfo) *floatValidator {

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
	Default    bool
	hasDefault bool
}

func (v *boolValidator) Validate(field reflect.Value, r *http.Request) error {
	return v.fieldValidator.Validate(field, r)
}

func (v *boolValidator) GetDefault() (interface{}, bool) {

	return v.Default, v.hasDefault
}

func newBoolValidator(pi ParamInfo) *boolValidator {

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
func NewRequestValidator(requestData reflect.Type) *RequestValidator {

	//if the user passes a pointer we walk the actual struct
	if requestData.Kind() == reflect.Ptr {
		requestData = requestData.Elem()
	}

	ret := &RequestValidator{
		fieldValidators: make([]validator, 0),
	}

	//iterate over the fields and create a validator for each
	for i := 0; i < requestData.NumField(); i++ {

		field := requestData.FieldByIndex([]int{i})

		if field.Name == "_" {
			continue
		}

		var vali validator
		switch field.Type.Kind() {
		case reflect.Struct:

			//for structs - we add validators recursively
			validator := NewRequestValidator(field.Type)
			if validator != nil && len(validator.fieldValidators) > 0 {
				ret.fieldValidators = append(ret.fieldValidators, validator.fieldValidators...)
			}
			continue

		case reflect.String:
			vali = newStringValidator(field)

		case reflect.Int, reflect.Int32, reflect.Int64:
			vali = newIntValidator(field)

		case reflect.Float32, reflect.Float64:
			vali = newFloatValidator(field)
		case reflect.Bool:
			vali = newBoolValidator(field)
		default:
			logging.Error("I don't know how to validate %s", field.Type.Kind())
			continue
		}

		if vali != nil {
			logging.Debug("Adding validator %v to request validator %v", vali, requestData)
			ret.fieldValidators = append(ret.fieldValidators, vali)
		}

	}

	return ret
}
