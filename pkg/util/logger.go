package util

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger

// Config 日志配置
type LogConfig struct {
	Level      string
	Filename   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
}

// InitLogger 初始化日志
func InitLogger(config *LogConfig) error {
	// 设置日志级别
	var level zapcore.Level
	switch config.Level {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	default:
		level = zap.InfoLevel
	}

	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 文件写入器
	fileWriter := &lumberjack.Logger{
		Filename:   config.Filename,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   true,
	}

	// 控制台写入器
	consoleWriter := zapcore.Lock(os.Stdout)

	// 创建核心
	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.AddSync(fileWriter), level),
		zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), zapcore.AddSync(consoleWriter), level),
	)

	// 创建logger
	Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return nil
}

// Info 信息日志
func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

// Debug 调试日志
func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

// Warn 警告日志
func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

// Error 错误日志
func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

// Fatal 致命错误日志
func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}

// Sync 同步日志
func Sync() {
	_ = Logger.Sync()
}

// WithField 添加字段
func WithField(key string, value interface{}) *zap.Logger {
	return Logger.With(zap.Any(key, value))
}

// WithFields 添加多个字段
func WithFields(fields map[string]interface{}) *zap.Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return Logger.With(zapFields...)
}

// WithError 添加错误字段
func WithError(err error) *zap.Logger {
	return Logger.With(zap.Error(err))
}

// WithDuration 添加耗时字段
func WithDuration(d time.Duration) *zap.Logger {
	return Logger.With(zap.Duration("duration", d))
}
