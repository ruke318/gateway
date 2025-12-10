# 并发安全说明

Gateway 使用数据库存储配置，通过 GORM 保证并发安全。

## 设计原则

| 特性 | 说明 |
|-----|------|
| 数据库存储 | 配置存储在 MySQL，由数据库保证事务一致性 |
| 按需加载 | 每次请求从数据库加载最新配置 |
| 无状态 | 网关本身不缓存配置，支持水平扩展 |
| VM 池 | JavaScript 执行器使用对象池，避免频繁创建 |

## JS VM 池

```go
// 初始化 VM 池
hook.InitVMPool(poolSize)

// 从池中获取 VM 执行脚本
vm := pool.Get()
defer pool.Put(vm)
```

## 公共函数库

公共函数库在启动时加载到内存，更新后需调用重载接口：

```bash
POST /admin/db/reload-library
```

## 测试

```bash
cd backend
go test ./...
```
