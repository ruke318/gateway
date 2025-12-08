package util

import (
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/gorilla/schema"
	"github.com/savsgio/atreugo/v11"
)

var decoder = schema.NewDecoder()

func init() {
	decoder.IgnoreUnknownKeys(true)
}

// BindQuery 绑定 query 参数到结构体
func BindQuery(ctx *atreugo.RequestCtx, dest interface{}) error {
	values := make(url.Values)
	ctx.QueryArgs().VisitAll(func(key, value []byte) {
		values.Add(string(key), string(value))
	})
	return decoder.Decode(dest, values)
}

// BindJSON 绑定 JSON body 到结构体
func BindJSON(ctx *atreugo.RequestCtx, dest interface{}) error {
	return json.Unmarshal(ctx.PostBody(), dest)
}

// GetString 获取字符串参数，UserValue 优先级高于 QueryArgs
func GetString(ctx *atreugo.RequestCtx, key string, def ...string) string {
	// 优先从 UserValue 获取（路由参数）
	if v := ctx.UserValue(key); v != nil {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	// 其次从 QueryArgs 获取
	if v := ctx.QueryArgs().Peek(key); len(v) > 0 {
		return string(v)
	}
	// 返回默认值
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

// GetInt 获取整数参数，UserValue 优先级高于 QueryArgs
func GetInt(ctx *atreugo.RequestCtx, key string, def ...int) int {
	s := GetString(ctx, key)
	if s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			return v
		}
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}

// GetInt64 获取 int64 参数
func GetInt64(ctx *atreugo.RequestCtx, key string, def ...int64) int64 {
	s := GetString(ctx, key)
	if s != "" {
		if v, err := strconv.ParseInt(s, 10, 64); err == nil {
			return v
		}
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}

// GetUint64 获取 uint64 参数
func GetUint64(ctx *atreugo.RequestCtx, key string, def ...uint64) uint64 {
	s := GetString(ctx, key)
	if s != "" {
		if v, err := strconv.ParseUint(s, 10, 64); err == nil {
			return v
		}
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
