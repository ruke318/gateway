package middleware

import "github.com/ruke318/gateway/hook"

type ErrorMiddleware struct {
	hookManager *hook.Manager
}

func NewErrorMiddleware(hookManager *hook.Manager) *ErrorMiddleware {
	return &ErrorMiddleware{
		hookManager: hookManager,
	}
}

func (m *ErrorMiddleware) Handle(ctx *hook.HookContext) error {
	return m.hookManager.Execute(hook.OnError, ctx)
}
