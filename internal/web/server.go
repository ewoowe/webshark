package web

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
	"webshark/internal/config"
	"webshark/internal/logger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Server struct {
	engine *gin.Engine
	config *config.Config
}

// GetEngine 获取 Gin 引擎（用于注册额外路由）
func (s *Server) GetEngine() *gin.Engine {
	return s.engine
}

// NewServer 创建新的 Web 服务器
func NewServer(conf *config.Config) *Server {
	// 设置 Gin 模式
	if conf.Log.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	// 使用中间件
	engine.Use(gin.Recovery())     // 恢复 panic
	engine.Use(LoggerMiddleware()) // 自定义日志中间件
	engine.Use(CORSMiddleware())   // CORS 中间件

	return &Server{
		engine: engine,
		config: conf,
	}
}

// SetupRoutes 配置路由
func (s *Server) SetupRoutes() {
	// 健康检查
	s.engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		})
	})

	// 配置 API 路由
	s.SetupAPIRoutes()
}

// validateVersion 校验 API 版本
func (s *Server) validateVersion(c *gin.Context) bool {
	version := c.Param("version")
	// 从配置文件中获取允许的 API 版本
	allowedVersion := s.config.App.Version

	if version != allowedVersion {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": fmt.Sprintf("不支持的 API 版本：%s，当前支持的版本：%s", version, allowedVersion),
			"data":    nil,
		})
		return false
	}
	return true
}

// Start 启动 Web 服务器
func (s *Server) Start() error {
	addr := s.config.Server.Host + ":" + strconv.Itoa(s.config.Server.Port)
	logger.Info("正在启动 Web 服务器",
		zap.String("address", addr),
		zap.String("mode", gin.Mode()))

	if err := s.engine.Run(addr); err != nil {
		logger.Error("Web 服务器启动失败", zap.Error(err))
		return err
	}

	return nil
}

// LoggerMiddleware 自定义日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

// CORSMiddleware CORS 中间件
func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: []string{
			"https://app.apifox.com",
			"http://localhost:38080",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Token"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour, // 缓存预检结果
	})
}
