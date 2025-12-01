package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ruke318/gateway/config"
	"github.com/ruke318/gateway/hook"
	"github.com/ruke318/gateway/middleware"
	"github.com/ruke318/gateway/proxy"
	"github.com/ruke318/gateway/router"
	"github.com/ruke318/gateway/transform"
)

type Gateway struct {
	hookManager    *hook.Manager
	forwarder      *proxy.Forwarder
	auth           *middleware.AuthMiddleware
	transform      *middleware.TransformMiddleware
	errorHandler   *middleware.ErrorMiddleware
	router         *router.Router
	dslTransformer *transform.DSLTransformer
}

func NewGateway(hookManager *hook.Manager, forwarder *proxy.Forwarder, auth *middleware.AuthMiddleware, transform *middleware.TransformMiddleware, errorHandler *middleware.ErrorMiddleware, router *router.Router, dslTransformer *transform.DSLTransformer) *Gateway {
	return &Gateway{
		hookManager:    hookManager,
		forwarder:      forwarder,
		auth:           auth,
		transform:      transform,
		errorHandler:   errorHandler,
		router:         router,
		dslTransformer: dslTransformer,
	}
}

// executeHook 执行 Hook，优先执行接口级别的 Hook，如果没有则执行全局 Hook
func (g *Gateway) executeHook(point hook.HookPoint, route *config.RouteConfig, ctx *hook.HookContext) error {
	// 如果路由配置中有接口级别的 Hook，优先执行
	if route != nil && route.Hooks != nil {
		hookPointName := hookPointToString(point)
		if script, exists := route.Hooks[hookPointName]; exists && script != "" {
			// 创建临时 Hook 执行器
			executor := hook.NewJSExecutor(script)
			return executor.Execute(ctx)
		}
	}

	// 否则执行全局 Hook
	return g.hookManager.Execute(point, ctx)
}

