package main

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type WebSocketManager struct {
	clients      map[string]*websocket.Conn
	clientsMutex *sync.RWMutex
	upgrader     websocket.Upgrader
	ctx          context.Context
}

// 初始化 WebSocketManager
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		clients:      make(map[string]*websocket.Conn),
		clientsMutex: &sync.RWMutex{},
		upgrader:     websocket.Upgrader{},
		ctx:          context.Background(),
	}
}

// 添加客戶端連接
func (wsm *WebSocketManager) AddClient(userId string, conn *websocket.Conn) {
	wsm.clientsMutex.Lock()
	defer wsm.clientsMutex.Unlock()
	wsm.clients[userId] = conn
}

// 刪除客戶端連接
func (wsm *WebSocketManager) RemoveClient(userId string) {
	wsm.clientsMutex.Lock()
	defer wsm.clientsMutex.Unlock()
	delete(wsm.clients, userId)
}

// 發送消息給指定客戶端
func (wsm *WebSocketManager) SendMessage(userId string, message string) error {
	wsm.clientsMutex.RLock()
	defer wsm.clientsMutex.RUnlock()

	conn, ok := wsm.clients[userId]
	if !ok {
		return nil // 用戶不在線
	}
	return conn.WriteMessage(websocket.TextMessage, []byte(message))
}

func main() {
	wsm := NewWebSocketManager()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		userId := r.URL.Query().Get("userId")
		if userId == "" {
			http.Error(w, "缺少 userId", http.StatusBadRequest)
			return
		}

		conn, err := wsm.upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("WebSocket 錯誤:", err)
			return
		}

		// 添加客戶端
		wsm.AddClient(userId, conn)

		defer func() {
			// 移除客戶端
			wsm.RemoveClient(userId)
			conn.Close()
		}()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				log.Println("WebSocket 讀取錯誤:", err)
				break
			}
		}
	})

	log.Println("Server started at :8080")
	http.ListenAndServe(":8080", nil)
}
