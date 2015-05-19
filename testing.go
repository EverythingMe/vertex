package web2

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"text/tabwriter"
	"time"

	"github.com/dvirsky/go-pylog/logging"
)

// TestCase represents a testcase the API runs for a certain API.
//
// Each API contains a list of integration tests that can be run to monitor it. Each test can have a category
// associated with it, and we can run tests by a specific category only.
//
// A test should fail or succeed, and can optionally write error output
type Tester interface {
	Category() string
	Run(*TestContext) error
}

type testFunc struct {
	name     string
	f        func(*TestContext) error
	category string
}

func (t testFunc) String() string {
	return t.name
}

func TestFunc(name, category string, f func(*TestContext) error) testFunc {
	return testFunc{
		name:     name,
		category: category,
		f:        f,
	}
}

func (t testFunc) Category() string {
	return t.category
}
func (t testFunc) Run(ctx *TestContext) error {
	return t.f(ctx)
}

func testCaseName(tc Tester) string {

	if s, ok := tc.(fmt.Stringer); ok {
		return s.String()
	}
	return reflect.TypeOf(tc).Name()

}

type TestContext struct {
	api      *API
	addr     string
	testName string
	messages []string
}

func (t *TestContext) Log(format string, params ...interface{}) {
	msg := fmt.Sprintf("[%v] %s", time.Now().Format("15:04:05.000"), fmt.Sprintf(format, params...))
	logging.Info(msg)
	t.messages = append(t.messages, msg)

}

func (t *TestContext) Fatal(format string, params ...interface{}) {
	panic("FATAL: " + fmt.Sprintf(format, params...))
}

func (t *TestContext) ServerUrl() string {
	return fmt.Sprintf("http://%s", t.addr)
}

func (t *TestContext) FormatUrl(route string) string {
	return fmt.Sprintf("http://%s/%s", t.addr, t.api.FullPath(route))
}

func (t *TestContext) NewRequest(method, relpath string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(method, t.FormatUrl(relpath), body)
}

type TestRunner struct {
	category   string
	serverAddr string
	a          *API
	output     *tabwriter.Writer
}

func newTestRunner(output io.Writer, a *API, addr string, category string) *TestRunner {
	return &TestRunner{
		serverAddr: addr,
		category:   category,
		a:          a,
		output:     tabwriter.NewWriter(output, 0, 8, 40, '\t', 0),
	}
}

func (t *TestRunner) runTest(tc Tester) (err error, msgs []string) {

	defer func() {
		e := recover()
		if e != nil {
			err = fmt.Errorf("%v", e)
			return
		}
	}()

	ctx := &TestContext{
		api:      t.a,
		addr:     t.serverAddr,
		messages: make([]string, 0),
		testName: testCaseName(tc),
	}

	err = tc.Run(ctx)

	return fmt.Errorf("ERROR: ", err), ctx.messages
}
func (t *TestRunner) Run(tests []Tester) error {

	success := true
	for n, tc := range tests {
		if t.category == "" || t.category == tc.Category() {
			if _, err := fmt.Fprintf(t.output, "%d. Running test %s ....\t", n+1, testCaseName(tc)); err != nil {
				return err
			}

			st := time.Now()
			err, msgs := t.runTest(tc)
			if err != nil {
				fmt.Fprintf(t.output, "[FAIL] (%v)\n", time.Since(st))
				fmt.Fprintln(t.output, err)
				if msgs != nil && len(msgs) > 0 {
					fmt.Fprintln(t.output, "Output:")
					for _, msg := range msgs {
						fmt.Fprintln(t.output, msg)
					}
				}
				success = false

			} else {
				fmt.Fprintf(t.output, "[PASS] (%v)\n", time.Since(st))

			}

			fmt.Fprintln(t.output, "--------------------------------------------\n")
			t.output.Flush()
		}

	}

	if !success {
		return errors.New("At least one tester failed")
	}
	return nil
}
