package handler

import (
	"log"
	"net/http"
	"sync"
	"webshark/internal/service"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源
	},
}

var clients = make(map[string]*websocket.Conn)
var clientsMutex sync.RWMutex

func captureWebSocket(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "Missing session_id", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// 注册客户端
	clientsMutex.Lock()
	clients[sessionID] = conn
	clientsMutex.Unlock()

	// 在服务层注册客户端
	service.RegisterClient(sessionID, conn)

	// 清理
	defer func() {
		clientsMutex.Lock()
		delete(clients, sessionID)
		clientsMutex.Unlock()
		service.UnregisterClient(sessionID)
	}()

	// 保持连接，等待客户端断开
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// BroadcastPacket 向指定会话广播数据包
func BroadcastPacket(sessionID string, packetData []byte) {
	clientsMutex.RLock()
	defer clientsMutex.RUnlock()

	if conn, ok := clients[sessionID]; ok {
		err := conn.WriteMessage(websocket.TextMessage, packetData)
		if err != nil {
			log.Printf("WebSocket write error for session %s: %v", sessionID, err)
			conn.Close()
		}
	}
}
