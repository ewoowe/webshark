package web

import (
	"context"
	"encoding/json"
	"webshark/internal/logger"
	"webshark/internal/websocket"

	"go.uber.org/zap"
)

// CaptureSessionHandler 处理抓包会话事件
type CaptureSessionHandler struct{}

// CaptureSessionParams 抓包会话参数
type CaptureSessionParams struct {
	SessionID string `json:"sessionId"`
	ClientID  string `json:"clientId"`
	Action    string `json:"action"` // "register" or "unregister"
}

// Handle 处理抓包会话事件
func (h *CaptureSessionHandler) Handle(ctx context.Context, event *websocket.WsEventMessage) error {
	var params CaptureSessionParams
	if err := json.Unmarshal(event.Params, &params); err != nil {
		logger.Error("解析抓包会话参数失败", zap.Error(err))
		return nil
	}

	switch params.Action {
	case "register":
		RegisterCaptureSession(params.ClientID, params.SessionID)
	case "unregister":
		UnregisterCaptureSession(params.SessionID)
	default:
		logger.Warn("未知的抓包会话动作", zap.String("action", params.Action))
	}

	return nil
}

// GetEventType 返回支持的事件类型
func (h *CaptureSessionHandler) GetEventType() []websocket.EventType {
	// 注册自定义事件类型
	captureSessionType := websocket.NewEventType("CaptureSession")
	websocket.RegisterEventType(captureSessionType)
	return []websocket.EventType{captureSessionType}
}
