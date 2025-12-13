package router

import (
	"github.com/ruke318/gateway/handler"
	"github.com/savsgio/atreugo/v11"
)

// RegisterOutRoutes 注册对外业务路由
// 包括：
// - /gateway/v1/invoke - 主动调用厂商接口
// - /gateway/v1/notify/{service_id}/{channel} - 接收厂商回调
func RegisterOutRoutes(app *atreugo.Atreugo, h *handler.InvokeHandler) {
	// 主动调用路由
	app.POST("/gateway/v1/invoke", h.Invoke)

	// 回调路由（支持 POST 和 GET）
	// URL 格式：/gateway/v1/notify/{service_id}/{channel}
	// 示例：
	//   POST /gateway/v1/notify/payNotify/1       - 数字渠道，走默认逻辑
	//   POST /gateway/v1/notify/payNotify/alipay  - 支付宝回调
	//   GET  /gateway/v1/notify/refundNotify/wechat - 微信回调（GET）
	app.POST("/gateway/v1/notify/{service_id}/{channel}", h.HandleNotify)
	app.GET("/gateway/v1/notify/{service_id}/{channel}", h.HandleNotify)
}
