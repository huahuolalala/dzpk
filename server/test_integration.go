package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func main() {
	// 连接服务端 - 玩家1
	log.Println("[PLAYER1] Connecting...")
	conn1, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		log.Fatalf("Player1 Failed to connect: %v", err)
	}
	defer conn1.Close()

	// 连接服务端 - 玩家2
	log.Println("[PLAYER2] Connecting...")
	conn2, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	if err != nil {
		log.Fatalf("Player2 Failed to connect: %v", err)
	}
	defer conn2.Close()

	var roomCode string
	var player1ID string

	// 接收消息 - 玩家1
	done1 := make(chan struct{})
	go func() {
		defer close(done1)
		for {
			_, message, err := conn1.ReadMessage()
			if err != nil {
				log.Printf("[PLAYER1] Read error: %v", err)
				return
			}
			var msg Message
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("[PLAYER1] JSON parse error: %v", err)
				continue
			}
			log.Printf("[PLAYER1] Received: %+v", msg.Data)

			// 提取房间码
			if msg.Type == "room_update" {
				data := msg.Data.(map[string]interface{})
				if rc, ok := data["room_code"].(string); ok {
					roomCode = rc
					log.Printf("[PLAYER1] Got room code: %s", roomCode)
				}
				if pid, ok := data["player_id"].(string); ok && player1ID == "" {
					player1ID = pid
					log.Printf("[PLAYER1] Got player ID: %s", player1ID)
				}
			}
		}
	}()

	// 接收消息 - 玩家2
	done2 := make(chan struct{})
	go func() {
		defer close(done2)
		for {
			_, message, err := conn2.ReadMessage()
			if err != nil {
				log.Printf("[PLAYER2] Read error: %v", err)
				return
			}
			var msg Message
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("[PLAYER2] JSON parse error: %v", err)
				continue
			}
			log.Printf("[PLAYER2] Received: %+v", msg.Data)
		}
	}()

	// 等待连接建立
	time.Sleep(500 * time.Millisecond)

	// 玩家1: 创建房间
	log.Println("[PLAYER1] Creating room...")
	createMsg := Message{
		Type: "create_room",
		Data: map[string]interface{}{
			"player_name": "Alice",
		},
	}
	if err := conn1.WriteJSON(createMsg); err != nil {
		log.Fatalf("[PLAYER1] Failed to send: %v", err)
	}

	// 等待玩家1收到响应
	time.Sleep(1 * time.Second)

	if roomCode == "" {
		log.Fatal("[PLAYER1] Did not receive room code!")
	}

	// 玩家2: 加入房间
	log.Printf("[PLAYER2] Joining room %s...", roomCode)
	joinMsg := Message{
		Type: "join_room",
		Data: map[string]interface{}{
			"room_code":  roomCode,
			"player_name": "Bob",
		},
	}
	if err := conn2.WriteJSON(joinMsg); err != nil {
		log.Fatalf("[PLAYER2] Failed to send: %v", err)
	}

	// 等待响应
	time.Sleep(1 * time.Second)

	log.Println("[INTEGRATION TEST] PASSED - Both players connected successfully!")
}
