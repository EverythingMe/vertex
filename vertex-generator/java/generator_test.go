package java

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/EverythingMe/vertex/swagger"
)

var swg = `
{"swagger":"2.0","info":{"version":"3.0","title":"EverythingMe3API","description":"EverythingMe's context aware API"},"host":"localhost:9944","basePath":"/evme/3.0","schemes":["http","https"],"consumes":["text/json"],"produces":["text/json"],"paths":{"/Context":{"get":{"description":"Get the extended context from the server","parameters":[{"$ref":"#/parameters/apiKey"},{"name":"X-Evme-Ctx","description":"Encoded Context Snapshot (base64 thrift)","type":"string","in":"header"}],"responses":{"default":{"description":"context.ExtendedContext","schema":{"$ref":"#/definitions/ExtendedContext"}}},"tags":["Context"]}},"/Echo":{"get":{"description":"Echo the request data and context for debugging","parameters":[{"$ref":"#/parameters/apiKey"},{"name":"X-Evme-Ctx","description":"Encoded Context Snapshot (base64 thrift)","type":"string","in":"header"}],"responses":{"default":{"description":"context.Context","schema":{"$ref":"#/definitions/Context"}}},"tags":["Echo"]}},"/ping":{"get":{"description":"test ping","responses":{"default":{"description":"string","schema":{"type":"string"}}},"tags":["Ping"]}}},"definitions":{"APIKey":{"type":"object","properties":{"enabled":{"type":"boolean"},"key":{"type":"string"},"name":{"type":"string"}},"additionalProperties":false,"required":["key","name","enabled"]},"Client":{"type":"object","properties":{"deviceId":{"type":"string"},"locale":{"type":"string"},"os":{"type":"string"},"version":{"type":"string"}},"additionalProperties":false,"required":["os","version","deviceId","locale"]},"ClientContext":{"type":"object","properties":{"activeInsights":{"type":"array","items":{"$ref":"#/definitions/Insight"}},"activeScenarios":{"type":"array","items":{"type":"string"}},"device":{"$ref":"#/definitions/DeviceContext"},"location":{"$ref":"#/definitions/GeoContext"},"motionState":{"$ref":"#/definitions/MotionState"},"network":{"$ref":"#/definitions/NetworkContext"},"phone":{"$ref":"#/definitions/PhoneContext"},"time":{"$ref":"#/definitions/TimeContext"}},"additionalProperties":false},"Context":{"type":"object","properties":{"client":{"$ref":"#/definitions/ClientContext"},"ext":{"$ref":"#/definitions/ExtendedContext"},"request":{"$ref":"#/definitions/RequestContext"}},"additionalProperties":false,"required":["client","request"]},"DeviceContext":{"type":"object","properties":{"battery":{"type":"integer"},"batteryPercent":{"type":"integer"},"deviceName":{"type":"string"},"dockType":{"type":"integer"},"headphonesConnected":{"type":"boolean"},"locale":{"type":"string"},"manufacturer":{"type":"string"},"os":{"type":"string"},"osVersion":{"type":"string"},"screenActive":{"type":"boolean"},"screenHeight":{"type":"integer"},"screenWidth":{"type":"integer"},"yearClass":{"type":"integer"}},"additionalProperties":false,"required":["battery","batteryPercent","headphonesConnected","screenActive","dockType","deviceName","manufacturer","os","osVersion","screenWidth","screenHeight","locale","yearClass"]},"ExtendedContext":{"type":"object","properties":{"insights":{"type":"array","items":{"$ref":"#/definitions/ServerInsight"}},"scenartios":{"type":"array","items":{"type":"string"}}},"additionalProperties":false,"required":["insights","scenartios"]},"GeoContext":{"type":"object","properties":{"accuracy":{"type":"number"},"homeCountry":{"type":"string"},"isRoaming":{"type":"boolean"},"isTravelling":{"type":"boolean"},"knownLocationCertainty":{"type":"number"},"knownLocationId":{"type":"string"},"lat":{"type":"number"},"lon":{"type":"number"},"simCountry":{"type":"string"}},"additionalProperties":false,"required":["lat","lon","accuracy","knownLocationCertainty","simCountry","homeCountry","isTravelling","isRoaming","knownLocationId"]},"Insight":{"type":"object","properties":{"confidence":{"type":"number"},"lastUpdatePeriod":{"type":"integer"},"params":{"type":"object","patternProperties":{".*":{"type":"string"}}},"type":{"type":"string"},"values":{"type":"array","items":{"type":"string"}}},"additionalProperties":false,"required":["type","confidence","lastUpdatePeriod"]},"MotionState":{"type":"object","properties":{"activity":{"type":"integer"},"confidence":{"type":"integer"}},"additionalProperties":false,"required":["activity","confidence"]},"NetworkContext":{"type":"object","properties":{"BTConnectedDevices":{"type":"array","items":{"type":"string"}},"BTEnabled":{"type":"boolean"},"activeNetwork":{"type":"string"},"availableWifiNetworks":{"type":"array","items":{"type":"string"}},"carrier":{"type":"string"},"mobileCountryCode":{"type":"integer"},"mobileNetworkCode":{"type":"integer"},"networkStrength":{"type":"integer"},"wifiName":{"type":"string"}},"additionalProperties":false,"required":["activeNetwork","networkStrength","wifiName","availableWifiNetworks","carrier","mobileNetworkCode","mobileCountryCode","BTEnabled","BTConnectedDevices"]},"PhoneContext":{"type":"object","properties":{"missedCall":{"type":"string"}},"additionalProperties":false,"required":["missedCall"]},"RequestContext":{"type":"object","properties":{"api":{"$ref":"#/definitions/APIKey"},"client":{"$ref":"#/definitions/Client"},"featureFlags":{"type":"array","items":{"type":"string"}},"homeCountryLocale":{"type":"string"},"ip":{"type":"string"},"requestId":{"type":"string"},"resolvedLocation":{"$ref":"#/definitions/TGeoLocation"},"trace":{"type":"array","items":{"$ref":"#/definitions/ServerHit"}},"user":{"$ref":"#/definitions/User"}},"additionalProperties":false,"required":["ip","requestId","featureFlags","homeCountryLocale"]},"ServerHit":{"type":"object","properties":{"cluster":{"type":"string"},"hostName":{"type":"string"},"timestamp":{"type":"integer"}},"additionalProperties":false,"required":["hostName","cluster","timestamp"]},"ServerInsight":{"type":"object","properties":{"TTL":{"type":"integer"},"confidence":{"type":"number"},"goefenceRadius":{"type":"number"},"jsonValue":{"type":"string"},"type":{"type":"string"}},"additionalProperties":false,"required":["type","jsonValue","confidence","TTL","goefenceRadius"]},"TGeoLocation":{"type":"object","properties":{"city":{"type":"string"},"country":{"type":"string"},"countryCode":{"type":"string"},"countryId":{"type":"integer"},"lat":{"type":"number"},"lon":{"type":"number"},"name":{"type":"string"},"state":{"type":"string"},"type":{"type":"string"},"zip":{"type":"string"}},"additionalProperties":false,"required":["lat","lon","name","type","city","zip","state","country","countryCode","countryId"]},"TimeContext":{"type":"object","properties":{"isWeekend":{"type":"boolean"},"localTime":{"type":"integer"},"timeOfDay":{"type":"integer"},"timeZone":{"type":"integer"}},"additionalProperties":false,"required":["localTime","timeZone","timeOfDay","isWeekend"]},"User":{"type":"object","properties":{"allowAdultContent":{"type":"boolean"},"credentials":{"type":"string"},"id":{"type":"string"},"internal":{"type":"boolean"}},"additionalProperties":false,"required":["id","credentials","internal","allowAdultContent"]}},"parameters":{"apiKey":{"name":"apiKey","description":"Given API Key","type":"string","in":"query"}}}`

