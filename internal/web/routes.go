package web

import (
	"webshark/internal/handler/web"
	"webshark/internal/utils"

	"github.com/gin-gonic/gin"
)

// SetupAPIRoutes 配置 API 路由
func (s *Server) SetupAPIRoutes() {
	// API 路由组（带版本参数）
	api := s.engine.Group("/api/:version/webshark")
	{
		// 版本校验中间件
		api.Use(func(c *gin.Context) {
			if !s.validateVersion(c) {
				c.Abort()
				return
			}
			c.Next()
		})

		// 获取远程网卡列表
		api.GET("/interfaces", web.GetInterfaces)

		// 开始抓包
		api.POST("/capture/start", web.StartCapture)

		// 停止抓包
		api.POST("/capture/stop", web.StopCapture)

		// 获取数据包详情
		api.GET("/capture/packet/detail", web.GetPacketDetail)

		// Host 管理路由组
		hosts := api.Group("/hosts")
		{
			// 创建主机
			hosts.POST("", web.CreateHost)

			// 获取主机列表
			hosts.GET("", web.ListHosts)

			// 搜索主机
			hosts.GET("/search", web.SearchHosts)

			// 获取单个主机
			hosts.GET("/:id", web.GetHost)

			// 更新主机
			hosts.PUT("/:id", web.UpdateHost)

			// 删除主机
			hosts.DELETE("/:id", web.DeleteHost)
		}

		// Task 管理路由组

	}

	// WebSocket 路由（带 clientID）
	ws := s.engine.Group("/websocket/:version/webshark")
	{
		ws.GET("/event/:ttype/:id", func(c *gin.Context) {
			taskType := c.Param("ttype")
			if taskType == "" {
				web.BadRequest(c, "Missing type parameter")
				return
			}
			id := c.Param("id")
			if id == "" {
				web.BadRequest(c, "Missing id parameter")
				return
			}

			// 获取 WebSocket 服务器
			wsServer := utils.GetWebSocketServer()
			if wsServer == nil {
				web.InternalErrorWithMsg(c, nil, "WebSocket server not initialized")
				return
			}

			// 处理 WebSocket 连接，传递 taskType 和 id
			wsServer.HandleWebSocket(c.Writer, c.Request, taskType, id)
		})
	}
}
