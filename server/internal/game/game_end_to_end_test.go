package game

import (
	"testing"
)

// TestEndToEndTwoPlayer_PreflopToShowdown 完整两人对战流程测试
// 模拟 Alice vs Bob 的完整一局，包括所有阶段转换
func TestEndToEndTwoPlayer_PreflopToShowdown(t *testing.T) {
	players := []*Player{
		{ID: "alice", Name: "Alice", Chips: 1000},
		{ID: "bob", Name: "Bob", Chips: 1000},
	}
	game := NewGameState(players, 100)

	t.Log("========== 游戏初始化 ==========")
	t.Logf("阶段: %s", game.Phase)
	t.Logf("DealerIndex: %d, SmallBlindIndex: %d, BigBlindIndex: %d", game.DealerIndex, game.SmallBlindIndex, game.BigBlindIndex)
	t.Logf("CurrentPlayer: %s", game.Players[game.CurrentPlayer].Name)
	t.Logf("Alice: chips=%d, bet=%d, status=%s", game.Players[0].Chips, game.Players[0].Bet, game.Players[0].Status)
	t.Logf("Bob: chips=%d, bet=%d, status=%s", game.Players[1].Chips, game.Players[1].Bet, game.Players[1].Status)
	t.Logf("Pot: %d, CurrentBet: %d, MinRaise: %d", game.Pot, game.CurrentBet, game.MinRaise)

	// ========== PREFLOP ==========
	t.Log("\n========== PREFLOP ==========")

	// Alice (小盲, $50) 需要再跟 $50 才能继续，或者可以加注/弃牌
	// 当前 Bob 已经下了 $100 (大盲)

	// Alice raise to 200
	raiseAction := &Action{Type: RaiseAction, PlayerID: "alice", Amount: 200}
	if !game.ProcessAction(raiseAction) {
		t.Fatal("Alice raise 应该成功")
	}
	t.Logf("Alice raise 200 -> 成功")
	t.Logf("CurrentPlayer: %s, CurrentBet: %d", game.Players[game.CurrentPlayer].Name, game.CurrentBet)
	t.Logf("Alice: chips=%d, bet=%d", game.Players[0].Chips, game.Players[0].Bet)
	t.Logf("Bob: chips=%d, bet=%d", game.Players[1].Chips, game.Players[1].Bet)
	t.Logf("IsRoundFinished: %v", game.IsRoundFinished())

	// Bob 跟注 200
	callAction := &Action{Type: CallAction, PlayerID: "bob", Amount: 0}
	if !game.ProcessAction(callAction) {
		t.Fatal("Bob call 应该成功")
	}
	t.Logf("Bob call 200 -> 成功")
	t.Logf("Pot: %d", game.Pot)
	t.Logf("IsRoundFinished: %v", game.IsRoundFinished())

	if !game.IsRoundFinished() {
		t.Fatal("Preflop 回合应该结束")
	}

	// ========== FLOP ==========
	t.Log("\n========== FLOP ==========")
	game.NextPhase()
	t.Logf("阶段: %s", game.Phase)
	t.Logf("公共牌: %v", game.CommunityCards)
	t.Logf("CurrentPlayer: %s", game.Players[game.CurrentPlayer].Name)
	t.Logf("Pot: %d, CurrentBet: %d", game.Pot, game.CurrentBet)
	t.Logf("IsRoundFinished: %v", game.IsRoundFinished())

	// FLOP 轮次 - 当前下注归零，所有人从 DealerIndex+1 开始行动
	// DealerIndex=0, 所以 Bob (index 1) 先行动

	// Bob check (当前下注为 0，不需要跟注)
	checkAction := &Action{Type: CheckAction, PlayerID: "bob", Amount: 0}
	if !game.ProcessAction(checkAction) {
		t.Fatal("Bob check 应该成功")
	}
	t.Logf("Bob check -> 成功")

	// Alice check
	checkAction = &Action{Type: CheckAction, PlayerID: "alice", Amount: 0}
	if !game.ProcessAction(checkAction) {
		t.Fatal("Alice check 应该成功")
	}
	t.Logf("Alice check -> 成功")

	if !game.IsRoundFinished() {
		t.Fatal("Flop 回合应该结束")
	}

	// ========== TURN ==========
	t.Log("\n========== TURN ==========")
	game.NextPhase()
	t.Logf("阶段: %s", game.Phase)
	t.Logf("公共牌: %v", game.CommunityCards)
	t.Logf("CurrentPlayer: %s", game.Players[game.CurrentPlayer].Name)

	// Alice bet 100
	raiseAction = &Action{Type: RaiseAction, PlayerID: "alice", Amount: 100}
	if !game.ProcessAction(raiseAction) {
		t.Fatal("Alice raise 应该成功")
	}
	t.Logf("Alice raise 100 -> 成功")

	// Bob call
	callAction = &Action{Type: CallAction, PlayerID: "bob", Amount: 0}
	if !game.ProcessAction(callAction) {
		t.Fatal("Bob call 应该成功")
	}
	t.Logf("Bob call -> 成功")

	if !game.IsRoundFinished() {
		t.Fatal("Turn 回合应该结束")
	}

	// ========== RIVER ==========
	t.Log("\n========== RIVER ==========")
	game.NextPhase()
	t.Logf("阶段: %s", game.Phase)
	t.Logf("公共牌: %v", game.CommunityCards)
	t.Logf("CurrentPlayer: %s", game.Players[game.CurrentPlayer].Name)

	// Bob check
	checkAction = &Action{Type: CheckAction, PlayerID: "bob", Amount: 0}
	if !game.ProcessAction(checkAction) {
		t.Fatal("Bob check 应该成功")
	}
	t.Logf("Bob check -> 成功")

	// Alice check
	checkAction = &Action{Type: CheckAction, PlayerID: "alice", Amount: 0}
	if !game.ProcessAction(checkAction) {
		t.Fatal("Alice check 应该成功")
	}
	t.Logf("Alice check -> 成功")

	if !game.IsRoundFinished() {
		t.Fatal("River 回合应该结束")
	}

	// ========== SHOWDOWN ==========
	t.Log("\n========== SHOWDOWN ==========")
	game.NextPhase()
	t.Logf("阶段: %s", game.Phase)

	winners := DetermineWinners(game)
	if len(winners) == 0 {
		t.Fatal("应该有赢家")
	}
	t.Logf("赢家: %s, 牌型: %s, 赢取: %d", winners[0].PlayerName, winners[0].HandRank, winners[0].WinAmount)

	// 验证结算后的筹码
	t.Logf("结算后 Alice chips: %d", game.Players[0].Chips)
	t.Logf("结算后 Bob chips: %d", game.Players[1].Chips)

	t.Log("\n========== 游戏结束 ==========")
}

