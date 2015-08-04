package swagger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"testing"
	"text/template"

	"github.com/alecthomas/jsonschema"
)

var swg = `{"swagger":"2.0","info":{"version":"3.0","title":"EverythingAPI","description":"EverythingMe's context aware API"},"host":"localhost:9944","basePath":"/evme/3.0","schemes":["http","https"],"consumes":["text/json"],"produces":["text/json"],"paths":{"/Context":{"get":{"description":"Get the extended context from the server","parameters":[{"name":"apiKey","description":"","required":false,"type":"string","in":"query"},{"name":"X-Evme-Ctx","description":"","required":false,"type":"string","in":"header"}],"responses":{"default":{"description":"context.ExtendedContext","schema":{"$ref":"#/definitions/ExtendedContext"}}},"tags":["Context"]}},"/Echo":{"get":{"description":"Echo the request data and context for debugging","parameters":[{"name":"apiKey","description":"","required":false,"type":"string","in":"query"},{"name":"X-Evme-Ctx","description":"","required":false,"type":"string","in":"header"}],"responses":{"default":{"description":"context.Context","schema":{"$ref":"#/definitions/Context"}}},"tags":["Echo"]}},"/ping":{"get":{"description":"test ping","responses":{"default":{"description":"string","schema":{"type":"string"}}},"tags":["Ping"]}}},"definitions":{"APIKey":{"type":"object","properties":{"enabled":{"type":"boolean"},"key":{"type":"string"},"name":{"type":"string"}},"additionalProperties":false,"required":["key","name","enabled"]},"Client":{"type":"object","properties":{"deviceId":{"type":"string"},"locale":{"type":"string"},"os":{"type":"string"},"version":{"type":"string"}},"additionalProperties":false,"required":["os","version","deviceId","locale"]},"ClientContext":{"type":"object","properties":{"activeInsights":{"type":"array","items":{"$ref":"#/definitions/Insight"}},"activeScenarios":{"type":"array","items":{"type":"string"}},"device":{"$ref":"#/definitions/DeviceContext"},"location":{"$ref":"#/definitions/GeoContext"},"motionState":{"$ref":"#/definitions/MotionState"},"network":{"$ref":"#/definitions/NetworkContext"},"phone":{"$ref":"#/definitions/PhoneContext"},"time":{"$ref":"#/definitions/TimeContext"}},"additionalProperties":false},"Context":{"type":"object","properties":{"client":{"$ref":"#/definitions/ClientContext"},"ext":{"$ref":"#/definitions/ExtendedContext"},"request":{"$ref":"#/definitions/RequestContext"}},"additionalProperties":false,"required":["client","request"]},"DeviceContext":{"type":"object","properties":{"battery":{"type":"integer"},"batteryPercent":{"type":"integer"},"deviceName":{"type":"string"},"dockType":{"type":"integer"},"headphonesConnected":{"type":"boolean"},"locale":{"type":"string"},"manufacturer":{"type":"string"},"os":{"type":"string"},"osVersion":{"type":"string"},"screenActive":{"type":"boolean"},"screenHeight":{"type":"integer"},"screenWidth":{"type":"integer"},"yearClass":{"type":"integer"}},"additionalProperties":false,"required":["battery","batteryPercent","headphonesConnected","screenActive","dockType","deviceName","manufacturer","os","osVersion","screenWidth","screenHeight","locale","yearClass"]},"ExtendedContext":{"type":"object","properties":{"insights":{"type":"array","items":{"$ref":"#/definitions/ServerInsight"}},"scenartios":{"type":"array","items":{"type":"string"}}},"additionalProperties":false,"required":["insights","scenartios"]},"GeoContext":{"type":"object","properties":{"accuracy":{"type":"number"},"homeCountry":{"type":"string"},"isRoaming":{"type":"boolean"},"isTravelling":{"type":"boolean"},"knownLocationCertainty":{"type":"number"},"knownLocationId":{"type":"string"},"lat":{"type":"number"},"lon":{"type":"number"},"simCountry":{"type":"string"}},"additionalProperties":false,"required":["lat","lon","accuracy","knownLocationCertainty","simCountry","homeCountry","isTravelling","isRoaming","knownLocationId"]},"Insight":{"type":"object","properties":{"confidence":{"type":"number"},"lastUpdatePeriod":{"type":"integer"},"params":{"type":"object","patternProperties":{".*":{"type":"string"}}},"type":{"type":"string"},"values":{"type":"array","items":{"type":"string"}}},"additionalProperties":false,"required":["type","confidence","lastUpdatePeriod"]},"MotionState":{"type":"object","properties":{"activity":{"type":"integer"},"confidence":{"type":"integer"}},"additionalProperties":false,"required":["activity","confidence"]},"NetworkContext":{"type":"object","properties":{"BTConnectedDevices":{"type":"array","items":{"type":"string"}},"BTEnabled":{"type":"boolean"},"activeNetwork":{"type":"string"},"availableWifiNetworks":{"type":"array","items":{"type":"string"}},"carrier":{"type":"string"},"mobileCountryCode":{"type":"integer"},"mobileNetworkCode":{"type":"integer"},"networkStrength":{"type":"integer"},"wifiName":{"type":"string"}},"additionalProperties":false,"required":["activeNetwork","networkStrength","wifiName","availableWifiNetworks","carrier","mobileNetworkCode","mobileCountryCode","BTEnabled","BTConnectedDevices"]},"PhoneContext":{"type":"object","properties":{"missedCall":{"type":"string"}},"additionalProperties":false,"required":["missedCall"]},"RequestContext":{"type":"object","properties":{"api":{"$ref":"#/definitions/APIKey"},"client":{"$ref":"#/definitions/Client"},"featureFlags":{"type":"array","items":{"type":"string"}},"homeCountryLocale":{"type":"string"},"ip":{"type":"string"},"requestId":{"type":"string"},"resolvedLocation":{"$ref":"#/definitions/TGeoLocation"},"trace":{"type":"array","items":{"$ref":"#/definitions/ServerHit"}},"user":{"$ref":"#/definitions/User"}},"additionalProperties":false,"required":["ip","requestId","featureFlags","homeCountryLocale"]},"ServerHit":{"type":"object","properties":{"cluster":{"type":"string"},"hostName":{"type":"string"},"timestamp":{"type":"integer"}},"additionalProperties":false,"required":["hostName","cluster","timestamp"]},"ServerInsight":{"type":"object","properties":{"TTL":{"type":"integer"},"confidence":{"type":"number"},"goefenceRadius":{"type":"number"},"jsonValue":{"type":"string"},"type":{"type":"string"}},"additionalProperties":false,"required":["type","jsonValue","confidence","TTL","goefenceRadius"]},"TGeoLocation":{"type":"object","properties":{"city":{"type":"string"},"country":{"type":"string"},"countryCode":{"type":"string"},"countryId":{"type":"integer"},"lat":{"type":"number"},"lon":{"type":"number"},"name":{"type":"string"},"state":{"type":"string"},"type":{"type":"string"},"zip":{"type":"string"}},"additionalProperties":false,"required":["lat","lon","name","type","city","zip","state","country","countryCode","countryId"]},"TimeContext":{"type":"object","properties":{"isWeekend":{"type":"boolean"},"localTime":{"type":"integer"},"timeOfDay":{"type":"integer"},"timeZone":{"type":"integer"}},"additionalProperties":false,"required":["localTime","timeZone","timeOfDay","isWeekend"]},"User":{"type":"object","properties":{"allowAdultContent":{"type":"boolean"},"credentials":{"type":"string"},"id":{"type":"string"},"internal":{"type":"boolean"}},"additionalProperties":false,"required":["id","credentials","internal","allowAdultContent"]}}}`

