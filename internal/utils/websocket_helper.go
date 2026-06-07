package utils

import (
	"webshark/internal/websocket"
)

// GetWebSocketServer 获取全局 WebSocket 服务器实例
// 该函数由 main.go 中的实际实现填充
var GetWebSocketServer func() *websocket.WebSocketServer
