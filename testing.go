package vertex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
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
	serverURl string
	routePath string
	category  string
	messages  []string
	startTime time.Time
}

func (t *TestContext) Log(format string, params ...interface{}) {
	msg := fmt.Sprintf("%v> %s", time.Now().Format("15:04:05.000"), fmt.Sprintf(format, params...))
	logging.Info(msg)
	t.messages = append(t.messages, msg)

}

func (t *TestContext) Fatal(format string, params ...interface{}) {

	res := newTestResult(resultFatal, fmt.Sprintf(format, params...), 2, t)
	panic(res)
}

func (t *TestContext) Skip() {
	panic(newTestResult(resultSkipped, "", 2, t))
}

func (t *TestContext) Fail(format string, params ...interface{}) {
	panic(newTestResult(resultFailed, fmt.Sprintf(format, params...), 2, t))
}

func (t *TestContext) ServerUrl() string {
	return t.serverURl
}

// FormatUrl returns a fully formatted URL for the context's route, with all path params replaced by
// their respective values in the pathParams map
func (t *TestContext) FormatUrl(pathParams Params) string {

	u := fmt.Sprintf("%s%s", t.serverURl, t.api.FullPath(FormatPath(t.routePath, pathParams)))

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
	category  string
	serverURL string
	api       *API
	output    io.Writer
	formatter resultFormatter
}
type resultFormatter interface {
	format(testResult) error
}

const (
	TestFormatText = "text"
	TestFormatJson = "json"
)

type jsonResultFormatter struct {
	encoder *json.Encoder
	w       io.Writer
}

func newJsonResultFormatter(out io.Writer) *jsonResultFormatter {
	return &jsonResultFormatter{
		encoder: json.NewEncoder(out),
		w:       out,
	}
}

func (f jsonResultFormatter) format(r testResult) error {

	if err := f.encoder.Encode(r); err != nil {
		return err
	}

	return nil
}

type textResultFormatter struct {
	w *tabwriter.Writer
}

func newTextResultFormatter(w io.Writer) *textResultFormatter {

	return &textResultFormatter{
		w: tabwriter.NewWriter(w, 40, 8, 0, '\t', tabwriter.DiscardEmptyColumns),
	}
}

func (f textResultFormatter) format(result testResult) error {

	if _, err := fmt.Fprintf(f.w, "- %s\t(category: %s)\t[%s]\t(%v)\n", result.Path, result.Category, result.Result, result.Duration); err != nil {
		return err
	}

	// Output log messages if we failed
	if result.isFailure() {
		if result.Message != "" {
			if _, err := fmt.Fprintf(f.w, " ERROR in %s: %s\n", result.FailPoint, result.Message); err != nil {
				return err
			}
		}
		if result.Log != nil && len(result.Log) > 0 {
			fmt.Fprintln(f.w, "  Messages:")
			for _, msg := range result.Log {
				if _, err := fmt.Fprintln(f.w, "  ", msg); err != nil {
					return err
				}
			}
		}

		fmt.Fprintln(f.w, "")

	}

	return f.w.Flush()
}

func newTestRunner(output io.Writer, a *API, serverURL string, category string, format string) *testRunner {

	var formatter resultFormatter
	switch format {
	case TestFormatJson:
		formatter = newJsonResultFormatter(output)
	case TestFormatText:
		fallthrough
	default:
		formatter = newTextResultFormatter(output)

	}
	return &testRunner{
		serverURL: serverURL,
		category:  category,
		api:       a,
		output:    output,
		formatter: formatter,
	}
}

// test status codes
const (
	resultFatal   = "FATAL"
	resultSkipped = "SKIP"
	resultMissing = "MISSING"
	resultFailed  = "FAIL"
	resultPass    = "PASS"
)

type testResult struct {
	Result    string        `json:"result"`
	Path      string        `json:"path,omitempty"`
	Category  string        `json:"category,omitempty"`
	Log       []string      `json:"log,omitempty"`
	Message   string        `json:"message,omitempty"`
	FailPoint string        `json:"failpoint,omitempty"`
	Duration  time.Duration `json:"duration,omitempty"`
}

func (r testResult) isFailure() bool {
	return !(r.Result == resultPass || r.Result == resultSkipped)
}

func newTestResult(result, message string, depth int, ctx *TestContext) testResult {

	ret := testResult{
		Path:     ctx.routePath,
		Category: ctx.category,
		Result:   result,
		Message:  message,
		Duration: time.Since(ctx.startTime),
	}

	if ret.isFailure() {
		if pc, _, line, ok := runtime.Caller(depth); ok {

			f := runtime.FuncForPC(pc)
			if f != nil {
				ret.FailPoint = fmt.Sprintf("%s:%d", path.Base(f.Name()), line)
			}

		}
	}

	return ret
}

// runTest safely runs a test and catches its output and panics
func (t *testRunner) runTest(tc Tester, path string) (res testResult) {

	// missing testers fail as missing
	if tc == nil {
		res = newTestResult(resultMissing, "", 1, &TestContext{routePath: path, category: tc.Category(), startTime: time.Now()})
		return
	}

	ctx := &TestContext{
		api:       t.api,
		serverURl: t.serverURL,
		routePath: path,
		messages:  make([]string, 0),
		category:  tc.Category(),
		startTime: time.Now(),
	}

	// recover from panics and analyze the input
	defer func() {

		e := recover()
		if e != nil {

			switch x := e.(type) {
			case testResult:
				res = x
			default:
				res = newTestResult(resultFatal, fmt.Sprintf("Panic handling test: %v", x), 6, ctx)
			}

		}
		return
	}()

	tc.Test(ctx)
	res = newTestResult(resultPass, "", 1, ctx)

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
func (t *testRunner) invokeTest(path string, tc Tester) *testResult {

	if t.shouldRun(tc) {

		var result testResult
		if tc == nil || t.shouldRun(tc) {
			result = t.runTest(tc, path)
			logging.Info("Test result for %s: %#v", path, result)
			if err := t.formatter.format(result); err != nil {
				logging.Error("Error running formatter: %s", err)
			}
			return &result
		}

	}

	return nil
}

func (t *testRunner) Run() bool {

	reschan := make(chan *testResult)
	wg := sync.WaitGroup{}
	for _, route := range t.api.Routes {
		wg.Add(1)

		go func(route Route) {

			reschan <- t.invokeTest(route.Path, route.Test)
			wg.Done()
		}(route)

	}

	go func() {
		wg.Wait()
		close(reschan)
	}()

	success := true
	for res := range reschan {
		if res == nil {
			continue
		}

		if res.isFailure() {
			success = false
		}
	}

	return success

}

func RunCLITest(apiName, serverAddr, category, format string) bool {

	builder, ok := apiBuilders[apiName]
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: API %s not found\n", apiName)
		return false
	}

	a := builder()

	tr := newTestRunner(os.Stdout, a, serverAddr, category, format)
	return tr.Run()
}
