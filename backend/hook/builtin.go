package hook

import (
	"fmt"
	"time"

	"github.com/dop251/goja"
	redisClient "github.com/ruke318/gateway/redis"
	"github.com/ruke318/gateway/utils"
)

// RegisterBuiltins 注册内置模块到 JS 运行时
func RegisterBuiltins(vm *goja.Runtime) {
	// crypto 模块 - 加密解密
	vm.Set("crypto", map[string]interface{}{
		"md5":           utils.MD5,
		"sha1":          utils.SHA1,
		"sha256":        utils.SHA256,
		"sha512":        utils.SHA512,
		"hmacMD5":       utils.HmacMD5,
		"hmacSHA1":      utils.HmacSHA1,
		"hmacSHA256":    utils.HmacSHA256,
		"aesEncrypt":    utils.AESEncrypt,
		"aesDecrypt":    utils.AESDecrypt,
		"aesCBCEncrypt": utils.AESCBCEncrypt,
		"aesCBCDecrypt": utils.AESCBCDecrypt,
		"aesECBEncrypt": utils.AESECBEncrypt,
		"aesECBDecrypt": utils.AESECBDecrypt,
		"desEncrypt":    utils.DESEncrypt,
		"desDecrypt":    utils.DESDecrypt,
		"des3Encrypt":   utils.DES3Encrypt,
		"des3Decrypt":   utils.DES3Decrypt,
		"rsaEncrypt":    utils.RSAEncrypt,
		"rsaDecrypt":    utils.RSADecrypt,
		"rsaSign":       utils.RSASign,
		"rsaVerify":     utils.RSAVerify,
		"randomBytes":   utils.RandomBytes,
	})

	// http 模块 - HTTP 请求
	vm.Set("http", map[string]interface{}{
		"get":      utils.HTTPGet,
		"post":     utils.HTTPPost,
		"postJSON": utils.HTTPPostJSON,
		"postForm": utils.HTTPPostForm,
		"request":  utils.HTTPRequest,
	})

	// encoding 模块 - 编解码
	vm.Set("encoding", map[string]interface{}{
		"base64Encode": utils.Base64Encode,
		"base64Decode": utils.Base64Decode,
		"hexEncode":    utils.HexEncode,
		"hexDecode":    utils.HexDecode,
		"jsonEncode":   utils.JSONEncode,
		"jsonDecode":   utils.JSONDecode,
		"xmlEncode":    utils.XMLEncode,
		"xmlDecode":    utils.XMLDecode,
		"urlEncode":    utils.URLEncode,
		"urlDecode":    utils.URLDecode,
	})

	// util 模块 - 工具函数
	vm.Set("util", map[string]interface{}{
		"uuid":       utils.UUID,
		"now":        utils.Now,
		"formatTime": utils.FormatTime,
		"parseTime":  utils.ParseTime,
		"sleep":      utils.Sleep,
	})

	// console 模块 - 控制台输出
	vm.Set("console", map[string]interface{}{
		"log": consoleLog,
	})

	// redis 模块 - Redis 操作（直接使用 redis.Client）
	if redisClient.IsEnabled() {
		vm.Set("redis", createRedisModule())
	}
}

// createRedisModule 创建 Redis 模块（直接包装 redis.Client 方法）
func createRedisModule() map[string]interface{} {
	return map[string]interface{}{
		"get": func(key string) interface{} {
			val, err := redisClient.Client.Get(redisClient.Ctx, key).Result()
			if err != nil {
				return nil
			}
			return val
		},
		"set": func(key string, value interface{}) bool {
			return redisClient.Client.Set(redisClient.Ctx, key, value, 0).Err() == nil
		},
		"setex": func(key string, value interface{}, seconds int) bool {
			return redisClient.Client.Set(redisClient.Ctx, key, value, time.Duration(seconds)*time.Second).Err() == nil
		},
		"del": func(key string) bool {
			return redisClient.Client.Del(redisClient.Ctx, key).Err() == nil
		},
		"exists": func(key string) bool {
			n, _ := redisClient.Client.Exists(redisClient.Ctx, key).Result()
			return n > 0
		},
		"expire": func(key string, seconds int) bool {
			return redisClient.Client.Expire(redisClient.Ctx, key, time.Duration(seconds)*time.Second).Err() == nil
		},
		"ttl": func(key string) int64 {
			duration, _ := redisClient.Client.TTL(redisClient.Ctx, key).Result()
			return int64(duration.Seconds())
		},
		"incr": func(key string) int64 {
			val, _ := redisClient.Client.Incr(redisClient.Ctx, key).Result()
			return val
		},
		"decr": func(key string) int64 {
			val, _ := redisClient.Client.Decr(redisClient.Ctx, key).Result()
			return val
		},
		"hget": func(key, field string) interface{} {
			val, err := redisClient.Client.HGet(redisClient.Ctx, key, field).Result()
			if err != nil {
				return nil
			}
			return val
		},
		"hset": func(key, field string, value interface{}) bool {
			return redisClient.Client.HSet(redisClient.Ctx, key, field, value).Err() == nil
		},
		"hgetall": func(key string) map[string]string {
			val, _ := redisClient.Client.HGetAll(redisClient.Ctx, key).Result()
			return val
		},
	}
}

// consoleLog 打印日志
func consoleLog(args ...interface{}) {
	fmt.Println(args...)
}
