package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/dz-poker/server/internal/game"
	"github.com/dz-poker/server/internal/model"
	"github.com/dz-poker/server/internal/repository"
	"github.com/dz-poker/server/internal/service"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源，生产环境应限制
	},
}

// Hub 管理所有连接和房间
type Hub struct {
	mu          sync.RWMutex
	clients     map[string]*Client            // clientID -> Client
	rooms       map[string]map[string]*Client // roomID -> (clientID -> Client)
	roomStore   *repository.RoomStore
	games       map[string]*game.GameState // roomID -> GameState
	authService *service.AuthService
}

// Client WebSocket 客户端连接
type Client struct {
	ID            string
	PlayerID      string
	UserID        string // Authenticated user ID
	Nickname      string // User's nickname
	Avatar        string // User's avatar
	RoomID        string
	Conn          *websocket.Conn
	Send          chan []byte
	Hub           *Hub
	Authenticated bool
}

// NewHub 创建 Hub
func NewHub(roomStore *repository.RoomStore, authService *service.AuthService) *Hub {
	return &Hub{
		clients:     make(map[string]*Client),
		rooms:       make(map[string]map[string]*Client),
		roomStore:   roomStore,
		games:       make(map[string]*game.GameState),
		authService: authService,
	}
}

// Run 启动 Hub
func (h *Hub) Run() {
	// 清理过期房间的 goroutine
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		for range ticker.C {
			h.cleanupRooms()
		}
	}()
}

func (h *Hub) cleanupRooms() {
	// 清理空房间
}

// HandleWebSocket 处理 WebSocket 连接
func (h *Hub) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}

	clientID := uuid.New().String()
	client := &Client{
		ID:   clientID,
		Conn: conn,
		Send: make(chan []byte, 256),
		Hub:  h,
	}

	h.mu.Lock()
	h.clients[clientID] = client
	h.mu.Unlock()

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.Hub.removeClient(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(4096)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("websocket read error: %v", err)
			}
			break
		}

		c.Hub.handleMessage(c, message)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *Hub) removeClient(c *Client) {
	h.mu.Lock()
	if roomClients, ok := h.rooms[c.RoomID]; ok {
		delete(roomClients, c.ID)
		if len(roomClients) == 0 {
			delete(h.rooms, c.RoomID)
		}
	}
	h.mu.Unlock()

	// 清理玩家在 roomStore 中的状态
	if c.RoomID != "" && c.PlayerID != "" {
		room, dismissed := h.roomStore.RemovePlayer(c.RoomID, c.PlayerID)

		// 检查是否是房主离开
		if room != nil && room.HostID == c.PlayerID {
			// 房主离开，广播房主离开消息，解散房间
			h.mu.Lock()
			h.broadcastToRoom(c.RoomID, model.MsgHostLeft, map[string]interface{}{
				"message": "房主已离开，房间已解散",
			}, "")
			delete(h.rooms, c.RoomID)
			h.mu.Unlock()
			// 删除房间
			h.roomStore.DeleteRoom(c.RoomID)
		} else if dismissed {
			// 房间解散（没有其他玩家），广播给所有客户端
			h.mu.Lock()
			h.broadcastToRoom(c.RoomID, model.MsgRoomDismissed, nil, "")
			delete(h.rooms, c.RoomID)
			h.mu.Unlock()
		} else if room != nil {
			// 广播玩家离开，包含完整玩家列表
			players := make([]*model.PlayerInfo, len(room.Players))
			for i, p := range room.Players {
				players[i] = &model.PlayerInfo{
					ID:    p.ID,
					Name:  p.Name,
					Seat:  p.Seat,
					Chips: p.Chips,
					Ready: p.Ready,
				}
			}
			h.mu.Lock()
			h.broadcastToRoom(c.RoomID, model.MsgRoomUpdate, map[string]interface{}{
				"action":    "player_left",
				"player_id": c.PlayerID,
				"players":   players,
			}, "")
			h.mu.Unlock()
		}
	}

	h.mu.Lock()
	delete(h.clients, c.ID)
	h.mu.Unlock()
	close(c.Send)
}

