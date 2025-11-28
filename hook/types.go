package hook

import "net/http"

type HookPoint int

const (
	BeforeAuth HookPoint = iota
	AfterAuth
	BeforeRequestTransform
	AfterRequestTransform
	BeforeForward
	AfterForward
	BeforeResponseTransform
	AfterResponseTransform
	OnError
)

type HookContext struct {
	Request        *http.Request
	Response       *http.Response
	RequestBody    []byte
	ResponseBody   []byte
	RequestHeaders map[string]string
	ResponseHeaders map[string]string
	Error          error
	Data           map[string]interface{}
}

type Hook interface {
	Execute(ctx *HookContext) error
}
