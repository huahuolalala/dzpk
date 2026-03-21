package game

import (
	"testing"
)

func TestCard_NewCard(t *testing.T) {
	cases := []struct {
		rank Rank
		suit Suit
		want string
	}{
		{Ace, Spade, "A♠"},
		{King, Heart, "K♥"},
		{Queen, Diamond, "Q♦"},
		{Jack, Club, "J♣"},
		{Ten, Spade, "10♠"},
		{Seven, Heart, "7♥"},
		{Deuce, Club, "2♣"},
	}

	for _, c := range cases {
		card := NewCard(c.rank, c.suit)
		if card.String() != c.want {
			t.Errorf("NewCard(%v, %v) = %v, want %v", c.rank, c.suit, card.String(), c.want)
		}
	}
}

func TestRank_String(t *testing.T) {
	cases := []struct {
		rank Rank
		want string
	}{
		{Ace, "A"},
		{King, "K"},
		{Queen, "Q"},
		{Jack, "J"},
		{Ten, "10"},
		{Nine, "9"},
		{Deuce, "2"},
	}

	for _, c := range cases {
		if c.rank.String() != c.want {
			t.Errorf("Rank(%v).String() = %v, want %v", c.rank, c.rank.String(), c.want)
		}
	}
}

func TestSuit_String(t *testing.T) {
	cases := []struct {
		suit Suit
		want string
	}{
		{Spade, "♠"},
		{Heart, "♥"},
		{Diamond, "♦"},
		{Club, "♣"},
	}

	for _, c := range cases {
		if c.suit.String() != c.want {
			t.Errorf("Suit(%v).String() = %v, want %v", c.suit, c.suit.String(), c.want)
		}
	}
}

func TestCard_IsValid(t *testing.T) {
	if !NewCard(Ace, Spade).IsValid() {
		t.Error("Ace of Spades should be valid")
	}
	if NewCard(Rank(15), Spade).IsValid() {
		t.Error("Rank 15 should be invalid")
	}
	if NewCard(Ace, Suit(4)).IsValid() {
		t.Error("Suit 4 should be invalid")
	}
}

func TestParseCard(t *testing.T) {
	cases := []struct {
		input  string
		wantOK bool
		want   Card
	}{
		{"As", true, NewCard(Ace, Spade)},
		{"Kh", true, NewCard(King, Heart)},
		{"Qd", true, NewCard(Queen, Diamond)},
		{"Jc", true, NewCard(Jack, Club)},
		{"Ts", true, NewCard(Ten, Spade)},
		{"9h", true, NewCard(Nine, Heart)},
		{"2c", true, NewCard(Deuce, Club)},
		{"AX", false, Card{}},
		{"", false, Card{}},
	}

	for _, c := range cases {
		card, ok := ParseCard(c.input)
		if ok != c.wantOK {
			t.Errorf("ParseCard(%q) ok = %v, want %v", c.input, ok, c.wantOK)
		}
		if ok && card != c.want {
			t.Errorf("ParseCard(%q) = %v, want %v", c.input, card, c.want)
		}
	}
}
