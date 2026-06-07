package logger

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	log  *zap.Logger
	once sync.Once
)

// InitLogger 初始化日志器（全局调用一次即可）
func InitLogger() {
	InitLoggerWithConfig(nil)
}

// InitLoggerWithConfig 使用自定义配置初始化日志器
func InitLoggerWithConfig(cfg *Config) {
	once.Do(func() {
		if cfg == nil {
			log = getDefaultLogger()
			log.Warn("日志配置为空，使用默认日志")
		} else {
			var err error
			log, err = cfg.Build()
			if err != nil {
				log = getDefaultLogger()
				log.Warn("构建日志失败，使用默认日志", zap.Error(err))
			}
		}
	})
}

// GetLogger 获取全局日志器实例
func GetLogger() *zap.Logger {
	if log == nil {
		return getDefaultLogger()
	}
	return log
}

// Sync 同步日志缓冲到磁盘
func Sync() error {
	if log != nil {
		return log.Sync()
	}
	return nil
}

// getDefaultLogger 获取默认配置的生产环境日志器
func getDefaultLogger() *zap.Logger {
	// 创建自定义编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 使用 JSON 编码器
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	// 设置日志级别
	level := zap.NewAtomicLevelAt(zap.InfoLevel)

	// 创建核心
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		level,
	)

	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
}

// Debug 记录调试级别日志
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

// Info 记录信息级别日志
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Warn 记录警告级别日志
func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

// Error 记录错误级别日志
func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

// DPanic 记录 panic 级别日志（开发环境会 panic）
func DPanic(msg string, fields ...zap.Field) {
	GetLogger().DPanic(msg, fields...)
}

// Panic 记录 panic 级别日志（总是 panic）
func Panic(msg string, fields ...zap.Field) {
	GetLogger().Panic(msg, fields...)
}

// Fatal 记录 fatal 级别日志（然后退出程序）
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

// With 添加字段到 logger
func With(fields ...zap.Field) *zap.Logger {
	return GetLogger().With(fields...)
}