func (h *Hub) handleMessage(c *Client, data []byte) {
	log.Printf("[WS] Received raw: %s", string(data))
	var msg model.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		c.sendError("invalid_message", "Invalid message format")
		return
	}

	log.Printf("[WS] Message parsed: type=%q, data=%v", msg.Type, msg.Data)

	switch msg.Type {
	case string(model.MsgWSLogin):
		log.Printf("[WS] Routing to HandleWSLogin")
		h.HandleWSLogin(c, msg.Data)
	case string(model.MsgCreateRoom):
		log.Printf("[WS] Routing to HandleCreateRoom")
		h.HandleCreateRoom(c, msg.Data)
	case string(model.MsgJoinRoom):
		log.Printf("[WS] Routing to HandleJoinRoom")
		h.HandleJoinRoom(c, msg.Data)
	case string(model.MsgLeaveRoom):
		log.Printf("[WS] Routing to HandleLeaveRoom")
		h.HandleLeaveRoom(c)
	case string(model.MsgStartGame):
		log.Printf("[WS] Routing to HandleStartGame")
		h.HandleStartGame(c, msg.Data)
	case string(model.MsgPlayerAct):
		log.Printf("[WS] Routing to HandlePlayerAction")
		h.HandlePlayerAction(c, msg.Data)
	case string(model.MsgReady):
		log.Printf("[WS] Routing to HandleReady, MsgReady=%s", model.MsgReady)
		h.HandleReady(c, msg.Data)
	case string(model.MsgCreateAIRoom):
		log.Printf("[WS] Routing to HandleCreateAIRoom")
		h.HandleCreateAIRoom(c, msg.Data)
	case string(model.MsgChat):
		h.HandleChat(c, msg.Data)
	default:
		log.Printf("[WS] Unknown message type: %s", msg.Type)
		c.sendError("unknown_type", "Unknown message type")
	}
}

// HandleWSLogin 处理 WebSocket 登录
func (h *Hub) HandleWSLogin(c *Client, data interface{}) {
	var req model.WSLoginReq
	if err := h.parseData(data, &req); err != nil {
		c.sendError("invalid_data", "Invalid login request")
		return
	}

	result, err := h.authService.ValidateToken(req.Token)
	if err != nil {
		c.sendError("auth_failed", "Invalid or expired token")
		return
	}

	c.Authenticated = true
	c.UserID = result.User.ID
	c.Nickname = result.User.Nickname
	c.Avatar = result.User.Avatar

	winRate := "0%"
	if result.Stats != nil && result.Stats.TotalGames > 0 {
		winRate = fmt.Sprintf("%.1f%%", float64(result.Stats.Wins)/float64(result.Stats.TotalGames)*100)
	}

	stats := &model.StatsResp{}
	if result.Stats != nil {
		stats = &model.StatsResp{
			TotalGames:  result.Stats.TotalGames,
			Wins:        result.Stats.Wins,
			WinRate:     winRate,
			TotalProfit: result.Stats.TotalProfit,
		}
	}

	resp := model.WSLoginResp{
		UserID:   result.User.ID,
		Username: result.User.Username,
		Nickname: result.User.Nickname,
		Avatar:   result.User.Avatar,
		Chips:    result.User.Chips,
		Stats:    stats,
	}
	c.sendMessage(model.MsgWSLoginResp, resp)
}

// HandleCreateRoom 处理创建房间（需要认证）
func (h *Hub) HandleCreateRoom(c *Client, data interface{}) {
	if !c.Authenticated {
		c.sendError("not_authenticated", "Please login first")
		return
	}

	var req model.CreateRoomReq
	if err := h.parseData(data, &req); err != nil {
		c.sendError("invalid_data", "Invalid create room request")
		return
	}

	playerID := c.UserID
	player := &model.Player{
		ID:        playerID,
		Name:      c.Nickname,
		Chips:     10000, // Will be updated from DB
		Connected: true,
	}

	// Get user's current chips from DB
	if user, err := repository.GetUserByID(c.UserID); err == nil && user != nil {
		player.Chips = user.Chips
	}

	room, err := h.roomStore.CreateRoom(player)
	if err != nil {
		log.Printf("[WS] CreateRoom failed: %v", err)
		c.sendError("create_failed", err.Error())
		return
	}

	log.Printf("[WS] Room created: code=%s, id=%s, host=%s", room.Code, room.ID, room.HostID)

	c.PlayerID = playerID
	c.RoomID = room.ID

	h.mu.Lock()
	if h.rooms[room.ID] == nil {
		h.rooms[room.ID] = make(map[string]*Client)
	}
	h.rooms[room.ID][c.ID] = c
	h.mu.Unlock()

	// 构建玩家列表
	players := make([]*model.PlayerInfo, len(room.Players))
	for i, p := range room.Players {
		players[i] = &model.PlayerInfo{
			ID:        p.ID,
			Name:      p.Name,
			Seat:      p.Seat,
			Chips:     p.Chips,
			Connected: p.Connected,
			Ready:     p.Ready,
		}
	}

	resp := model.CreateRoomResp{
		RoomCode:   room.Code,
		PlayerID:   playerID,
		PlayerName: player.Name,
		HostID:     room.HostID,
		Players:    players,
	}
	c.sendMessage(model.MsgRoomUpdate, resp)
}