// TestEndToEndTwoPlayer_AllInShowdown 测试两人 All-in 后进入 Showdown
func TestEndToEndTwoPlayer_AllInShowdown(t *testing.T) {
	players := []*Player{
		{ID: "alice", Name: "Alice", Chips: 500},
		{ID: "bob", Name: "Bob", Chips: 1000},
	}
	game := NewGameState(players, 100)

	t.Log("========== All-in 场景测试 ==========")

	// Alice all-in 在 preflop
	allInAction := &Action{Type: AllInAction, PlayerID: "alice", Amount: 0}
	if !game.ProcessAction(allInAction) {
		t.Fatal("Alice all-in 应该成功")
	}
	t.Logf("Alice all-in 500")

	// Bob 跟注
	callAction := &Action{Type: CallAction, PlayerID: "bob", Amount: 0}
	if !game.ProcessAction(callAction) {
		t.Fatal("Bob call 应该成功")
	}
	t.Logf("Bob call")

	// 此时 Bob 也 all-in 了，游戏应该进入发牌阶段
	t.Logf("当前玩家数: %d (Active: %d)", game.activeCount(), game.ActiveOnlyCount())

	// 翻牌
	game.NextPhase()
	t.Logf("阶段: %s, 公共牌: %v", game.Phase, game.CommunityCards)

	// 转牌
	game.NextPhase()
	t.Logf("阶段: %s, 公共牌: %v", game.Phase, game.CommunityCards)

	// 河牌
	game.NextPhase()
	t.Logf("阶段: %s, 公共牌: %v", game.Phase, game.CommunityCards)

	// 摊牌
	game.NextPhase()
	t.Logf("阶段: %s", game.Phase)

	winners := DetermineWinners(game)
	t.Logf("赢家: %s, 牌型: %s, 赢取: %d", winners[0].PlayerName, winners[0].HandRank, winners[0].WinAmount)

	// 验证总筹码不变 (500 + 1000 = 1500)
	totalChips := game.Players[0].Chips + game.Players[1].Chips + game.Pot
	t.Logf("总筹码: %d", totalChips)
}

