package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
	"webshark/internal/logger"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// WebSocketServer WebSocket 服务端结构
type WebSocketServer struct {
	host       string
	port       int
	upgrader   websocket.Upgrader
	clients    map[string]*ClientConnection // 使用 clientId 作为 key
	clientMu   sync.RWMutex
	eventChan  chan *WsEventMessage
	closeChan  chan struct{}
	wg         sync.WaitGroup // 用于等待服务端所有协程结束
	dispatcher EventDispatcherInterface
	// 标记是否已经打印过无客户端连接的警告日志
	noClientWarningLogged bool
}

// ClientConnection 客户端连接信息
type ClientConnection struct {
	conn      *websocket.Conn
	clientID  string
	sessionID string
	sendChan  chan []byte
	closeChan chan struct{}
}

// EventDispatcherInterface 事件分发器接口
type EventDispatcherInterface interface {
	Dispatch(event *WsEventMessage) error
}

// ServerOption 服务端配置选项
type ServerOption func(*WebSocketServer)

// WithEventDispatcher 设置事件分发器
func WithEventDispatcher(dispatcher EventDispatcherInterface) ServerOption {
	return func(s *WebSocketServer) {
		s.dispatcher = dispatcher
	}
}

// WithEventChannelSize 设置事件通道大小
func WithEventChannelSize(size int) ServerOption {
	return func(s *WebSocketServer) {
		s.eventChan = make(chan *WsEventMessage, size)
	}
}

// NewWebSocketServer 创建新的 WebSocket 服务端
func NewWebSocketServer(host string, port int, options ...ServerOption) *WebSocketServer {
	server := &WebSocketServer{
		host:      host,
		port:      port,
		clients:   make(map[string]*ClientConnection),
		eventChan: make(chan *WsEventMessage, 256),
		closeChan: make(chan struct{}),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// 允许所有来源
				return true
			},
		},
	}

	// 应用配置选项
	for _, option := range options {
		option(server)
	}

	return server
}

// GetEventChan 获取事件通道
func (s *WebSocketServer) GetEventChan() chan *WsEventMessage {
	return s.eventChan
}

// Start 启动 WebSocket 服务端（仅启动事件处理协程，HTTP 服务器由Web服务管理）
func (s *WebSocketServer) Start() {
	// 启动事件处理协程
	s.wg.Add(1)
	go s.eventHandler()
}

// HandleWebSocket 处理 WebSocket 连接请求（导出为公开方法）
func (s *WebSocketServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 升级为 WebSocket 连接
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("WebSocket 升级失败", zap.Error(err))
		return
	}

	// 生成客户端 ID 和 Session ID
	clientID := generateClientID()
	sessionID := generateSessionID()

	// 创建客户端连接
	client := &ClientConnection{
		conn:      conn,
		clientID:  clientID,
		sessionID: sessionID,
		sendChan:  make(chan []byte, 256),
		closeChan: make(chan struct{}),
	}

	// 注册客户端
	s.registerClient(client)

	logger.Info("新客户端连接",
		zap.String("clientID", clientID),
		zap.String("sessionID", sessionID),
		zap.String("remoteAddr", r.RemoteAddr))

	// 发送连接成功事件
	s.sendConnectedEvent(client)

	// 启动读写协程
	s.wg.Add(2)
	go s.readPump(client)
	go s.writePump(client)
}

// HandleWebSocketWithClientID 处理 WebSocket 连接请求，使用指定的 clientID
func (s *WebSocketServer) HandleWebSocketWithClientID(w http.ResponseWriter, r *http.Request, clientID string) {
	// 升级为 WebSocket 连接
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("WebSocket 升级失败", zap.Error(err))
		return
	}

	// 生成 Session ID
	sessionID := generateSessionID()

	// 创建客户端连接
	client := &ClientConnection{
		conn:      conn,
		clientID:  clientID,
		sessionID: sessionID,
		sendChan:  make(chan []byte, 256),
		closeChan: make(chan struct{}),
	}

	// 注册客户端
	s.registerClient(client)

	logger.Info("新客户端连接（自定义 clientID）",
		zap.String("clientID", clientID),
		zap.String("sessionID", sessionID),
		zap.String("remoteAddr", r.RemoteAddr))

	// 发送连接成功事件
	s.sendConnectedEvent(client)

	// 启动读写协程
	s.wg.Add(2)
	go s.readPump(client)
	go s.writePump(client)
}

// registerClient 注册客户端
func (s *WebSocketServer) registerClient(client *ClientConnection) {
	s.clientMu.Lock()
	defer s.clientMu.Unlock()

	// 检查是否已存在相同 clientID 的连接
	if oldClient, exists := s.clients[client.clientID]; exists {
		logger.Info("检测到重复 clientID，关闭旧连接",
			zap.String("clientID", client.clientID),
			zap.String("oldSessionID", oldClient.sessionID),
			zap.String("newSessionID", client.sessionID))

		// 关闭旧连接
		select {
		case <-oldClient.closeChan:
			// 已经关闭
		default:
			close(oldClient.closeChan)
		}
		// 关闭 WebSocket 连接
		_ = oldClient.conn.Close()
		// 关闭发送通道
		close(oldClient.sendChan)
	}

	// 注册新客户端
	s.clients[client.clientID] = client
}