// HandleJoinRoom 处理加入房间（需要认证）
func (h *Hub) HandleJoinRoom(c *Client, data interface{}) {
	if !c.Authenticated {
		c.sendError("not_authenticated", "Please login first")
		return
	}

	var req model.JoinRoomReq
	if err := h.parseData(data, &req); err != nil {
		c.sendError("invalid_data", "Invalid join room request")
		return
	}

	room, ok := h.roomStore.GetRoomByCode(req.RoomCode)
	if !ok {
		c.sendError("room_not_found", "Room not found")
		return
	}

	log.Printf("[JoinRoom] UserID=%s Nickname=%s joining room %s (host=%s)", c.UserID, c.Nickname, room.ID, room.HostID)

	if room.Status != "waiting" {
		c.sendError("room_not_waiting", "Game already started")
		return
	}

	playerID := c.UserID
	player := &model.Player{
		ID:        playerID,
		Name:      c.Nickname,
		Chips:     10000,
		Connected: true,
	}

	// Get user's current chips from DB
	if user, err := repository.GetUserByID(c.UserID); err == nil && user != nil {
		player.Chips = user.Chips
	}

	if err := h.roomStore.AddPlayer(room.ID, player); err != nil {
		c.sendError("join_failed", err.Error())
		return
	}

	log.Printf("[JoinRoom] Player %s (ID=%s) added to room %s, total players: %d", c.Nickname, playerID, room.ID, len(room.Players))

	c.PlayerID = playerID
	c.RoomID = room.ID

	h.mu.Lock()
	if h.rooms[room.ID] == nil {
		h.rooms[room.ID] = make(map[string]*Client)
	}
	h.rooms[room.ID][c.ID] = c
	h.mu.Unlock()

	// 返回房间信息给新加入玩家
	players := make([]*model.PlayerInfo, len(room.Players))
	for i, p := range room.Players {
		players[i] = &model.PlayerInfo{
			ID:        p.ID,
			Name:      p.Name,
			Seat:      p.Seat,
			Chips:     p.Chips,
			Connected: p.Connected,
			Ready:     p.Ready,
		}
	}

	resp := model.JoinRoomResp{
		RoomCode: room.Code,
		PlayerID: playerID,
		Players:  players,
		HostID:   room.HostID,
	}
	c.sendMessage(model.MsgRoomUpdate, resp)

	// 广播给房间内其他玩家 - 包含完整 players 列表
	h.broadcastToRoom(room.ID, model.MsgRoomUpdate, map[string]interface{}{
		"action":  "player_joined",
		"players": players,
	}, c.ID)
}

// HandleLeaveRoom 处理离开房间
func (h *Hub) HandleLeaveRoom(c *Client) {
	if c.RoomID == "" {
		return
	}

	h.mu.Lock()
	if roomClients, ok := h.rooms[c.RoomID]; ok {
		delete(roomClients, c.ID)
		if len(roomClients) == 0 {
			delete(h.rooms, c.RoomID)
		}
	}
	h.mu.Unlock()

	room, dismissed := h.roomStore.RemovePlayer(c.RoomID, c.PlayerID)

	if dismissed {
		h.broadcastToRoom(c.RoomID, model.MsgRoomDismissed, nil, "")
		h.mu.Lock()
		delete(h.rooms, c.RoomID)
		h.mu.Unlock()
	} else if room != nil {
		// 广播完整的 players 列表（包含 ready 状态）
		players := make([]*model.PlayerInfo, len(room.Players))
		for i, p := range room.Players {
			players[i] = &model.PlayerInfo{
				ID:    p.ID,
				Name:  p.Name,
				Seat:  p.Seat,
				Chips: p.Chips,
				Ready: p.Ready,
			}
		}
		h.broadcastToRoom(c.RoomID, model.MsgRoomUpdate, map[string]interface{}{
			"action":    "player_left",
			"player_id": c.PlayerID,
			"players":   players,
		}, "")
	}

	c.RoomID = ""
	c.PlayerID = ""
}

