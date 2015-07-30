package vertex

import "net/http"

// Middleware are pre/post processors that can inspect, change, or fail the request. e.g. authentication, logging, etc
//
// Each middleware needs to call next(w,r) so its next-in-line middleware will work, or return without it if it wishes to
// terminate the processing chain
type Middleware interface {
	Handle(w http.ResponseWriter, r *Request, next HandlerFunc) (interface{}, error)
}

// MiddlewareChain just wraps a variadic list of middlewares to make your code less ugly :)
func MiddlewareChain(mw ...Middleware) []Middleware {
	return mw
}

// MiddlewareFunc is a wrapper that allows functions to act as middleware
type MiddlewareFunc func(http.ResponseWriter, *Request, HandlerFunc) (interface{}, error)

// Handle runs the underlying func
func (f MiddlewareFunc) Handle(w http.ResponseWriter, r *Request, next HandlerFunc) (interface{}, error) {
	return f(w, r, next)
}

type step struct {
	mw   Middleware
	next *step
}

func (s *step) handle(w http.ResponseWriter, r *Request) (interface{}, error) {

	return s.mw.Handle(w, r, HandlerFunc(s.next.handle))
}

func (s *step) append(mw Middleware) {

	if s.next == nil {
		s.next = &step{
			mw:   mw,
			next: nil,
		}
	} else {
		s.next.append(mw)
	}
}

func buildChain(mws ...Middleware) *step {
	if mws == nil {
		return nil
	}
	switch len(mws) {
	case 0:
		return nil
	case 1:
		return &step{
			mw:   mws[0],
			next: nil,
		}
	default:
		return &step{
			mw:   mws[0],
			next: buildChain(mws[1:]...),
		}
	}

}
