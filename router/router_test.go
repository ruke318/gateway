package router

import (
	"net/http"
	"sync"
	"testing"

	"github.com/ruke318/gateway/config"
)

// TestConcurrentAccess 测试并发访问的安全性
func TestConcurrentAccess(t *testing.T) {
	// 创建路由器
	routes := []config.RouteConfig{
		{
			Path:       "/api/test",
			Method:     "GET",
			BackendURL: "http://localhost:9090",
			ResponseTransform: map[string]interface{}{
				"code": "200",
				"data": "$.result",
			},
		},
	}

	router := NewRouter(routes, "http://localhost:9090")

	// 并发测试：同时进行路由匹配和路由更新
	var wg sync.WaitGroup
	errChan := make(chan error, 100)

	// 启动 50 个 goroutine 进行路由匹配
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				req, _ := http.NewRequest("GET", "/api/test", nil)
				route, err := router.Match(req)
				if err != nil {
					errChan <- err
					return
				}
				// 读取 ResponseTransform map（如果是浅拷贝，这里会并发读写 panic）
				_ = route.ResponseTransform["code"]
				_ = route.ResponseTransform["data"]
			}
		}()
	}

	// 同时启动 10 个 goroutine 进行路由更新
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				newRoute := config.RouteConfig{
					Path:       "/api/test",
					Method:     "GET",
					BackendURL: "http://localhost:9090",
					ResponseTransform: map[string]interface{}{
						"code":    "200",
						"message": "updated",
						"index":   index,
					},
				}
				// 更新路由（如果是浅拷贝，这里会导致 concurrent map write）
				_ = router.UpdateRoute(newRoute)
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	// 检查是否有错误
	for err := range errChan {
		t.Errorf("Concurrent access error: %v", err)
	}
}

// TestAddRouteDuplication 测试重复路由检测
func TestAddRouteDuplication(t *testing.T) {
	router := NewRouter([]config.RouteConfig{}, "http://localhost:9090")

	route1 := config.RouteConfig{
		Path:       "/api/users",
		Method:     "POST",
		BackendURL: "http://localhost:9090",
	}

	// 第一次添加应该成功
	err := router.AddRoute(route1)
	if err != nil {
		t.Errorf("First AddRoute should succeed, got error: %v", err)
	}

	// 第二次添加相同路由应该失败
	err = router.AddRoute(route1)
	if err == nil {
		t.Error("Second AddRoute should fail with duplicate error")
	}

	// 验证只有一个路由
	routes := router.GetAllRoutes()
	if len(routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(routes))
	}
}

// TestDeepCopy 测试深拷贝功能
func TestDeepCopy(t *testing.T) {
	routes := []config.RouteConfig{
		{
			Path:       "/api/test",
			Method:     "GET",
			BackendURL: "http://localhost:9090",
			ResponseTransform: map[string]interface{}{
				"code": "200",
				"data": "$.result",
			},
		},
	}

	router := NewRouter(routes, "http://localhost:9090")

	// 获取路由配置
	req, _ := http.NewRequest("GET", "/api/test", nil)
	route1, _ := router.Match(req)
	route2, _ := router.Match(req)

	// 修改第一个返回的配置
	route1.ResponseTransform["code"] = "500"
	route1.ResponseTransform["new_field"] = "added"

	// 验证第二个返回的配置没有被修改
	if route2.ResponseTransform["code"] != "200" {
		t.Error("Deep copy failed: modifications affected other copies")
	}

	if _, exists := route2.ResponseTransform["new_field"]; exists {
		t.Error("Deep copy failed: new field appeared in other copies")
	}
}
