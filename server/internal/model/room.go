package model

import "time"

// Room 房间状态
type Room struct {
	ID        string    `json:"id"`
	Code      string    `json:"code"` // 6位数字房间码
	HostID    string    `json:"host_id"`
	Status    string    `json:"status"` // waiting/playing/settling
	Players   []*Player `json:"players"`
	MaxSeats  int       `json:"max_seats"`
	CreatedAt time.Time `json:"created_at"`
}

// Player 玩家
type Player struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Seat      int    `json:"seat"` // 座位号 0-8
	Chips     int64  `json:"chips"`
	Connected bool   `json:"connected"`
	Ready     bool   `json:"ready"` // 是否准备
	// AI 相关字段
	IsAI    bool   `json:"is_ai"`    // 是否是AI
	AILevel string `json:"ai_level"` // AI等级: easy/normal/hard
}

// RoomUpdate 房间更新消息
type RoomUpdate struct {
	Type    string        `json:"type"`
	Room    *Room         `json:"room"`
	Players []*PlayerInfo `json:"players"`
}
