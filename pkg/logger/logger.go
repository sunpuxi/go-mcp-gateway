package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Log 全局 SugaredLogger，用于便捷调用
var Log *zap.SugaredLogger

// Config 日志配置
type Config struct {
	Level  string // debug / info / warn / error
	Format string // json / console
}

// Init 初始化全局日志器
// 应在 main 函数中最早调用
func Init(cfg Config) error {
	level := parseLevel(cfg.Level)
	encoderCfg := buildEncoderConfig(cfg.Format)

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(os.Stdout),
		level,
	)

	Log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()
	return nil
}

// Sync 刷新缓冲，应在 main 退出前调用
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}

// --- 便捷函数，保持与 log 包类似的调用风格 ---

// Info 信息日志
func Info(msg string, keysAndValues ...interface{}) {
	Log.Infow(msg, keysAndValues...)
}

// Error 错误日志
func Error(msg string, keysAndValues ...interface{}) {
	Log.Errorw(msg, keysAndValues...)
}

// Warn 警告日志
func Warn(msg string, keysAndValues ...interface{}) {
	Log.Warnw(msg, keysAndValues...)
}

// Debug 调试日志
func Debug(msg string, keysAndValues ...interface{}) {
	Log.Debugw(msg, keysAndValues...)
}

// Fatal 致命错误日志（会调用 os.Exit(1)）
func Fatal(msg string, keysAndValues ...interface{}) {
	Log.Fatalw(msg, keysAndValues...)
}

// --- 内部辅助 ---

func parseLevel(s string) zapcore.Level {
	switch s {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func buildEncoderConfig(format string) zapcore.EncoderConfig {
	cfg := zap.NewProductionEncoderConfig()
	cfg.TimeKey = "time"
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	if format == "console" {
		cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		cfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	}
	return cfg
}
