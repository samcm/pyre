package server

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/samcm/pyre/internal/api"
	"github.com/sirupsen/logrus"
)

// Server defines the interface for the HTTP server
type Server interface {
	Start(ctx context.Context) error
	Stop() error
}

// server implements the HTTP server
type server struct {
	host       string
	port       int
	handler    *api.APIHandler
	frontend   embed.FS
	httpServer *http.Server
	log        logrus.FieldLogger
}

var _ Server = (*server)(nil)

// NewServer creates a new HTTP server
func NewServer(host string, port int, handler *api.APIHandler, frontend embed.FS, log logrus.FieldLogger) Server {
	return &server{
		host:     host,
		port:     port,
		handler:  handler,
		frontend: frontend,
		log:      log.WithField("package", "server"),
	}
}

// Start starts the HTTP server
func (s *server) Start(ctx context.Context) error {
	s.log.WithFields(logrus.Fields{
		"host": s.host,
		"port": s.port,
	}).Info("starting HTTP server")

	// Create router
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS middleware for development
	r.Use(corsMiddleware)

	// Mount API routes under /api/v1
	r.Route("/api/v1", func(r chi.Router) {
		api.HandlerFromMux(s.handler, r)
	})

	// Serve SPA for all other routes
	r.Get("/*", s.spaHandler())

	// Create HTTP server
	addr := net.JoinHostPort(s.host, fmt.Sprintf("%d", s.port))
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.log.WithError(err).Error("HTTP server error")
		}
	}()

	s.log.WithField("addr", addr).Info("HTTP server started")
	return nil
}

// Stop stops the HTTP server
func (s *server) Stop() error {
	s.log.Info("stopping HTTP server")

	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.httpServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown HTTP server: %w", err)
		}
	}

	s.log.Info("HTTP server stopped")
	return nil
}

// spaHandler serves the SPA frontend
func (s *server) spaHandler() http.HandlerFunc {
	// Pre-compute the dist filesystem
	distFS, err := fs.Sub(s.frontend, "frontend/dist")
	if err != nil {
		s.log.WithError(err).Fatal("failed to get dist subdirectory")
	}

	// Pre-read index.html for SPA fallback
	indexHTML, err := fs.ReadFile(distFS, "index.html")
	if err != nil {
		s.log.WithError(err).Fatal("failed to read index.html")
	}

	fileServer := http.FileServer(http.FS(distFS))

	return func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the requested file
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		// Check if file exists
		file, err := distFS.Open(path)
		if err == nil {
			stat, statErr := file.Stat()
			file.Close()
			if statErr == nil && !stat.IsDir() {
				// File exists, serve it
				fileServer.ServeHTTP(w, r)
				return
			}
		}

		// File not found or is directory, serve index.html for SPA routing
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(indexHTML)
	}
}

// corsMiddleware adds CORS headers for development
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
