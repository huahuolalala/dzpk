package game

import (
	"testing"
)

func TestHandRank_RoyalFlush(t *testing.T) {
	cards := []Card{
		NewCard(Ace, Spade),
		NewCard(King, Spade),
		NewCard(Queen, Spade),
		NewCard(Jack, Spade),
		NewCard(Ten, Spade),
	}
	rank := JudgeHand(cards)
	if rank != RoyalFlush {
		t.Errorf("RoyalFlush expected, got %v", rank)
	}
}

func TestHandRank_StraightFlush(t *testing.T) {
	cards := []Card{
		NewCard(Nine, Heart),
		NewCard(Eight, Heart),
		NewCard(Seven, Heart),
		NewCard(Six, Heart),
		NewCard(Five, Heart),
	}
	rank := JudgeHand(cards)
	if rank != StraightFlush {
		t.Errorf("StraightFlush expected, got %v", rank)
	}
}

func TestHandRank_FourOfAKind(t *testing.T) {
	cards := []Card{
		NewCard(Ace, Spade),
		NewCard(Ace, Heart),
		NewCard(Ace, Diamond),
		NewCard(Ace, Club),
		NewCard(King, Spade),
	}
	rank := JudgeHand(cards)
	if rank != FourOfAKind {
		t.Errorf("FourOfAKind expected, got %v", rank)
	}
}

func TestHandRank_FullHouse(t *testing.T) {
	cards := []Card{
		NewCard(Ace, Spade),
		NewCard(Ace, Heart),
		NewCard(Ace, Diamond),
		NewCard(King, Spade),
		NewCard(King, Heart),
	}
	rank := JudgeHand(cards)
	if rank != FullHouse {
		t.Errorf("FullHouse expected, got %v", rank)
	}
}

func TestHandRank_Flush(t *testing.T) {
	cards := []Card{
		NewCard(Ace, Spade),
		NewCard(Jack, Spade),
		NewCard(Eight, Spade),
		NewCard(Six, Spade),
		NewCard(Three, Spade),
	}
	rank := JudgeHand(cards)
	if rank != Flush {
		t.Errorf("Flush expected, got %v", rank)
	}
}

func TestHandRank_Straight(t *testing.T) {
	cards := []Card{
		NewCard(Nine, Spade),
		NewCard(Eight, Heart),
		NewCard(Seven, Diamond),
		NewCard(Six, Club),
		NewCard(Five, Spade),
	}
	rank := JudgeHand(cards)
	if rank != Straight {
		t.Errorf("Straight expected, got %v", rank)
	}
}

func TestHandRank_A2345_Straight(t *testing.T) {
	// A-2-3-4-5 顺子
	cards := []Card{
		NewCard(Ace, Spade),
		NewCard(Deuce, Heart),
		NewCard(Three, Diamond),
		NewCard(Four, Club),
		NewCard(Five, Spade),
	}
	rank := JudgeHand(cards)
	if rank != Straight {
		t.Errorf("Straight (A-2-3-4-5) expected, got %v", rank)
	}
}

func TestHandRank_ThreeOfAKind(t *testing.T) {
	cards := []Card{
		NewCard(Ace, Spade),
		NewCard(Ace, Heart),
		NewCard(Ace, Diamond),
		NewCard(King, Spade),
		NewCard(Queen, Heart),
	}
	rank := JudgeHand(cards)
	if rank != ThreeOfAKind {
		t.Errorf("ThreeOfAKind expected, got %v", rank)
	}
}

func TestHandRank_TwoPair(t *testing.T) {
	cards := []Card{
		NewCard(Ace, Spade),
		NewCard(Ace, Heart),
		NewCard(King, Spade),
		NewCard(King, Heart),
		NewCard(Queen, Club),
	}
	rank := JudgeHand(cards)
	if rank != TwoPair {
		t.Errorf("TwoPair expected, got %v", rank)
	}
}

func TestHandRank_OnePair(t *testing.T) {
	cards := []Card{
		NewCard(Ace, Spade),
		NewCard(Ace, Heart),
		NewCard(King, Spade),
		NewCard(Queen, Club),
		NewCard(Jack, Heart),
	}
	rank := JudgeHand(cards)
	if rank != OnePair {
		t.Errorf("OnePair expected, got %v", rank)
	}
}

func TestHandRank_HighCard(t *testing.T) {
	cards := []Card{
		NewCard(Ace, Spade),
		NewCard(Jack, Heart),
		NewCard(Eight, Diamond),
		NewCard(Six, Club),
		NewCard(Three, Spade),
	}
	rank := JudgeHand(cards)
	if rank != HighCard {
		t.Errorf("HighCard expected, got %v", rank)
	}
}

func TestCompareHands(t *testing.T) {
	// 皇家同花顺 > 同花顺 > 四条 > 葫芦 > 同花 > 顺子 > 三条 > 两对 > 一对 > 高牌
	hands := []struct {
		cards []Card
		rank  HandRank
	}{
		{[]Card{NewCard(Ace, Spade), NewCard(King, Spade), NewCard(Queen, Spade), NewCard(Jack, Spade), NewCard(Ten, Spade)}, RoyalFlush},
		{[]Card{NewCard(Nine, Heart), NewCard(Eight, Heart), NewCard(Seven, Heart), NewCard(Six, Heart), NewCard(Five, Heart)}, StraightFlush},
		{[]Card{NewCard(Ace, Spade), NewCard(Ace, Heart), NewCard(Ace, Diamond), NewCard(Ace, Club), NewCard(King, Spade)}, FourOfAKind},
		{[]Card{NewCard(Ace, Spade), NewCard(Ace, Heart), NewCard(Ace, Diamond), NewCard(King, Spade), NewCard(King, Heart)}, FullHouse},
		{[]Card{NewCard(Ace, Spade), NewCard(Jack, Spade), NewCard(Eight, Spade), NewCard(Six, Spade), NewCard(Three, Spade)}, Flush},
		{[]Card{NewCard(Nine, Spade), NewCard(Eight, Heart), NewCard(Seven, Diamond), NewCard(Six, Club), NewCard(Five, Spade)}, Straight},
		{[]Card{NewCard(Ace, Spade), NewCard(Ace, Heart), NewCard(Ace, Diamond), NewCard(King, Spade), NewCard(Queen, Heart)}, ThreeOfAKind},
		{[]Card{NewCard(Ace, Spade), NewCard(Ace, Heart), NewCard(King, Spade), NewCard(King, Heart), NewCard(Queen, Club)}, TwoPair},
		{[]Card{NewCard(Ace, Spade), NewCard(Ace, Heart), NewCard(King, Spade), NewCard(Queen, Club), NewCard(Jack, Heart)}, OnePair},
		{[]Card{NewCard(Ace, Spade), NewCard(Jack, Heart), NewCard(Eight, Diamond), NewCard(Six, Club), NewCard(Three, Spade)}, HighCard},
	}

	for i := 0; i < len(hands)-1; i++ {
		for j := i + 1; j < len(hands); j++ {
			if hands[i].rank <= hands[j].rank {
				t.Errorf("Hand %d (%v) should beat Hand %d (%v)", i, hands[i].rank, j, hands[j].rank)
			}
		}
	}
}
