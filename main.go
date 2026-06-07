package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	"webshark/internal/config"
	"webshark/internal/gorm"
	eventhandler "webshark/internal/handler/event"
	webhandler "webshark/internal/handler/web"
	"webshark/internal/logger"
	"webshark/internal/service"
	"webshark/internal/utils"
	"webshark/internal/web"
	"webshark/internal/websocket"

	"go.uber.org/zap"
)

// 全局 WebSocket 服务器实例（用于在其他包中访问）
var globalWebSocketServer *websocket.WebSocketServer

// GetGlobalWebSocketServer 获取全局 WebSocket 服务器实例
func GetGlobalWebSocketServer() *websocket.WebSocketServer {
	return globalWebSocketServer
}

// init 初始化 utils 包中的全局函数
func init() {
	utils.GetWebSocketServer = GetGlobalWebSocketServer
}

// App 应用主结构
type App struct {
	conf       *config.Config             // 全局配置文件
	dispatcher *websocket.EventDispatcher // 事件分发器
	webServer  *web.Server                // Web 服务端
	wsServer   *websocket.WebSocketServer // WebSocket 服务端
	sigChan    chan os.Signal             // 信号通道
}

// NewApp 创建新应用实例
func NewApp() *App {
	return &App{
		sigChan: make(chan os.Signal, 1),
	}
}

func main() {
	fmt.Println("Starting WebShark...")
	app := NewApp()

	if err := app.Run(); err != nil {
		logger.Fatal("应用启动失败", zap.Error(err))
	}

	logger.Info("Started WebShark...")

	// 等待停止信号
	app.WaitSignal()

	logger.Info("Shut down WebShark...")
}

// Run 运行应用
func (a *App) Run() (err error) {
	// 1. 初始化配置和日志
	if err = a.initConfigAndLogger(); err != nil {
		return
	}

	// 2. 初始化事件系统和DB（在连接/启动服务前初始化）
	if err = a.initEventAndDb(); err != nil {
		return
	}

	// 3. 启动Web组件
	if err = a.runWeb(); err != nil {
		return
	}

	return nil
}

// initConfigAndLogger 初始化配置和日志
func (a *App) initConfigAndLogger() error {
	var err error
	a.conf, err = config.InitConfig("")
	if err != nil {
		return err
	}

	logConfig := &logger.Config{
		Level:        a.conf.Log.Level,
		Format:       a.conf.Log.Format,
		OutputPaths:  a.conf.Log.OutputPaths,
		RotateConfig: convertRotateConfig(a.conf.Log.RotateConfig),
	}
	logger.InitLoggerWithConfig(logConfig)

	logger.Info("配置加载成功",
		zap.String("log_level", a.conf.Log.Level),
		zap.String("log_format", a.conf.Log.Format))

	// 初始化 Nacos配置客户端
	if err := config.InitNacosConfig(); err != nil {
		logger.Error("初始化 Nacos客户端失败", zap.Error(err))
		return err
	}
	logger.Info("Nacos配置客户端已初始化")

	// 启动动态配置管理器
	dynamicConfigMgr := config.NewDynamicConfigManager()
	if err := dynamicConfigMgr.Start(); err != nil {
		logger.Warn("启动动态配置管理失败", zap.Error(err))
		return err
	}
	logger.Info("动态配置管理器已启动")

	return nil
}

// convertRotateConfig 转换日志切割配置
func convertRotateConfig(rotateConfig *config.LogRotateConfig) *logger.LogRotateConfig {
	if rotateConfig == nil {
		return nil
	}
	return &logger.LogRotateConfig{
		Filename:   rotateConfig.Filename,
		MaxSize:    rotateConfig.MaxSize,
		MaxBackups: rotateConfig.MaxBackups,
		MaxAge:     rotateConfig.MaxAge,
		Compress:   rotateConfig.Compress,
	}
}

