// Package handler 提供操作日志相关的 HTTP 处理器
package handler

import (
	"github.com/ruke318/gateway/database"
	"github.com/ruke318/gateway/model"
	"github.com/ruke318/gateway/util"
	"github.com/savsgio/atreugo/v11"
)

// OperationLogHandler 操作日志处理器
type OperationLogHandler struct{}

// NewOperationLogHandler 创建操作日志处理器
func NewOperationLogHandler() *OperationLogHandler {
	return &OperationLogHandler{}
}

// ListOperationLogs 获取操作日志列表
// GET /api/operation-logs?page=1&size=20&username=admin&operation=create&resource=user
func (h *OperationLogHandler) ListOperationLogs(ctx *atreugo.RequestCtx) error {
	page := util.GetInt(ctx, "page", 1)
	size := util.GetInt(ctx, "size", 20)
	username := util.GetString(ctx, "username")
	operation := util.GetString(ctx, "operation")
	resource := util.GetString(ctx, "resource")

	query := database.DB.Model(&model.OperationLog{})

	// 过滤条件
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if operation != "" {
		query = query.Where("operation = ?", operation)
	}
	if resource != "" {
		query = query.Where("resource = ?", resource)
	}

	// 统计总数
	var total int64
	query.Count(&total)

	// 分页查询
	var logs []model.OperationLog
	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Order("id DESC").Find(&logs).Error; err != nil {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    500,
			"message": "查询失败",
		}, 500)
	}

	return ctx.JSONResponse(map[string]interface{}{
		"code":    200,
		"message": "success",
		"data": map[string]interface{}{
			"list":  logs,
			"total": total,
			"page":  page,
			"size":  size,
		},
	}, 200)
}

// GetOperationLog 获取操作日志详情
// GET /api/operation-logs/:id
func (h *OperationLogHandler) GetOperationLog(ctx *atreugo.RequestCtx) error {
	id := util.GetUint64(ctx, "id")

	var log model.OperationLog
	if err := database.DB.First(&log, id).Error; err != nil {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    404,
			"message": "日志不存在",
		}, 404)
	}

	return ctx.JSONResponse(map[string]interface{}{
		"code":    200,
		"message": "success",
		"data":    log,
	}, 200)
}

// GetOperationStatistics 获取操作统计
// GET /api/operation-logs/statistics
func (h *OperationLogHandler) GetOperationStatistics(ctx *atreugo.RequestCtx) error {
	// 统计各操作类型的数量
	var operationStats []struct {
		Operation string `json:"operation"`
		Count     int64  `json:"count"`
	}
	database.DB.Model(&model.OperationLog{}).
		Select("operation, COUNT(*) as count").
		Group("operation").
		Find(&operationStats)

	// 统计各资源类型的数量
	var resourceStats []struct {
		Resource string `json:"resource"`
		Count    int64  `json:"count"`
	}
	database.DB.Model(&model.OperationLog{}).
		Select("resource, COUNT(*) as count").
		Group("resource").
		Find(&resourceStats)

	// 统计活跃用户（最近7天）
	var activeUsers []struct {
		Username string `json:"username"`
		Count    int64  `json:"count"`
	}
	database.DB.Model(&model.OperationLog{}).
		Select("username, COUNT(*) as count").
		Where("created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)").
		Group("username").
		Order("count DESC").
		Limit(10).
		Find(&activeUsers)

	return ctx.JSONResponse(map[string]interface{}{
		"code":    200,
		"message": "success",
		"data": map[string]interface{}{
			"operation_stats": operationStats,
			"resource_stats":  resourceStats,
			"active_users":    activeUsers,
		},
	}, 200)
}
