package model

// MessageType 消息类型
type MessageType string

const (
	// 客户端 -> 服务端
	MsgCreateRoom   MessageType = "create_room"
	MsgJoinRoom     MessageType = "join_room"
	MsgLeaveRoom    MessageType = "leave_room"
	MsgStartGame    MessageType = "start_game"
	MsgPlayerAct    MessageType = "player_action"
	MsgWSLogin      MessageType = "ws_login"
	MsgReady        MessageType = "ready"
	MsgCreateAIRoom MessageType = "create_ai_room" // 创建AI房间

	// 服务端 -> 客户端
	MsgRoomUpdate    MessageType = "room_update"
	MsgGameState     MessageType = "game_state"
	MsgGameResult    MessageType = "game_result"
	MsgRoomDismissed MessageType = "room_dismissed"
	MsgHostLeft      MessageType = "host_left"    // 房主离开
	MsgChipsUpdate   MessageType = "chips_update" // 筹码更新
	MsgError         MessageType = "error"
	MsgWSLoginResp   MessageType = "ws_login_resp"
	MsgAuthRequired  MessageType = "auth_required"
	MsgPlayerReady   MessageType = "player_ready"
	MsgChat          MessageType = "chat" // 聊天消息
)

// Message 通用消息结构
type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
	Seq  int64       `json:"seq,omitempty"`
}

// CreateRoomReq 创建房间请求（已认证用户不需要player_name）
type CreateRoomReq struct {
	RoomCode string `json:"room_code,omitempty"`
}

// CreateAIRoomReq 创建AI房间请求
type CreateAIRoomReq struct {
	AILevel string `json:"ai_level,omitempty"` // easy/normal/hard，默认normal
}

// CreateRoomResp 创建房间响应
type CreateRoomResp struct {
	RoomCode   string        `json:"room_code"`
	PlayerID   string        `json:"player_id"`
	PlayerName string        `json:"player_name"`
	HostID     string        `json:"host_id"`
	Players    []*PlayerInfo `json:"players"`
}

// JoinRoomReq 加入房间请求（已认证用户只需要房间码）
type JoinRoomReq struct {
	RoomCode string `json:"room_code" validate:"required,len=6"`
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
	Ready     bool   `json:"ready"`
	// AI 相关字段
	IsAI    bool   `json:"is_ai"`
	AILevel string `json:"ai_level"`
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

// ChatMessage 聊天消息请求
type ChatMessage struct {
	Content string `json:"content"`
}

// ChatBroadcast 广播的聊天消息
type ChatBroadcast struct {
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
	Avatar     string `json:"avatar"`
	Content    string `json:"content"`
}
