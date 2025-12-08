package router

import (
	"github.com/ruke318/gateway/handler"
	"github.com/savsgio/atreugo/v11"
)

// RegisterOutRoutes 注册对外业务路由
func RegisterOutRoutes(app *atreugo.Atreugo, h *handler.InvokeHandler) {
	app.POST("/gateway/v1/invoke", h.Invoke)
}
