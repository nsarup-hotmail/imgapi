package httpapi

import (
	"net/http"

	"github.com/nsarup/imgapi/internal/config"
	"github.com/nsarup/imgapi/internal/logging"
	"github.com/nsarup/imgapi/internal/service"
)

// Server encapsulates the HTTP layer.
type Server struct {
	cfg config.Config
	log *logging.Logger
	svc *service.Service
	mux *http.ServeMux
}

// NewServer constructs a new HTTP server with routes wired.
func NewServer(cfg config.Config, log *logging.Logger, svc *service.Service) *Server {
	s := &Server{cfg: cfg, log: log, svc: svc, mux: http.NewServeMux()}
	s.routes()
	return s
}

// Handler returns the http.Handler for this server.
func (s *Server) Handler() http.Handler { return s.mux }

func (s *Server) routes() {
	s.mux.HandleFunc("/healthz", s.handleHealth)
	s.mux.HandleFunc("/images", s.handleImages)    // POST
	s.mux.HandleFunc("/images/", s.handleGetImage) // GET
}
