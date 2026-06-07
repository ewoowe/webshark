package event

import (
	"context"
	"encoding/json"
	"webshark/internal/logger"
	"webshark/internal/websocket"

	"go.uber.org/zap"
)

// SubscriptionEventHandler 订阅事件处理器
type SubscriptionEventHandler struct{}

// SubscriptionParams 订阅事件参数
type SubscriptionParams struct {
	Message            string   `json:"message"`
	SubscribedEvents   []string `json:"subscribedEvents,omitempty"`
	UnsubscribedEvents []string `json:"unsubscribedEvents,omitempty"`
}

// Handle 处理订阅事件
func (h *SubscriptionEventHandler) Handle(ctx context.Context, event *websocket.WsEventMessage) error {
	var params SubscriptionParams
	if err := json.Unmarshal(event.Params, &params); err != nil {
		logger.Error("解析订阅参数失败", zap.Error(err))
		return nil
	}

	switch event.EventType {
	case "Subscribe":
		logger.Info("订阅成功",
			zap.String("message", params.Message),
			zap.Strings("events", params.SubscribedEvents))
	case "Unsubscribe":
		logger.Info("取消订阅成功",
			zap.String("message", params.Message),
			zap.Strings("events", params.UnsubscribedEvents))
	}

	return nil
}

// GetEventType 返回支持的事件类型
func (h *SubscriptionEventHandler) GetEventType() []websocket.EventType {
	return []websocket.EventType{websocket.EventTypeSubscribe, websocket.EventTypeUnsubscribe}
}