// unregisterClient 注销客户端
func (s *WebSocketServer) unregisterClient(client *ClientConnection) {
	s.clientMu.Lock()
	defer s.clientMu.Unlock()
	delete(s.clients, client.clientID)
}

// SendToClient 发送消息到指定客户端
func (s *WebSocketServer) SendToClient(clientID string, message []byte) error {
	s.clientMu.RLock()
	client, exists := s.clients[clientID]
	s.clientMu.RUnlock()

	if !exists {
		return fmt.Errorf("客户端不存在: %s", clientID)
	}

	select {
	case client.sendChan <- message:
		return nil
	default:
		return fmt.Errorf("客户端发送通道已满: %s", clientID)
	}
}

// sendConnectedEvent 发送连接成功事件
func (s *WebSocketServer) sendConnectedEvent(client *ClientConnection) {
	params := map[string]interface{}{
		"message":   "连接成功",
		"clientId":  client.clientID,
		"sessionId": client.sessionID,
	}

	event, err := NewWebSocketEventMessage(EventTypeConnected.String(), params)
	if err != nil {
		logger.Error("创建连接事件失败", zap.Error(err))
		return
	}

	eventData, _ := json.Marshal(event)
	select {
	case client.sendChan <- eventData:
	default:
		logger.Warn("发送连接事件失败，通道已满")
	}
}

// readPump 从客户端读取消息
func (s *WebSocketServer) readPump(client *ClientConnection) {
	defer s.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("readPump 发生异常", zap.Any("panic", r))
		}
	}()

	// 限制读取大小为 10MB
	client.conn.SetReadLimit(10 * 1024 * 1024)

	for {
		select {
		case <-client.closeChan:
			logger.Debug("readPump: closeChan 已关闭", zap.String("clientID", client.clientID))
			return
		default:
			messageType, message, err := client.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logger.Error("WebSocket 错误", zap.Error(err), zap.String("clientID", client.clientID))
				}
				s.handleDisconnect(client)
				return
			}

			if messageType == websocket.PingMessage {
				// 自动回复 Pong
				continue
			}

			// 解析消息
			var eventMsg WsEventMessage
			if err := json.Unmarshal(message, &eventMsg); err != nil {
				logger.Debug("解析消息失败",
					zap.String("message", string(message)),
					zap.Error(err),
					zap.String("clientID", client.clientID))
				continue
			}

			// 根据事件类型处理
			s.handleMessage(client, &eventMsg)
		}
	}
}

// writePump 向客户端写入消息
func (s *WebSocketServer) writePump(client *ClientConnection) {
	defer s.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			logger.Error("writePump 发生异常", zap.Any("panic", r))
		}
	}()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-client.closeChan:
			logger.Debug("writePump: closeChan 已关闭", zap.String("clientID", client.clientID))
			return
		case message, ok := <-client.sendChan:
			if !ok {
				return
			}

			err := client.conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				logger.Error("发送消息失败", zap.Error(err), zap.String("clientID", client.clientID))
				s.handleDisconnect(client)
				return
			}
		case <-ticker.C:
			// 发送 ping 消息保持连接
			err := client.conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				logger.Error("发送 ping 消息失败", zap.Error(err), zap.String("clientID", client.clientID))
				s.handleDisconnect(client)
				return
			}
		}
	}
}

// handleMessage 处理接收到的消息
func (s *WebSocketServer) handleMessage(client *ClientConnection, event *WsEventMessage) {
	// 根据事件类型处理
	switch event.EventType {
	case EventTypePing.eventType:
		s.handlePing(client)
	case EventTypeSubscribe.eventType:
		s.handleSubscribe(client, event)
	default:
		// 转发给事件分发器处理
		if s.dispatcher != nil {
			if err := s.dispatcher.Dispatch(event); err != nil {
				logger.Error("事件分发失败",
					zap.String("eventType", event.EventType),
					zap.Error(err))
			}
		}
	}
}

// handlePing 处理 Ping 消息
func (s *WebSocketServer) handlePing(client *ClientConnection) {
	params := map[string]interface{}{
		"message":    "pong",
		"clientId":   client.clientID,
		"serverTime": time.Now().Format(time.RFC3339),
	}

	event, err := NewWebSocketEventMessage(EventTypePong.String(), params)
	if err != nil {
		logger.Error("创建 Pong 事件失败", zap.Error(err))
		return
	}

	eventData, _ := json.Marshal(event)
	select {
	case client.sendChan <- eventData:
	default:
		logger.Warn("发送 Pong 消息失败，通道已满")
	}
}

