package config

import (
	"log"

	"github.com/spf13/viper"
)

type RouteConfig struct {
	Path              string                 `mapstructure:"path"`
	Method            string                 `mapstructure:"method"`
	BackendURL        string                 `mapstructure:"backendUrl"`
	BackendPath       string                 `mapstructure:"backendPath"`
	BackendMethod     string                 `mapstructure:"backendMethod"`
	RequestTransform  map[string]interface{} `mapstructure:"requestTransform"`
	ResponseTransform map[string]interface{} `mapstructure:"responseTransform"`
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