func TestGenerate(t *testing.T) {

	var api swagger.API

	if err := json.Unmarshal([]byte(swg), &api); err != nil {
		t.Fatal(err)
	}

	g := &Generator{
		substitutions: map[string]string{
			"Client": "me.everything.context.thrift.Client",
		},
	}

	b, err := g.Generate(&api)
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, len(b) > 1000)

	// TODO: Real validation of the output. Right now it's not stable enough to validate
	fmt.Println(string(b))

}

func TestJavaAPI(t *testing.T) {
	var api swagger.API

	if err := json.Unmarshal([]byte(swg), &api); err != nil {
		t.Fatal(err)
	}

	g := &Generator{
		substitutions: map[string]string{
			"Client": "me.everything.context.thrift.Client",
		},
	}

	japi := g.newJavaAPI(&api)

	assert.Equal(t, api.Info.Title, japi.Name)
	assert.Equal(t, api.Info.Description, japi.Doc)
	assert.Equal(t, len(api.Paths), len(japi.Methods))
	assert.Equal(t, len(api.Definitions), len(japi.Types))

	for _, tp := range japi.Types {

		assert.NotEmpty(t, tp.Name)
		if sub, found := g.substitutions[tp.Name]; found {
			assert.Equal(t, sub, tp.Extends)
		} else {
			assert.Empty(t, tp.Extends)
		}

		assert.True(t, len(tp.Members) > 0)
	}

	for _, method := range japi.Methods {
		assert.NotEmpty(t, method.Name)
		assert.NotEmpty(t, method.Path)
		assert.NotNil(t, method.Returns)
	}

}

func TestFormatMethodName(t *testing.T) {

	tests := [][]string{
		{"GET", "//User/byId", "getUserById"},
		{"POst", "/User/byId", "postUserById"},
		{"GET", "User/ByID", "getUserByID"},
		{"GET", "/User/by_id.foo", "getUserByidfoo"},
		{"GET", "/User/by/id/and/name", "getUserByIdAndName"},
	}

	for _, args := range tests {

		name := formatMethodName(args[1], args[0])
		assert.Equal(t, args[2], name)

	}
}
