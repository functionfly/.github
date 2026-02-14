package api

import (
	"context"
	"net/http"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Server struct {
	db     *storage.PostgresDB
	router *mux.Router
}

func NewServer(db *storage.PostgresDB) *Server {
	s := &Server{
		db:     db,
		router: mux.NewRouter(),
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// API versioning
	api := s.router.PathPrefix("/v1").Subrouter()

	// Auth routes
	api.HandleFunc("/auth/login", s.handleLogin).Methods("POST")

	// App routes
	api.HandleFunc("/apps", s.handleCreateApp).Methods("POST")
	api.HandleFunc("/apps/{appId}", s.handleGetApp).Methods("GET")

	// Backend routes
	api.HandleFunc("/apps/{appId}/backends", s.handleCreateBackend).Methods("POST")
	api.HandleFunc("/apps/{appId}/backends", s.handleListBackends).Methods("GET")

	// Routing routes
	api.HandleFunc("/apps/{appId}/route", s.handleGetRoute).Methods("GET")

	// Health check
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")
}

func (s *Server) ListenAndServe(addr string) error {
	logrus.WithField("addr", addr).Info("API server listening")
	return http.ListenAndServe(addr, s.router)
}

func (s *Server) Shutdown(ctx context.Context) error {
	// For now, just return nil as we're using the basic HTTP server
	// TODO: Implement graceful shutdown with proper HTTP server
	return nil
}

// Placeholder handlers - to be implemented
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Login not implemented"))
}

func (s *Server) handleCreateApp(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Create app not implemented"))
}

func (s *Server) handleGetApp(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Get app not implemented"))
}

func (s *Server) handleCreateBackend(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Create backend not implemented"))
}

func (s *Server) handleListBackends(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("List backends not implemented"))
}

func (s *Server) handleGetRoute(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Get route not implemented"))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok"}`))
}