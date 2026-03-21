package repository

import (
	"testing"

	"github.com/dz-poker/server/internal/model"
)

func TestRoomStore_Create(t *testing.T) {
	store := NewRoomStore()
	host := &model.Player{ID: "h1", Name: "Host"}

	room, err := store.CreateRoom(host)
	if err != nil {
		t.Fatalf("CreateRoom failed: %v", err)
	}

	if room.Code == "" {
		t.Error("Room code should not be empty")
	}
	if len(room.Code) != 6 {
		t.Errorf("Room code should be 6 digits, got %d", len(room.Code))
	}
	if room.HostID != "h1" {
		t.Errorf("HostID should be 'h1', got %s", room.HostID)
	}
	if room.Status != "waiting" {
		t.Errorf("Status should be 'waiting', got %s", room.Status)
	}
	if len(room.Players) != 1 {
		t.Errorf("Should have 1 player, got %d", len(room.Players))
	}
}

func TestRoomStore_Get(t *testing.T) {
	store := NewRoomStore()
	host := &model.Player{ID: "h1", Name: "Host"}

	room, _ := store.CreateRoom(host)

	found, ok := store.GetRoom(room.ID)
	if !ok {
		t.Error("GetRoom should find created room")
	}
	if found.ID != room.ID {
		t.Errorf("Got wrong room: %s != %s", found.ID, room.ID)
	}
}

func TestRoomStore_GetRoomByCode(t *testing.T) {
	store := NewRoomStore()
	host := &model.Player{ID: "h1", Name: "Host"}

	room, _ := store.CreateRoom(host)

	found, ok := store.GetRoomByCode(room.Code)
	if !ok {
		t.Error("GetRoomByCode should find created room")
	}
	if found.ID != room.ID {
		t.Errorf("Got wrong room: %s != %s", found.ID, room.ID)
	}
}

func TestRoomStore_AddPlayer(t *testing.T) {
	store := NewRoomStore()
	host := &model.Player{ID: "h1", Name: "Host"}
	player := &model.Player{ID: "p1", Name: "Player"}

	room, _ := store.CreateRoom(host)
	err := store.AddPlayer(room.ID, player)
	if err != nil {
		t.Fatalf("AddPlayer failed: %v", err)
	}

	if len(room.Players) != 2 {
		t.Errorf("Should have 2 players, got %d", len(room.Players))
	}
	if player.Seat != 1 {
		t.Errorf("Player seat should be 1, got %d", player.Seat)
	}
}

func TestRoomStore_AddPlayer_MaxSeats(t *testing.T) {
	store := NewRoomStore()
	host := &model.Player{ID: "h1", Name: "Host"}
	room, _ := store.CreateRoom(host)

	// 尝试加入 8 个玩家 (room.MaxSeats=9, 已有1人)
	for i := 0; i < 8; i++ {
		p := &model.Player{ID: string(rune('0' + i)), Name: string(rune('0' + i))}
		if err := store.AddPlayer(room.ID, p); err != nil {
			t.Fatalf("AddPlayer failed at %d: %v", i, err)
		}
	}

	// 第9个玩家应该失败
	p := &model.Player{ID: "overflow", Name: "Overflow"}
	err := store.AddPlayer(room.ID, p)
	if err == nil {
		t.Error("Should fail when room is full")
	}
}

func TestRoomStore_RemovePlayer(t *testing.T) {
	store := NewRoomStore()
	host := &model.Player{ID: "h1", Name: "Host"}
	player := &model.Player{ID: "p1", Name: "Player"}

	room, _ := store.CreateRoom(host)
	store.AddPlayer(room.ID, player)

	updatedRoom, dismissed := store.RemovePlayer(room.ID, "p1")
	if dismissed {
		t.Error("Room should not be dismissed when non-host leaves")
	}
	if len(updatedRoom.Players) != 1 {
		t.Errorf("Should have 1 player left, got %d", len(updatedRoom.Players))
	}
}

func TestRoomStore_RemovePlayer_HostLeaves(t *testing.T) {
	store := NewRoomStore()
	host := &model.Player{ID: "h1", Name: "Host"}
	player := &model.Player{ID: "p1", Name: "Player"}

	room, _ := store.CreateRoom(host)
	store.AddPlayer(room.ID, player)

	_, dismissed := store.RemovePlayer(room.ID, "h1")
	if !dismissed {
		t.Error("Room should be dismissed when host leaves")
	}
}

func TestRoomStore_Delete(t *testing.T) {
	store := NewRoomStore()
	host := &model.Player{ID: "h1", Name: "Host"}
	room, _ := store.CreateRoom(host)

	store.DeleteRoom(room.ID)

	_, ok := store.GetRoom(room.ID)
	if ok {
		t.Error("Room should be deleted")
	}
}

func TestRoomStore_MaxSeats(t *testing.T) {
	store := NewRoomStore()
	host := &model.Player{ID: "h1", Name: "Host"}
	room, _ := store.CreateRoom(host)

	if room.MaxSeats != 9 {
		t.Errorf("MaxSeats should be 9, got %d", room.MaxSeats)
	}
}
