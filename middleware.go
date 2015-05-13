package web2

import "net/http"

type Middleware interface {
	Handle(w http.ResponseWriter, r *http.Request, next HandlerFunc) (interface{}, error)
}

type MiddlewareFunc func(http.ResponseWriter, *http.Request, HandlerFunc) (interface{}, error)

type HandlerFunc func(http.ResponseWriter, *http.Request) (interface{}, error)

func (f MiddlewareFunc) Handle(w http.ResponseWriter, r *http.Request, next HandlerFunc) (interface{}, error) {
	return f(w, r, next)
}

type step struct {
	mw   Middleware
	next *step
}

func (s *step) Handle(w http.ResponseWriter, r *http.Request) (interface{}, error) {

	return s.mw.Handle(w, r, HandlerFunc(func(http.ResponseWriter, *http.Request) (interface{}, error) {

		if s.next != nil {
			return s.next.Handle(w, r)
		}

		return nil, nil

	}))
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
