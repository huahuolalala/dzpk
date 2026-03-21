package model

import "time"

// User represents a user in the system
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Nickname     string    `json:"nickname"`
	Avatar       string    `json:"avatar"`
	Chips        int64     `json:"chips"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UserStats represents user game statistics
type UserStats struct {
	UserID      string    `json:"user_id"`
	TotalGames  int64     `json:"total_games"`
	Wins        int64     `json:"wins"`
	TotalProfit int64     `json:"total_profit"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GameHistory represents a single game record
type GameHistory struct {
	ID           string    `json:"id"`
	RoomCode     string    `json:"room_code"`
	UserID       string    `json:"user_id"`
	InitialChips int64     `json:"initial_chips"`
	FinalChips   int64     `json:"final_chips"`
	Profit       int64     `json:"profit"`
	FinalRank    int       `json:"final_rank"`
	BestHand     string    `json:"best_hand"`
	CreatedAt    time.Time `json:"created_at"`
}

// Avatar options (9 animals)
var Avatars = []string{
	"🐱", // cat
	"🐶", // dog
	"🐰", // rabbit
	"🦊", // fox
	"🐼", // panda
	"🦁", // lion
	"🐸", // frog
	"🐵", // monkey
	"🐷", // pig
}

// RegisterReq registration request
type RegisterReq struct {
	Username string `json:"username" validate:"required,min=3,max=20"`
	Password string `json:"password" validate:"required,min=6,max=20"`
	Nickname string `json:"nickname" validate:"required,min=1,max=20"`
	Avatar   string `json:"avatar" validate:"required"`
}

// RegisterResp registration response
type RegisterResp struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Chips    int64  `json:"chips"`
	Token    string `json:"token"`
}

// LoginReq login request
type LoginReq struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResp login response
type LoginResp struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Chips    int64  `json:"chips"`
	Token    string `json:"token"`
}

// UserInfoResp user info response
type UserInfoResp struct {
	UserID   string     `json:"user_id"`
	Username string     `json:"username"`
	Nickname string     `json:"nickname"`
	Avatar   string     `json:"avatar"`
	Chips    int64      `json:"chips"`
	Stats    *StatsResp `json:"stats,omitempty"`
}

// StatsResp user statistics response
type StatsResp struct {
	TotalGames  int64  `json:"total_games"`
	Wins        int64  `json:"wins"`
	WinRate     string `json:"win_rate"`
	TotalProfit int64  `json:"total_profit"`
}

// GameHistoryResp game history response
type GameHistoryResp struct {
	Total int           `json:"total"`
	Games []GameHistory `json:"games"`
}

// WSLoginReq WebSocket login request (after initial connection)
type WSLoginReq struct {
	Token string `json:"token"`
}

// WSLoginResp WebSocket login response
type WSLoginResp struct {
	UserID   string     `json:"user_id"`
	Username string     `json:"username"`
	Nickname string     `json:"nickname"`
	Avatar   string     `json:"avatar"`
	Chips    int64      `json:"chips"`
	Stats    *StatsResp `json:"stats,omitempty"`
}

// AuthResult authentication result
type AuthResult struct {
	User  *User
	Stats *UserStats
}

// AdminUserResp admin user info response
type AdminUserResp struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Chips    int64  `json:"chips"`
}

// UpdateChipsReq update chips request
type UpdateChipsReq struct {
	UserID string `json:"user_id" binding:"required"`
	Chips  int64  `json:"chips" binding:"required,min=0"`
}

// UpdateChipsResp update chips response
type UpdateChipsResp struct {
	UserID string `json:"user_id"`
	Chips  int64  `json:"chips"`
}
