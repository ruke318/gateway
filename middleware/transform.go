package middleware

import "github.com/ruke318/gateway/hook"

type TransformMiddleware struct {
	hookManager *hook.Manager
}

func NewTransformMiddleware(hookManager *hook.Manager) *TransformMiddleware {
	return &TransformMiddleware{
		hookManager: hookManager,
	}
}

func (m *TransformMiddleware) TransformRequest(ctx *hook.HookContext) error {
	if err := m.hookManager.Execute(hook.BeforeRequestTransform, ctx); err != nil {
		return err
	}

	if err := m.hookManager.Execute(hook.AfterRequestTransform, ctx); err != nil {
		return err
	}

	return nil
}

func (m *TransformMiddleware) TransformResponse(ctx *hook.HookContext) error {
	if err := m.hookManager.Execute(hook.BeforeResponseTransform, ctx); err != nil {
		return err
	}

	if err := m.hookManager.Execute(hook.AfterResponseTransform, ctx); err != nil {
		return err
	}

	return nil
}
