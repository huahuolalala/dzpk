package game

// Suit 花色
type Suit int

const (
	Spade   Suit = 0
	Heart   Suit = 1
	Diamond Suit = 2
	Club    Suit = 3
)

func (s Suit) String() string {
	switch s {
	case Spade:
		return "♠"
	case Heart:
		return "♥"
	case Diamond:
		return "♦"
	case Club:
		return "♣"
	default:
		return "?"
	}
}

// Rank 点数
type Rank int

const (
	Deuce Rank = 2
	Three      = 3
	Four       = 4
	Five       = 5
	Six        = 6
	Seven      = 7
	Eight      = 8
	Nine       = 9
	Ten        = 10
	Jack       = 11
	Queen      = 12
	King       = 13
	Ace        = 14
)

func (r Rank) String() string {
	switch r {
	case Ace:
		return "A"
	case King:
		return "K"
	case Queen:
		return "Q"
	case Jack:
		return "J"
	case Ten:
		return "10"
	default:
		if r >= Deuce && r <= Nine {
			return string(rune('0' + r))
		}
		return "?"
	}
}

// Card 一张扑克牌
type Card struct {
	Rank Rank
	Suit Suit
}

func NewCard(rank Rank, suit Suit) Card {
	return Card{Rank: rank, Suit: suit}
}

func (c Card) String() string {
	return c.Rank.String() + c.Suit.String()
}

func (c Card) IsValid() bool {
	return c.Rank >= Deuce && c.Rank <= Ace && c.Suit >= Spade && c.Suit <= Club
}

// ParseCard 解析字符串为 Card，如 "As" -> Ace of Spades
func ParseCard(s string) (Card, bool) {
	if len(s) < 2 {
		return Card{}, false
	}

	var rank Rank
	var suit Suit

	// 解析花色
	switch s[len(s)-1] {
	case 's', 'S':
		suit = Spade
	case 'h', 'H':
		suit = Heart
	case 'd', 'D':
		suit = Diamond
	case 'c', 'C':
		suit = Club
	default:
		return Card{}, false
	}

	// 解析点数
	switch s[0] {
	case 'a', 'A':
		rank = Ace
	case 'k', 'K':
		rank = King
	case 'q', 'Q':
		rank = Queen
	case 'j', 'J':
		rank = Jack
	case 't', 'T':
		rank = Ten
	case '9':
		rank = Nine
	case '8':
		rank = Eight
	case '7':
		rank = Seven
	case '6':
		rank = Six
	case '5':
		rank = Five
	case '4':
		rank = Four
	case '3':
		rank = Three
	case '2':
		rank = Deuce
	default:
		return Card{}, false
	}

	return NewCard(rank, suit), true
}
