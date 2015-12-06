package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/EverythingMe/vertex"
)

var mockkHandler = vertex.HandlerFunc(func(w http.ResponseWriter, r *vertex.Request) (interface{}, error) {
	return nil, nil
})

func TestIPFilter(t *testing.T) {

	flt := NewIPRangeFilter().AllowPrivate()
	flt.Allow("8.8.8.4")
	flt.Deny("127.0.0.2")
	hr, _ := http.NewRequest("GET", "/foo", nil)
	r := vertex.NewRequest(hr)
	checkAddr := func(addr string) error {
		r.RemoteIP = addr
		_, err := flt.Handle(httptest.NewRecorder(), r, mockkHandler)
		return err
	}

	assert.NoError(t, checkAddr("127.0.0.1"))
	assert.NoError(t, checkAddr("172.16.25.46"))
	assert.Error(t, checkAddr("8.8.8.8"))
	assert.NoError(t, checkAddr("8.8.8.4"))
	assert.Error(t, checkAddr("127.0.0.2"))

}

func TestAPIKeyValidator(t *testing.T) {

	v := NewAPIKeyValidator("apiKey", "foo", "bar")
	hr, _ := http.NewRequest("GET", "/foo", nil)
	r := vertex.NewRequest(hr)
	check := func(k string) error {
		r.Form.Set(v.paramName, k)
		_, err := v.Handle(httptest.NewRecorder(), r, mockkHandler)
		return err
	}

	assert.NoError(t, check("foo"))
	assert.NoError(t, check("bar"))
	assert.Error(t, check(""))
	assert.Error(t, check("sdfsdfsd"))

}
