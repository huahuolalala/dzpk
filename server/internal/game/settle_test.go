package game

import (
	"testing"
)

func TestSettle_TwoPlayers(t *testing.T) {
	// Alice 有一对 A，Bob 有一对 K
	players := []*Player{
		{ID: "p1", Name: "Alice", Chips: 1000},
		{ID: "p2", Name: "Bob", Chips: 1000},
	}
	game := &GameState{
		Phase:   Showdown,
		Players: players,
		Pot:     200,
		CommunityCards: []Card{
			NewCard(Ace, Diamond),
			NewCard(King, Club),
			NewCard(Queen, Heart),
			NewCard(Jack, Spade),
			NewCard(Ten, Diamond),
		},
	}
	game.Players[0].HoleCards = []Card{NewCard(Ace, Spade), NewCard(Ace, Heart)}
	game.Players[1].HoleCards = []Card{NewCard(King, Spade), NewCard(King, Heart)}
	game.Players[0].Status = Active
	game.Players[1].Status = Active

	results := DetermineWinners(game)

	if len(results) != 1 {
		t.Errorf("Should have 1 winner, got %d", len(results))
	}
	if results[0].PlayerID != "p1" {
		t.Errorf("Alice should win, got %s", results[0].PlayerID)
	}
}

func TestSettle_SplitPot(t *testing.T) {
	// 两人平分底池
	players := []*Player{
		{ID: "p1", Name: "Alice", Chips: 1000},
		{ID: "p2", Name: "Bob", Chips: 1000},
	}
	game := NewGameState(players, 100)
	game.Phase = Showdown
	game.Pot = 200

	// 设置相同手牌
	game.Players[0].HoleCards = []Card{NewCard(Ace, Spade), NewCard(King, Heart)}
	game.Players[1].HoleCards = []Card{NewCard(Ace, Heart), NewCard(King, Spade)}
	game.Players[0].Status = Active
	game.Players[1].Status = Active

	results := DetermineWinners(game)

	if len(results) != 2 {
		t.Errorf("Should have 2 winners (split), got %d", len(results))
	}
}

func TestSettle_AllInSidePot(t *testing.T) {
	// Alice 全下 100，Bob 跟注 100，Charlie 跟注 100
	players := []*Player{
		{ID: "p1", Name: "Alice", Chips: 0, Status: AllIn, Bet: 100},
		{ID: "p2", Name: "Bob", Chips: 900, Status: Active, Bet: 100},
		{ID: "p3", Name: "Charlie", Chips: 900, Status: Active, Bet: 100},
	}
	game := &GameState{
		Players: players,
		Pot:     300,
	}

	// 主池 = 300 (3 x 100)
	sidePots, mainPot := CalculatePots(game)

	if mainPot != 300 {
		t.Errorf("Main pot should be 300, got %d", mainPot)
	}
	if len(sidePots) != 0 {
		t.Errorf("Should have no side pots, got %d", len(sidePots))
	}
}
