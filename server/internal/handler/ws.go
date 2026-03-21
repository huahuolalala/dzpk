package handler

import (
	"encoding/json"
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
	mu      sync.RWMutex
	clients map[string]*Client            // clientID -> Client
	rooms   map[string]map[string]*Client // roomID -> (clientID -> Client)

	roomStore *repository.RoomStore
	games     map[string]*game.GameState // roomID -> GameState
}

// Client WebSocket 客户端连接
type Client struct {
	ID       string
	PlayerID string
	RoomID   string
	Conn     *websocket.Conn
	Send     chan []byte
	Hub      *Hub
}

// NewHub 创建 Hub
func NewHub(roomStore *repository.RoomStore) *Hub {
	return &Hub{
		clients:   make(map[string]*Client),
		rooms:     make(map[string]map[string]*Client),
		roomStore: roomStore,
		games:     make(map[string]*game.GameState),
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
	defer h.mu.Unlock()

	if c.RoomID != "" {
		if roomClients, ok := h.rooms[c.RoomID]; ok {
			delete(roomClients, c.ID)
			if len(roomClients) == 0 {
				delete(h.rooms, c.RoomID)
			}
		}
	}

	delete(h.clients, c.ID)
	close(c.Send)
}

func (h *Hub) handleMessage(c *Client, data []byte) {
	log.Printf("[WS] Received message: %s", string(data))
	var msg model.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		c.sendError("invalid_message", "Invalid message format")
		return
	}

	log.Printf("[WS] Message type: %s", msg.Type)

	switch msg.Type {
	case string(model.MsgCreateRoom):
		h.HandleCreateRoom(c, msg.Data)
	case string(model.MsgJoinRoom):
		h.HandleJoinRoom(c, msg.Data)
	case string(model.MsgLeaveRoom):
		h.HandleLeaveRoom(c)
	case string(model.MsgStartGame):
		h.HandleStartGame(c, msg.Data)
	case string(model.MsgPlayerAct):
		h.HandlePlayerAction(c, msg.Data)
	default:
		c.sendError("unknown_type", "Unknown message type")
	}
}

// HandleCreateRoom 处理创建房间
func (h *Hub) HandleCreateRoom(c *Client, data interface{}) {
	var req model.CreateRoomReq
	if err := h.parseData(data, &req); err != nil {
		c.sendError("invalid_data", "Invalid create room request")
		return
	}

	playerID := uuid.New().String()
	player := &model.Player{
		ID:        playerID,
		Name:      req.PlayerName,
		Chips:     10000,
		Connected: true,
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

// HandleJoinRoom 处理加入房间
func (h *Hub) HandleJoinRoom(c *Client, data interface{}) {
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

	if room.Status != "waiting" {
		c.sendError("room_not_waiting", "Game already started")
		return
	}

	playerID := uuid.New().String()
	player := &model.Player{
		ID:        playerID,
		Name:      req.PlayerName,
		Chips:     10000,
		Connected: true,
	}

	if err := h.roomStore.AddPlayer(room.ID, player); err != nil {
		c.sendError("join_failed", err.Error())
		return
	}

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
		}
	}

	resp := model.JoinRoomResp{
		RoomCode: room.Code,
		PlayerID: playerID,
		Players:  players,
		HostID:   room.HostID,
	}
	c.sendMessage(model.MsgRoomUpdate, resp)

	// 广播给房间内其他玩家
	h.broadcastToRoom(room.ID, model.MsgRoomUpdate, map[string]interface{}{
		"action": "player_joined",
		"player": players[len(players)-1],
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
		h.broadcastToRoom(c.RoomID, model.MsgRoomUpdate, map[string]interface{}{
			"action":    "player_left",
			"player_id": c.PlayerID,
		}, "")
	}

	c.RoomID = ""
	c.PlayerID = ""
}

// HandleStartGame 处理开始游戏
func (h *Hub) HandleStartGame(c *Client, data interface{}) {
	if c.RoomID == "" {
		c.sendError("not_in_room", "Not in a room")
		return
	}

	room, ok := h.roomStore.GetRoom(c.RoomID)
	if !ok {
		c.sendError("room_not_found", "Room not found")
		return
	}

	if room.HostID != c.PlayerID {
		c.sendError("not_host", "Only host can start game")
		return
	}

	if len(room.Players) < 2 {
		c.sendError("not_enough_players", "Need at least 2 players")
		return
	}

	room.Status = "playing"

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
	h.games[c.RoomID] = gs

	// 广播游戏状态
	h.broadcastGameState(c.RoomID)
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

	// 检查是否需要进入下一阶段
	if gs.IsHandFinished() {
		h.settleGame(c.RoomID)
		return
	}

	if gs.IsRoundFinished() {
		if gs.Phase == game.Showdown {
			h.settleGame(c.RoomID)
			return
		}
		gs.NextPhase()
		for gs.Phase != game.Showdown && gs.ActiveOnlyCount() == 0 {
			gs.NextPhase()
		}
		h.broadcastGameState(c.RoomID)
		if gs.Phase == game.Showdown {
			h.settleGame(c.RoomID)
		}
	}
}

func (h *Hub) broadcastGameState(roomID string) {
	gs, ok := h.games[roomID]
	if !ok {
		return
	}

	players := make([]map[string]interface{}, len(gs.Players))
	for i, p := range gs.Players {
		cards := make([]string, len(p.HoleCards))
		for j, c := range p.HoleCards {
			cards[j] = c.String()
		}
		players[i] = map[string]interface{}{
			"id":     p.ID,
			"name":   p.Name,
			"seat":   i,
			"chips":  p.Chips,
			"bet":    p.Bet,
			"status": p.Status.String(),
			"cards":  cards,
		}
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
				bestCards := make([]string, len(res.AllCards))
				for k, c := range res.AllCards {
					bestCards[k] = c.String()
				}
				players[i]["best_hand_rank"] = res.Rank.String()
				players[i]["best_hand_cards"] = bestCards
			}
		}
	}

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

	h.broadcastToRoom(roomID, model.MsgGameResult, map[string]interface{}{
		"winners":       results,
		"final_players": finalPlayers,
	}, "")

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
		return
	}

	select {
	case c.Send <- jsonData:
	default:
		log.Printf("client %s send buffer full", c.ID)
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
