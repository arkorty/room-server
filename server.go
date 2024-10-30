package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type MessageType string

const (
	TextUpdate     MessageType = "text-update"
	InitialContent MessageType = "initial-content"
	JoinRoom       MessageType = "join-room"
)

type BaseMessage struct {
	Type MessageType `json:"type"`
	Code string     `json:"code"`
}

type TextUpdateMessage struct {
	BaseMessage
	Content string `json:"content"`
}

type InitialContentMessage struct {
	BaseMessage
	Content string `json:"content"`
}

type JoinRoomMessage struct {
	BaseMessage
}

type Room struct {
	Content  string
	Clients  map[*websocket.Conn]bool
	mutex    sync.RWMutex
}

var (
	rooms = make(map[string]*Room)
	roomsMutex sync.RWMutex
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func main() {
	port := 8080
	http.HandleFunc("/textrt/api/v1", handleWebSocket)
	
	log.Printf("WebSocket server starting on port %d...", port)
	if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), nil); err != nil {
		log.Fatal("ListenAndServe error:", err)
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("New client connected from %s", conn.RemoteAddr())
	
	var currentRoom string
	
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message from %s: %v", conn.RemoteAddr(), err)
			handleClientDisconnection(conn, &currentRoom)
			break
		}

		if messageType != websocket.TextMessage {
			continue
		}

		var baseMsg BaseMessage
		if err := json.Unmarshal(message, &baseMsg); err != nil {
			log.Printf("Error unmarshaling message from %s: %v", conn.RemoteAddr(), err)
			continue
		}

		switch baseMsg.Type {
		case JoinRoom:
			handleJoinRoom(conn, message, &currentRoom)
		case TextUpdate:
			handleTextUpdate(conn, message, currentRoom)
		default:
			log.Printf("Unknown message type received: %s", baseMsg.Type)
		}
	}
}

func handleJoinRoom(conn *websocket.Conn, message []byte, currentRoom *string) {
	var joinMsg JoinRoomMessage
	if err := json.Unmarshal(message, &joinMsg); err != nil {
		log.Printf("Error unmarshaling join room message: %v", err)
		return
	}

	if *currentRoom != "" {
		leaveRoom(conn, *currentRoom)
	}

	*currentRoom = joinMsg.Code
	log.Printf("Client %s joining room: %s", conn.RemoteAddr(), *currentRoom)

	roomsMutex.Lock()
	if _, exists := rooms[*currentRoom]; !exists {
		rooms[*currentRoom] = &Room{
			Content: "",
			Clients: make(map[*websocket.Conn]bool),
			mutex:   sync.RWMutex{},
		}
		log.Printf("Created new room: %s", *currentRoom)
	}
	room := rooms[*currentRoom]
	room.mutex.Lock()
	room.Clients[conn] = true
	room.mutex.Unlock()
	roomsMutex.Unlock()

	initialMsg := InitialContentMessage{
		BaseMessage: BaseMessage{
			Type: InitialContent,
			Code: *currentRoom,
		},
		Content: room.Content,
	}
	
	if err := conn.WriteJSON(initialMsg); err != nil {
		log.Printf("Error sending initial content to %s: %v", conn.RemoteAddr(), err)
	} else {
		log.Printf("Sent initial content to client %s in room %s", conn.RemoteAddr(), *currentRoom)
	}
}

func handleTextUpdate(conn *websocket.Conn, message []byte, currentRoom string) {
	if currentRoom == "" {
		return
	}

	var updateMsg TextUpdateMessage
	if err := json.Unmarshal(message, &updateMsg); err != nil {
		log.Printf("Error unmarshaling text update message: %v", err)
		return
	}

	roomsMutex.RLock()
	room, exists := rooms[currentRoom]
	roomsMutex.RUnlock()
	
	if !exists {
		return
	}

	room.mutex.Lock()
	room.Content = updateMsg.Content
	room.mutex.Unlock()

	log.Printf("Broadcasting text update in room %s from client %s", currentRoom, conn.RemoteAddr())

	room.mutex.RLock()
	for client := range room.Clients {
		if client != conn {
			if err := client.WriteJSON(updateMsg); err != nil {
				log.Printf("Error broadcasting to client %s: %v", client.RemoteAddr(), err)
			}
		}
	}
	room.mutex.RUnlock()
}

func leaveRoom(conn *websocket.Conn, roomCode string) {
	roomsMutex.Lock()
	defer roomsMutex.Unlock()

	if room, exists := rooms[roomCode]; exists {
		room.mutex.Lock()
		delete(room.Clients, conn)
		clientCount := len(room.Clients)
		room.mutex.Unlock()

		if clientCount == 0 {
			delete(rooms, roomCode)
			log.Printf("Room %s deleted (no clients remaining)", roomCode)
		}
		log.Printf("Client %s left room %s", conn.RemoteAddr(), roomCode)
	}
}

func handleClientDisconnection(conn *websocket.Conn, currentRoom *string) {
	if *currentRoom != "" {
		leaveRoom(conn, *currentRoom)
	}
	log.Printf("Client %s disconnected", conn.RemoteAddr())
}
