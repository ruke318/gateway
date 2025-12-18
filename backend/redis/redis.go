package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	// Client 全局 Redis 客户端（供 Go 代码直接使用）
	Client *redis.Client
	Ctx    = context.Background()
)

// Init 初始化 Redis 连接
func Init(addr, password string, db, poolSize int) error {
	if addr == "" {
		return nil
	}

	Client = redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     poolSize,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	return Client.Ping(Ctx).Err()
}

// IsEnabled 检查是否已启用
func IsEnabled() bool {
	return Client != nil
}
