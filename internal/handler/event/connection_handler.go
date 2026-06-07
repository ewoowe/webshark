package event

import (
	"context"
	"encoding/json"
	"webshark/internal/logger"
	"webshark/internal/websocket"

	"go.uber.org/zap"
)

// ConnectionEventHandler 连接事件处理器
type ConnectionEventHandler struct{}

// ConnectionParams 连接事件参数
type ConnectionParams struct {
	Message   string `json:"message"`
	ClientID  string `json:"clientId,omitempty"`
	SessionID string `json:"sessionId,omitempty"`
}

// Handle 处理连接事件
func (h *ConnectionEventHandler) Handle(ctx context.Context, event *websocket.WsEventMessage) error {
	var params ConnectionParams
	if err := json.Unmarshal(event.Params, &params); err != nil {
		logger.Error("解析连接参数失败", zap.Error(err))
		return nil // 连接事件解析失败不影响其他处理
	}

	switch event.EventType {
	case "Connected":
		logger.Info("客户端已连接",
			zap.String("message", params.Message),
			zap.String("clientID", params.ClientID))
	case "Disconnected":
		logger.Info("客户端已断开",
			zap.String("message", params.Message),
			zap.String("clientID", params.ClientID))
	}

	return nil
}

// GetEventType 返回支持的事件类型
func (h *ConnectionEventHandler) GetEventType() string {
	return ""
}
