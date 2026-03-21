package game

import (
	"testing"
)

// TestTwoPlayerFullGame 两人对战完整流程测试
// 场景：Alice vs Bob，追踪完整的 preflop -> flop -> turn -> river -> showdown
func TestTwoPlayerFullGame(t *testing.T) {
	players := []*Player{
		{ID: "alice", Name: "Alice", Chips: 1000},
		{ID: "bob", Name: "Bob", Chips: 1000},
	}
	game := NewGameState(players, 100) // BigBlind=100

	t.Logf("=== 游戏开始 ===")
	t.Logf("阶段: %s", game.Phase)
	t.Logf("Dealer: Alice, SB: Alice(50), BB: Bob(100)")
	t.Logf("当前下注: %d, 当前玩家: %s", game.CurrentBet, game.Players[game.CurrentPlayer].Name)
	t.Logf("Alice: chips=%d, bet=%d, status=%s", game.Players[0].Chips, game.Players[0].Bet, game.Players[0].Status)
	t.Logf("Bob: chips=%d, bet=%d, status=%s", game.Players[1].Chips, game.Players[1].Bet, game.Players[1].Status)

	// ========== PREFLOP ==========
	t.Logf("\n=== PREFLOP 阶段 ===")

	// Alice (小盲) 需要决定：call 50, raise, or fold
	// 当前 CurrentPlayer = Alice (index 0)
	// Alice 的 bet = 50 (小盲), Bob 的 bet = 100 (大盲)
	// Alice 需要再跟 50 才能继续

	// Alice raise to 200
	action := &Action{Type: RaiseAction, PlayerID: "alice", Amount: 200}
	ok := game.ProcessAction(action)
	if !ok {
		t.Errorf("Alice raise 200 应该成功")
	}
	t.Logf("Alice raise 200 -> ok")
	t.Logf("Alice: chips=%d, bet=%d", game.Players[0].Chips, game.Players[0].Bet)
	t.Logf("Bob: chips=%d, bet=%d", game.Players[1].Chips, game.Players[1].Bet)
	t.Logf("CurrentBet=%d, CurrentPlayer=%s", game.CurrentBet, game.Players[game.CurrentPlayer].Name)

	// Bob 跟注
	action = &Action{Type: CallAction, PlayerID: "bob", Amount: 0}
	ok = game.ProcessAction(action)
	if !ok {
		t.Errorf("Bob call 应该成功")
	}
	t.Logf("Bob call -> ok")
	t.Logf("Pot=%d", game.Pot)

	// 检查 preflop 是否结束
	if !game.IsRoundFinished() {
		t.Errorf("Preflop 应该结束")
	}

	// 进入 FLOP
	t.Logf("\n=== 进入 FLOP ===")
	game.NextPhase()
	t.Logf("阶段: %s", game.Phase)
	t.Logf("公共牌: %v", game.CommunityCards)
	t.Logf("CurrentPlayer=%s", game.Players[game.CurrentPlayer].Name)

	// FLOP: Alice 过牌, Bob 过牌
	action = &Action{Type: CheckAction, PlayerID: "alice", Amount: 0}
	ok = game.ProcessAction(action)
	if !ok {
		t.Errorf("Alice check 应该成功")
	}
	t.Logf("Alice check -> ok")

	action = &Action{Type: CheckAction, PlayerID: "bob", Amount: 0}
	ok = game.ProcessAction(action)
	if !ok {
		t.Errorf("Bob check 应该成功")
	}
	t.Logf("Bob check -> ok")

	if !game.IsRoundFinished() {
		t.Errorf("Flop betting 应该结束")
	}

	// 进入 TURN
	t.Logf("\n=== 进入 TURN ===")
	game.NextPhase()
	t.Logf("阶段: %s", game.Phase)
	t.Logf("公共牌: %v", game.CommunityCards)

	// TURN: Alice 下注 100, Bob 跟注
	action = &Action{Type: RaiseAction, PlayerID: "alice", Amount: 100}
	ok = game.ProcessAction(action)
	if !ok {
		t.Errorf("Alice raise 100 应该成功")
	}
	t.Logf("Alice raise 100 -> ok, CurrentBet=%d", game.CurrentBet)

	action = &Action{Type: CallAction, PlayerID: "bob", Amount: 0}
	ok = game.ProcessAction(action)
	if !ok {
		t.Errorf("Bob call 应该成功")
	}
	t.Logf("Bob call -> ok")

	if !game.IsRoundFinished() {
		t.Errorf("Turn betting 应该结束")
	}

	// 进入 RIVER
	t.Logf("\n=== 进入 RIVER ===")
	game.NextPhase()
	t.Logf("阶段: %s", game.Phase)
	t.Logf("公共牌: %v", game.CommunityCards)

	// RIVER: Alice 过牌, Bob 过牌
	action = &Action{Type: CheckAction, PlayerID: "alice", Amount: 0}
	ok = game.ProcessAction(action)
	if !ok {
		t.Errorf("Alice check 应该成功")
	}
	t.Logf("Alice check -> ok")

	action = &Action{Type: CheckAction, PlayerID: "bob", Amount: 0}
	ok = game.ProcessAction(action)
	if !ok {
		t.Errorf("Bob check 应该成功")
	}
	t.Logf("Bob check -> ok")

	// 进入 SHOWDOWN
	t.Logf("\n=== 进入 SHOWDOWN ===")
	game.NextPhase()
	t.Logf("阶段: %s", game.Phase)

	// 检查赢家
	winners := DetermineWinners(game)
	if len(winners) == 0 {
		t.Errorf("应该有赢家")
	}
	t.Logf("赢家: %s, 牌型: %s", winners[0].PlayerName, winners[0].HandRank)

	t.Logf("\n=== 游戏结束 ===")
}

