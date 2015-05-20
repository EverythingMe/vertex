package web2

import (
	"errors"
	"fmt"
	"net/http"
	"path"

	"github.com/julienschmidt/httprouter"
)

// Server represents a multi-API http server with a single router
type Server struct {
	addr   string
	apis   []*API
	router *httprouter.Router
}

// NewServer creates a new blank server to add APIs to
func NewServer(addr string) *Server {
	return &Server{
		addr:   addr,
		apis:   make([]*API, 0),
		router: httprouter.New(),
	}
}

// Add an API to the server
func (s *Server) AddAPI(a *API) {
	a.configure(s.router)

	s.router.PanicHandler = func(w http.ResponseWriter, r *http.Request, v interface{}) {
		http.Error(w, fmt.Sprintf("PANIC handling request: %v", v), http.StatusInternalServerError)
	}
	fmt.Println(path.Join("/test", a.root(), ":category"))
	s.router.Handle("GET", path.Join("/test", a.root(), ":category"), a.testHandler("127.0.0.1"+s.addr))

	s.apis = append(s.apis, a)
}

// Handler returns the underlying router, mainly for testing
func (s *Server) Handler() http.Handler {
	return s.router
}

// Run runs the server if it has any APIs registered on it
func (s *Server) Run() error {
	if len(s.apis) == 0 {
		return errors.New("No APIs defined for server")
	}

	// Server the console swagger UI
	s.router.ServeFiles("/console/*filepath", http.Dir("./console"))

	return http.ListenAndServe(s.addr, s.router)
}