func render(templateString string, api *JavaAPI) ([]byte, error) {
	//	tpl, err := template.New("schema").Funcs(template.FuncMap{
	//		"getDefault": preprocessDefault,
	//	}).Parse(templateString)

	templateString = strings.Replace(
		strings.Replace(
			strings.Replace(
				strings.Replace(templateString, "\\\n", "", -1),
				"\\n", "\n", -1),
			"~", "`", -1),
		"\\ ", "", -1)

	tpl, err := template.New("schema").Parse(templateString)
	if err != nil {
		return nil, fmt.Errorf("Could not parse template: %s", err)
	}

	buf := bytes.NewBuffer(nil)

	if err = tpl.Execute(buf, api); err != nil {
		return nil, fmt.Errorf("Could not execute template: %s", err)
	}

	return buf.Bytes(), nil

}

func TestSwagger(t *testing.T) {

	var api API

	if err := json.Unmarshal([]byte(swg), &api); err != nil {
		t.Fatal(err)
	}

	japi := NewJavaAPI(&api)

	fmt.Printf("%#v\n", japi)

	b, err := render(tpl, &japi)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(b))

}

/*
SwaggerVersion string            `json:"swagger"`
	Info           Info              `json:"info,omitempty"`
	Host           string            `json:"host"`
	Basepath       string            `json:"basePath"`
	Schemes        []string          `json:"schemes"`
	Consumes       []string          `json:"consumes"`
	Produces       []string          `json:"produces"`
	Paths          map[string]Path   `json:"paths"`
	Definitions    map[string]Schema `json:"definitions,omitempty"`
*/

func javaType(t *jsonschema.Type, namespace string) string {

	switch Type(t.Type) {
	case String:
		return "String"
	case Number:
		return "float"
	case Boolean:
		return "boolean"
	case Integer:
		return "int"
	case Array:
		return fmt.Sprintf("List<%s>", javaType(t.Items, namespace))
	case Object:
		return "Object"
	default:
		if t.Ref != "" {
			if namespace != "" {
				return namespace + "." + path.Base(t.Ref)
			}
			return path.Base(t.Ref)
		}
		if namespace != "" {
			return namespace + "." + t.Type
		}
		return t.Type

	}
}

func methodName(path, method string) string {

	return method + strings.Title(strings.Replace(path, "/", "", -1))
}

type JavaClass struct {
	Name       string
	Members    []JavaMember
	Extends    string
	Implements string
}

