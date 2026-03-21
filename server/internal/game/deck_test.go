package game

import (
	"testing"
)

func TestNewDeck(t *testing.T) {
	deck := NewDeck()
	if len(deck.cards) != 52 {
		t.Errorf("NewDeck() should have 52 cards, got %d", len(deck.cards))
	}

	// 检查没有重复牌
	seen := make(map[Card]bool)
	for _, c := range deck.cards {
		if seen[c] {
			t.Errorf("Duplicate card found: %v", c)
		}
		seen[c] = true
	}
}

func TestDeck_Shuffle(t *testing.T) {
	deck1 := NewDeck()
	deck2 := NewDeck()

	// 洗牌后应该不同
	deck1.Shuffle()
	if deck1.cards[0] == deck2.cards[0] && deck1.cards[1] == deck2.cards[1] {
		// 极低概率发生，仅警告
		t.Log("Warning: shuffle may not be random")
	}
}

func TestDeck_Deal(t *testing.T) {
	deck := NewDeck()
	_, ok := deck.Deal()
	if !ok {
		t.Error("Deal() should return a card from non-empty deck")
	}
	if len(deck.cards) != 51 {
		t.Errorf("After Deal(), deck should have 51 cards, got %d", len(deck.cards))
	}

	// 从空牌堆取牌
	emptyDeck := &Deck{}
	_, ok = emptyDeck.Deal()
	if ok {
		t.Error("Deal() from empty deck should return ok=false")
	}
}

func TestDeck_DealAll(t *testing.T) {
	deck := NewDeck()
	cards := deck.DealAll(5)

	if len(cards) != 5 {
		t.Errorf("DealAll(5) should return 5 cards, got %d", len(cards))
	}
	if len(deck.cards) != 47 {
		t.Errorf("After DealAll(5), deck should have 47 cards, got %d", len(deck.cards))
	}

	// 请求超过剩余数量
	cards = deck.DealAll(50)
	if len(cards) != 47 {
		t.Errorf("DealAll(50) should return remaining 47 cards, got %d", len(cards))
	}
	if len(deck.cards) != 0 {
		t.Errorf("After DealAll, deck should be empty")
	}
}

func TestDeck_Reset(t *testing.T) {
	deck := NewDeck()
	deck.Deal()
	deck.Deal()
	deck.Deal()

	deck.Reset()

	if len(deck.cards) != 52 {
		t.Errorf("After Reset(), deck should have 52 cards, got %d", len(deck.cards))
	}
}
