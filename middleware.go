package vertex

import "net/http"

type Middleware interface {
	Handle(w http.ResponseWriter, ctx *Request, next HandlerFunc) (interface{}, error)
}

type MiddlewareFunc func(http.ResponseWriter, *Request, HandlerFunc) (interface{}, error)

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
