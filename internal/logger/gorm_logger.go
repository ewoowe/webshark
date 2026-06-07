package logger

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"go.uber.org/zap"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

// GormLogger GORM 日志适配器
type GormLogger struct {
	zapLogger            *zap.Logger
	slowThreshold        time.Duration
	ignoreRecordNotFound bool
	colorful             bool
	logLevel             gormlogger.LogLevel // 当前日志级别
}

// NewGormLogger 创建 GORM 日志适配器
func NewGormLogger(zapLogger *zap.Logger, config gormlogger.Config) *GormLogger {
	if zapLogger == nil {
		zapLogger = GetLogger()
	}

	return &GormLogger{
		zapLogger:            zapLogger,
		slowThreshold:        config.SlowThreshold,
		ignoreRecordNotFound: config.IgnoreRecordNotFoundError,
		colorful:             config.Colorful,
		logLevel:             config.LogLevel,
	}
}

// LogMode 返回具有指定日志级别的 GORM 日志实例
func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.logLevel = level
	newLogger.zapLogger = l.zapLogger.With(zap.String("gorm_level", fmt.Sprintf("%v", level)))
	return &newLogger
}

// Info 记录信息级别日志
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.log(l.zapLogger.Info, msg, data...)
}

// Warn 记录警告级别日志
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.log(l.zapLogger.Warn, msg, data...)
}

// Error 记录错误级别日志
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.log(l.zapLogger.Error, msg, data...)
}

// Trace 追踪 SQL 执行
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)

	// 检查是否为慢查询
	slow := elapsed > l.slowThreshold

	sql, rows := fc()

	var fields []zap.Field
	fields = append(fields,
		zap.Duration("duration", elapsed),
		zap.Int64("rows", rows),
		zap.String("caller", utils.FileWithLineNum()),
	)

	if err != nil && !(l.ignoreRecordNotFound && errors.Is(err, gormlogger.ErrRecordNotFound)) {
		fields = append(fields, zap.Error(err))
	}

	if slow {
		// 慢查询 - 记录详细调用栈和 SQL 语句（仅慢查询时）
		// SQL 语句只打印前 100 个字符
		sqlPreview := sql
		if len(sql) > 100 {
			sqlPreview = sql[:100] + "..."
		}
		fields = append(fields,
			zap.String("sql", sqlPreview),
			zap.String("stack", captureStack(3)),
		)
		l.zapLogger.Warn("SLOW SQL", fields...)
	} else if err != nil && !(l.ignoreRecordNotFound && errors.Is(err, gormlogger.ErrRecordNotFound)) {
		// 错误
		l.zapLogger.Error("SQL Error", fields...)
	} else if l.logLevel == gormlogger.Info {
		// 普通 SQL 执行 - 仅在日志级别严格等于 Info 时输出（即开启了 SQL 日志）
		// SQL 语句只打印前 100 个字符
		sqlPreview := sql
		if len(sql) > 100 {
			sqlPreview = sql[:100] + "..."
		}
		fields = append(fields, zap.String("sql", sqlPreview))
		l.zapLogger.Info("SQL", fields...)
	}
}

// log 通用的日志记录方法
func (l *GormLogger) log(logFunc func(string, ...zap.Field), msg string, data ...interface{}) {
	if len(data) == 0 {
		logFunc(msg)
		return
	}

	// 格式化数据
	formatted := fmt.Sprintf(msg, data...)
	logFunc(formatted)
}

// captureStack 捕获调用栈信息
func captureStack(skip int) string {
	var buf strings.Builder

	for i := skip; i < 10; i++ { // 最多捕获 10 层调用栈
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)
		var fnName string
		if fn != nil {
			fnName = fn.Name()
		} else {
			fnName = "unknown"
		}

		buf.WriteString(fmt.Sprintf("\n\t%s:%d in %s", file, line, fnName))
	}

	return buf.String()
}