// HandleReady 处理玩家准备
func (h *Hub) HandleReady(c *Client, data interface{}) {
	log.Printf("[Ready] HandleReady called: PlayerID=%s, RoomID=%s, Nickname=%s", c.PlayerID, c.RoomID, c.Nickname)

	if c.RoomID == "" {
		log.Printf("[Ready] Error: not in room")
		c.sendError("not_in_room", "Not in a room")
		return
	}

	room, ok := h.roomStore.GetRoom(c.RoomID)
	if !ok {
		log.Printf("[Ready] Error: room not found")
		c.sendError("room_not_found", "Room not found")
		return
	}

	if room.Status != "waiting" {
		log.Printf("[Ready] Error: game already started, status=%s", room.Status)
		c.sendError("game_already_started", "Game already started")
		return
	}

	// 获取当前准备状态
	currentReady := false
	for _, p := range room.Players {
		if p.ID == c.PlayerID {
			currentReady = p.Ready
			break
		}
	}

	// 房主不能取消准备
	if room.HostID == c.PlayerID && currentReady {
		log.Printf("[Ready] Host cannot cancel ready")
		c.sendError("host_cannot_cancel", "Host cannot cancel ready")
		return
	}

	newReady := !currentReady
	log.Printf("[Ready] Toggling ready: %v -> %v", currentReady, newReady)

	if ok := h.roomStore.SetPlayerReady(c.RoomID, c.PlayerID, newReady); !ok {
		log.Printf("[Ready] Error: failed to set ready status")
		c.sendError("ready_failed", "Failed to set ready status")
		return
	}

	// 广播准备状态给房间内所有玩家
	h.broadcastToRoom(c.RoomID, model.MsgPlayerReady, map[string]interface{}{
		"player_id": c.PlayerID,
		"nickname":  c.Nickname,
		"ready":     newReady,
	}, "")
	log.Printf("[Ready] Broadcasted player_ready to room")
}

// HandleStartGame 处理开始游戏
func (h *Hub) HandleStartGame(c *Client, data interface{}) {
	log.Printf("[StartGame] PlayerID=%s, RoomID=%s", c.PlayerID, c.RoomID)

	if c.RoomID == "" {
		log.Printf("[StartGame] Error: not in room")
		c.sendError("not_in_room", "Not in a room")
		return
	}

	room, ok := h.roomStore.GetRoom(c.RoomID)
	if !ok {
		log.Printf("[StartGame] Error: room not found, roomID=%s", c.RoomID)
		c.sendError("room_not_found", "Room not found")
		return
	}

	log.Printf("[StartGame] Room status: host=%s, playerCount=%d", room.HostID, len(room.Players))

	if room.HostID != c.PlayerID {
		log.Printf("[StartGame] Error: not host, playerID=%s", c.PlayerID)
		c.sendError("not_host", "Only host can start game")
		return
	}

	if room.Status != "waiting" {
		log.Printf("[StartGame] Error: room not waiting, status=%s", room.Status)
		c.sendError("room_not_waiting", "Room not waiting")
		return
	}

	if len(room.Players) < 2 {
		log.Printf("[StartGame] Error: not enough players, count=%d", len(room.Players))
		c.sendError("not_enough_players", "Need at least 2 players")
		return
	}

	// 检查所有玩家是否都已准备
	log.Printf("[StartGame] Checking AllPlayersReady...")
	if !h.roomStore.AllPlayersReady(c.RoomID) {
		log.Printf("[StartGame] Not all players ready, rejecting start_game")
		players := make([]*model.PlayerInfo, len(room.Players))
		for i, p := range room.Players {
			players[i] = &model.PlayerInfo{
				ID:    p.ID,
				Name:  p.Name,
				Seat:  p.Seat,
				Chips: p.Chips,
				Ready: p.Ready,
			}
		}
		c.sendMessage(model.MsgRoomUpdate, map[string]interface{}{
			"action":  "sync_players",
			"players": players,
		})
		c.sendError("not_all_ready", "All players must be ready")
		return
	}
	log.Printf("[StartGame] All players ready, starting game!")

	// 重置所有玩家准备状态
	h.roomStore.ResetAllPlayersReady(c.RoomID)

	room.Status = "playing"

	// 记录每个玩家的初始筹码（扣除大小盲之前）
	initialChips := make(map[string]int64)
	for _, p := range room.Players {
		initialChips[p.ID] = p.Chips
	}

	// 创建游戏玩家
	gamePlayers := make([]*game.Player, len(room.Players))
	for i, p := range room.Players {
		gamePlayers[i] = &game.Player{
			ID:    p.ID,
			Name:  p.Name,
			Chips: p.Chips,
		}
	}

	// 初始化游戏
	gs := game.NewGameState(gamePlayers, 100)
	gs.InitialChips = initialChips

	h.games[c.RoomID] = gs

	// 广播游戏状态
	h.broadcastGameState(c.RoomID)

	// 检查当前玩家是否是AI
	h.triggerAIIfNeeded(c.RoomID)
}

