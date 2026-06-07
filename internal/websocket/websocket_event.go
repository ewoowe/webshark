package websocket

import (
	"encoding/json"
	"time"
)

// WsEventMessage WebSocket 事件消息结构
type WsEventMessage struct {
	Timestamp int64           `json:"timestamp"` // UTC 时间戳（秒）
	EventType string          `json:"eventType"` // 事件类型
	Datetime  string          `json:"datetime"`  // 可读时间字符串
	Params    json.RawMessage `json:"params"`    // 事件详细参数（JSON 格式）
}

// NewWebSocketEventMessage 创建新的事件消息
func NewWebSocketEventMessage(eventType string, params interface{}) (*WsEventMessage, error) {
	// 序列化 params 为 JSON
	paramsBytes, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &WsEventMessage{
		Timestamp: now.Unix(),
		EventType: eventType,
		Datetime:  now.Format("2006-01-02 15:04:05"),
		Params:    paramsBytes,
	}, nil
}

// GetParams 解析 params 字段到指定的结构体
func (m *WsEventMessage) GetParams(target interface{}) error {
	return json.Unmarshal(m.Params, target)
}

// String 返回消息的字符串表示
func (m *WsEventMessage) String() string {
	data, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(data)
}

// IsValid 验证消息是否有效
func (m *WsEventMessage) IsValid() bool {
	return m.Timestamp > 0 && m.EventType != ""
}
