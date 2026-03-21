package model

// MessageType 消息类型
type MessageType string

const (
	// 客户端 -> 服务端
	MsgCreateRoom MessageType = "create_room"
	MsgJoinRoom   MessageType = "join_room"
	MsgLeaveRoom  MessageType = "leave_room"
	MsgStartGame  MessageType = "start_game"
	MsgPlayerAct  MessageType = "player_action"

	// 服务端 -> 客户端
	MsgRoomUpdate    MessageType = "room_update"
	MsgGameState     MessageType = "game_state"
	MsgGameResult    MessageType = "game_result"
	MsgRoomDismissed MessageType = "room_dismissed"
	MsgError         MessageType = "error"
)

// Message 通用消息结构
type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
	Seq  int64       `json:"seq,omitempty"`
}

// CreateRoomReq 创建房间请求
type CreateRoomReq struct {
	PlayerName string `json:"player_name" validate:"required,min=1,max=20"`
}

// CreateRoomResp 创建房间响应
type CreateRoomResp struct {
	RoomCode   string        `json:"room_code"`
	PlayerID   string        `json:"player_id"`
	PlayerName string        `json:"player_name"`
	HostID     string        `json:"host_id"`
	Players    []*PlayerInfo `json:"players"`
}

// JoinRoomReq 加入房间请求
type JoinRoomReq struct {
	RoomCode   string `json:"room_code" validate:"required,len=6"`
	PlayerName string `json:"player_name" validate:"required,min=1,max=20"`
}

// JoinRoomResp 加入房间响应
type JoinRoomResp struct {
	RoomCode string        `json:"room_code"`
	PlayerID string        `json:"player_id"`
	Players  []*PlayerInfo `json:"players"`
	HostID   string        `json:"host_id"`
}

// PlayerInfo 玩家信息
type PlayerInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Seat      int    `json:"seat"`
	Chips     int64  `json:"chips"`
	Connected bool   `json:"connected"`
}

// StartGameReq 开始游戏请求
type StartGameReq struct {
	RoomCode string `json:"room_code"`
}

// PlayerAction 玩家动作
type PlayerAction struct {
	Action string `json:"action"`           // check/call/raise/fold/allin
	Amount int64  `json:"amount,omitempty"` // 加注金额
}

// ErrorResp 错误响应
type ErrorResp struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
