package config

import (
	"encoding/json"
	"log"

	"github.com/spf13/viper"
)

type RouteConfig struct {
	Path              string                 `mapstructure:"path" json:"path"`
	Method            string                 `mapstructure:"method" json:"method"`
	BackendURL        string                 `mapstructure:"backendUrl" json:"backendUrl"`
	BackendPath       string                 `mapstructure:"backendPath" json:"backendPath"`
	BackendMethod     string                 `mapstructure:"backendMethod" json:"backendMethod"`
	RequestTransform  map[string]interface{} `mapstructure:"requestTransform" json:"requestTransform"`
	ResponseTransform map[string]interface{} `mapstructure:"responseTransform" json:"responseTransform"`
}

// DeepCopy 返回 RouteConfig 的深拷贝
// 使用 JSON 序列化/反序列化方式，确保 map 字段也被深拷贝
// 这样可以避免并发修改导致的 panic
func (r *RouteConfig) DeepCopy() RouteConfig {
	// 序列化为 JSON
	data, err := json.Marshal(r)
	if err != nil {
		// 如果序列化失败，返回原值（理论上不应该失败）
		log.Printf("Warning: failed to marshal RouteConfig: %v", err)
		return *r
	}

	// 反序列化为新对象
	var copy RouteConfig
	if err := json.Unmarshal(data, &copy); err != nil {
		// 如果反序列化失败，返回原值
		log.Printf("Warning: failed to unmarshal RouteConfig: %v", err)
		return *r
	}

	return copy
}

type Config struct {
	Port       string
	BackendURL string
	AuthToken  string
	Routes     []RouteConfig
}

func Load() *Config {
	viper.SetDefault("port", ":8080")
	viper.SetDefault("backendURL", "http://localhost:9090")
	viper.SetDefault("authToken", "default-token")

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: config file not found, using defaults: %v", err)
	}

	var cfg Config
	cfg.Port = viper.GetString("port")
	cfg.BackendURL = viper.GetString("backendURL")
	cfg.AuthToken = viper.GetString("authToken")

	if err := viper.UnmarshalKey("routes", &cfg.Routes); err != nil {
		log.Printf("Warning: failed to parse routes: %v", err)
	}

	return &cfg
}
