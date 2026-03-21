package game

import "github.com/google/uuid"

// Phase 游戏阶段
type Phase int

const (
	PreFlop Phase = iota
	Flop
	Turn
	River
	Showdown
)

func (p Phase) String() string {
	switch p {
	case PreFlop:
		return "preflop"
	case Flop:
		return "flop"
	case Turn:
		return "turn"
	case River:
		return "river"
	default:
		return "showdown"
	}
}

// PlayerStatus 玩家状态
type PlayerStatus int

const (
	Active PlayerStatus = iota
	Folded
	AllIn
)

func (s PlayerStatus) String() string {
	switch s {
	case Active:
		return "active"
	case Folded:
		return "fold"
	default:
		return "all-in"
	}
}

// ActionType 动作类型
type ActionType int

const (
	CheckAction ActionType = iota
	CallAction
	RaiseAction
	FoldAction
	AllInAction
	SmallBlindAction
	BigBlindAction
)

func (a ActionType) String() string {
	switch a {
	case CheckAction:
		return "check"
	case CallAction:
		return "call"
	case RaiseAction:
		return "raise"
	case FoldAction:
		return "fold"
	case SmallBlindAction:
		return "small_blind"
	case BigBlindAction:
		return "big_blind"
	default:
		return "allin"
	}
}

// Action 玩家动作
type Action struct {
	Type     ActionType
	PlayerID string
	Amount   int64
}

// Player 游戏玩家
type Player struct {
	ID        string
	Name      string
	Chips     int64
	Status    PlayerStatus
	HoleCards []Card // 手牌
	Bet       int64  // 本轮下注
	TotalBet  int64  // 累计投入
}

// GameState 游戏状态
type GameState struct {
	ID                 string
	Phase              Phase
	Players            []*Player
	CommunityCards     []Card
	CurrentPlayer      int
	CurrentBet         int64
	MinRaise           int64
	Pot                int64
	SmallBlind         int64
	BigBlind           int64
	DealerIndex        int
	SmallBlindIndex    int
	BigBlindIndex      int
	LastAggressorIndex int
	ActedThisRound     []bool
	Deck               *Deck
	Actions            []Action
	InitialChips       map[string]int64 // 玩家初始筹码（游戏开始前）
}

func NewGameState(players []*Player, bigBlind int64) *GameState {
	game := &GameState{
		ID:             uuid.New().String(),
		Phase:          PreFlop,
		Players:        players,
		CommunityCards: make([]Card, 0, 5),
		CurrentPlayer:  0,
		CurrentBet:     0,
		MinRaise:       bigBlind,
		Pot:            0,
		SmallBlind:     bigBlind / 2,
		BigBlind:       bigBlind,
		Deck:           NewDeck(),
	}

	// 初始化玩家状态
	for _, p := range players {
		p.Status = Active
		p.HoleCards = nil
		p.Bet = 0
		p.TotalBet = 0
	}

	// 洗牌
	game.Deck.Shuffle()

	// 发手牌
	for _, p := range players {
		cards := game.Deck.DealAll(2)
		p.HoleCards = cards
	}

	game.DealerIndex = 0
	if len(players) == 2 {
		game.SmallBlindIndex = game.DealerIndex
		game.BigBlindIndex = (game.DealerIndex + 1) % len(players)
	} else {
		game.SmallBlindIndex = (game.DealerIndex + 1) % len(players)
		game.BigBlindIndex = (game.DealerIndex + 2) % len(players)
	}

	game.postBlind(game.SmallBlindIndex, game.SmallBlind)
	game.postBlind(game.BigBlindIndex, game.BigBlind)

	game.CurrentBet = maxInt64(game.Players[game.SmallBlindIndex].Bet, game.Players[game.BigBlindIndex].Bet)
	game.MinRaise = game.BigBlind
	game.LastAggressorIndex = game.BigBlindIndex
	game.ActedThisRound = make([]bool, len(players))
	game.CurrentPlayer = game.nextActiveFrom(game.BigBlindIndex)

	return game
}

