// Package middleware 提供中间件
package middleware

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/ruke318/gateway/auth"
	"github.com/ruke318/gateway/database"
	"github.com/ruke318/gateway/model"
	"github.com/savsgio/atreugo/v11"
)

// OperationLogMiddleware 操作日志记录中间件
// 记录所有管理后台的增删改操作，包括修改前后的数据对比
func OperationLogMiddleware(ctx *atreugo.RequestCtx) error {
	// 获取请求方法和路径
	method := string(ctx.Method())
	path := string(ctx.Path())

	// 只记录 POST/PUT/DELETE 操作
	if method != "POST" && method != "PUT" && method != "DELETE" {
		return ctx.Next()
	}

	// 排除登录等不需要记录的接口
	if strings.HasPrefix(path, "/api/auth/login") {
		return ctx.Next()
	}

	// 解析操作类型和资源
	operation, resource, resourceID := parseOperation(method, path)
	if operation == "" {
		return ctx.Next() // 无法识别的操作不记录
	}

	// 获取当前用户
	userID, username := getCurrentUser(ctx)
	if userID == 0 {
		return ctx.Next() // 没有用户信息，不记录
	}

	// 对于 UPDATE 和 DELETE 操作，在执行前获取旧数据
	var beforeData string
	if (operation == "update" || operation == "delete") && resourceID != "" {
		beforeData = getResourceData(resource, resourceID)
	}

	// 执行实际请求
	err := ctx.Next()

	// 只记录成功的请求
	if err != nil || ctx.Response.StatusCode() >= 400 {
		return err
	}

	// 对于 CREATE 和 UPDATE 操作，获取请求 body 作为新数据
	var afterData string
	if operation == "create" || operation == "update" {
		if len(ctx.PostBody()) > 0 {
			afterData = filterSensitiveFields(ctx.PostBody())
		}
	}

	// 获取客户端IP
	ip := getClientIP(ctx)

	// 异步记录日志（不阻塞请求响应）
	go func() {
		log := &model.OperationLog{
			UserID:     userID,
			Username:   username,
			Operation:  operation,
			Resource:   resource,
			ResourceID: resourceID,
			BeforeData: beforeData,
			AfterData:  afterData,
			IP:         ip,
			CreatedAt:  model.LocalTime(time.Now()),
		}
		database.DB.Create(log)
	}()

	return err
}

// getCurrentUser 获取当前用户信息
func getCurrentUser(ctx *atreugo.RequestCtx) (uint64, string) {
	// 先尝试从 JWT 获取
	userID, username, _ := auth.GetCurrentUser(ctx)
	if userID != 0 {
		return userID, username
	}

	// 检查是否是 admin token 认证
	adminToken := string(ctx.Request.Header.Peek("X-Admin-Token"))
	if adminToken != "" {
		return 1, "admin" // 默认管理员
	}

	return 0, ""
}

// getResourceData 根据资源类型和ID查询数据
func getResourceData(resource, resourceID string) string {
	var result map[string]interface{}

	switch resource {
	case "user":
		var user model.User
		if err := database.DB.First(&user, resourceID).Error; err == nil {
			result = structToMap(user)
		}
	case "vendor":
		var vendor model.Vendor
		if err := database.DB.First(&vendor, resourceID).Error; err == nil {
			result = structToMap(vendor)
		}
	case "organization":
		var org model.Organization
		if err := database.DB.First(&org, resourceID).Error; err == nil {
			result = structToMap(org)
		}
	case "service":
		var service model.Service
		if err := database.DB.First(&service, resourceID).Error; err == nil {
			result = structToMap(service)
		}
	case "script":
		var script model.ScriptLibrary
		if err := database.DB.First(&script, resourceID).Error; err == nil {
			result = structToMap(script)
		}
	case "hook_script":
		var hookScript model.HookScript
		if err := database.DB.First(&hookScript, resourceID).Error; err == nil {
			result = structToMap(hookScript)
		}
	case "service_hook":
		var serviceHook model.ServiceHook
		if err := database.DB.First(&serviceHook, resourceID).Error; err == nil {
			result = structToMap(serviceHook)
		}
	default:
		return ""
	}

	if result == nil {
		return ""
	}

	// 过滤敏感字段
	delete(result, "password")
	delete(result, "Password")

	if jsonData, err := json.Marshal(result); err == nil {
		return string(jsonData)
	}

	return ""
}

// structToMap 将结构体转换为 map
func structToMap(obj interface{}) map[string]interface{} {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil
	}

	return result
}

// filterSensitiveFields 过滤敏感字段
func filterSensitiveFields(body []byte) string {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return ""
	}

	// 移除密码等敏感字段
	delete(data, "password")
	delete(data, "old_password")
	delete(data, "new_password")

	if jsonData, err := json.Marshal(data); err == nil {
		return string(jsonData)
	}

	return ""
}

// parseOperation 解析操作类型和资源
// 例如：POST /api/users -> (create, user, "")
// 例如：PUT /api/users/123 -> (update, user, "123")
// 例如：DELETE /api/users/123 -> (delete, user, "123")
func parseOperation(method, path string) (operation, resource, resourceID string) {
	// 移除 /api/ 前缀
	path = strings.TrimPrefix(path, "/api/")
	path = strings.TrimPrefix(path, "/admin/db/")

	// 分割路径
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return
	}

	// 确定操作类型
	switch method {
	case "POST":
		operation = "create"
	case "PUT":
		operation = "update"
	case "DELETE":
		operation = "delete"
	}

	// 确定资源类型
	if len(parts) >= 1 {
		resource = normalizeResource(parts[0])
	}

	// 提取资源ID（如果存在）
	if len(parts) >= 2 {
		resourceID = parts[1]
	}

	return
}

// normalizeResource 规范化资源名称
func normalizeResource(s string) string {
	// 移除复数形式的 s，统一为单数
	if strings.HasSuffix(s, "s") && s != "service" && s != "services" {
		s = strings.TrimSuffix(s, "s")
	}

	// 特殊资源映射
	switch s {
	case "vendors", "vendor":
		return "vendor"
	case "organizations", "organization":
		return "organization"
	case "services", "service":
		return "service"
	case "scripts", "script":
		return "script"
	case "hook-scripts", "hook-script":
		return "hook_script"
	case "service-hooks", "service-hook":
		return "service_hook"
	case "users", "user":
		return "user"
	default:
		return s
	}
}

// getClientIP 获取客户端真实IP
func getClientIP(ctx *atreugo.RequestCtx) string {
	// 尝试从 X-Real-IP 获取
	if ip := string(ctx.Request.Header.Peek("X-Real-IP")); ip != "" {
		return ip
	}

	// 尝试从 X-Forwarded-For 获取
	if xff := string(ctx.Request.Header.Peek("X-Forwarded-For")); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 返回远程地址
	return ctx.RemoteIP().String()
}