// initEventAndDb 初始化事件系统
func (a *App) initEventAndDb() error {
	a.dispatcher = websocket.NewEventDispatcher(websocket.WithEventBufferSize(512))

	// 初始化数据服务（从配置文件读取数据库配置）
	if a.conf.Database != nil && a.conf.Database.Host != "" {
		repo, err := gorm.NewWebSharkRepository(a.conf.Database)
		if err != nil {
			logger.Error("创建数据仓库失败",
				zap.String("host", a.conf.Database.Host),
				zap.Int("port", a.conf.Database.Port),
				zap.String("dbname", a.conf.Database.DBName),
				zap.Error(err))
			return err
		}

		webhandler.InitHostService(service.NewHostService(repo))

		logger.Info("数据服务已初始化",
			zap.String("host", a.conf.Database.Host),
			zap.String("dbname", a.conf.Database.DBName))
	} else {
		noDbCfg := fmt.Errorf("未配置数据库，跳过数据服务初始化")
		logger.Error(noDbCfg.Error())
		return noDbCfg
	}

	// 初始化默认的事件处理器
	eventhandler.InitDefaultHandlers(a.dispatcher)
	a.dispatcher.Start()

	return nil
}

// runWeb 运行Web相关组件
func (a *App) runWeb() error {
	// 启动 websocket 服务端
	a.startWebSocketServer()

	// 启动 Web 服务端
	if err := a.startWebServer(); err != nil {
		return err
	}

	// 启动后台协程
	a.startBackgroundWorkers()

	return nil
}

// startWebSocketServer 启动 WebSocket 服务端
func (a *App) startWebSocketServer() {
	logger.Info("正在启动 WebSocket 服务端",
		zap.String("host", a.conf.Server.Host),
		zap.Int("port", a.conf.Server.Port),
		zap.String("url", a.conf.Server.URL))

	// 创建服务端，使用当前的事件分发器
	a.wsServer = websocket.NewWebSocketServer(
		a.conf.Server.Host,
		a.conf.Server.Port,
		websocket.WithEventDispatcher(a.dispatcher),
		websocket.WithEventChannelSize(512),
	)

	// 保存到全局变量
	globalWebSocketServer = a.wsServer

	// 启动 WebSocket 服务端（scc事件处理协程）
	a.wsServer.Start()

	logger.Info("WebSocket 服务端已创建",
		zap.String("address", a.conf.Server.Host+":"+strconv.Itoa(a.conf.Server.Port)),
		zap.String("path", a.conf.Server.URL))

}

// startWebServer 启动 Web 服务端
func (a *App) startWebServer() error {
	// 创建 Web 服务器
	a.webServer = web.NewServer(a.conf)

	// 设置路由
	a.webServer.SetupRoutes()

	var webServerErr error
	// 在协程中启动
	go func() {
		if err := a.webServer.Start(); err != nil {
			webServerErr = err
		}
	}()

	// 等待一小段时间确保服务端启动
	time.Sleep(1000 * time.Millisecond)

	if webServerErr != nil {
		return webServerErr
	}

	logger.Info("Web 服务端已启动",
		zap.String("address", a.conf.Server.Host+":"+strconv.Itoa(a.conf.Server.Port)))

	return nil
}

// startBackgroundWorkers 启动后台协程
func (a *App) startBackgroundWorkers() {
	// 所有后台工作已移到客户端内部
}

// WaitSignal 等待停止信号
func (a *App) WaitSignal() {
	signal.Notify(a.sigChan, syscall.SIGINT, syscall.SIGTERM)
	logger.Info("等待停止信号...", zap.String("signal", "SIGINT, SIGTERM"))
	capSig := <-a.sigChan
	signal.Stop(a.sigChan)

	a.ShutDown(capSig)
}

func (a *App) ShutDown(capSig os.Signal) {
	logger.Info("接收到信号，正在关闭连接...", zap.String("signal", capSig.String()))

	// 停止事件分发器
	a.dispatcher.Stop()

	// 关闭 WebSocket 服务端（如果已启动）
	if a.wsServer != nil {
		logger.Info("正在关闭 WebSocket 服务端...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel() // 确保一定会释放资源
		if err := a.wsServer.Stop(ctx); err != nil {
			logger.Error("关闭 WebSocket 服务端失败", zap.Error(err))
		}
	}

	// 关闭 Web 服务端（如果已启动）
	if a.webServer != nil {
		logger.Info("正在关闭 Web 服务端...")
	}

	logger.Info("已退出")
	_ = logger.Sync()
}
