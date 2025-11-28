package router

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/ruke318/gateway/config"
)

type Router struct {
	routes         []config.RouteConfig
	defaultBackend string
	mu             sync.RWMutex
}

func NewRouter(routes []config.RouteConfig, defaultBackend string) *Router {
	return &Router{
		routes:         routes,
		defaultBackend: defaultBackend,
	}
}

// AddRoute 动态添加路由
// 如果路由已存在（相同 path 和 method），返回错误
func (r *Router) AddRoute(route config.RouteConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 检查路由是否已存在
	for _, existingRoute := range r.routes {
		if existingRoute.Path == route.Path && existingRoute.Method == route.Method {
			return fmt.Errorf("route already exists: %s %s", route.Method, route.Path)
		}
	}

	r.routes = append(r.routes, route)
	return nil
}

// UpdateRoute 动态更新路由（根据 path 和 method 匹配）
func (r *Router) UpdateRoute(route config.RouteConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, existingRoute := range r.routes {
		if existingRoute.Path == route.Path && existingRoute.Method == route.Method {
			r.routes[i] = route
			return nil
		}
	}
	return fmt.Errorf("route not found: %s %s", route.Method, route.Path)
}

// DeleteRoute 动态删除路由
func (r *Router) DeleteRoute(path, method string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, route := range r.routes {
		if route.Path == path && route.Method == method {
			r.routes = append(r.routes[:i], r.routes[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("route not found: %s %s", method, path)
}

// GetAllRoutes 获取所有路由配置
// 返回深拷贝，避免外部修改影响内部状态
func (r *Router) GetAllRoutes() []config.RouteConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 返回深拷贝，避免外部修改
	routes := make([]config.RouteConfig, len(r.routes))
	for i, route := range r.routes {
		routes[i] = route.DeepCopy()
	}
	return routes
}

func (r *Router) Match(req *http.Request) (*config.RouteConfig, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, route := range r.routes {
		if r.matchRoute(req, &route) {
			// 返回深拷贝，避免并发修改 map 导致 panic
			routeCopy := route.DeepCopy()
			return &routeCopy, nil
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
