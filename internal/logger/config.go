package logger

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LogRotateConfig 日志切割配置
type LogRotateConfig struct {
	Filename   string `json:"filename"`    // 日志文件路径
	MaxSize    int    `json:"max_size"`    // 单个文件最大大小 (MB)
	MaxBackups int    `json:"max_backups"` // 保留的旧日志文件数量
	MaxAge     int    `json:"max_age"`     // 日志文件保留天数
	Compress   bool   `json:"compress"`    // 是否压缩旧日志
}

// Config 日志配置
type Config struct {
	Level        string           // 日志级别：debug, info, warn, error
	Format       string           // 输出格式：json, console
	OutputPaths  []string         // 多个输出路径：stdout, stderr, 或文件路径（用于同时输出到多个地方）
	RotateConfig *LogRotateConfig // 日志切割配置（仅当输出到文件时有效）
}

// NewConfig 创建默认配置
func NewConfig() *Config {
	return &Config{
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{"stdout"},
		RotateConfig: &LogRotateConfig{
			Filename:   "logs/webshark.log",
			MaxSize:    10,    // 10MB
			MaxBackups: 5,     // 保留 5 个旧文件
			MaxAge:     30,    // 保留 30 天
			Compress:   false, // 不压缩
		},
	}
}

// Build 根据配置构建日志器
func (c *Config) Build() (*zap.Logger, error) {
	// 解析日志级别
	level := zap.NewAtomicLevel()
	switch c.Level {
	case "debug":
		level.SetLevel(zap.DebugLevel)
	case "info":
		level.SetLevel(zap.InfoLevel)
	case "warn":
		level.SetLevel(zap.WarnLevel)
	case "error":
		level.SetLevel(zap.ErrorLevel)
	default:
		level.SetLevel(zap.InfoLevel)
	}

	// 创建编码器配置
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.CallerKey = "caller"
	encoderConfig.MessageKey = "message"
	encoderConfig.LevelKey = "level"
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	// 使用自定义的短路径编码器
	encoderConfig.EncodeCaller = shortenCaller
	// 禁用 HTML 转义
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder

	// 选择编码器
	var encoder zapcore.Encoder
	if c.Format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 确定输出路径列表
	outputPaths := c.OutputPaths
	if len(outputPaths) == 0 {
		outputPaths = []string{"stdout"}
	}

	// 创建多个输出核心
	var cores []zapcore.Core
	for _, path := range outputPaths {
		core := zapcore.NewCore(
			encoder,
			c.getWriteSyncer(path),
			level,
		)
		cores = append(cores, core)
	}

	// 合并多个核心
	combinedCore := zapcore.NewTee(cores...)

	// 构建日志器
	logger := zap.New(combinedCore, zap.AddCaller(), zap.AddCallerSkip(1))
	return logger, nil
}

// getWriteSyncer 获取输出写入器
func (c *Config) getWriteSyncer(path string) zapcore.WriteSyncer {
	switch path {
	case "stdout":
		return zapcore.AddSync(os.Stdout)
	case "stderr":
		return zapcore.AddSync(os.Stderr)
	default:
		// 文件输出，使用 lumberjack 实现自动切割
		if c.RotateConfig == nil {
			// 如果没有配置切割，使用默认配置
			c.RotateConfig = &LogRotateConfig{
				Filename:   path,
				MaxSize:    10,
				MaxBackups: 5,
				MaxAge:     30,
				Compress:   false,
			}
		} else {
			// 如果配置了路径，使用配置的路径
			if c.RotateConfig.Filename == "" {
				c.RotateConfig.Filename = path
			}
		}

		lumberjackLogger := &lumberjack.Logger{
			Filename:   c.RotateConfig.Filename,
			MaxSize:    c.RotateConfig.MaxSize,
			MaxBackups: c.RotateConfig.MaxBackups,
			MaxAge:     c.RotateConfig.MaxAge,
			Compress:   c.RotateConfig.Compress,
		}

		return zapcore.AddSync(lumberjackLogger)
	}
}

// getProjectName 从 go.mod 文件中获取项目名称
func getProjectName() string {
	// 尝试多种方法获取项目名
	var projectName string

	// 方法 1: 尝试在当前工作目录查找 go.mod
	if wd, err := os.Getwd(); err == nil {
		projectName = findGoModInDir(wd)
	}

	// 方法 2: 如果方法 1 失败，尝试在可执行文件所在目录查找
	if projectName == "" {
		if exePath, err := os.Executable(); err == nil {
			exeDir := filepath.Dir(exePath)
			projectName = findGoModInDir(exeDir)
		}
	}

	// 方法 3: 使用默认值
	if projectName == "" {
		projectName = "webshark" // 降级处理：如果无法获取项目名，使用默认值
	}

	return projectName
}

// findGoModInDir 在指定目录及其父目录中查找 go.mod 并提取项目名
func findGoModInDir(startDir string) string {
	// 最多向上查找 5 层
	currentDir := startDir
	for i := 0; i < 5; i++ {
		goModPath := filepath.Join(currentDir, "go.mod")

		// 检查 go.mod 是否存在
		if _, err := os.Stat(goModPath); err == nil {
			// 找到 go.mod，解析项目名
			file, err := os.Open(goModPath)
			if err != nil {
				continue
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				// 查找 module 开头的行
				if strings.HasPrefix(line, "module ") {
					// 提取模块名（去掉 "module " 前缀）
					moduleName := strings.TrimSpace(strings.TrimPrefix(line, "module "))
					// 只取最后一段作为项目名（例如：github.com/user/scc -> scc）
					parts := strings.Split(moduleName, "/")
					return parts[len(parts)-1]
				}
			}
		}

		// 向上一级目录继续查找
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// 已经到达根目录
			break
		}
		currentDir = parentDir
	}

	return ""
}

// shortenCaller 自定义的短路径编码器，保持路径一致性
func shortenCaller(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	// 获取完整路径
	fullPath := caller.FullPath()

	// 动态获取项目名称
	projectName := getProjectName()
	if projectName == "" {
		projectName = "webshark" // 降级处理：如果无法获取项目名，使用默认值
	}

	// 策略：找到最后一个 "/{projectName}/" 目录，保留从该目录开始的路径
	// 这样可以确保所有路径都以 {projectName}/ 开头
	projectMarker := string(filepath.Separator) + projectName + string(filepath.Separator)
	if idx := strings.LastIndex(fullPath, projectMarker); idx != -1 {
		// 找到了项目目录，保留从项目名开始的路径
		fullPath = fullPath[idx+1:] // +1 是为了包含项目名前面的斜杠
	} else {
		// 没找到项目目录，尝试其他方法
		// 如果是 GOPATH 模式，去掉 GOPATH/src 前缀
		if gopath := os.Getenv("GOPATH"); gopath != "" {
			prefix := filepath.Join(gopath, "src") + string(filepath.Separator)
			if strings.HasPrefix(fullPath, prefix) {
				fullPath = strings.TrimPrefix(fullPath, prefix)
			}
		}
	}

	// 确保路径格式一致（使用正斜杠）
	fullPath = filepath.ToSlash(fullPath)

	enc.AppendString(fullPath)
}