func (h *Hub) HandlePlayerAction(c *Client, data interface{}) {
	if c.RoomID == "" {
		c.sendError("not_in_room", "Not in a room")
		return
	}

	gs, ok := h.games[c.RoomID]
	if !ok {
		c.sendError("game_not_started", "Game not started")
		return
	}

	var req model.PlayerAction
	if err := h.parseData(data, &req); err != nil {
		c.sendError("invalid_data", "Invalid action data")
		return
	}

	// 验证是否是当前玩家
	currentPlayer := gs.Players[gs.CurrentPlayer]
	if currentPlayer.ID != c.PlayerID {
		c.sendError("not_your_turn", "Not your turn")
		return
	}

	// 转换动作类型
	var actionType game.ActionType
	switch req.Action {
	case "check":
		actionType = game.CheckAction
	case "call":
		actionType = game.CallAction
	case "raise":
		actionType = game.RaiseAction
	case "fold":
		actionType = game.FoldAction
	case "allin":
		actionType = game.AllInAction
	default:
		c.sendError("invalid_action", "Unknown action type")
		return
	}

	action := &game.Action{
		Type:     actionType,
		PlayerID: c.PlayerID,
		Amount:   req.Amount,
	}

	// 处理动作
	ok = gs.ProcessAction(action)
	if !ok {
		c.sendError("invalid_action", "Cannot perform this action")
		return
	}

	// 广播更新后的游戏状态
	h.broadcastGameState(c.RoomID)

	// 检查游戏进度并触发AI行动
	h.checkGameProgress(c.RoomID)
}

func (h *Hub) broadcastGameState(roomID string) {
	gs, ok := h.games[roomID]
	if !ok {
		return
	}

	// 获取房间信息以识别AI玩家
	room, _ := h.roomStore.GetRoom(roomID)
	aiPlayerMap := make(map[string]*model.Player)
	if room != nil {
		for _, p := range room.Players {
			if p.IsAI {
				aiPlayerMap[p.ID] = p
			}
		}
	}

	players := make([]map[string]interface{}, len(gs.Players))
	for i, p := range gs.Players {
		cards := make([]string, len(p.HoleCards))
		for j, c := range p.HoleCards {
			cards[j] = c.String()
		}
		playerInfo := map[string]interface{}{
			"id":            p.ID,
			"name":          p.Name,
			"seat":          i,
			"chips":         p.Chips,
			"initial_chips": gs.InitialChips[p.ID],
			"bet":           p.Bet,
			"status":        p.Status.String(),
			"cards":         cards,
		}
		// 添加AI标识
		if aiPlayer, isAI := aiPlayerMap[p.ID]; isAI {
			playerInfo["is_ai"] = true
			playerInfo["ai_level"] = aiPlayer.AILevel
		} else {
			playerInfo["is_ai"] = false
		}
		players[i] = playerInfo
	}

	communityCards := make([]string, len(gs.CommunityCards))
	for i, c := range gs.CommunityCards {
		communityCards[i] = c.String()
	}

	toCall := int64(0)
	if len(gs.Players) > 0 {
		current := gs.Players[gs.CurrentPlayer]
		if current != nil && gs.CurrentBet > current.Bet {
			toCall = gs.CurrentBet - current.Bet
		}
	}

	state := map[string]interface{}{
		"phase":             gs.Phase.String(),
		"pot":               gs.Pot,
		"community_cards":   communityCards,
		"current_player":    gs.CurrentPlayer,
		"current_bet":       gs.CurrentBet,
		"min_raise":         gs.MinRaise,
		"small_blind":       gs.SmallBlind,
		"big_blind":         gs.BigBlind,
		"dealer_index":      gs.DealerIndex,
		"small_blind_index": gs.SmallBlindIndex,
		"big_blind_index":   gs.BigBlindIndex,
		"to_call":           toCall,
		"players":           players,
	}

	// 为当前轮到行动的玩家（或者所有玩家）计算当前最大牌型，这里为了简化，我们在服务端计算每个玩家当前的 BestHand 下发
	for i, p := range gs.Players {
		if p.Status != game.Folded && len(p.HoleCards) > 0 {
			res := game.EvaluateBestHand(p, gs.CommunityCards)
			if res != nil {
				highlightCards := game.BestHandHighlightCards(*res)
				bestCards := make([]string, len(highlightCards))
				for k, c := range highlightCards {
					bestCards[k] = c.String()
				}
				players[i]["best_hand_rank"] = res.Rank.String()
				players[i]["best_hand_cards"] = bestCards
			}
		}
	}

	// 添加操作历史
	actions := make([]map[string]interface{}, len(gs.Actions))
	for i, a := range gs.Actions {
		playerName := ""
		for _, p := range gs.Players {
			if p.ID == a.PlayerID {
				playerName = p.Name
				break
			}
		}
		actions[i] = map[string]interface{}{
			"player_name": playerName,
			"action":      a.Type.String(),
			"amount":      a.Amount,
			"phase":       gs.Phase.String(),
		}
	}
	state["actions"] = actions

	h.broadcastToRoom(roomID, model.MsgGameState, state, "")
}

