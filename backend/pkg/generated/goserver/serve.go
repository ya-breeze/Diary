// Hand-written server bootstrap: CustomControllers, Serve, and CORS middleware.
// This replaces the previously template-generated main.go.
package goserver

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/ya-breeze/diary.be/pkg/config"
)

// CustomControllers holds the concrete service implementations.
type CustomControllers struct {
	AssetsAPIService AssetsAPIService
	AuthAPIService   AuthAPIService
	HealthAPIService HealthAPIService
	ItemsAPIService  ItemsAPIService
	SyncAPIService   SyncAPIService
	UserAPIService   UserAPIService
}

// Serve starts the HTTP server and returns the listening address and a finish channel.
func Serve(
	ctx context.Context,
	logger *slog.Logger,
	cfg *config.Config,
	controllers CustomControllers,
	extraRouters []Router,
	middlewares ...mux.MiddlewareFunc,
) (net.Addr, chan int, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen: %w", err)
	}
	logger.Info(fmt.Sprintf("Listening at port %d...", listener.Addr().(*net.TCPAddr).Port))

	// Build the base mux and register extra (custom) routes first so they take
	// priority over any generated routes for the same path.
	router := mux.NewRouter().StrictSlash(true)
	for _, r := range extraRouters {
		for name, route := range r.Routes() {
			var h http.Handler = route.HandlerFunc
			h = Logger(h, name)
			router.Methods(route.Method).Path(route.Pattern).Name(name).Handler(h)
		}
	}

	// Register generated routes via the strict handler.
	strictImpl := newStrictServerImpl(controllers)
	strictHandler := NewStrictHandlerWithOptions(strictImpl, nil, StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			logger.Error("API request error", "error", err, "path", r.URL.Path)
			http.Error(w, err.Error(), http.StatusBadRequest)
		},
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			logger.Error("API response error", "error", err, "path", r.URL.Path)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		},
	})
	HandlerWithOptions(strictHandler, GorillaServerOptions{BaseRouter: router})

	router.Use(middlewares...)

	allowedOrigins := []string{"http://localhost:3000"}
	if cfg.AllowedOrigins != "" {
		parts := strings.Split(cfg.AllowedOrigins, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		allowedOrigins = parts
	}

	server := &http.Server{
		Handler: createCORSMiddleware(allowedOrigins)(router),
	}

	go func() {
		_ = server.Serve(listener)
	}()

	finishChan := make(chan int, 1)
	go func() {
		<-ctx.Done()
		logger.Info("Shutting down server...")
		timeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(timeout)
		finishChan <- 1
		logger.Info("Server stopped")
	}()

	return listener.Addr(), finishChan, nil
}

// createCORSMiddleware creates a CORS middleware that only allows specific origins.
func createCORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[o] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && allowed[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Authorization")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
