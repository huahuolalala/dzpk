package game

import (
	"testing"
)

func TestGameState_NewGame(t *testing.T) {
	players := []*Player{
		{ID: "p1", Name: "Alice", Chips: 1000},
		{ID: "p2", Name: "Bob", Chips: 1000},
	}
	game := NewGameState(players, 100)

	if game.Phase != PreFlop {
		t.Errorf("Initial phase should be PreFlop, got %v", game.Phase)
	}
	if len(game.Players) != 2 {
		t.Errorf("Should have 2 players, got %d", len(game.Players))
	}
	if game.CurrentPlayer != 0 {
		t.Errorf("CurrentPlayer should be 0, got %d", game.CurrentPlayer)
	}
	if game.SmallBlind != 50 {
		t.Errorf("SmallBlind should be 50, got %d", game.SmallBlind)
	}
	if game.BigBlind != 100 {
		t.Errorf("BigBlind should be 100, got %d", game.BigBlind)
	}
	if game.CurrentBet != 100 {
		t.Errorf("CurrentBet should be 100, got %d", game.CurrentBet)
	}
}

func TestGameState_ProcessAction_Check(t *testing.T) {
	players := []*Player{
		{ID: "p1", Name: "Alice", Chips: 1000},
		{ID: "p2", Name: "Bob", Chips: 1000},
	}
	game := NewGameState(players, 100)
	game.CurrentBet = 0
	for _, p := range game.Players {
		p.Bet = 0
	}

	action := &Action{Type: CheckAction, PlayerID: "p1"}
	ok := game.ProcessAction(action)
	if !ok {
		t.Error("Check action should succeed when CurrentBet=0")
	}
}

func TestGameState_ProcessAction_Call(t *testing.T) {
	players := []*Player{
		{ID: "p1", Name: "Alice", Chips: 1000},
		{ID: "p2", Name: "Bob", Chips: 1000},
	}
	game := NewGameState(players, 100)
	game.CurrentBet = 100
	game.Pot = 0
	game.ActedThisRound = make([]bool, len(game.Players))
	for _, p := range game.Players {
		p.Bet = 0
		p.Chips = 1000
		p.Status = Active
	}

	action := &Action{Type: CallAction, PlayerID: "p1"}
	ok := game.ProcessAction(action)
	if !ok {
		t.Error("Call action should succeed")
	}
	if game.Players[0].Chips != 900 {
		t.Errorf("Alice chips should be 900, got %d", game.Players[0].Chips)
	}
}

func TestGameState_ProcessAction_Raise(t *testing.T) {
	players := []*Player{
		{ID: "p1", Name: "Alice", Chips: 1000},
		{ID: "p2", Name: "Bob", Chips: 1000},
	}
	game := NewGameState(players, 100)
	game.CurrentBet = 100
	game.MinRaise = 100
	game.Pot = 0
	game.ActedThisRound = make([]bool, len(game.Players))
	for _, p := range game.Players {
		p.Bet = 0
		p.Chips = 1000
		p.Status = Active
	}

	action := &Action{Type: RaiseAction, PlayerID: "p1", Amount: 300}
	ok := game.ProcessAction(action)
	if !ok {
		t.Error("Raise action should succeed")
	}
	if game.CurrentBet != 300 {
		t.Errorf("CurrentBet should be 300, got %d", game.CurrentBet)
	}
	if game.Players[0].Chips != 700 {
		t.Errorf("Alice chips should be 700, got %d", game.Players[0].Chips)
	}
}

func TestGameState_ProcessAction_Fold(t *testing.T) {
	players := []*Player{
		{ID: "p1", Name: "Alice", Chips: 1000},
		{ID: "p2", Name: "Bob", Chips: 1000},
	}
	game := NewGameState(players, 100)

	action := &Action{Type: FoldAction, PlayerID: "p1"}
	ok := game.ProcessAction(action)
	if !ok {
		t.Error("Fold action should succeed")
	}
	if game.Players[0].Status != Folded {
		t.Errorf("Alice status should be Folded, got %v", game.Players[0].Status)
	}
}

func TestGameState_ProcessAction_AllIn(t *testing.T) {
	players := []*Player{
		{ID: "p1", Name: "Alice", Chips: 500},
		{ID: "p2", Name: "Bob", Chips: 1000},
	}
	game := NewGameState(players, 100)

	action := &Action{Type: AllInAction, PlayerID: "p1"}
	ok := game.ProcessAction(action)
	if !ok {
		t.Error("AllIn action should succeed")
	}
	if game.Players[0].Status != AllIn {
		t.Errorf("Alice status should be AllIn, got %v", game.Players[0].Status)
	}
	if game.Players[0].Chips != 0 {
		t.Errorf("Alice chips should be 0, got %d", game.Players[0].Chips)
	}
}

func TestGameState_NextPhase(t *testing.T) {
	players := []*Player{
		{ID: "p1", Name: "Alice", Chips: 1000},
		{ID: "p2", Name: "Bob", Chips: 1000},
	}
	game := NewGameState(players, 100)

	// PreFlop -> Flop
	game.NextPhase()
	if game.Phase != Flop {
		t.Errorf("Next phase should be Flop, got %v", game.Phase)
	}
	if len(game.CommunityCards) != 3 {
		t.Errorf("Flop should have 3 community cards, got %d", len(game.CommunityCards))
	}

	// Flop -> Turn
	game.NextPhase()
	if game.Phase != Turn {
		t.Errorf("Next phase should be Turn, got %v", game.Phase)
	}
	if len(game.CommunityCards) != 4 {
		t.Errorf("Turn should have 4 community cards, got %d", len(game.CommunityCards))
	}

	// Turn -> River
	game.NextPhase()
	if game.Phase != River {
		t.Errorf("Next phase should be River, got %v", game.Phase)
	}
	if len(game.CommunityCards) != 5 {
		t.Errorf("River should have 5 community cards, got %d", len(game.CommunityCards))
	}

	// River -> Showdown
	game.NextPhase()
	if game.Phase != Showdown {
		t.Errorf("Next phase should be Showdown, got %v", game.Phase)
	}
}
