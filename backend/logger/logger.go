package logger

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

// Init 初始化日志，支持配置文件路径和日志级别
func Init(path, level string) error {
	// 解析日志级别
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// 编码配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var core zapcore.Core

	if path == "" {
		// 输出到控制台
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			zapLevel,
		)
	} else {
		// 输出到文件
		file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(file),
			zapLevel,
		)
	}

	log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return nil
}

func init() {
	// 默认初始化，输出到控制台
	Init("", "info")
}

// GenerateLogID 生成 LogID: bizNo_随机串
func GenerateLogID(bizNo string) string {
	b := make([]byte, 8)
	rand.Read(b)
	randStr := hex.EncodeToString(b)
	if bizNo == "" {
		return randStr
	}
	return fmt.Sprintf("%s_%s", bizNo, randStr)
}

// Info 记录信息日志
func Info(logID, stage, msg string, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("logID", logID),
		zap.String("stage", stage),
	}, fields...)
	log.Info(msg, allFields...)
}

// Error 记录错误日志
func Error(logID, stage, msg string, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("logID", logID),
		zap.String("stage", stage),
	}, fields...)
	log.Error(msg, allFields...)
}

// Debug 记录调试日志
func Debug(logID, stage, msg string, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("logID", logID),
		zap.String("stage", stage),
	}, fields...)
	log.Debug(msg, allFields...)
}

// Warn 记录警告日志
func Warn(logID, stage, msg string, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("logID", logID),
		zap.String("stage", stage),
	}, fields...)
	log.Warn(msg, allFields...)
}

// Sync 刷新日志缓冲
func Sync() {
	log.Sync()
}
