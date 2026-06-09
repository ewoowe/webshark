package web

import (
	"net/http"
	"webshark/internal/entity"
	"webshark/internal/logger"
	"webshark/internal/service"
	"webshark/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CaptureWebSocketHandler 处理抓包 WebSocket 连接
func CaptureWebSocketHandler(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Missing session_id parameter",
		})
		return
	}

	// 获取 WebSocket 服务器
	wsServer := utils.GetWebSocketServer()
	if wsServer == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "WebSocket server not initialized",
		})
		return
	}

	// 升级为 WebSocket 连接（HandleWebSocket 内部已经处理了升级）
	wsServer.HandleWebSocket(c.Writer, c.Request)

	// 注意：由于 HandleWebSocket 是异步的，我们需要在连接建立后注册会话
	// 这里我们通过事件处理器来完成注册
	logger.Info("抓包 WebSocket 连接请求", zap.String("sessionID", sessionID))
}

// RegisterCaptureSession 注册抓包会话与客户端的关联
func RegisterCaptureSession(clientID, sessionID string) {
	broadcaster := service.GetPacketBroadcaster()
	broadcaster.RegisterSessionClient(sessionID, clientID)
	logger.Info("注册抓包会话",
		zap.String("clientID", clientID),
		zap.String("sessionID", sessionID))
}

// UnregisterCaptureSession 注销抓包会话与客户端的关联
func UnregisterCaptureSession(sessionID string) {
	broadcaster := service.GetPacketBroadcaster()
	broadcaster.UnregisterSessionClient(sessionID)
	logger.Info("注销抓包会话", zap.String("sessionID", sessionID))
}

// StartCapture 开始抓包（Gin handler）
func StartCapture(c *gin.Context) {
	var req service.CaptureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorWithStruct(c, err, req)
		return
	}
	if req.DetailFormat == "" {
		req.DetailFormat = "normal"
	}

	taskInfo, err := service.StartCapture(req)
	if err != nil {
		logger.Error("开启抓包失败", zap.Any("CaptureRequest", req), zap.Error(err))
		InternalErrorWithMsg(c, err, "开启抓包失败")
		return
	}

	SuccessWithMsg(c, *taskInfo, "开启抓包成功")
}

// StopCapture 停止抓包（Gin handler）
func StopCapture(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, entity.ApiResponse[any]{
			Code: entity.Failure,
			Msg:  "Missing session_id",
		})
		return
	}

	err := service.StopCapture(sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, entity.ApiResponse[any]{
			Code: entity.Failure,
			Msg:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, entity.ApiResponse[any]{
		Code: entity.Success,
		Msg:  "Capture stopped",
	})
}