func (h *Hub) settleGame(roomID string) {
	gs := h.games[roomID]
	if gs == nil {
		return
	}

	winners := game.DetermineWinners(gs)

	results := make([]map[string]interface{}, len(winners))
	for i, w := range winners {
		results[i] = map[string]interface{}{
			"player_id":   w.PlayerID,
			"player_name": w.PlayerName,
			"hand_rank":   w.HandRank.String(),
			"win_amount":  w.WinAmount,
		}
	}

	chips := make(map[string]int64, len(gs.Players))
	finalPlayers := make([]map[string]interface{}, len(gs.Players))
	for i, p := range gs.Players {
		chips[p.ID] = p.Chips
		finalPlayers[i] = map[string]interface{}{
			"id":    p.ID,
			"chips": p.Chips,
		}
	}
	h.roomStore.UpdateRoomPlayers(roomID, chips, "waiting")

	// 获取房间信息
	room, _ := h.roomStore.GetRoom(roomID)

	// 记录游戏结果
	for i, p := range gs.Players {
		initialChips := gs.InitialChips[p.ID]
		profit := p.Chips - initialChips
		rank := i + 1
		bestHand := ""

		// Find best hand for this player
		if res := game.EvaluateBestHand(p, gs.CommunityCards); res != nil {
			bestHand = res.Rank.String()
		}

		// Check if player won
		won := false
		for _, w := range winners {
			if w.PlayerID == p.ID {
				won = true
				break
			}
		}

		// Record to database
		repository.RecordGameResult(roomID, p.ID, initialChips, p.Chips, rank, bestHand)
		repository.UpdateUserStats(p.ID, profit, won)
		repository.UpdateUserChips(p.ID, p.Chips)
	}

	h.broadcastToRoom(roomID, model.MsgGameResult, map[string]interface{}{
		"winners":       results,
		"final_players": finalPlayers,
		"initial_chips": gs.InitialChips,
	}, "")

	// 重置所有玩家准备状态
	h.roomStore.ResetAllPlayersReady(roomID)

	// 调试日志：确认重置后的状态
	log.Printf("[settleGame] After ResetAllPlayersReady:")
	roomCheck, _ := h.roomStore.GetRoom(roomID)
	if roomCheck != nil {
		for _, p := range roomCheck.Players {
			log.Printf("  player %s ready=%v", p.ID, p.Ready)
		}
	}

	// 获取更新后的房间信息用于广播
	if room != nil {
		room.Status = "waiting"
		for _, p := range room.Players {
			// 房主和AI玩家自动准备
			if p.ID == room.HostID || p.IsAI {
				p.Ready = true
			}
		}
		players := make([]*model.PlayerInfo, len(room.Players))
		for i, p := range room.Players {
			players[i] = &model.PlayerInfo{
				ID:      p.ID,
				Name:    p.Name,
				Seat:    p.Seat,
				Chips:   p.Chips,
				Ready:   p.Ready,
				IsAI:    p.IsAI,
				AILevel: p.AILevel,
			}
		}
		log.Printf("[settleGame] Broadcasting room_update with players ready states:")
		for _, p := range players {
			log.Printf("  player %s ready=%v", p.ID, p.Ready)
		}
		// 广播房间更新，通知所有客户端游戏结束、状态重置
		h.broadcastToRoom(roomID, model.MsgRoomUpdate, map[string]interface{}{
			"room_code": room.Code,
			"action":    "game_ended",
			"players":   players,
		}, "")
	}

	// 清理游戏状态
	delete(h.games, roomID)
}

func (h *Hub) parseData(data interface{}, v interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, v)
}

func (c *Client) sendMessage(msgType model.MessageType, data interface{}) {
	msg := model.Message{
		Type: string(msgType),
		Data: data,
	}
	jsonData, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[WS] sendMessage marshal error: %v", err)
		return
	}

	select {
	case c.Send <- jsonData:
		log.Printf("[WS] sendMessage: type=%s to client=%s", msgType, c.ID)
	default:
		log.Printf("[WS] client %s send buffer full, dropping message %s", c.ID, msgType)
	}
}

func (c *Client) sendError(code, message string) {
	c.sendMessage(model.MsgError, model.ErrorResp{
		Code:    code,
		Message: message,
	})
}

func (h *Hub) broadcastToRoom(roomID string, msgType model.MessageType, data interface{}, excludeID string) {
	msg := model.Message{
		Type: string(msgType),
		Data: data,
	}
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.mu.RLock()
	clients := h.rooms[roomID]
	h.mu.RUnlock()

	for id, client := range clients {
		if id != excludeID {
			select {
			case client.Send <- jsonData:
			default:
				log.Printf("client %s send buffer full", id)
			}
		}
	}
}

