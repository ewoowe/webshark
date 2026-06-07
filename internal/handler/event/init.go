package event

import (
	"webshark/internal/websocket"
)

// InitDefaultHandlers 注册所有默认处理器
func InitDefaultHandlers(dispatcher *websocket.EventDispatcher) {
	// 连接相关处理器
	connHandler := &ConnectionEventHandler{}
	dispatcher.RegisterFunc(websocket.EventTypeConnected, connHandler.Handle, 10)
	dispatcher.RegisterFunc(websocket.EventTypeDisconnected, connHandler.Handle, 10)

	// 订阅相关处理器
	subHandler := &SubscriptionEventHandler{}
	dispatcher.RegisterFunc(websocket.EventTypeSubscribe, subHandler.Handle, 10)
	dispatcher.RegisterFunc(websocket.EventTypeUnsubscribe, subHandler.Handle, 10)
}
