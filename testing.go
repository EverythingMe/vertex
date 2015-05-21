package vertex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
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
	Test(*TestContext) error
	Category() string
}

// test categories
const (
	CriticalTests = "critical"
	WarningTests  = "warning"
	AllTests      = "all"
)

type testFunc struct {
	f        func(*TestContext) error
	category string
}

func (f testFunc) Test(ctx *TestContext) error {
	return f.f(ctx)
}

func (f testFunc) Category() string {
	return f.category
}

// CrititcalTest wraps testers to signify that the tester is considered critical
func CriticalTest(f func(ctx *TestContext) error) Tester {
	return testFunc{f, CriticalTests}
}

// WarningTest wraps testers to signify that the tester is a warning test
func WarningTest(f func(ctx *TestContext) error) Tester {
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
	panic("FATAL: " + fmt.Sprintf(format, params...))
}

func (t *TestContext) Skip() {
	panic(skipped)
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
	skipped = "SKIP"
	missing = "MISSING"
	failed  = "FAIL"
	passed  = "PASS"
)

// runTest safely runs a test and catches its output and panics
func (t *testRunner) runTest(tc Tester, path string) (status string, err error, msgs []string) {

	// recover from panics and analyze the input
	defer func() {
		e := recover()
		if e != nil {
			switch e {

			// t.Skip was called
			case skipped:
				status = skipped
				err = nil
				return
			// we paniced
			default:
				status = failed
				err = fmt.Errorf("%v", e)
				return
			}
		}
	}()

	// missing testers fail as missing
	if tc == nil {
		status = missing
		return
	}

	ctx := &TestContext{
		api:       t.api,
		addr:      t.serverAddr,
		routePath: path,
		messages:  make([]string, 0),
	}

	err = tc.Test(ctx)
	if err != nil {
		err = fmt.Errorf("ERROR: %v", err)
		status = failed
	} else {
		status = passed

	}
	msgs = ctx.messages
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

		var status string
		var err error
		var msgs []string

		st := time.Now()
		if tc == nil {
			status = missing
		} else {
			if t.shouldRun(tc) {
				status, err, msgs = t.runTest(tc, path)
			}
		}

		fmt.Fprintf(buf, "[%s]\t(%v)\n", status, time.Since(st))

		// Output log messages if we failed
		if err != nil {
			fmt.Fprintln(buf, " ", err)
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

	for path, route := range t.api.Routes {

		if parallel {
			wg.Add(1)
			go t.invokeTest(path, route.Test, wg)
		} else {
			t.invokeTest(path, route.Test, nil)
		}

	}

	wg.Wait()
	t.output.Flush()

	return nil

}
