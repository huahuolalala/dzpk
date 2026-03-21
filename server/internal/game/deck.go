package game

import "math/rand"

type Deck struct {
	cards []Card
}

func NewDeck() *Deck {
	cards := make([]Card, 0, 52)
	for s := Spade; s <= Club; s++ {
		for r := Deuce; r <= Ace; r++ {
			cards = append(cards, NewCard(r, s))
		}
	}
	return &Deck{cards: cards}
}

func (d *Deck) Shuffle() {
	for i := len(d.cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
	}
}

func (d *Deck) Deal() (Card, bool) {
	if len(d.cards) == 0 {
		return Card{}, false
	}
	card := d.cards[len(d.cards)-1]
	d.cards = d.cards[:len(d.cards)-1]
	return card, true
}

func (d *Deck) DealAll(n int) []Card {
	if n > len(d.cards) {
		n = len(d.cards)
	}
	cards := make([]Card, n)
	for i := 0; i < n; i++ {
		cards[i], _ = d.Deal()
	}
	return cards
}

func (d *Deck) Reset() {
	*d = *NewDeck()
}
