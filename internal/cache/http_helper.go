package cache

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/1055373165/ggcache/pkg/common/logger"
)

// Server represents an HTTP cache server.
type HTTPServer struct {
	srv     *http.Server
	handler *HTTPPool
	cache   *Group
	stopCh  chan struct{}
	wg      sync.WaitGroup
}

// NewServer creates a new HTTP cache server instance.
func NewHTTPServer(currentSrvAddr string, peers []string, cache *Group) *HTTPServer {
	h := NewHTTPPool(currentSrvAddr)
	h.UpdatePeers(peers...)
	cache.RegisterServer(h)

	return &HTTPServer{
		handler: h,
		cache:   cache,
		stopCh:  make(chan struct{}),
	}
}

// StartHTTPCacheServer starts a cache server with graceful shutdown support.
func StartHTTPCacheServer(currentSrvAddr string, peers []string, cache *Group) error {
	if !strings.HasPrefix(currentSrvAddr, "http://") {
		return fmt.Errorf("invalid address format: %s, must start with http://", currentSrvAddr)
	}

	server := NewHTTPServer(currentSrvAddr, peers, cache)
	if err := server.Start(currentSrvAddr[7:]); err != nil {
		return fmt.Errorf("failed to start HTTP cache server: %w", err)
	}

	// Setup signal handling
	server.handleSignals()
	return nil
}

// Start initializes and starts the HTTP server
func (s *HTTPServer) Start(addr string) error {
	s.srv = &http.Server{
		Addr:    addr,
		Handler: s.handler,
	}

	// Start server in a goroutine
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		logger.LogrusObj.Infof("cache service is running at %v", addr)
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.LogrusObj.Errorf("HTTP server error: %v", err)
		}
	}()

	return nil
}

// Stop gracefully shuts down the server
func (s *HTTPServer) Stop(ctx context.Context) error {
	select {
	case <-s.stopCh:
		// Already stopped
		return nil
	default:
		close(s.stopCh)
	}

	if err := s.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	// Wait for all goroutines to finish
	s.wg.Wait()
	return nil
}

// handleSignals sets up signal handling for graceful shutdown
func (s *HTTPServer) handleSignals() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		logger.LogrusObj.Infof("received signal: %v", sig)

		// Create a context with timeout for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.Stop(ctx); err != nil {
			logger.LogrusObj.Errorf("server shutdown error: %v", err)
		}
	}()
}

// APIServer represents an HTTP API server
type APIServer struct {
	srv    *http.Server
	cache  *Group
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewAPIServer creates a new HTTP API server instance
func NewAPIServer(cache *Group) *APIServer {
	return &APIServer{
		cache:  cache,
		stopCh: make(chan struct{}),
	}
}

// StartHTTPAPIServer starts an API server with graceful shutdown support
func StartHTTPAPIServer(apiAddr string, cache *Group) error {
	if !strings.HasPrefix(apiAddr, "http://") {
		return fmt.Errorf("invalid address format: %s, must start with http://", apiAddr)
	}

	server := NewAPIServer(cache)
	if err := server.Start(apiAddr[7:]); err != nil {
		return fmt.Errorf("failed to start HTTP API server: %w", err)
	}

	// Setup signal handling
	server.handleSignals()
	return nil
}

// Start initializes and starts the API server
func (s *APIServer) Start(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api", s.handleAPIRequest)

	s.srv = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Start server in a goroutine
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		logger.LogrusObj.Infof("API server is running at %v", addr)
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.LogrusObj.Errorf("API server error: %v", err)
		}
	}()

	return nil
}

// Stop gracefully shuts down the API server
func (s *APIServer) Stop(ctx context.Context) error {
	select {
	case <-s.stopCh:
		// Already stopped
		return nil
	default:
		close(s.stopCh)
	}

	if err := s.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("API server shutdown failed: %w", err)
	}

	// Wait for all goroutines to finish
	s.wg.Wait()
	return nil
}

// handleSignals sets up signal handling for graceful shutdown
func (s *APIServer) handleSignals() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		logger.LogrusObj.Infof("received signal: %v", sig)

		// Create a context with timeout for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := s.Stop(ctx); err != nil {
			logger.LogrusObj.Errorf("API server shutdown error: %v", err)
		}
	}()
}

// handleAPIRequest handles the API requests
func (s *APIServer) handleAPIRequest(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "missing key parameter", http.StatusBadRequest)
		return
	}

	view, err := s.cache.Get(key)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get cache: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	if _, err := w.Write(view.Bytes()); err != nil {
		logger.LogrusObj.Errorf("failed to write response: %v", err)
	}
}
