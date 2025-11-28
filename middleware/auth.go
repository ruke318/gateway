package middleware

import (
	"fmt"
	"net/http"

	"github.com/ruke318/gateway/hook"
)

type AuthMiddleware struct {
	hookManager *hook.Manager
	authToken   string
}

func NewAuthMiddleware(hookManager *hook.Manager, authToken string) *AuthMiddleware {
	return &AuthMiddleware{
		hookManager: hookManager,
		authToken:   authToken,
	}
}

func (m *AuthMiddleware) Handle(ctx *hook.HookContext) error {
	if err := m.hookManager.Execute(hook.BeforeAuth, ctx); err != nil {
		fmt.Println("BeforeAuth error", err)
		return err
	}

	token := ctx.Request.Header.Get("Authorization")
	if token != "Bearer "+m.authToken {
		ctx.Error = http.ErrAbortHandler
		return ctx.Error
	}

	if err := m.hookManager.Execute(hook.AfterAuth, ctx); err != nil {
		return err
	}

	return nil
}
