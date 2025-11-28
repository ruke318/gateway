package main

import (
	"log"
	"net/http"

	"github.com/ruke318/gateway/config"
	"github.com/ruke318/gateway/handler"
	"github.com/ruke318/gateway/hook"
	"github.com/ruke318/gateway/middleware"
	"github.com/ruke318/gateway/proxy"
	"github.com/ruke318/gateway/router"
	"github.com/ruke318/gateway/transform"
)

func main() {
	cfg := config.Load()
	hookManager := hook.NewManager()
	hookManager.RegisterScript(hook.BeforeAuth, "scripts/examples/auth.js")
	hookManager.RegisterScript(hook.AfterRequestTransform, "scripts/examples/transform.js")
	hookManager.RegisterScript(hook.OnError, "scripts/examples/error.js")

	forwarder := proxy.NewForwarder(cfg.BackendURL)
	auth := middleware.NewAuthMiddleware(hookManager, cfg.AuthToken)
	transformMiddleware := middleware.NewTransformMiddleware(hookManager)
	errorHandler := middleware.NewErrorMiddleware(hookManager)

	routerInstance := router.NewRouter(cfg.Routes, cfg.BackendURL)
	dslTransformer := transform.NewDSLTransformer()

	gateway := handler.NewGateway(hookManager, forwarder, auth, transformMiddleware, errorHandler, routerInstance, dslTransformer)

	if err := http.ListenAndServe(cfg.Port, gateway); err != nil {
		log.Fatal(err)
	}
}
