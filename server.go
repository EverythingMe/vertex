package vertex

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"path"
	"sync"
	"time"

	"github.com/hydrogen18/stoppableListener"
	"github.com/julienschmidt/httprouter"
)

// Server represents a multi-API http server with a single router
type Server struct {
	addr     string
	apis     []*API
	router   *httprouter.Router
	listener net.Listener
	wg       sync.WaitGroup
}

type builderFunc func() *API

var apiBuilders = map[string]builderFunc{}

// Register lest you automatically add an API to the server from your module's init() function.
//
// name is a unique name for your API (doesn't have to match the API name exactly).
//
// builder is a func that creates the API when we are ready to start the server.
//
// Optionally, you can pass a pointer to a config struct, or nil if you don't need to. This way, we can read the config struct's values
// from a unified config file BEFORE we call the builder, so the builder can use values in the config struct.
func Register(name string, builder func() *API, config interface{}) {
	//logging.Info("Adding api builder %s", name)
	apiBuilders[name] = builderFunc(builder)

	if config != nil {
		registerAPIConfig(name, config)
	}
}

// NewServer creates a new blank server to add APIs to
func NewServer(addr string) *Server {
	return &Server{
		addr:   addr,
		apis:   make([]*API, 0),
		router: httprouter.New(),
	}
}

// AddAPI adds an API to the server manually. It's preferred to use Register in an init() function
func (s *Server) AddAPI(a *API) {
	a.configure(s.router)

	s.router.PanicHandler = func(w http.ResponseWriter, r *http.Request, v interface{}) {
		http.Error(w, fmt.Sprintf("PANIC handling request: %v", v), http.StatusInternalServerError)
	}

	s.router.Handle("GET", path.Join("/test", a.root(), ":category"), a.testHandler)

	s.apis = append(s.apis, a)
}

// Handler returns the underlying router, mainly for testing
func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) InitAPIs() {
	for _, builder := range apiBuilders {
		s.AddAPI(builder())
	}
}

// Run runs the server if it has any APIs registered on it
func (s *Server) Run() (err error) {

	if len(s.apis) == 0 {
		return errors.New("No APIs defined for server")
	}

	// Server the console swagger UI
	s.router.ServeFiles("/console/*filepath", http.Dir("../console"))

	// Start a stoppable listener
	var l net.Listener
	if l, err = net.Listen("tcp", s.addr); err != nil {
		return fmt.Errorf("Could not listen in server: %s", err)
	}

	if s.listener, err = stoppableListener.New(l); err != nil {
		return fmt.Errorf("Could not start stoppable listener in server: %s", err)
	}

	s.wg.Add(1)
	defer func() {
		s.wg.Done()
		// don't return an error on server stopped
		if err == stoppableListener.StoppedError {
			err = nil
		}
	}()

	srv := http.Server{
		Handler:      s.router,
		ReadTimeout:  time.Duration(Config.Server.ClientTimeout) * time.Second,
		WriteTimeout: time.Duration(Config.Server.ClientTimeout) * time.Second, // maximum duration before timing out write of the response
	}
	return srv.Serve(s.listener)

}

// Stop waits up to a second and closes the server
func (s *Server) Stop() {

	s.listener.(*stoppableListener.StoppableListener).Stop()
	s.wg.Wait()
}
