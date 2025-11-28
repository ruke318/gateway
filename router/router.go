package router

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ruke318/gateway/config"
)

type Router struct {
	routes        []config.RouteConfig
	defaultBackend string
}

func NewRouter(routes []config.RouteConfig, defaultBackend string) *Router {
	return &Router{
		routes:        routes,
		defaultBackend: defaultBackend,
	}
}

func (r *Router) Match(req *http.Request) (*config.RouteConfig, error) {
	for _, route := range r.routes {
		if r.matchRoute(req, &route) {
			return &route, nil
		}
	}
	return nil, fmt.Errorf("no matching route found for %s %s", req.Method, req.URL.Path)
}

func (r *Router) matchRoute(req *http.Request, route *config.RouteConfig) bool {
	if route.Method != "" && !strings.EqualFold(route.Method, req.Method) {
		return false
	}

	if route.Path == "" {
		return false
	}

	if route.Path == req.URL.Path {
		return true
	}

	if strings.HasSuffix(route.Path, "*") {
		prefix := strings.TrimSuffix(route.Path, "*")
		return strings.HasPrefix(req.URL.Path, prefix)
	}

	return false
}

func (r *Router) GetBackendURL(route *config.RouteConfig) string {
	if route.BackendURL != "" {
		return route.BackendURL
	}
	return r.defaultBackend
}

func (r *Router) GetBackendPath(route *config.RouteConfig, originalPath string) string {
	if route.BackendPath != "" {
		return route.BackendPath
	}
	return originalPath
}

func (r *Router) GetBackendMethod(route *config.RouteConfig, originalMethod string) string {
	if route.BackendMethod != "" {
		return route.BackendMethod
	}
	return originalMethod
}