// hookPointToString 将 HookPoint 转换为字符串
func hookPointToString(point hook.HookPoint) string {
	switch point {
	case hook.BeforeAuth:
		return "BeforeAuth"
	case hook.AfterAuth:
		return "AfterAuth"
	case hook.BeforeRequestTransform:
		return "BeforeRequestTransform"
	case hook.AfterRequestTransform:
		return "AfterRequestTransform"
	case hook.BeforeForward:
		return "BeforeForward"
	case hook.AfterForward:
		return "AfterForward"
	case hook.BeforeResponseTransform:
		return "BeforeResponseTransform"
	case hook.AfterResponseTransform:
		return "AfterResponseTransform"
	case hook.OnError:
		return "OnError"
	default:
		return ""
	}
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &hook.HookContext{
		Request:         r,
		RequestHeaders:  make(map[string]string),
		ResponseHeaders: make(map[string]string),
		Data:            make(map[string]interface{}),
	}

	for k, v := range r.Header {
		if len(v) > 0 {
			ctx.RequestHeaders[k] = v[0]
		}
	}

	body, _ := io.ReadAll(r.Body)
	ctx.RequestBody = body

	// 将请求体解析为 JSON 并添加到 ctx.Data，便于在 DSL 中访问
	var requestBodyData interface{}
	if len(body) > 0 {
		json.Unmarshal(body, &requestBodyData)
	}

	ctx.Data["request"] = map[string]interface{}{
		"method": r.Method,
		"path":   r.URL.Path,
		"query":  r.URL.RawQuery,
		"host":   r.Host,
		"header": ctx.RequestHeaders,
		"body":   requestBodyData,
	}

	var matchedRoute *config.RouteConfig
	if g.router != nil {
		route, err := g.router.Match(r)
		if err == nil {
			matchedRoute = route
			ctx.Data["route"] = map[string]interface{}{
				"path":          route.Path,
				"method":        route.Method,
				"backendUrl":    route.BackendURL,
				"backendPath":   route.BackendPath,
				"backendMethod": route.BackendMethod,
			}
		}
	}

	// BeforeAuth Hook（接口级别 Hook 优先）
	if err := g.executeHook(hook.BeforeAuth, matchedRoute, ctx); err != nil {
		ctx.Error = err
		g.executeHook(hook.OnError, matchedRoute, ctx)
		http.Error(w, "BeforeAuth error", http.StatusUnauthorized)
		return
	}

	// 认证逻辑
	if err := g.auth.Handle(ctx); err != nil {
		ctx.Error = err
		g.executeHook(hook.OnError, matchedRoute, ctx)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// AfterAuth Hook（接口级别 Hook 优先）
	if err := g.executeHook(hook.AfterAuth, matchedRoute, ctx); err != nil {
		ctx.Error = err
		g.executeHook(hook.OnError, matchedRoute, ctx)
		http.Error(w, "AfterAuth error", http.StatusInternalServerError)
		return
	}

	// BeforeRequestTransform Hook（接口级别 Hook 优先）
	if err := g.executeHook(hook.BeforeRequestTransform, matchedRoute, ctx); err != nil {
		ctx.Error = err
		g.executeHook(hook.OnError, matchedRoute, ctx)
		http.Error(w, "BeforeRequestTransform error", http.StatusInternalServerError)
		return
	}

	// AfterRequestTransform Hook（接口级别 Hook 优先）
	if err := g.executeHook(hook.AfterRequestTransform, matchedRoute, ctx); err != nil {
		ctx.Error = err
		g.executeHook(hook.OnError, matchedRoute, ctx)
		http.Error(w, "AfterRequestTransform error", http.StatusInternalServerError)
		return
	}

	if matchedRoute != nil && len(matchedRoute.RequestTransform) > 0 {
		transformed, err := g.dslTransformer.TransformWithContext(ctx.RequestBody, matchedRoute.RequestTransform, ctx.Data)
		if err != nil {
			ctx.Error = err
			g.executeHook(hook.OnError, matchedRoute, ctx)
			http.Error(w, fmt.Sprintf("DSL transform error: %v", err), http.StatusInternalServerError)
			return
		}
		ctx.RequestBody = transformed
	}

	// BeforeForward Hook（接口级别 Hook 优先）
	if err := g.executeHook(hook.BeforeForward, matchedRoute, ctx); err != nil {
		ctx.Error = err
		g.executeHook(hook.OnError, matchedRoute, ctx)
		http.Error(w, "BeforeForward error", http.StatusInternalServerError)
		return
	}

	var resp *http.Response
	var respBody []byte
	var err error

	if matchedRoute != nil {
		backendURL := g.router.GetBackendURL(matchedRoute)
		backendPath := g.router.GetBackendPath(matchedRoute, r.URL.Path)
		backendMethod := g.router.GetBackendMethod(matchedRoute, r.Method)
		resp, respBody, err = g.forwarder.ForwardWithOptions(backendMethod, backendURL, backendPath, ctx.RequestBody, r.Header)
	} else {
		resp, respBody, err = g.forwarder.Forward(r, ctx.RequestBody)
	}

	if err != nil {
		ctx.Error = err
		g.executeHook(hook.OnError, matchedRoute, ctx)
		http.Error(w, "Forward error", http.StatusBadGateway)
		return
	}

	ctx.Response = resp
	ctx.ResponseBody = respBody

	responseHeaders := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			responseHeaders[k] = v[0]
		}
	}
	ctx.Data["response"] = map[string]interface{}{
		"status": resp.StatusCode,
		"header": responseHeaders,
	}

	// AfterForward Hook（接口级别 Hook 优先）
	if err := g.executeHook(hook.AfterForward, matchedRoute, ctx); err != nil {
		ctx.Error = err
		g.executeHook(hook.OnError, matchedRoute, ctx)
		http.Error(w, "AfterForward error", http.StatusInternalServerError)
		return
	}

	// BeforeResponseTransform Hook（接口级别 Hook 优先）
	if err := g.executeHook(hook.BeforeResponseTransform, matchedRoute, ctx); err != nil {
		ctx.Error = err
		g.executeHook(hook.OnError, matchedRoute, ctx)
		http.Error(w, "BeforeResponseTransform error", http.StatusInternalServerError)
		return
	}

	// AfterResponseTransform Hook（接口级别 Hook 优先）
	if err := g.executeHook(hook.AfterResponseTransform, matchedRoute, ctx); err != nil {
		ctx.Error = err
		g.executeHook(hook.OnError, matchedRoute, ctx)
		http.Error(w, "AfterResponseTransform error", http.StatusInternalServerError)
		return
	}

	if matchedRoute != nil && len(matchedRoute.ResponseTransform) > 0 {
		transformed, err := g.dslTransformer.TransformWithContext(ctx.ResponseBody, matchedRoute.ResponseTransform, ctx.Data)
		if err != nil {
			ctx.Error = err
			g.executeHook(hook.OnError, matchedRoute, ctx)
			http.Error(w, fmt.Sprintf("DSL transform error: %v", err), http.StatusInternalServerError)
			return
		}
		ctx.ResponseBody = transformed
	}

	for k, v := range ctx.ResponseHeaders {
		w.Header().Set(k, v)
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(ctx.ResponseBody)
}
