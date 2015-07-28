package middleware

import (
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/dvirsky/go-pylog/logging"
	"gitlab.doit9.com/backend/instrument"
	"gitlab.doit9.com/server/vertex"
)

// ConnectionLimiter limits the maximum allowed open connections (actually concurrent running requests)
// on an API.
//
// If applied to the whole API, it limits the API as a whole. If an instance of the limiter is applied
// to a specific route, it limits the concurrent running requests of that route. A combination of the two
// can be applied - say 1000 concurrent requests on the whole API, and 100 concurrent on a specific route
type ConnectionLimiter struct {
	max     int32
	running int32
}

func NewConnectionLimiter(max int32) *ConnectionLimiter {
	ret := &ConnectionLimiter{
		max:     max,
		running: 0,
	}

	go ret.sampleInstrumentation()
	return ret
}

func (b *ConnectionLimiter) sampleInstrumentation() {

	host, _ := os.Hostname()

	key := fmt.Sprintf("concurrent_requests.%s", host)

	for range time.Tick(time.Second) {
		instrument.Gauge(key, int64(b.running))
	}

}

func (b *ConnectionLimiter) Handle(w http.ResponseWriter, r *vertex.Request, next vertex.HandlerFunc) (interface{}, error) {

	num := atomic.AddInt32(&b.running, 1)
	defer atomic.AddInt32(&b.running, -1)
	if num > b.max {
		instrument.Hit("over_capacity")
		logging.Warning("Connection limit exceeded: %d/%d", num, b.max)
		return nil, vertex.ResourceUnavailableError("Connection Limit Exceeded")
	}

	return next(w, r)

}
