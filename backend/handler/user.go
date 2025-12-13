// Package handler 提供用户管理相关的 HTTP 处理器
package handler

import (
	"time"

	"github.com/ruke318/gateway/auth"
	"github.com/ruke318/gateway/database"
	"github.com/ruke318/gateway/model"
	"github.com/ruke318/gateway/util"
	"github.com/savsgio/atreugo/v11"
)

// UserHandler 用户管理处理器
type UserHandler struct{}

// NewUserHandler 创建用户管理处理器
func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

// Login 用户登录
// POST /api/auth/login
// Body: {"username": "admin", "password": "admin123"}
func (h *UserHandler) Login(ctx *atreugo.RequestCtx) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := util.BindJSON(ctx, &req); err != nil {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    400,
			"message": "请求参数错误",
		}, 400)
	}

	// 查询用户
	var user model.User
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    401,
			"message": "用户名或密码错误",
		}, 401)
	}

	// 检查状态
	if user.Status == 0 {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    403,
			"message": "账号已被禁用",
		}, 403)
	}

	// 验证密码
	if !user.CheckPassword(req.Password) {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    401,
			"message": "用户名或密码错误",
		}, 401)
	}

	// 生成 Token
	token, err := auth.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    500,
			"message": "生成 Token 失败",
		}, 500)
	}

	// 更新最后登录时间
	now := model.LocalTime(time.Now())
	database.DB.Model(&user).Update("last_login", &now)

	return ctx.JSONResponse(map[string]interface{}{
		"code":    200,
		"message": "登录成功",
		"data": map[string]interface{}{
			"token":    token,
			"user_id":  user.ID,
			"username": user.Username,
			"real_name": user.RealName,
			"role":     user.Role,
		},
	}, 200)
}

// GetCurrentUserInfo 获取当前登录用户信息
// GET /api/auth/me
func (h *UserHandler) GetCurrentUserInfo(ctx *atreugo.RequestCtx) error {
	userID, username, role := auth.GetCurrentUser(ctx)

	var user model.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    404,
			"message": "用户不存在",
		}, 404)
	}

	return ctx.JSONResponse(map[string]interface{}{
		"code":    200,
		"message": "success",
		"data": map[string]interface{}{
			"id":         user.ID,
			"username":   username,
			"real_name":  user.RealName,
			"role":       role,
			"status":     user.Status,
			"last_login": user.LastLogin,
			"created_at": user.CreatedAt,
		},
	}, 200)
}

// ListUsers 获取用户列表（管理员）
// GET /api/users?page=1&size=10
func (h *UserHandler) ListUsers(ctx *atreugo.RequestCtx) error {
	page := util.GetInt(ctx, "page", 1)
	size := util.GetInt(ctx, "size", 10)

	var users []model.User
	var total int64

	query := database.DB.Model(&model.User{})
	query.Count(&total)

	offset := (page - 1) * size
	if err := query.Offset(offset).Limit(size).Order("id DESC").Find(&users).Error; err != nil {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    500,
			"message": "查询失败",
		}, 500)
	}

	return ctx.JSONResponse(map[string]interface{}{
		"code":    200,
		"message": "success",
		"data": map[string]interface{}{
			"list":  users,
			"total": total,
			"page":  page,
			"size":  size,
		},
	}, 200)
}

// CreateUser 创建用户（管理员）
// POST /api/users
// Body: {"username": "user1", "password": "123456", "real_name": "张三", "role": "user"}
func (h *UserHandler) CreateUser(ctx *atreugo.RequestCtx) error {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		RealName string `json:"real_name"`
		Role     string `json:"role"`
	}

	if err := util.BindJSON(ctx, &req); err != nil {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    400,
			"message": "请求参数错误",
		}, 400)
	}

	// 检查用户名是否已存在
	var count int64
	database.DB.Model(&model.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    400,
			"message": "用户名已存在",
		}, 400)
	}

	// 创建用户
	user := &model.User{
		Username: req.Username,
		RealName: req.RealName,
		Role:     req.Role,
		Status:   1,
	}

	if err := user.SetPassword(req.Password); err != nil {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    500,
			"message": "密码加密失败",
		}, 500)
	}

	if err := database.DB.Create(user).Error; err != nil {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    500,
			"message": "创建用户失败",
		}, 500)
	}

	return ctx.JSONResponse(map[string]interface{}{
		"code":    200,
		"message": "创建成功",
		"data":    user,
	}, 200)
}

