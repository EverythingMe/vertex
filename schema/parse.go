package schema

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/dvirsky/go-pylog/logging"
)

// parseDefault takes the default string of a paramInfo and parses it according to the param's type
func parseDefault(val string, t reflect.Kind) (interface{}, bool) {

	if val == "" {
		return nil, false
	}

	switch t {
	case reflect.Int, reflect.Int32, reflect.Int16, reflect.Int64:
		if i, err := parseInt(val); err == nil {
			return i, true
		} else {
			logging.Error("Error parsing int default '%s': %s", val, err)
		}
	case reflect.Float32, reflect.Float64:
		if f, err := parseFloat(val); err == nil {
			return f, true
		} else {
			logging.Error("Error parsing float default '%s': %s", val, err)
		}
	case reflect.String:
		return val, true
	case reflect.Bool:
		if b, err := parseBool(val); err == nil {
			return b, true
		} else {
			logging.Error("Error parsing bool default '%s': %s", val, err)
		}
	case reflect.Slice:
		if l, err := parseList(val); err == nil {
			return l, true
		} else {
			logging.Error("Error parsing string list '%s': %s", val, err)
		}
	}

	return nil, false
}

func parseInt(val string) (int64, error) {

	return strconv.ParseInt(val, 10, 64)

}

func parseFloat(val string) (float64, error) {
	return strconv.ParseFloat(val, 64)
}

func parseBool(val string) (bool, error) {
	val = strings.ToLower(val)

	switch val {
	case "false", "0":
		return false, nil
	case "true", "1":
		return true, nil

	}

	return false, fmt.Errorf("invalid value for bool: %s", val)
}

func parseList(val string) ([]string, error) {

	arr := strings.Split(val, ",")
	for i := range arr {
		arr[i] = strings.TrimSpace(arr[i])
	}

	return arr, nil

}
