# Concurrency Safety

Gateway uses MySQL database for configuration storage, ensuring concurrency safety through GORM.

## Design Principles

| Feature | Description |
|---------|-------------|
| **Database Storage** | Configurations stored in MySQL with transaction consistency |
| **On-Demand Loading** | Each request loads latest configuration from database |
| **Stateless** | Gateway does not cache configurations, supports horizontal scaling |
| **VM Pool** | JavaScript executors use object pooling to avoid frequent creation |

## JavaScript VM Pool

```go
// Initialize VM pool
hook.InitVMPool(poolSize)

// Get VM from pool and execute script
vm := pool.Get()
defer pool.Put(vm)
```

## Common Script Library

Common scripts are loaded into memory at startup. After updates, call the reload endpoint:

```bash
POST /admin/db/reload-library
```

## Horizontal Scaling

Since the gateway is stateless and loads configurations from the database on each request, you can run multiple gateway instances behind a load balancer for high availability and scalability.