// UpdateUser 更新用户（管理员）
// PUT /api/users/:id
func (h *UserHandler) UpdateUser(ctx *atreugo.RequestCtx) error {
	id := util.GetUint64(ctx, "id")

	var req struct {
		RealName string `json:"real_name"`
		Role     string `json:"role"`
		Status   int    `json:"status"`
	}

	if err := util.BindJSON(ctx, &req); err != nil {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    400,
			"message": "请求参数错误",
		}, 400)
	}

	var user model.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    404,
			"message": "用户不存在",
		}, 404)
	}

	// 更新字段
	updates := map[string]interface{}{
		"real_name": req.RealName,
		"role":      req.Role,
		"status":    req.Status,
	}

	if err := database.DB.Model(&user).Updates(updates).Error; err != nil {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    500,
			"message": "更新失败",
		}, 500)
	}

	return ctx.JSONResponse(map[string]interface{}{
		"code":    200,
		"message": "更新成功",
	}, 200)
}

// DeleteUser 删除用户（管理员）
// DELETE /api/users/:id
func (h *UserHandler) DeleteUser(ctx *atreugo.RequestCtx) error {
	id := util.GetUint64(ctx, "id")

	// 不能删除自己
	currentUserID, _, _ := auth.GetCurrentUser(ctx)
	if currentUserID == id {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    400,
			"message": "不能删除当前登录用户",
		}, 400)
	}

	if err := database.DB.Delete(&model.User{}, id).Error; err != nil {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    500,
			"message": "删除失败",
		}, 500)
	}

	return ctx.JSONResponse(map[string]interface{}{
		"code":    200,
		"message": "删除成功",
	}, 200)
}

// ChangePassword 修改密码
// POST /api/auth/change-password
// Body: {"old_password": "old", "new_password": "new"}
func (h *UserHandler) ChangePassword(ctx *atreugo.RequestCtx) error {
	userID, _, _ := auth.GetCurrentUser(ctx)

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := util.BindJSON(ctx, &req); err != nil {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    400,
			"message": "请求参数错误",
		}, 400)
	}

	var user model.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    404,
			"message": "用户不存在",
		}, 404)
	}

	// 验证旧密码
	if !user.CheckPassword(req.OldPassword) {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    400,
			"message": "原密码错误",
		}, 400)
	}

	// 设置新密码
	if err := user.SetPassword(req.NewPassword); err != nil {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    500,
			"message": "密码加密失败",
		}, 500)
	}

	if err := database.DB.Model(&user).Update("password", user.Password).Error; err != nil {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    500,
			"message": "修改密码失败",
		}, 500)
	}

	return ctx.JSONResponse(map[string]interface{}{
		"code":    200,
		"message": "修改密码成功",
	}, 200)
}

// ResetPassword 重置用户密码（管理员）
// POST /api/users/:id/reset-password
// Body: {"new_password": "123456"}
func (h *UserHandler) ResetPassword(ctx *atreugo.RequestCtx) error {
	id := util.GetUint64(ctx, "id")

	var req struct {
		NewPassword string `json:"new_password"`
	}

	if err := util.BindJSON(ctx, &req); err != nil {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    400,
			"message": "请求参数错误",
		}, 400)
	}

	var user model.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    404,
			"message": "用户不存在",
		}, 404)
	}

	// 设置新密码
	if err := user.SetPassword(req.NewPassword); err != nil {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    500,
			"message": "密码加密失败",
		}, 500)
	}

	if err := database.DB.Model(&user).Update("password", user.Password).Error; err != nil {
		return ctx.JSONResponse(map[string]interface{}{
			"code":    500,
			"message": "重置密码失败",
		}, 500)
	}

	return ctx.JSONResponse(map[string]interface{}{
		"code":    200,
		"message": "重置密码成功",
	}, 200)
}
