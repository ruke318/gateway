package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ruke318/gateway/config"
	"github.com/ruke318/gateway/hook"
	"github.com/ruke318/gateway/router"
)

// AdminHandler 提供管理接口
type AdminHandler struct {
	router      *router.Router
	hookManager *hook.Manager
	adminToken  string // 管理 API 的访问 Token
}

func NewAdminHandler(router *router.Router, hookManager *hook.Manager, adminToken string) *AdminHandler {
	return &AdminHandler{
		router:      router,
		hookManager: hookManager,
		adminToken:  adminToken,
	}
}

// ServeHTTP 处理管理请求
func (h *AdminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 验证管理 Token
	token := r.Header.Get("X-Admin-Token")
	if token != h.adminToken {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	switch r.URL.Path {
	// 路由管理
	case "/admin/routes":
		h.handleRoutes(w, r)
	case "/admin/routes/add":
		h.handleAddRoute(w, r)
	case "/admin/routes/update":
		h.handleUpdateRoute(w, r)
	case "/admin/routes/delete":
		h.handleDeleteRoute(w, r)

	// Hook 管理
	case "/admin/hooks/update":
		h.handleUpdateHook(w, r)
	case "/admin/hooks/clear":
		h.handleClearHook(w, r)

	default:
		http.Error(w, "not found", http.StatusNotFound)
	}
}

// 路由管理接口

type RouteRequest struct {
	Route config.RouteConfig `json:"route"`
}

type DeleteRouteRequest struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

func (h *AdminHandler) handleRoutes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	routes := h.router.GetAllRoutes()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    routes,
	})
}

func (h *AdminHandler) handleAddRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
		return
	}

	if err := h.router.AddRoute(req.Route); err != nil {
		http.Error(w, fmt.Sprintf("failed to add route: %v", err), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "route added successfully",
	})
}

func (h *AdminHandler) handleUpdateRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
		return
	}

	if err := h.router.UpdateRoute(req.Route); err != nil {
		http.Error(w, fmt.Sprintf("failed to update route: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "route updated successfully",
	})
}

func (h *AdminHandler) handleDeleteRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DeleteRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
		return
	}

	if err := h.router.DeleteRoute(req.Path, req.Method); err != nil {
		http.Error(w, fmt.Sprintf("failed to delete route: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "route deleted successfully",
	})
}

// Hook 管理接口

type UpdateHookRequest struct {
	HookPoint string `json:"hookPoint"` // "BeforeAuth", "AfterAuth", etc.
	Script    string `json:"script"`    // JavaScript 脚本内容
}

type ClearHookRequest struct {
	HookPoint string `json:"hookPoint"`
}

func (h *AdminHandler) handleUpdateHook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UpdateHookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// 将字符串转换为 HookPoint
	hookPoint, err := parseHookPoint(req.HookPoint)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid hook point: %v", err), http.StatusBadRequest)
		return
	}

	if err := h.hookManager.UpdateHook(hookPoint, req.Script); err != nil {
		http.Error(w, fmt.Sprintf("failed to update hook: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "hook updated successfully",
	})
}

func (h *AdminHandler) handleClearHook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ClearHookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request: %v", err), http.StatusBadRequest)
		return
	}

	hookPoint, err := parseHookPoint(req.HookPoint)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid hook point: %v", err), http.StatusBadRequest)
		return
	}

	h.hookManager.ClearHook(hookPoint)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "hook cleared successfully",
	})
}

// parseHookPoint 将字符串转换为 HookPoint
func parseHookPoint(s string) (hook.HookPoint, error) {
	switch s {
	case "BeforeAuth":
		return hook.BeforeAuth, nil
	case "AfterAuth":
		return hook.AfterAuth, nil
	case "BeforeRequestTransform":
		return hook.BeforeRequestTransform, nil
	case "AfterRequestTransform":
		return hook.AfterRequestTransform, nil
	case "BeforeForward":
		return hook.BeforeForward, nil
	case "AfterForward":
		return hook.AfterForward, nil
	case "BeforeResponseTransform":
		return hook.BeforeResponseTransform, nil
	case "AfterResponseTransform":
		return hook.AfterResponseTransform, nil
	case "OnError":
		return hook.OnError, nil
	default:
		return 0, fmt.Errorf("unknown hook point: %s", s)
	}
}
