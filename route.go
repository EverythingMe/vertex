package web2

type Route struct {
	Description string
	Handler     RequestHandler
	Methods     MethodFlag
	Security    SecurityScheme
	Middleware  []Middleware
}

//describe the API for a request
func (r Route) toSwagger() swagger.Method {

	ret := swagger.Method{
		Description: r.Description,
		Parameters:  describeStructFields(reflect.TypeOf(r.Handler)),
		Responses: map[string]swagger.Response{
			"default": {"", swagger.Schema{}},
		},
	}

	return ret
}