func (g *GameState) ProcessAction(action *Action) bool {
	playerIndex := g.indexOfPlayer(action.PlayerID)
	if playerIndex == -1 {
		return false
	}
	player := g.Players[playerIndex]
	if player.Status != Active {
		return false
	}

	requiredToCall := g.CurrentBet - player.Bet
	if requiredToCall < 0 {
		requiredToCall = 0
	}

	switch action.Type {
	case CheckAction:
		if requiredToCall > 0 {
			return false // 不能过牌，必须跟注
		}
		g.ActedThisRound[playerIndex] = true
	case CallAction:
		if requiredToCall == 0 {
			return false
		}
		// 筹码不足以支付跟注，转为 all-in
		if player.Chips < requiredToCall {
			action.Type = AllInAction
			return g.processAllIn(player, playerIndex, player.Bet+player.Chips)
		}
		callAmount := minInt64(player.Chips, requiredToCall)
		player.Chips -= callAmount
		player.Bet += callAmount
		player.TotalBet += callAmount
		g.Pot += callAmount
		action.Amount = callAmount // 记录跟注金额
		if player.Chips == 0 {
			player.Status = AllIn
		}
		g.ActedThisRound[playerIndex] = true
	case RaiseAction:
		if action.Amount <= g.CurrentBet {
			return false
		}
		totalAvailable := player.Bet + player.Chips
		minRaiseTo := g.CurrentBet + g.MinRaise
		// 金额超出总可用筹码时，如果玩家想用全部筹码，则转为 all-in
		if action.Amount > totalAvailable {
			if action.Amount == totalAvailable && player.Chips > 0 {
				// 全部跟注的情况：转为 all-in
				action.Type = AllInAction
				action.Amount = totalAvailable
				return g.processAllIn(player, playerIndex, totalAvailable)
			}
			return false
		}
		if action.Amount < minRaiseTo && action.Amount != totalAvailable {
			return false
		}
		contribution := action.Amount - player.Bet
		if contribution <= 0 {
			return false
		}
		player.Chips -= contribution
		player.Bet = action.Amount
		player.TotalBet += contribution
		g.Pot += contribution
		if player.Chips == 0 {
			player.Status = AllIn
		}
		oldBet := g.CurrentBet
		if action.Amount > oldBet {
			g.MinRaise = action.Amount - oldBet
			g.CurrentBet = action.Amount
			g.LastAggressorIndex = playerIndex
			g.resetActedAfterRaise(playerIndex)
		} else {
			g.ActedThisRound[playerIndex] = true
		}
	case FoldAction:
		player.Status = Folded
		g.ActedThisRound[playerIndex] = true
	case AllInAction:
		return g.processAllIn(player, playerIndex, player.Bet+player.Chips)
	}

	g.Actions = append(g.Actions, *action)
	g.advanceToNextPlayer()
	return true
}

func (g *GameState) NextPhase() {
	switch g.Phase {
	case PreFlop:
		g.Phase = Flop
		g.CommunityCards = append(g.CommunityCards, g.Deck.DealAll(3)...)
	case Flop:
		g.Phase = Turn
		if card, ok := g.Deck.Deal(); ok {
			g.CommunityCards = append(g.CommunityCards, card)
		}
	case Turn:
		g.Phase = River
		if card, ok := g.Deck.Deal(); ok {
			g.CommunityCards = append(g.CommunityCards, card)
		}
	case River:
		g.Phase = Showdown
	}

	firstToAct := g.nextActiveFrom(g.DealerIndex)
	g.startBettingRound(firstToAct)
}

func (g *GameState) getPlayer(id string) *Player {
	for _, p := range g.Players {
		if p.ID == id {
			return p
		}
	}
	return nil
}