// TestTwoPlayerAllInScenario 测试两人All-in场景
func TestTwoPlayerAllInScenario(t *testing.T) {
	players := []*Player{
		{ID: "alice", Name: "Alice", Chips: 500},
		{ID: "bob", Name: "Bob", Chips: 1000},
	}
	game := NewGameState(players, 100)

	t.Logf("=== All-in 场景测试 ===")

	// Alice 全下 (在翻牌前)
	action := &Action{Type: AllInAction, PlayerID: "alice", Amount: 0}
	ok := game.ProcessAction(action)
	if !ok {
		t.Errorf("Alice all-in 应该成功")
	}
	t.Logf("Alice all-in 500")
	t.Logf("Alice: chips=%d, status=%s, bet=%d", game.Players[0].Chips, game.Players[0].Status, game.Players[0].Bet)
	t.Logf("Bob: chips=%d, bet=%d", game.Players[1].Chips, game.Players[1].Bet)
	t.Logf("Pot=%d, CurrentBet=%d", game.Pot, game.CurrentBet)

	// Bob 决策：跟注还是弃牌
	// 如果 Bob 跟注，应该进入 showdown
	// 如果 Bob 弃牌，Alice 立即获胜

	// Bob 跟注
	action = &Action{Type: CallAction, PlayerID: "bob", Amount: 0}
	ok = game.ProcessAction(action)
	if !ok {
		t.Errorf("Bob call 应该成功")
	}
	t.Logf("Bob call -> ok")
	t.Logf("Pot=%d", game.Pot)

	// 此时两人都是 All-in，游戏应该结束并进入 showdown
	if !game.IsRoundFinished() {
		t.Errorf("All-in 场景下回合应该结束")
	}

	if !game.IsHandFinished() {
		t.Logf("手牌还没结束（因为还有公共牌未发），应该进入下一阶段")
	}

	// 进入 showdown
	for game.Phase != Showdown {
		game.NextPhase()
		t.Logf("进入阶段: %s, 公共牌: %v", game.Phase, game.CommunityCards)
	}

	winners := DetermineWinners(game)
	if len(winners) == 0 {
		t.Errorf("应该有赢家")
	}
	t.Logf("赢家: %s, 赢取: %d", winners[0].PlayerName, winners[0].WinAmount)
}

// TestTwoPlayerFoldScenario 测试弃牌场景
func TestTwoPlayerFoldScenario(t *testing.T) {
	players := []*Player{
		{ID: "alice", Name: "Alice", Chips: 1000},
		{ID: "bob", Name: "Bob", Chips: 1000},
	}
	game := NewGameState(players, 100)

	t.Logf("=== 弃牌场景测试 ===")

	// Alice 全下
	action := &Action{Type: AllInAction, PlayerID: "alice", Amount: 0}
	ok := game.ProcessAction(action)
	if !ok {
		t.Errorf("Alice all-in 应该成功")
	}
	t.Logf("Alice all-in")

	// Bob 弃牌
	action = &Action{Type: FoldAction, PlayerID: "bob", Amount: 0}
	ok = game.ProcessAction(action)
	if !ok {
		t.Errorf("Bob fold 应该成功")
	}
	t.Logf("Bob fold")

	// 此时 Alice 应该立即获胜
	if !game.IsHandFinished() {
		t.Errorf("Bob 弃牌后，手牌应该结束")
	}

	t.Logf("Alice 获胜！chips=%d", game.Players[0].Chips)
}

