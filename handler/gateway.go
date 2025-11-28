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

	if err := g.auth.Handle(ctx); err != nil {
		ctx.Error = err
		g.errorHandler.Handle(ctx)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := g.transform.TransformRequest(ctx); err != nil {
		ctx.Error = err
		g.errorHandler.Handle(ctx)
		http.Error(w, "Transform error", http.StatusInternalServerError)
		return
	}

	if matchedRoute != nil && len(matchedRoute.RequestTransform) > 0 {
		transformed, err := g.dslTransformer.TransformWithContext(ctx.RequestBody, matchedRoute.RequestTransform, ctx.Data)
		if err != nil {
			ctx.Error = err
			g.errorHandler.Handle(ctx)
			http.Error(w, fmt.Sprintf("DSL transform error: %v", err), http.StatusInternalServerError)
			return
		}
		ctx.RequestBody = transformed
	}

	if err := g.hookManager.Execute(hook.BeforeForward, ctx); err != nil {
		ctx.Error = err
		g.errorHandler.Handle(ctx)
		http.Error(w, "Hook error", http.StatusInternalServerError)
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
		g.errorHandler.Handle(ctx)
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

	if err := g.hookManager.Execute(hook.AfterForward, ctx); err != nil {
		ctx.Error = err
		g.errorHandler.Handle(ctx)
		http.Error(w, "Hook error", http.StatusInternalServerError)
		return
	}

	if err := g.transform.TransformResponse(ctx); err != nil {
		ctx.Error = err
		g.errorHandler.Handle(ctx)
		http.Error(w, "Transform error", http.StatusInternalServerError)
		return
	}

	if matchedRoute != nil && len(matchedRoute.ResponseTransform) > 0 {
		transformed, err := g.dslTransformer.TransformWithContext(ctx.ResponseBody, matchedRoute.ResponseTransform, ctx.Data)
		if err != nil {
			ctx.Error = err
			g.errorHandler.Handle(ctx)
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
