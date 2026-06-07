package service

import (
	"encoding/json"
	"sync"
	"webshark/internal/logger"
	"webshark/internal/utils"
	"webshark/internal/websocket"

	"go.uber.org/zap"
)

// PacketBroadcaster 负责将抓包数据广播到 WebSocket 客户端
type PacketBroadcaster struct {
	mu             sync.RWMutex
	sessionClients map[string]string // sessionID -> clientID 映射
}

var (
	broadcaster     *PacketBroadcaster
	broadcasterOnce sync.Once
)

// GetPacketBroadcaster 获取全局 PacketBroadcaster 实例
func GetPacketBroadcaster() *PacketBroadcaster {
	broadcasterOnce.Do(func() {
		broadcaster = &PacketBroadcaster{
			sessionClients: make(map[string]string),
		}
	})
	return broadcaster
}

// RegisterSessionClient 注册会话与客户端的关联
func (pb *PacketBroadcaster) RegisterSessionClient(sessionID, clientID string) {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	pb.sessionClients[sessionID] = clientID
	logger.Debug("注册会话-客户端映射",
		zap.String("sessionID", sessionID),
		zap.String("clientID", clientID))
}

// UnregisterSessionClient 注销会话与客户端的关联
func (pb *PacketBroadcaster) UnregisterSessionClient(sessionID string) {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	delete(pb.sessionClients, sessionID)
	logger.Debug("注销会话-客户端映射", zap.String("sessionID", sessionID))
}

// SendPacketToSession 发送数据包到指定会话的客户端
func (pb *PacketBroadcaster) SendPacketToSession(sessionID string, packet interface{}) error {
	pb.mu.RLock()
	clientID, exists := pb.sessionClients[sessionID]
	pb.mu.RUnlock()

	if !exists {
		logger.Warn("会话没有关联的客户端", zap.String("sessionID", sessionID))
		return nil
	}

	// 序列化数据包
	_, err := json.Marshal(packet)
	if err != nil {
		logger.Error("序列化数据包失败", zap.Error(err))
		return err
	}

	// 获取 WebSocket 服务器并发送消息
	wsServer := utils.GetWebSocketServer()
	if wsServer == nil {
		logger.Error("WebSocket 服务器未初始化")
		return nil
	}

	// 创建数据包事件
	params := map[string]interface{}{
		"sessionId": sessionID,
		"packet":    packet,
	}

	event, err := websocket.NewWebSocketEventMessage("PacketData", params)
	if err != nil {
		logger.Error("创建数据包事件失败", zap.Error(err))
		return err
	}

	eventBytes, _ := json.Marshal(event)

	// 发送到指定客户端
	err = wsServer.SendToClient(clientID, eventBytes)
	if err != nil {
		logger.Error("发送数据包到客户端失败",
			zap.String("clientID", clientID),
			zap.String("sessionID", sessionID),
			zap.Error(err))
		return err
	}

	return nil
}

// HasActiveClient 检查会话是否有活跃的客户端连接
func (pb *PacketBroadcaster) HasActiveClient(sessionID string) bool {
	pb.mu.RLock()
	defer pb.mu.RUnlock()
	_, exists := pb.sessionClients[sessionID]
	return exists
}
