package middleware

import (
	"net/http"

	"gitlab.doit9.com/server/vertex"
)

// APIKeyValidator is a simple request validator middleware that looks for an API key in the request form.
// If a key exists and it's in the list of allowed keys - the request is approved
type APIKeyValidator struct {
	paramName string
	validKeys map[string]struct{}
}

// NewAPIKeyValidator creates a new validator middleware. paramName is the GET/POST parameter name
// we look for in requests. validKeys are a list of keys this filter will approve
func NewAPIKeyValidator(paramName string, validKeys ...string) *APIKeyValidator {
	ret := &APIKeyValidator{
		paramName: paramName,
		validKeys: make(map[string]struct{}, len(validKeys)),
	}

	ret.Add(validKeys...)
	return ret
}

// Add a new key(s) to the validator
func (v *APIKeyValidator) Add(keys ...string) {

	for _, k := range keys {
		v.validKeys[k] = struct{}{}
	}
}

func (v *APIKeyValidator) Handle(w http.ResponseWriter, r *vertex.Request, next vertex.HandlerFunc) (interface{}, error) {

	if _, found := v.validKeys[r.FormValue(v.paramName)]; !found {
		return nil, vertex.UnauthorizedError("missing or invalid api key '%s'", r.FormValue(v.paramName))
	}

	return next(w, r)

}