// NotifyUserChipsUpdate 通知用户筹码更新
func (h *Hub) NotifyUserChipsUpdate(userID string, chips int64) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	msg := model.Message{
		Type: string(model.MsgChipsUpdate),
		Data: map[string]interface{}{
			"user_id": userID,
			"chips":   chips,
		},
	}
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return
	}

	// 遍历所有客户端，找到对应用户的连接
	for _, client := range h.clients {
		if client.UserID == userID {
			select {
			case client.Send <- jsonData:
				log.Printf("[WS] Notified user %s about chips update: %d", userID, chips)
			default:
				log.Printf("client %s send buffer full", client.ID)
			}
		}
	}
}

// AI相关常量
const (
	aiPlayerCount = 5 // AI房间中AI玩家数量
	aiNamePrefix  = "AI "
)

// AI玩家名字池
var aiNames = []string{
	"Alice", "Bob", "Charlie", "Diana", "Eve",
	"Frank", "Grace", "Henry", "Ivy", "Jack",
}

// HandleCreateAIRoom 处理创建AI房间（和普通房间一样，只是添加AI玩家）
func (h *Hub) HandleCreateAIRoom(c *Client, data interface{}) {
	if !c.Authenticated {
		c.sendError("not_authenticated", "Please login first")
		return
	}

	var req model.CreateAIRoomReq
	if err := h.parseData(data, &req); err != nil {
		c.sendError("invalid_data", "Invalid create AI room request")
		return
	}

	// 默认AI等级
	aiLevel := req.AILevel
	if aiLevel == "" {
		aiLevel = string(game.AILevelNormal)
	}

	// 创建玩家 - 使用真实筹码
	playerID := c.UserID
	player := &model.Player{
		ID:        playerID,
		Name:      c.Nickname,
		Chips:     10000, // Will be updated from DB
		Connected: true,
		Ready:     false,
		IsAI:      false,
	}

	// 获取用户真实筹码
	if user, err := repository.GetUserByID(c.UserID); err == nil && user != nil {
		player.Chips = user.Chips
	}

	// 创建房间
	room, err := h.roomStore.CreateRoom(player)
	if err != nil {
		log.Printf("[WS] CreateAI Room failed: %v", err)
		c.sendError("create_failed", err.Error())
		return
	}

	log.Printf("[WS] AI Room created: code=%s, id=%s, host=%s", room.Code, room.ID, room.HostID)

	c.PlayerID = playerID
	c.RoomID = room.ID

	h.mu.Lock()
	if h.rooms[room.ID] == nil {
		h.rooms[room.ID] = make(map[string]*Client)
	}
	h.rooms[room.ID][c.ID] = c
	h.mu.Unlock()

	// 添加AI玩家
	usedNames := make(map[string]bool)
	usedNames[c.Nickname] = true

	for i := 0; i < aiPlayerCount; i++ {
		// 选择一个未使用的AI名字
		aiName := ""
		for _, name := range aiNames {
			if !usedNames[name] {
				aiName = name
				usedNames[name] = true
				break
			}
		}
		if aiName == "" {
			aiName = fmt.Sprintf("Bot%d", i+1)
		}

		aiPlayer := &model.Player{
			ID:        fmt.Sprintf("ai_%s_%d", room.ID[:8], i),
			Name:      aiName,
			Chips:     player.Chips, // AI使用相同的初始筹码
			Connected: true,
			Ready:     true, // AI自动准备
			IsAI:      true,
			AILevel:   aiLevel,
		}

		if err := h.roomStore.AddPlayer(room.ID, aiPlayer); err != nil {
			log.Printf("[WS] Failed to add AI player: %v", err)
			continue
		}
		log.Printf("[WS] Added AI player %s to room %s", aiPlayer.Name, room.ID)
	}

	// 重新获取房间信息
	room, _ = h.roomStore.GetRoom(room.ID)

	// 构建玩家列表
	players := make([]*model.PlayerInfo, len(room.Players))
	for i, p := range room.Players {
		players[i] = &model.PlayerInfo{
			ID:        p.ID,
			Name:      p.Name,
			Seat:      p.Seat,
			Chips:     p.Chips,
			Connected: p.Connected,
			Ready:     p.Ready,
			IsAI:      p.IsAI,
			AILevel:   p.AILevel,
		}
	}

	resp := model.CreateRoomResp{
		RoomCode:   room.Code,
		PlayerID:   playerID,
		PlayerName: player.Name,
		HostID:     room.HostID,
		Players:    players,
	}
	c.sendMessage(model.MsgRoomUpdate, resp)
}

