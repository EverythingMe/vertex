package vertex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"runtime"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/dvirsky/go-pylog/logging"
)

// Tester represents a testcase the API runs for a certain API.
//
// Each API contains a list of integration tests that can be run to monitor it. Each test can have a category
// associated with it, and we can run tests by a specific category only.
//
// A test should fail or succeed, and can optionally write error output
type Tester interface {
	Test(*TestContext)
	Category() string
}

// test categories
const (
	CriticalTests = "critical"
	WarningTests  = "warning"
	AllTests      = "all"
)

type testFunc struct {
	f        func(*TestContext)
	category string
}

func (f testFunc) Test(ctx *TestContext) {
	f.f(ctx)
}

func (f testFunc) Category() string {
	return f.category
}

// CrititcalTest wraps testers to signify that the tester is considered critical
func CriticalTest(f func(ctx *TestContext)) Tester {
	return testFunc{f, CriticalTests}
}

// WarningTest wraps testers to signify that the tester is a warning test
func WarningTest(f func(ctx *TestContext)) Tester {
	return testFunc{f, WarningTests}
}

type TestContext struct {
	api       *API
	addr      string
	routePath string
	messages  []string
}

func (t *TestContext) Log(format string, params ...interface{}) {
	msg := fmt.Sprintf("%v> %s", time.Now().Format("15:04:05.000"), fmt.Sprintf(format, params...))
	logging.Info(msg)
	t.messages = append(t.messages, msg)

}

func (t *TestContext) Fatal(format string, params ...interface{}) {

	res := newTestResult(fatal, fmt.Sprintf(format, params...), 2)
	panic(res)
}

func (t *TestContext) Skip() {
	panic(newTestResult(skipped, "", 2))
}

func (t *TestContext) Fail(format string, params ...interface{}) {
	panic(newTestResult(failed, fmt.Sprintf(format, params...), 2))
}

func (t *TestContext) ServerUrl() string {
	return fmt.Sprintf("http://%s", t.addr)
}

// FormatUrl returns a fully formatted URL for the context's route, with all path params replaced by
// their respective values in the pathParams map
func (t *TestContext) FormatUrl(pathParams Params) string {

	u := fmt.Sprintf("http://%s%s", t.addr, t.api.FullPath(FormatPath(t.routePath, pathParams)))

	logging.Debug("Formatted url: %s", u)
	return u
}

func (t *TestContext) NewRequest(method string, values url.Values, pathParams Params) (*http.Request, error) {

	var body io.Reader
	u := t.FormatUrl(pathParams)

	if values != nil && len(values) > 0 {
		if method == "POST" {

			body = bytes.NewReader([]byte(values.Encode()))
		} else {
			u += "?" + values.Encode()
		}
	}

	req, err := http.NewRequest(method, u, body)

	// for POST requests we need to correctly set the content type
	if err == nil && body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	return req, err
}

func (t *TestContext) JsonRequest(r *http.Request, v interface{}) (*http.Response, error) {

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return resp, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	// We replace the request's body with a fake one if the caller wants to peek inside
	resp.Body = ioutil.NopCloser(bytes.NewReader(b))

	err = json.Unmarshal(b, v)
	if err == nil && resp.StatusCode >= 400 {
		err = fmt.Errorf("Bad HTTP response code: %s", resp.Status)
	}
	return resp, err

}

type testRunner struct {
	category   string
	serverAddr string
	api        *API
	output     *tabwriter.Writer
}

func newTestRunner(output io.Writer, a *API, addr string, category string) *testRunner {
	return &testRunner{
		serverAddr: addr,
		category:   category,
		api:        a,
		output:     tabwriter.NewWriter(output, 20, 8, 0, '\t', tabwriter.DiscardEmptyColumns),
	}
}

// test status codes
const (
	fatal   = "FATAL"
	skipped = "SKIP"
	missing = "MISSING"
	failed  = "FAIL"
	passed  = "PASS"
)

type testResult struct {
	result  string
	message string
	fun     string
	file    string
	line    int
}

func (r testResult) isFailure() bool {
	return !(r.result == passed || r.result == skipped)
}

func newTestResult(result, message string, depth int) testResult {

	ret := testResult{result: result, message: message}

	if pc, file, line, ok := runtime.Caller(depth); ok {
		ret.line = line
		ret.file = path.Base(file)
		f := runtime.FuncForPC(pc)
		if f != nil {
			ret.fun = path.Base(f.Name())
		}

	}

	return ret
}

// runTest safely runs a test and catches its output and panics
func (t *testRunner) runTest(tc Tester, path string) (res testResult, msgs []string) {

	// missing testers fail as missing
	if tc == nil {
		res = newTestResult(missing, "", 1)
		return
	}

	ctx := &TestContext{
		api:       t.api,
		addr:      t.serverAddr,
		routePath: path,
		messages:  make([]string, 0),
	}

	// recover from panics and analyze the input
	defer func() {
		msgs = ctx.messages

		e := recover()
		if e != nil {

			switch x := e.(type) {
			case testResult:
				res = x
			default:
				res = newTestResult(fatal, fmt.Sprintf("Panic handling test: %v", x), 4)
			}

		}
		return
	}()

	tc.Test(ctx)
	res = newTestResult(passed, "", 1)

	return

}

// Determine whether a test shuold run based on the context
func (t *testRunner) shouldRun(tc Tester) bool {

	if t.category == "" || t.category == AllTests {
		return true
	}
	if tc == nil {
		return false
	}

	return getTestCategory(tc) == t.category

}

func getTestCategory(tc Tester) string {
	if tc != nil {
		return tc.Category()
	}

	return AllTests
}

// invokeTest runs a tester and prints the output
func (t *testRunner) invokeTest(path string, tc Tester, wg *sync.WaitGroup) {

	if t.shouldRun(tc) {

		buf := bytes.NewBuffer(nil)

		fmt.Fprintf(buf, "Testing %s\t\t(category: %s)\t......\t", path, getTestCategory(tc))

		var result testResult
		var msgs []string

		st := time.Now()

		if tc == nil {
			result = newTestResult(missing, "", 1)
		} else {
			if t.shouldRun(tc) {
				result, msgs = t.runTest(tc, path)
			}
		}
		logging.Info("Test result for %s: %#v", path, result)

		fmt.Fprintf(buf, "[%s]\t(%v)\n", result.result, time.Since(st))

		// Output log messages if we failed
		if result.isFailure() {
			if result.message != "" {
				fmt.Fprintf(buf, " ERROR in %s:%d: %s\n", result.fun, result.line, result.message)
			}
			if msgs != nil && len(msgs) > 0 {
				fmt.Fprintln(buf, "  Messages:")
				for _, msg := range msgs {
					fmt.Fprintln(buf, "  ", msg)
				}
			}

		}
		fmt.Fprintln(buf, "")

		t.output.Write(buf.Bytes())
	}
	//t.output.Flush()

	if wg != nil {
		wg.Done()
	}
}

func (t *testRunner) Run(parallel bool) error {

	wg := &sync.WaitGroup{}

	for _, route := range t.api.Routes {

		if parallel {
			wg.Add(1)
			go t.invokeTest(route.Path, route.Test, wg)
		} else {
			t.invokeTest(route.Path, route.Test, nil)
		}

	}

	wg.Wait()
	t.output.Flush()

	return nil

}
