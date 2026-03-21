package repository

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/dz-poker/server/internal/model"
)

// RoomStore 房间内存存储
type RoomStore struct {
	mu    sync.RWMutex
	rooms map[string]*model.Room
	codes map[string]string // code -> roomID
}

func NewRoomStore() *RoomStore {
	return &RoomStore{
		rooms: make(map[string]*model.Room),
		codes: make(map[string]string),
	}
}

// GenerateRoomCode 生成6位数字房间码（调用者需持有锁）
func (s *RoomStore) GenerateRoomCode() string {
	for i := 0; i < 100; i++ {
		code := fmt.Sprintf("%06d", rand.Intn(1000000))
		if _, ok := s.codes[code]; !ok {
			s.codes[code] = ""
			return code
		}
	}
	return ""
}

// CreateRoom 创建房间
func (s *RoomStore) CreateRoom(host *model.Player) (*model.Room, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	code := s.GenerateRoomCode()
	if code == "" {
		return nil, fmt.Errorf("failed to generate room code")
	}

	room := &model.Room{
		ID:        fmt.Sprintf("room_%d", time.Now().UnixNano()),
		Code:      code,
		HostID:    host.ID,
		Status:    "waiting",
		Players:   []*model.Player{host},
		MaxSeats:  9,
		CreatedAt: time.Now(),
	}

	s.rooms[room.ID] = room
	s.codes[code] = room.ID
	host.Seat = 0
	host.Ready = true // 房主创建时默认已准备

	return room, nil
}

// GetRoom 获取房间
func (s *RoomStore) GetRoom(roomID string) (*model.Room, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	room, ok := s.rooms[roomID]
	return room, ok
}

// GetRoomByCode 通过房间码获取房间
func (s *RoomStore) GetRoomByCode(code string) (*model.Room, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	roomID, ok := s.codes[code]
	if !ok {
		return nil, false
	}

	room, ok := s.rooms[roomID]
	return room, ok
}

// AddPlayer 添加玩家
func (s *RoomStore) AddPlayer(roomID string, player *model.Player) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, ok := s.rooms[roomID]
	if !ok {
		return fmt.Errorf("room not found")
	}

	if len(room.Players) >= room.MaxSeats {
		return fmt.Errorf("room is full")
	}

	// 分配座位
	seat := len(room.Players)
	player.Seat = seat
	player.Connected = true
	room.Players = append(room.Players, player)

	return nil
}

// RemovePlayer 移除玩家
func (s *RoomStore) RemovePlayer(roomID string, playerID string) (*model.Room, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, ok := s.rooms[roomID]
	if !ok {
		return nil, false
	}

	newPlayers := make([]*model.Player, 0, len(room.Players))
	var removed *model.Player
	for _, p := range room.Players {
		if p.ID == playerID {
			removed = p
		} else {
			newPlayers = append(newPlayers, p)
		}
	}

	if removed == nil {
		return room, false
	}

	room.Players = newPlayers

	// 房主离开，解散房间
	if removed.ID == room.HostID {
		delete(s.rooms, roomID)
		delete(s.codes, room.Code)
		return nil, true // true 表示房间已解散
	}

	return room, false
}

// DeleteRoom 删除房间
func (s *RoomStore) DeleteRoom(roomID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if room, ok := s.rooms[roomID]; ok {
		delete(s.codes, room.Code)
		delete(s.rooms, roomID)
	}
}

// ListRooms 列出所有房间
func (s *RoomStore) ListRooms() []*model.Room {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rooms := make([]*model.Room, 0, len(s.rooms))
	for _, room := range s.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

func (s *RoomStore) UpdateRoomPlayers(roomID string, chips map[string]int64, status string) (*model.Room, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, ok := s.rooms[roomID]
	if !ok {
		return nil, false
	}

	for _, p := range room.Players {
		if value, ok := chips[p.ID]; ok {
			p.Chips = value
		}
	}
	if status != "" {
		room.Status = status
	}
	return room, true
}

// SetPlayerReady 设置玩家准备状态
func (s *RoomStore) SetPlayerReady(roomID string, playerID string, ready bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, ok := s.rooms[roomID]
	if !ok {
		log.Printf("[RoomStore] SetPlayerReady: room %s not found", roomID)
		return false
	}

	for _, p := range room.Players {
		if p.ID == playerID {
			log.Printf("[RoomStore] SetPlayerReady: player %s ready=%v (room %s)", playerID, ready, roomID)
			p.Ready = ready
			return true
		}
	}
	log.Printf("[RoomStore] SetPlayerReady: player %s not found in room %s", playerID, roomID)
	return false
}

// AllPlayersReady 检查是否所有玩家都准备了（至少2人）
func (s *RoomStore) AllPlayersReady(roomID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, ok := s.rooms[roomID]
	if !ok {
		log.Printf("[RoomStore] AllPlayersReady: room %s not found", roomID)
		return false
	}

	log.Printf("[RoomStore] AllPlayersReady: room %s has %d players", roomID, len(room.Players))
	for _, p := range room.Players {
		log.Printf("[RoomStore] AllPlayersReady:   player id=%s name=%s ready=%v", p.ID, p.Name, p.Ready)
		if !p.Ready {
			return false
		}
	}
	if len(room.Players) < 2 {
		log.Printf("[RoomStore] AllPlayersReady: room %s needs 2 players", roomID)
		return false
	}
	log.Printf("[RoomStore] AllPlayersReady: room %s all ready!", roomID)
	return true
}

// ResetAllPlayersReady 重置所有玩家准备状态
func (s *RoomStore) ResetAllPlayersReady(roomID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, ok := s.rooms[roomID]
	if !ok {
		return
	}

	for _, p := range room.Players {
		p.Ready = false
	}
}