type JavaMember struct {
	Name string
	Type JavaTypeRef
}
type JavaTypeRef struct {
	Namespace string
	Type      string
	Contained []JavaTypeRef
}

func (t JavaTypeRef) String() string {

	ret := ""
	if t.Namespace != "" {
		ret = t.Namespace + "."
	}
	ret += t.Type
	if t.Contained != nil && len(t.Contained) > 0 {
		ret += "<"
		for _, sub := range t.Contained {
			ret += sub.String() + ","
		}

		ret = strings.TrimRight(ret, ",") + ">"
	}

	return ret
}

type JavaMethod struct {
	Name     string
	Returns  JavaTypeRef
	Params   []JavaParam
	HttpVerb string
	Doc      string
	Path     string
}

type JavaParam struct {
	Name string
	Doc  string
	Type JavaTypeRef
}

type JavaAPI struct {
	Name    string
	Doc     string
	Root    string
	Types   []JavaClass
	Methods []JavaMethod
}

func newJavaTypeRef(t *jsonschema.Type) JavaTypeRef {
	ret := JavaTypeRef{}
	switch Type(t.Type) {
	case String:
		ret.Type = "String"
	case Number:
		ret.Type = "Float"
	case Boolean:
		ret.Type = "Boolean"
	case Integer:
		ret.Type = "Integer"
	case Array:
		ret.Type = "List"
		ret.Contained = []JavaTypeRef{newJavaTypeRef(t.Items)}
	case Object:
		ret.Type = "Object"
	default:
		ret.Namespace = "Types"
		if t.Ref != "" {
			ret.Type = path.Base(t.Ref)
		} else {
			ret.Type = t.Type
		}
	}
	return ret
}

func newJavaClass(name string, t *jsonschema.Type) JavaClass {
	ret := JavaClass{
		Name:    name,
		Members: []JavaMember{},
	}

	for k, prop := range t.Properties {
		ret.Members = append(ret.Members, JavaMember{
			Name: k,
			Type: newJavaTypeRef(prop),
		})
	}
	return ret
}

func newJavaTypeSwagger(t Type) JavaTypeRef {
	ret := JavaTypeRef{}
	switch t {

	case Number:
		ret.Type = "Float"
	case Boolean:
		ret.Type = "Boolean"
	case Integer:
		ret.Type = "Integer"
		//	case Array:
		//		ret.Type = "List"
		//		ret.Contained = []JavaType{newJavaType(t.Items)}
	case Object:

		ret.Type = "Object"
	case String:
		fallthrough
	default:
		ret.Type = "String"
	}
	return ret
}

func formatMethodName(path, verb string) string {
	return strings.ToLower(verb) + strings.Title(strings.Replace(path, "/", "", -1))
}
func newJavaMethod(path, verb string, method Method) JavaMethod {

	ret := JavaMethod{
		Name:     formatMethodName(path, verb),
		Returns:  newJavaTypeRef(method.Responses["default"].Schema.Type),
		Params:   make([]JavaParam, 0, len(method.Parameters)),
		HttpVerb: strings.ToUpper(verb),
		Path:     path,
		Doc:      method.Description,
	}

	for _, param := range method.Parameters {

		jparm := JavaParam{
			Name: param.Name,
			Type: newJavaTypeSwagger(param.Type),
			Doc:  param.Description,
		}

		ret.Params = append(ret.Params, jparm)

	}
	return ret
}

func NewJavaAPI(swapi *API) JavaAPI {

	api := JavaAPI{
		Name:    swapi.Info.Title,
		Doc:     swapi.Info.Description,
		Root:    swapi.Basepath,
		Types:   make([]JavaClass, 0, len(swapi.Definitions)),
		Methods: make([]JavaMethod, 0, len(swapi.Paths)),
	}

	for name, tp := range swapi.Definitions {
		api.Types = append(api.Types, newJavaClass(name, tp.Type))
	}

	for path, methods := range swapi.Paths {
		for verb, method := range methods {

			api.Methods = append(api.Methods, newJavaMethod(path, verb, method))
		}
	}

	return api
}

var tpl = `
/** 
* Autogenerated Class {{.Name}}
*
* {{.Doc}}
*/
public class {{.Name}} extends BaseAPI {

    static final public String ROOT = "{{.Root}}";
  
    public static class Types {
        {{ range .Types }}
        public static class {{ .Name }} {\
            {{ range .Members }}\
               public {{ .Type.Type }} {{ .Name }};\
            {{ end }}
        }
        {{ end }}
    }
    
    
    public {{.Name}}(boolean secure, String host, Decoder decoder, Client client) {
        super(secure, host, decoder, client);
    }
    
    {{ range .Methods }}
        / **
        * {{ .Name }}
        * {{ .Doc }}\
        {{ range .Params }}\
        * @param {{ .Name }} 
        {{ end }}\
        **/
        public CompletableFuture<{{ .Returns }}> {{ .Name }}() {
            return perform(Request.Method.{{ .HttpVerb }}, ROOT, "{{.Path}}", new Request.ParamMap(),
                    parser({{ .Returns }}.class));
        }
    {{ end }}
}
`
