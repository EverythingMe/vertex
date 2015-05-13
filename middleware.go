package web2

import "net/http"

type Middleware interface {
	Handle(w http.ResponseWriter, r *http.Request, next Middleware) (interface{}, error)
}

type MiddlewareFunc func(http.ResponseWriter, *http.Request, Middleware) (interface{}, error)

func (f MiddlewareFunc) Handle(w http.ResponseWriter, r *http.Request, next Middleware) (interface{}, error) {
	return f(w, r, next)
}

type step struct {
	mw   Middleware
	next *step
}

func (s *step) Handle(w http.ResponseWriter, r *http.Request, next Middleware) (interface{}, error) {

	return s.mw.Handle(w, r, next)
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

func buildChain(mws []Middleware) *step {
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
			next: buildChain(mws[1:]),
		}
	}

}
