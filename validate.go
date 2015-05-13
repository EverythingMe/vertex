package web2

import (
	"fmt"
	"github.com/dvirsky/go-pylog/logging"
	"net/http"
	"reflect"
	"strconv"
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
	Key       string
	ParamName string
	Optional  bool
}

func newValidationError(msg string, args ...interface{}) error {
	return NewErrorCode(fmt.Sprintf(msg, args...), ErrInvalidInput)
}

func (v *fieldValidator) Validate(field reflect.Value, r *http.Request) error {

	//validate required fields
	if !v.Optional {

		if r.FormValue(v.ParamName) == "" || !field.IsValid() {
			return newValidationError("missing required param %s", v.Key)
		}

	}

	return nil
}

func getTag(f reflect.StructField, key, def string) string {
	ret := f.Tag.Get(key)
	if ret == "" {
		return def
	}
	return ret
}

func (v *fieldValidator) GetKey() string {
	return v.Key
}

func (v *fieldValidator) GetParamName() string {
	return v.ParamName
}
func (v *fieldValidator) IsOptional() bool {
	return v.Optional
}

func newFieldValidator(field reflect.StructField) *fieldValidator {
	ret := &fieldValidator{
		Key:       field.Name,
		ParamName: field.Name,
		Optional:  true,
	}

	//allow schema overrides of fields
	schemaName := field.Tag.Get("schema")
	if schemaName != "" {
		ret.ParamName = schemaName
	}

	//see if this is optional
	req := field.Tag.Get(K_REQUIRED)
	if req == "true" || req == "1" {
		ret.Optional = false
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

func (v *intValidator) GetDefault() (interface{}, bool) {

	if !v.Optional {
		return 0, false
	}
	return int(v.Default), true

}

func newIntValidator(field reflect.StructField) *intValidator {

	ret := &intValidator{
		fieldValidator: newFieldValidator(field),
	}

	var err error

	if field.Tag.Get(K_MIN) != "" {
		ret.Min = new(int64)
		*ret.Min, err = strconv.ParseInt(getTag(field, K_MIN, "0"), 10, 64)
		if err != nil {
			logging.Panic("Invalid default value for int: %s", field.Tag.Get(K_MIN))

		}
	}

	if field.Tag.Get(K_MAX) != "" {
		ret.Max = new(int64)
		*ret.Max, err = strconv.ParseInt(getTag(field, K_MAX, "0"), 10, 64)
		if err != nil {
			logging.Panic("Invalid default value for int: %s", field.Tag.Get(K_MIN))
		}
	}

	if field.Tag.Get(K_DEFAULT) != "" {

		ret.Default, err = strconv.ParseInt(field.Tag.Get(K_DEFAULT), 10, 64)
		if err != nil {
			logging.Panic("Invalid default value for int: %s", field.Tag.Get(K_DEFAULT))
		}

	}

	return ret
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

func (v *stringValidator) GetDefault() (interface{}, bool) {
	if v.Optional {
		return v.Default, v.Default != ""
	}
	return "", false
}

func newStringValidator(field reflect.StructField) *stringValidator {

	ret := &stringValidator{
		fieldValidator: newFieldValidator(field),
	}

	var err error

	if field.Tag.Get(K_MAXLEN) != "" {
		ret.MaxLen = new(int64)
		*ret.MaxLen, err = strconv.ParseInt(getTag(field, K_MAXLEN, "0"), 10, 32)
		if err != nil {
			logging.Panic("Invalid value for maxlen: %s", field.Tag.Get(K_MAXLEN))

		}
	}

	if field.Tag.Get(K_MINLEN) != "" {
		ret.MinLen = new(int64)
		*ret.MinLen, err = strconv.ParseInt(getTag(field, K_MINLEN, "0"), 10, 32)
		if err != nil {
			logging.Panic("Invalid value for minlen: %s", field.Tag.Get(K_MAXLEN))

		}
	}

	ret.Default = field.Tag.Get(K_DEFAULT)

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

func newFloatValidator(field reflect.StructField) *floatValidator {

	ret := &floatValidator{
		fieldValidator: newFieldValidator(field),
	}

	var err error

	if field.Tag.Get(K_MIN) != "" {
		ret.Min = new(float64)
		*ret.Min, err = strconv.ParseFloat(getTag(field, K_MIN, "0"), 64)
		if err != nil {
			logging.Panic("Invalid min value for float: %s", field.Tag.Get(K_MIN))

		}
	}

	if field.Tag.Get(K_MAX) != "" {
		ret.Max = new(float64)
		*ret.Max, err = strconv.ParseFloat(getTag(field, K_MAX, "0"), 64)
		if err != nil {
			logging.Panic("Invalid max value for float: %s", field.Tag.Get(K_MIN))
		}
	}

	if field.Tag.Get(K_DEFAULT) != "" {

		ret.Default, err = strconv.ParseFloat(field.Tag.Get(K_DEFAULT), 64)
		if err != nil {
			logging.Panic("Invalid default value for float: %s", field.Tag.Get(K_DEFAULT))
		}
		ret.hasDefault = true

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

func newBoolValidator(field reflect.StructField) *boolValidator {

	return &boolValidator{
		fieldValidator: newFieldValidator(field),
		Default:        field.Tag.Get(K_DEFAULT) == "true" || field.Tag.Get(K_DEFAULT) == "1",
		hasDefault:     field.Tag.Get(K_DEFAULT) != "",
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