// handleAIAction 处理AI玩家的行动
func (h *Hub) handleAIAction(roomID string) {
	gs, ok := h.games[roomID]
	if !ok {
		return
	}

	// 检查当前玩家是否是AI
	currentPlayer := gs.Players[gs.CurrentPlayer]
	if currentPlayer == nil {
		return
	}

	// 获取AI等级
	room, ok := h.roomStore.GetRoom(roomID)
	if !ok {
		return
	}

	var aiLevel game.AILevel = game.AILevelNormal
	for _, p := range room.Players {
		if p.ID == currentPlayer.ID && p.IsAI {
			aiLevel = game.AILevel(p.AILevel)
			if aiLevel == "" {
				aiLevel = game.AILevelNormal
			}
			break
		}
	}

	// AI思考延迟
	thinkDelay := game.GetAIThinkDelay(aiLevel)

	go func() {
		time.Sleep(thinkDelay)

		// 重新检查游戏状态
		h.mu.Lock()
		gs, ok := h.games[roomID]
		if !ok {
			h.mu.Unlock()
			return
		}

		// 再次检查当前玩家是否是AI（状态可能已改变）
		currentPlayer := gs.Players[gs.CurrentPlayer]
		if currentPlayer == nil {
			h.mu.Unlock()
			return
		}

		// 检查是否是AI
		isAI := false
		for _, p := range room.Players {
			if p.ID == currentPlayer.ID && p.IsAI {
				isAI = true
				break
			}
		}
		if !isAI {
			h.mu.Unlock()
			return
		}

		// AI决策
		decision := game.DecideAction(gs, gs.CurrentPlayer, aiLevel)

		action := &game.Action{
			Type:     decision.Action,
			PlayerID: currentPlayer.ID,
			Amount:   decision.Amount,
		}

		log.Printf("[AI] Player %s decides: %s amount=%d", currentPlayer.Name, decision.Action.String(), decision.Amount)

		// 处理动作
		ok = gs.ProcessAction(action)
		if !ok {
			log.Printf("[AI] Action failed for player %s", currentPlayer.Name)
			h.mu.Unlock()
			return
		}
		h.mu.Unlock()

		// 广播更新后的游戏状态
		h.broadcastGameState(roomID)

		// 检查是否需要进入下一阶段
		h.checkGameProgress(roomID)
	}()
}

// checkGameProgress 检查游戏进度
func (h *Hub) checkGameProgress(roomID string) {
	gs, ok := h.games[roomID]
	if !ok {
		return
	}

	// 检查是否只剩一个玩家
	if gs.IsHandFinished() {
		h.settleGame(roomID)
		return
	}

	// 检查是否需要进入下一阶段
	if gs.IsRoundFinished() {
		if gs.Phase == game.Showdown {
			h.settleGame(roomID)
			return
		}
		gs.NextPhase()
		for gs.Phase != game.Showdown && gs.ActiveOnlyCount() == 0 {
			gs.NextPhase()
		}
		h.broadcastGameState(roomID)
		if gs.Phase == game.Showdown {
			h.settleGame(roomID)
			return
		}
	}

	// 检查下一个玩家是否是AI
	h.triggerAIIfNeeded(roomID)
}

// triggerAIIfNeeded 如果当前玩家是AI则触发AI行动
func (h *Hub) triggerAIIfNeeded(roomID string) {
	gs, ok := h.games[roomID]
	if !ok {
		return
	}

	currentPlayer := gs.Players[gs.CurrentPlayer]
	if currentPlayer == nil {
		return
	}

	// 检查是否是AI
	room, ok := h.roomStore.GetRoom(roomID)
	if !ok {
		return
	}

	for _, p := range room.Players {
		if p.ID == currentPlayer.ID && p.IsAI && currentPlayer.Status == game.Active {
			log.Printf("[AI] Triggering AI action for player %s", currentPlayer.Name)
			go h.handleAIAction(roomID)
			return
		}
	}
}

// HandleChat 处理聊天消息
func (h *Hub) HandleChat(c *Client, data interface{}) {
	// 验证用户已认证
	if !c.Authenticated {
		c.sendError("not_authenticated", "Please login first")
		return
	}

	// 验证用户在房间内
	if c.RoomID == "" {
		c.sendError("not_in_room", "Not in a room")
		return
	}

	// 解析消息内容
	var req model.ChatMessage
	if err := h.parseData(data, &req); err != nil {
		c.sendError("invalid_data", "Invalid chat message")
		return
	}

	// 内容长度限制
	content := req.Content
	if len(content) == 0 {
		c.sendError("empty_message", "Message cannot be empty")
		return
	}
	if len(content) > 100 {
		content = content[:100]
	}

	// 广播聊天消息到房间所有玩家
	broadcast := model.ChatBroadcast{
		PlayerID:   c.PlayerID,
		PlayerName: c.Nickname,
		Avatar:     c.Avatar,
		Content:    content,
	}
	h.broadcastToRoom(c.RoomID, model.MsgChat, broadcast, "")
}