func (g *GameState) IsRoundFinished() bool {
	if g.activeCount() <= 1 {
		return true
	}
	for i, p := range g.Players {
		if p.Status == Active {
			if !g.ActedThisRound[i] {
				return false
			}
			if p.Bet < g.CurrentBet {
				return false
			}
		}
	}
	return true
}

func (g *GameState) IsHandFinished() bool {
	return g.activeCount() <= 1
}

func (g *GameState) startBettingRound(firstToAct int) {
	for _, p := range g.Players {
		p.Bet = 0
	}
	g.CurrentBet = 0
	g.MinRaise = g.BigBlind
	g.LastAggressorIndex = firstToAct
	g.ActedThisRound = make([]bool, len(g.Players))
	g.CurrentPlayer = firstToAct
}

func (g *GameState) resetActedAfterRaise(raiserIndex int) {
	for i, p := range g.Players {
		if p.Status == Active {
			g.ActedThisRound[i] = false
		}
	}
	g.ActedThisRound[raiserIndex] = true
}

func (g *GameState) activeCount() int {
	count := 0
	for _, p := range g.Players {
		if p.Status == Active || p.Status == AllIn {
			count++
		}
	}
	return count
}

func (g *GameState) ActiveOnlyCount() int {
	count := 0
	for _, p := range g.Players {
		if p.Status == Active {
			count++
		}
	}
	return count
}

func (g *GameState) nextActiveFrom(start int) int {
	if len(g.Players) == 0 {
		return 0
	}
	for i := 0; i < len(g.Players); i++ {
		idx := (start + 1 + i) % len(g.Players)
		if g.Players[idx].Status == Active {
			return idx
		}
	}
	return start % len(g.Players)
}

func (g *GameState) advanceToNextPlayer() {
	if g.activeCount() <= 1 {
		return
	}
	for i := 0; i < len(g.Players); i++ {
		g.CurrentPlayer = (g.CurrentPlayer + 1) % len(g.Players)
		if g.Players[g.CurrentPlayer].Status == Active {
			return
		}
	}
}

func (g *GameState) indexOfPlayer(id string) int {
	for i, p := range g.Players {
		if p.ID == id {
			return i
		}
	}
	return -1
}

func (g *GameState) postBlind(index int, amount int64) {
	if index < 0 || index >= len(g.Players) {
		return
	}
	player := g.Players[index]
	post := minInt64(player.Chips, amount)
	player.Chips -= post
	player.Bet += post
	player.TotalBet += post
	g.Pot += post
	if player.Chips == 0 {
		player.Status = AllIn
	}
	// 记录盲注动作
	actionType := SmallBlindAction
	if amount == g.BigBlind {
		actionType = BigBlindAction
	}
	g.Actions = append(g.Actions, Action{
		Type:     actionType,
		PlayerID: player.ID,
		Amount:   post,
	})
}

// processAllIn 处理 all-in 动作，total 是玩家本轮总下注额（Bet + Chips）
func (g *GameState) processAllIn(player *Player, playerIndex int, total int64) bool {
	if player.Chips == 0 {
		return false
	}
	contribution := player.Chips
	oldBet := g.CurrentBet
	player.Chips = 0
	player.Status = AllIn
	player.Bet = total
	player.TotalBet += contribution
	g.Pot += contribution

	if total > oldBet {
		// all-in 形成加注：更新下注线，让其他玩家依次决策
		g.CurrentBet = total
		g.MinRaise = maxInt64(g.MinRaise, total-oldBet)
		g.LastAggressorIndex = playerIndex
		// 只标记 all-in 玩家已行动，不重置其他玩家的行动状态
		g.ActedThisRound[playerIndex] = true
	} else {
		// all-in 金额 <= 当前下注，只需跟注
		g.ActedThisRound[playerIndex] = true
	}

	// 记录动作
	g.Actions = append(g.Actions, Action{
		Type:     AllInAction,
		PlayerID: player.ID,
		Amount:   total,
	})

	// 推进到下一家
	g.advanceToNextPlayer()
	return true
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