// TestCurrentPlayerAfterAllIn 测试 All-in 后的 CurrentPlayer 设置
func TestCurrentPlayerAfterAllIn(t *testing.T) {
	players := []*Player{
		{ID: "alice", Name: "Alice", Chips: 1000},
		{ID: "bob", Name: "Bob", Chips: 1000},
	}
	game := NewGameState(players, 100)

	t.Logf("=== CurrentPlayer 测试 ===")
	t.Logf("初始: CurrentPlayer=%s", game.Players[game.CurrentPlayer].Name)

	// Alice raise
	action := &Action{Type: RaiseAction, PlayerID: "alice", Amount: 300}
	ok := game.ProcessAction(action)
	if !ok {
		t.Errorf("Alice raise 应该成功")
	}
	t.Logf("Alice raise 300, CurrentPlayer=%s", game.Players[game.CurrentPlayer].Name)

	// Bob all-in
	action = &Action{Type: AllInAction, PlayerID: "bob", Amount: 0}
	ok = game.ProcessAction(action)
	if !ok {
		t.Errorf("Bob all-in 应该成功")
	}
	t.Logf("Bob all-in, CurrentPlayer=%s", game.Players[game.CurrentPlayer].Name)

	// Alice 跟注 (她已经被设置为了 CurrentPlayer)
	action = &Action{Type: CallAction, PlayerID: "alice", Amount: 0}
	ok = game.ProcessAction(action)
	if !ok {
		t.Errorf("Alice call 应该成功")
	}
	t.Logf("Alice call, CurrentPlayer=%s", game.Players[game.CurrentPlayer].Name)

	// 此时 Bob 是 All-in, Alice 是 Active
	// CurrentPlayer 应该被设置为一个有效的玩家

	t.Logf("最终状态:")
	for i, p := range game.Players {
		t.Logf("  %s: chips=%d, status=%s, bet=%d, acted=%v",
			p.Name, p.Chips, p.Status, p.Bet, game.ActedThisRound[i])
	}
}

// TestCheckAfterAllIn 测试 All-in 后对手过牌的问题
func TestCheckAfterAllIn(t *testing.T) {
	players := []*Player{
		{ID: "alice", Name: "Alice", Chips: 1000},
		{ID: "bob", Name: "Bob", Chips: 1000},
	}
	game := NewGameState(players, 100)

	t.Logf("=== All-in 后过牌场景 ===")

	// 进入 FLOP 阶段
	game.NextPhase()
	t.Logf("进入 FLOP, 公共牌: %v", game.CommunityCards)

	// Alice 下注 100
	action := &Action{Type: RaiseAction, PlayerID: "alice", Amount: 100}
	ok := game.ProcessAction(action)
	if !ok {
		t.Errorf("Alice raise 应该成功")
	}
	t.Logf("Alice raise 100")

	// Bob 全下 (all-in)
	action = &Action{Type: AllInAction, PlayerID: "bob", Amount: 0}
	ok = game.ProcessAction(action)
	if !ok {
		t.Errorf("Bob all-in 应该成功")
	}
	t.Logf("Bob all-in")

	// Alice 跟注
	action = &Action{Type: CallAction, PlayerID: "alice", Amount: 0}
	ok = game.ProcessAction(action)
	if !ok {
		t.Errorf("Alice call 应该成功")
	}
	t.Logf("Alice call")

	t.Logf("回合状态检查:")
	t.Logf("  IsRoundFinished=%v", game.IsRoundFinished())
	t.Logf("  IsHandFinished=%v", game.IsHandFinished())
	t.Logf("  activeCount=%d", game.activeCount())
	t.Logf("  ActiveOnlyCount=%d", game.ActiveOnlyCount())

	// 关键检查：当 Bob 全下后 Alice 跟注，两人下注相等
	// 回合应该结束
	if !game.IsRoundFinished() {
		t.Errorf("Bob 全下 Alice 跟注后，回合应该结束")
	}

	// 但是！此时是否应该立即进入 showdown 还是继续发牌？
	// 在真实的扑克中，如果还有公共牌没发，应该继续发牌然后 showdown
	t.Logf("当前阶段: %s", game.Phase)
}

// TestCheckActionRequiresCall 测试需要跟注时不能过牌
func TestCheckActionRequiresCall(t *testing.T) {
	players := []*Player{
		{ID: "alice", Name: "Alice", Chips: 1000},
		{ID: "bob", Name: "Bob", Chips: 1000},
	}
	game := NewGameState(players, 100)

	t.Logf("=== Check 需要跟注时不能过牌 ===")

	// 当前玩家是 Alice (小盲位)
	// Bob (大盲) 下了 100, Alice 下了 50 (小盲)
	// Alice 需要跟 50

	// Alice 尝试过牌 (应该失败)
	action := &Action{Type: CheckAction, PlayerID: "alice", Amount: 0}
	ok := game.ProcessAction(action)
	if ok {
		t.Errorf("Alice 在需要跟注时不能过牌")
	} else {
		t.Logf("正确: Alice 不能过牌 (需要跟注 50)")
	}

	// Alice 跟注
	action = &Action{Type: CallAction, PlayerID: "alice", Amount: 0}
	ok = game.ProcessAction(action)
	if !ok {
		t.Errorf("Alice call 应该成功")
	}
	t.Logf("Alice call 50 -> ok")

	// Bob 过牌 (不需要跟注，可以过牌)
	action = &Action{Type: CheckAction, PlayerID: "bob", Amount: 0}
	ok = game.ProcessAction(action)
	if !ok {
		t.Errorf("Bob check 应该成功")
	}
	t.Logf("Bob check -> ok")

	// 回合应该结束
	if !game.IsRoundFinished() {
		t.Errorf("Preflop 应该结束")
	}
	t.Logf("Preflop 回合结束, Pot=%d", game.Pot)
}
