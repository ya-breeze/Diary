// Hand-written router helpers: the Route/Routes/Router types and NewRouter
// constructor allow extra (non-generated) routers to plug into the server.
package goserver

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Route defines the parameters for an API endpoint.
type Route struct {
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// Routes is a map of named API endpoints.
type Routes map[string]Route

// Router must be implemented by any extra (non-generated) router.
type Router interface {
	Routes() Routes
}

// NewRouter creates a gorilla/mux router from a list of Router implementations.
func NewRouter(routers ...Router) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, api := range routers {
		for name, route := range api.Routes() {
			var handler http.Handler = route.HandlerFunc
			handler = Logger(handler, name)
			router.
				Methods(route.Method).
				Path(route.Pattern).
				Name(name).
				Handler(handler)
		}
	}
	return router
}

// Logger wraps an http.Handler with request/response logging.
func Logger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		inner.ServeHTTP(w, r)
		slog.Debug(fmt.Sprintf(
			"%s %s %s %s",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		))
	})
}
