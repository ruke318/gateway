package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Port       string
	BackendURL string
	AdminToken string

	// 数据库配置
	DB DBConfig

	// 日志配置
	Log LogConfig

	// JS VM 池配置
	VMPool VMPoolConfig
}

type DBConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	Database     string
	MaxIdleConns int
	MaxOpenConns int
	Debug        bool
}

type LogConfig struct {
	Path  string // 日志文件路径，为空则输出到控制台
	Level string // 日志级别: debug, info, warn, error
}

type VMPoolConfig struct {
	Size int // VM 池大小
}

func Load() *Config {
	viper.SetDefault("port", ":8080")
	viper.SetDefault("backendURL", "http://localhost:9090")
	viper.SetDefault("adminToken", "admin-secret-token")

	// 数据库默认配置
	viper.SetDefault("db.host", "localhost")
	viper.SetDefault("db.port", 3306)
	viper.SetDefault("db.user", "root")
	viper.SetDefault("db.password", "")
	viper.SetDefault("db.database", "gateway")
	viper.SetDefault("db.maxIdleConns", 10)
	viper.SetDefault("db.maxOpenConns", 100)
	viper.SetDefault("db.debug", false)

	// 日志默认配置
	viper.SetDefault("log.path", "")    // 为空输出到控制台
	viper.SetDefault("log.level", "info")

	// VM 池默认配置
	viper.SetDefault("vmPool.size", 100)

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
	cfg.AdminToken = viper.GetString("adminToken")

	// 数据库配置
	cfg.DB.Host = viper.GetString("db.host")
	cfg.DB.Port = viper.GetInt("db.port")
	cfg.DB.User = viper.GetString("db.user")
	cfg.DB.Password = viper.GetString("db.password")
	cfg.DB.Database = viper.GetString("db.database")
	cfg.DB.MaxIdleConns = viper.GetInt("db.maxIdleConns")
	cfg.DB.MaxOpenConns = viper.GetInt("db.maxOpenConns")
	cfg.DB.Debug = viper.GetBool("db.debug")

	// 日志配置
	cfg.Log.Path = viper.GetString("log.path")
	cfg.Log.Level = viper.GetString("log.level")

	// VM 池配置
	cfg.VMPool.Size = viper.GetInt("vmPool.size")

	return &cfg
}
