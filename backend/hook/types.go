package hook

import "net/http"

// HookContext Hook 执行上下文
type HookContext struct {
	LogID           string // 日志追踪ID
	Request         *http.Request
	Response        *http.Response
	RequestBody     []byte
	ResponseBody    []byte
	RequestHeaders  map[string]string
	ResponseHeaders map[string]string
	Error           error
	Data            map[string]interface{}
}

// Hook 接口
type Hook interface {
	Execute(ctx *HookContext) error
}
