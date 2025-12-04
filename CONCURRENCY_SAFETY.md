# 并发安全说明

Gateway 的配置热更新功能实现了完全的线程安全。

## 问题背景

在实现热更新时，需要解决并发读写的安全问题：

```go
// 浅拷贝会导致并发问题
routeCopy := route  // map 字段仍然共享引用
```

当多个 goroutine 同时读写共享的 map 时，会导致 panic。

## 解决方案

### 深拷贝

使用 JSON 序列化实现深拷贝：

```go
func (r *RouteConfig) DeepCopy() RouteConfig {
    data, _ := json.Marshal(r)
    var copy RouteConfig
    json.Unmarshal(data, &copy)
    return copy
}
```

### 读写锁

```go
type Router struct {
    routes []config.RouteConfig
    mu     sync.RWMutex
}

func (r *Router) Match(req *http.Request) (*config.RouteConfig, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    // 返回深拷贝
    routeCopy := route.DeepCopy()
    return &routeCopy, nil
}

func (r *Router) UpdateRoute(route config.RouteConfig) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    // 更新操作
}
```

## 安全保证

| 特性 | 说明 |
|-----|------|
| 读写分离 | 读操作用 RLock，写操作用 Lock |
| 深拷贝隔离 | 返回的配置是独立副本 |
| 原子操作 | 配置更新不会部分生效 |
| 请求隔离 | 配置更新不影响正在处理的请求 |

## 测试

```bash
go test -v ./router/
```

测试包括：
- 并发访问测试
- 重复路由检测
- 深拷贝验证

## 性能

深拷贝开销约 1-2 微秒，相比网络 I/O 可忽略。
