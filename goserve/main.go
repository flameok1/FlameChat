package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Room struct {
	ID      string `json:"roomid"`
	Name    string `json:"roomname"`
	Clients map[*Client]bool
	mu      sync.Mutex
}

type Client struct {
	conn *websocket.Conn
	room *Room
}

type Message struct {
	Protocol string `json:"protocol"`
	RoomID   string `json:"roomid,omitempty"`
	RoomName string `json:"roomname,omitempty"`
	Status   string `json:"status,omitempty"`
	Nickname string `json:"nickname,omitempty"`
	Message  string `json:"message,omitempty"`
}

var (
	rooms    = make(map[string]*Room)
	roomsMu  sync.Mutex
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允許所有來源的連接
		},
	}
)

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/getrooms", handleGetRooms)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleGetRooms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	roomsMu.Lock()
	roomList := make([]map[string]string, 0)
	for _, room := range rooms {
		roomList = append(roomList, map[string]string{
			"roomid":   room.ID,
			"roomname": room.Name,
		})
	}
	roomsMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(roomList)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	client := &Client{conn: conn}

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		switch msg.Protocol {
		case "openroom":
			handleOpenRoom(client, &msg)
		case "joinroom":
			handleJoinRoom(client, &msg)
		case "message":
			handleChatMessage(client, &msg)
		}
	}
}

func handleOpenRoom(client *Client, msg *Message) {
	roomID := generateRoomID()
	room := &Room{
		ID:      roomID,
		Name:    msg.RoomName,
		Clients: make(map[*Client]bool),
	}

	roomsMu.Lock()
	rooms[roomID] = room
	roomsMu.Unlock()

	room.mu.Lock()
	room.Clients[client] = true
	room.mu.Unlock()

	client.room = room

	response := Message{
		Protocol: "resopenroom",
		Status:   "ok",
		RoomID:   roomID,
	}
	client.conn.WriteJSON(response)
}

func handleJoinRoom(client *Client, msg *Message) {
	roomsMu.Lock()
	room, exists := rooms[msg.RoomID]
	roomsMu.Unlock()

	if !exists {
		client.conn.WriteJSON(Message{
			Protocol: "resjoinroom",
			Status:   "error",
			Message:  "Room not found",
		})
		return
	}

	room.mu.Lock()
	room.Clients[client] = true
	room.mu.Unlock()

	client.room = room

	client.conn.WriteJSON(Message{
		Protocol: "resjoinroom",
		Status:   "ok",
	})
}

func handleChatMessage(client *Client, msg *Message) {
	if client.room == nil {
		return
	}

	client.room.mu.Lock()
	for c := range client.room.Clients {
		c.conn.WriteJSON(Message{
			Protocol: "message",
			Nickname: msg.Nickname,
			Message:  msg.Message,
		})
	}
	client.room.mu.Unlock()
}

func generateRoomID() string {
	return fmt.Sprintf("room_%d", time.Now().UnixNano())
}
