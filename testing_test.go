package vertex

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var api = &API{
	Root:          "/mock",
	Name:          "testung",
	Version:       "1.0",
	Doc:           "Our fancy testung API",
	Title:         "Testung API!",
	Renderer:      JSONRenderer{},
	AllowInsecure: true,
	Middleware:    []Middleware{makeMockMW("Global middleware")},
	Routes: Routes{
		{
			Path:        "/test",
			Description: "test",
			Handler:     VoidHandler{},
			Methods:     GET,
			Test:        WarningTest(func(t *TestContext) {}),
		},
	},
}

func TestTestContext(t *testing.T) {

	tc := &TestContext{
		api:       api,
		addr:      "localhost:1277",
		routePath: "/test",
		messages:  []string{},
	}

	assert.Equal(t, tc.FormatUrl(nil), "http://localhost:1277/mock/test", "Badly formatted url")
	assert.Equal(t, tc.ServerUrl(), "http://localhost:1277", "Bad server url")

	tc.Log("Wor dawg")
	assert.Len(t, tc.messages, 1)

	vals := url.Values{}
	vals.Set("foo", "bar")
	req, err := tc.NewRequest("GET", vals, nil)
	assert.NoError(t, err, "Error creating request")
	assert.NotNil(t, req)
	assert.Nil(t, req.Body)
	assert.Equal(t, req.URL.String(), "http://localhost:1277/mock/test?foo=bar")

	testResults := func(f func()) (ret testResult) {

		defer func() {
			x := recover()
			if x != nil {
				ret = x.(testResult)
			}
		}()

		f()

		t.Errorf("We shuold have paniced during f()")
		return
	}

	res := testResults(func() {
		tc.Fail("WAT %s", "WAT")
	})

	assert.Equal(t, res.result, failed)
	assert.Equal(t, res.message, "WAT WAT")
	assert.Equal(t, res.file, "testing_test.go")
	assert.True(t, res.line > 0)

	res = testResults(func() {
		tc.Fatal("WAT %s", "WAT")
	})

	assert.Equal(t, res.result, fatal)
	assert.Equal(t, res.message, "WAT WAT")
	assert.Equal(t, res.file, "testing_test.go")
	assert.True(t, res.line > 0)

	res = testResults(func() {
		tc.Skip()
	})

	assert.Equal(t, res.result, skipped)
	assert.Equal(t, res.message, "")
	assert.Equal(t, res.file, "testing_test.go")
	assert.True(t, res.line > 0)

}

func TestTestRunner(t *testing.T) {

	outbuf := bytes.NewBuffer(nil)
	runner := newTestRunner(outbuf, api, "127.0.0.1:9947", WarningTests)
	err := runner.Run(false)
	if err != nil {
		t.Error(err)
	}

	os := outbuf.String()

	assert.True(t, strings.Contains(os, "Testing /test"))
	assert.True(t, strings.Contains(os, "category: warning"))
	assert.True(t, strings.Contains(os, "[PASS]"))

	fmt.Println(outbuf.String())
}