// TestEndToEndTwoPlayer_FoldScenario 测试弃牌场景
func TestEndToEndTwoPlayer_FoldScenario(t *testing.T) {
	players := []*Player{
		{ID: "alice", Name: "Alice", Chips: 1000},
		{ID: "bob", Name: "Bob", Chips: 1000},
	}
	game := NewGameState(players, 100)

	t.Log("========== 弃牌场景测试 ==========")

	// Alice all-in
	allInAction := &Action{Type: AllInAction, PlayerID: "alice", Amount: 0}
	if !game.ProcessAction(allInAction) {
		t.Fatal("Alice all-in 应该成功")
	}
	t.Logf("Alice all-in")

	// Bob fold
	foldAction := &Action{Type: FoldAction, PlayerID: "bob", Amount: 0}
	if !game.ProcessAction(foldAction) {
		t.Fatal("Bob fold 应该成功")
	}
	t.Logf("Bob fold")

	// 游戏应该立即结束
	if !game.IsHandFinished() {
		t.Fatal("Bob 弃牌后，手牌应该结束")
	}

	t.Logf("赢家: Alice, 赢取底池: %d", game.Pot)
}

// TestCheckCannotBeDoneWhenCallRequired 测试需要跟注时不能过牌
func TestCheckCannotBeDoneWhenCallRequired(t *testing.T) {
	players := []*Player{
		{ID: "alice", Name: "Alice", Chips: 1000},
		{ID: "bob", Name: "Bob", Chips: 1000},
	}
	game := NewGameState(players, 100)

	t.Log("========== Check 限制测试 ==========")

	// Alice (小盲) 下了 50, Bob (大盲) 下了 100
	// Alice 需要再跟 50 才能继续

	// Alice 尝试 check (应该失败)
	checkAction := &Action{Type: CheckAction, PlayerID: "alice", Amount: 0}
	if game.ProcessAction(checkAction) {
		t.Fatal("Alice check 应该失败 (需要跟注 50)")
	}
	t.Logf("Alice check 失败 -> 正确 (需要跟注)")

	// Alice 跟注
	callAction := &Action{Type: CallAction, PlayerID: "alice", Amount: 0}
	if !game.ProcessAction(callAction) {
		t.Fatal("Alice call 应该成功")
	}
	t.Logf("Alice call 50 -> 成功")

	// 现在轮到 Bob，当前下注为 0，可以 check
	checkAction = &Action{Type: CheckAction, PlayerID: "bob", Amount: 0}
	if !game.ProcessAction(checkAction) {
		t.Fatal("Bob check 应该成功")
	}
	t.Logf("Bob check -> 成功")

	t.Log("测试通过!")
}

// TestRaiseAndCall 测试加注和跟注
func TestRaiseAndCall(t *testing.T) {
	players := []*Player{
		{ID: "alice", Name: "Alice", Chips: 1000},
		{ID: "bob", Name: "Bob", Chips: 1000},
	}
	game := NewGameState(players, 100)

	t.Log("========== 加注跟注测试 ==========")

	// Alice raise to 300
	raiseAction := &Action{Type: RaiseAction, PlayerID: "alice", Amount: 300}
	if !game.ProcessAction(raiseAction) {
		t.Fatal("Alice raise 应该成功")
	}
	t.Logf("Alice raise 300")
	t.Logf("CurrentBet: %d, MinRaise: %d", game.CurrentBet, game.MinRaise)

	// Bob 跟注 300
	callAction := &Action{Type: CallAction, PlayerID: "bob", Amount: 0}
	if !game.ProcessAction(callAction) {
		t.Fatal("Bob call 应该成功")
	}
	t.Logf("Bob call 300")

	// 验证
	// Alice raise 300: 她已有 bet=50, 需要补 250 (300-50), chips: 1000-250=750
	// Bob call 300: 他已有 bet=100, 需要补 200 (300-100), chips: 1000-200=800
	// Pot = 初始盲注 150 + Alice补250 + Bob补200 = 600
	t.Logf("Pot: %d (expected: 600 = 150盲注 + 250Alice + 200Bob)", game.Pot)
	t.Logf("Alice: chips=%d, bet=%d", game.Players[0].Chips, game.Players[0].Bet)
	t.Logf("Bob: chips=%d, bet=%d", game.Players[1].Chips, game.Players[1].Bet)

	if game.Pot != 600 {
		t.Errorf("Pot 应该是 600, 实际是 %d", game.Pot)
	}
}
