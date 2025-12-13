package auth

import (
	"strings"

	"github.com/savsgio/atreugo/v11"
)

// AuthMiddleware JWT 认证中间件
func AuthMiddleware(ctx *atreugo.RequestCtx) error {
	// 从 Header 中获取 Token
	authHeader := string(ctx.Request.Header.Peek("Authorization"))
	if authHeader == "" {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    401,
			"message": "未登录或登录已过期",
		}, 401)
	}

	// 解析 Bearer Token
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    401,
			"message": "Token 格式错误",
		}, 401)
	}

	// 验证 Token
	claims, err := ParseToken(parts[1])
	if err != nil {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    401,
			"message": "Token 无效或已过期",
		}, 401)
	}

	// 将用户信息存储到上下文
	ctx.SetUserValue("user_id", claims.UserID)
	ctx.SetUserValue("username", claims.Username)
	ctx.SetUserValue("role", claims.Role)

	return ctx.Next()
}

// AdminMiddleware 管理员权限中间件
func AdminMiddleware(ctx *atreugo.RequestCtx) error {
	role := ctx.UserValue("role")
	if role != "admin" {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    403,
			"message": "权限不足",
		}, 403)
	}
	return ctx.Next()
}

// GetCurrentUser 从上下文获取当前用户信息
func GetCurrentUser(ctx *atreugo.RequestCtx) (userID uint64, username, role string) {
	if val := ctx.UserValue("user_id"); val != nil {
		userID = val.(uint64)
	}
	if val := ctx.UserValue("username"); val != nil {
		username = val.(string)
	}
	if val := ctx.UserValue("role"); val != nil {
		role = val.(string)
	}
	return
}