// handleSubscribe 处理订阅消息
func (s *WebSocketServer) handleSubscribe(client *ClientConnection, event *WsEventMessage) {
	var params struct {
		Events []string `json:"events"`
	}

	if err := json.Unmarshal(event.Params, &params); err != nil {
		logger.Error("解析订阅参数失败", zap.Error(err))
		return
	}

	logger.Info("客户端订阅事件",
		zap.String("clientID", client.clientID),
		zap.Strings("events", params.Events))

	// 回复订阅成功
	responseParams := map[string]interface{}{
		"message":  "订阅成功",
		"clientId": client.clientID,
		"events":   params.Events,
	}

	responseEvent, err := NewWebSocketEventMessage("Subscribed", responseParams)
	if err != nil {
		logger.Error("创建订阅响应失败", zap.Error(err))
		return
	}

	eventData, _ := json.Marshal(responseEvent)
	select {
	case client.sendChan <- eventData:
	default:
		logger.Warn("发送订阅响应失败，通道已满")
	}
}

// handleDisconnect 处理断开连接
func (s *WebSocketServer) handleDisconnect(client *ClientConnection) {
	select {
	case <-client.closeChan:
		return // 已经关闭
	default:
		close(client.closeChan)
	}

	s.unregisterClient(client)
	err := client.conn.Close()
	if err != nil {
		return
	}

	// 发送断开连接事件
	s.sendDisconnectedEvent(client)

	logger.Info("客户端断开连接",
		zap.String("clientID", client.clientID),
		zap.String("sessionID", client.sessionID))
}

// sendDisconnectedEvent 发送断开连接事件
func (s *WebSocketServer) sendDisconnectedEvent(client *ClientConnection) {
	params := map[string]interface{}{
		"message":   "连接已断开",
		"clientId":  client.clientID,
		"sessionId": client.sessionID,
	}

	event, err := NewWebSocketEventMessage(EventTypeDisconnected.String(), params)
	if err != nil {
		logger.Error("创建断开事件失败", zap.Error(err))
		return
	}

	// 广播给所有客户端
	s.broadcastEvent(event)
}

// eventHandler 事件处理协程
func (s *WebSocketServer) eventHandler() {
	defer s.wg.Done()
	logger.Info("事件处理协程已启动")
	defer logger.Info("事件处理协程已退出")

	for {
		select {
		case event := <-s.eventChan:
			// 广播给所有客户端
			s.broadcastEvent(event)
		case <-s.closeChan:
			return
		}
	}
}

// broadcastEvent 广播事件给所有客户端
func (s *WebSocketServer) broadcastEvent(event *WsEventMessage) {
	eventData, err := json.Marshal(event)
	if err != nil {
		logger.Error("序列化事件失败", zap.Error(err))
		return
	}

	s.clientMu.Lock()
	clientCount := len(s.clients)

	if clientCount == 0 {
		// 只有当之前没有打印过警告时才打印
		if !s.noClientWarningLogged {
			// 序列化事件为JSON字符串用于日志输出
			eventJSON, _ := json.Marshal(event)
			logger.Warn("当前没有客户端连接，事件将被丢弃",
				zap.String("eventType", event.EventType),
				zap.String("event", string(eventJSON)))
			s.noClientWarningLogged = true
		}
		s.clientMu.Unlock()
		return
	}

	// 如果有客户端连接，重置警告标志，以便下次无客户端时再次打印
	if s.noClientWarningLogged {
		s.noClientWarningLogged = false
	}

	s.clientMu.Unlock()

	s.clientMu.RLock()
	defer s.clientMu.RUnlock()

	for clientID, client := range s.clients {
		select {
		case client.sendChan <- eventData:
		default:
			logger.Warn("广播消息失败，通道已满",
				zap.String("eventType", event.EventType),
				zap.String("clientID", clientID))
		}
	}
}

// Stop 停止 WebSocket 服务端
func (s *WebSocketServer) Stop(ctx context.Context) error {
	logger.Info("正在关闭 WebSocket 服务端...")

	// 先关闭事件处理协程，防止新的事件加入队列
	close(s.closeChan)

	// 关闭所有客户端连接
	s.clientMu.Lock()
	for _, client := range s.clients {
		// 先关闭发送通道，确保 writePump 立即退出
		close(client.sendChan)
		// 再处理断开逻辑
		s.handleDisconnect(client)
	}
	// 清空客户端映射
	s.clients = make(map[string]*ClientConnection)
	s.clientMu.Unlock()

	// 等待所有协程结束
	s.wg.Wait()

	logger.Info("WebSocket 服务端已关闭")
	return nil
}

// GetClientCount 获取客户端数量
func (s *WebSocketServer) GetClientCount() int {
	s.clientMu.RLock()
	defer s.clientMu.RUnlock()
	return len(s.clients)
}

// generateClientID 生成客户端 ID
func generateClientID() string {
	return fmt.Sprintf("client_%d", time.Now().UnixNano())
}

// generateSessionID 生成 Session ID
func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}
