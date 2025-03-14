package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Room struct {
	RoomID      string `json:"roomid"`
	RoomName    string `json:"roomname"`
	Clients     map[*websocket.Conn]bool
	ChatHistory []ChatMessage
	mutex       sync.Mutex
}

type ChatMessage struct {
	Protocol string `json:"protocol"`
	Nickname string `json:"nickname"`
	Message  string `json:"message"`
	Time     string `json:"time"`
}

var (
	rooms = make(map[string]*Room)
	mutex sync.Mutex

	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func main() {
	http.HandleFunc("/getrooms", getRooms)
	http.HandleFunc("/ws", handleWebSocket)

	log.Println("Server starting on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func getRooms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	roomList := []map[string]string{}
	mutex.Lock()
	for id, room := range rooms {
		roomList = append(roomList, map[string]string{
			"roomid":   id,
			"roomname": room.RoomName,
		})
	}
	mutex.Unlock()

	json.NewEncoder(w).Encode(roomList)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Read error:", err)
			removeClientFromRooms(conn)
			break
		}

		protocol, _ := msg["protocol"].(string)
		switch protocol {
		case "openroom":
			handleOpenRoom(conn, msg)
		case "joinroom":
			handleJoinRoom(conn, msg)
		case "message":
			handleMessage(conn, msg)
		}
	}
}

func handleOpenRoom(conn *websocket.Conn, msg map[string]interface{}) {
	roomName, _ := msg["roomname"].(string)
	roomID := generateRoomID()

	mutex.Lock()
	rooms[roomID] = &Room{
		RoomID:      roomID,
		RoomName:    roomName,
		Clients:     make(map[*websocket.Conn]bool),
		ChatHistory: make([]ChatMessage, 0),
	}
	room := rooms[roomID]
	mutex.Unlock()

	room.mutex.Lock()
	room.Clients[conn] = true
	room.mutex.Unlock()

	conn.WriteJSON(map[string]interface{}{
		"protocol": "resopenroom",
		"status":   "ok",
		"roomid":   roomID,
	})
}

func handleJoinRoom(conn *websocket.Conn, msg map[string]interface{}) {
	roomID, _ := msg["roomid"].(string)

	mutex.Lock()
	room, exists := rooms[roomID]
	mutex.Unlock()

	if !exists {
		conn.WriteJSON(map[string]interface{}{
			"protocol": "resjoinroom",
			"status":   "error",
			"message":  "Room not found",
		})
		return
	}

	room.mutex.Lock()
	room.Clients[conn] = true
	chatHistory := room.ChatHistory
	room.mutex.Unlock()

	conn.WriteJSON(map[string]interface{}{
		"protocol":    "resjoinroom",
		"status":      "ok",
		"chathistory": chatHistory,
	})
}

func handleMessage(conn *websocket.Conn, msg map[string]interface{}) {
	roomID := findRoomIDByClient(conn)
	if roomID == "" {
		return
	}

	mutex.Lock()
	room := rooms[roomID]
	mutex.Unlock()

	chatMsg := ChatMessage{
		Protocol: "message",
		Nickname: msg["nickname"].(string),
		Message:  msg["message"].(string),
		Time:     time.Now().Format("15:04:05"),
	}

	room.mutex.Lock()
	// 管理历史记录，最多保存200条
	if len(room.ChatHistory) >= 200 {
		room.ChatHistory = room.ChatHistory[1:]
	}
	room.ChatHistory = append(room.ChatHistory, chatMsg)

	// 广播消息给所有客户端
	for client := range room.Clients {
		err := client.WriteJSON(chatMsg)
		if err != nil {
			log.Printf("Broadcast error: %v", err)
			delete(room.Clients, client)
			client.Close()
		}
	}
	room.mutex.Unlock()
}

func removeClientFromRooms(conn *websocket.Conn) {
	mutex.Lock()
	for _, room := range rooms {
		room.mutex.Lock()
		delete(room.Clients, conn)
		room.mutex.Unlock()
	}
	mutex.Unlock()
}

func findRoomIDByClient(conn *websocket.Conn) string {
	mutex.Lock()
	defer mutex.Unlock()

	for roomID, room := range rooms {
		room.mutex.Lock()
		if _, exists := room.Clients[conn]; exists {
			room.mutex.Unlock()
			return roomID
		}
		room.mutex.Unlock()
	}
	return ""
}

func generateRoomID() string {
	return time.Now().Format("20060102150405")
}
