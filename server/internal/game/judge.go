package game

import "sort"

// HandRank 牌型大小
type HandRank int

const (
	HighCard      HandRank = 0
	OnePair       HandRank = 1
	TwoPair       HandRank = 2
	ThreeOfAKind  HandRank = 3
	Straight      HandRank = 4
	Flush         HandRank = 5
	FullHouse     HandRank = 6
	FourOfAKind   HandRank = 7
	StraightFlush HandRank = 8
	RoyalFlush    HandRank = 9
)

func (r HandRank) String() string {
	switch r {
	case RoyalFlush:
		return "皇家同花顺"
	case StraightFlush:
		return "同花顺"
	case FourOfAKind:
		return "四条"
	case FullHouse:
		return "葫芦"
	case Flush:
		return "同花"
	case Straight:
		return "顺子"
	case ThreeOfAKind:
		return "三条"
	case TwoPair:
		return "两对"
	case OnePair:
		return "一对"
	default:
		return "高牌"
	}
}

// HandResult 手牌判定结果
type HandResult struct {
	Rank      HandRank
	RankCards []Card // 用于比较的关键牌
	AllCards  []Card // 完整的5张牌
}

func (r HandResult) Beats(other HandResult) bool {
	if r.Rank != other.Rank {
		return r.Rank > other.Rank
	}
	// 比较关键牌
	for i := 0; i < len(r.RankCards) && i < len(other.RankCards); i++ {
		if r.RankCards[i].Rank != other.RankCards[i].Rank {
			return r.RankCards[i].Rank > other.RankCards[i].Rank
		}
	}
	return false
}

// EvaluateHand 判断手牌牌型，返回完整结果
func EvaluateHand(cards []Card) HandResult {
	if len(cards) < 5 {
		return HandResult{Rank: HighCard, AllCards: copyCards(cards)}
	}

	// 复制并按点数排序（从大到小）
	sorted := make([]Card, len(cards))
	copy(sorted, cards)
	sortByRankDesc(sorted)

	// 检查同花
	isFlush := isFlush(sorted)

	// 检查顺子
	isStraight, straightHighCard := isStraight(sorted)

	// 同花顺
	if isFlush && isStraight {
		result := HandResult{AllCards: copyCards(cards)}
		if straightHighCard == Ace {
			result.Rank = RoyalFlush
		} else {
			result.Rank = StraightFlush
		}
		result.RankCards = straightRankCards(sorted, straightHighCard)
		return result
	}

	// 统计每种点数出现次数
	countMap := make(map[Rank]int)
	for _, c := range sorted {
		countMap[c.Rank]++
	}

	// 提取计数并排序
	counts := make([]int, 0, len(countMap))
	rankList := make([]Rank, 0, len(countMap))
	for r, c := range countMap {
		counts = append(counts, c)
		rankList = append(rankList, r)
	}
	bubbleSortDescWithRanks(counts, rankList)

	result := HandResult{AllCards: copyCards(cards)}

	if counts[0] == 4 {
		result.Rank = FourOfAKind
		// 四条的关键牌是四条的点数
		for _, c := range sorted {
			if c.Rank == rankList[0] {
				result.RankCards = append(result.RankCards, c)
				if len(result.RankCards) == 4 {
					break
				}
			}
		}
		// 添加最大的边牌
		for _, c := range sorted {
			if c.Rank != rankList[0] {
				result.RankCards = append(result.RankCards, c)
				break
			}
		}
		return result
	}
	if counts[0] == 3 && counts[1] == 2 {
		result.Rank = FullHouse
		// 三条 + 对子
		for _, c := range sorted {
			if c.Rank == rankList[0] && len(result.RankCards) < 3 {
				result.RankCards = append(result.RankCards, c)
			}
		}
		for _, c := range sorted {
			if c.Rank == rankList[1] && len(result.RankCards) < 5 {
				result.RankCards = append(result.RankCards, c)
			}
		}
		return result
	}
	if isFlush {
		result.Rank = Flush
		result.RankCards = copyCards(sorted)
		return result
	}
	if isStraight {
		result.Rank = Straight
		result.RankCards = straightRankCards(sorted, straightHighCard)
		return result
	}
	if counts[0] == 3 {
		result.Rank = ThreeOfAKind
		// 三条的关键牌
		for _, c := range sorted {
			if c.Rank == rankList[0] {
				result.RankCards = append(result.RankCards, c)
				if len(result.RankCards) == 3 {
					break
				}
			}
		}
		// 添加最大的两张边牌
		cnt := 0
		for _, c := range sorted {
			if c.Rank != rankList[0] {
				result.RankCards = append(result.RankCards, c)
				cnt++
				if cnt == 2 {
					break
				}
			}
		}
		return result
	}
	if counts[0] == 2 && counts[1] == 2 {
		result.Rank = TwoPair
		// 两对的关键牌
		for _, c := range sorted {
			if c.Rank == rankList[0] && len(result.RankCards) < 2 {
				result.RankCards = append(result.RankCards, c)
			}
		}
		for _, c := range sorted {
			if c.Rank == rankList[1] && len(result.RankCards) < 4 {
				result.RankCards = append(result.RankCards, c)
			}
		}
		// 添加最大的边牌
		for _, c := range sorted {
			if c.Rank != rankList[0] && c.Rank != rankList[1] {
				result.RankCards = append(result.RankCards, c)
				break
			}
		}
		return result
	}
	if counts[0] == 2 {
		result.Rank = OnePair
		// 一对的关键牌
		for _, c := range sorted {
			if c.Rank == rankList[0] {
				result.RankCards = append(result.RankCards, c)
				if len(result.RankCards) == 2 {
					break
				}
			}
		}
		// 添加最大的三张边牌
		cnt := 0
		for _, c := range sorted {
			if c.Rank != rankList[0] {
				result.RankCards = append(result.RankCards, c)
				cnt++
				if cnt == 3 {
					break
				}
			}
		}
		return result
	}
	result.Rank = HighCard
	result.RankCards = copyCards(sorted)
	return result
}

// EvaluateBestHand 对任意数量的牌（如 2张手牌 + N张公共牌）找出最大牌型
func EvaluateBestHand(player *Player, community []Card) *HandResult {
	allCards := append(player.HoleCards, community...)
	if len(allCards) < 5 {
		if len(allCards) == 0 {
			return nil
		}
		// 不足5张也走 EvaluateHand 简化判定（主要返回高牌或对子）
		res := EvaluateHand(allCards)
		return &res
	}

	// 找出所有5张牌的组合
	var best HandResult
	hasBest := false
	combinations := getCombinations(len(allCards), 5)
	for _, combo := range combinations {
		cards := make([]Card, 5)
		for i, idx := range combo {
			cards[i] = allCards[idx]
		}
		result := EvaluateHand(cards)
		if !hasBest || result.Rank > best.Rank || (result.Rank == best.Rank && result.Beats(best)) {
			best = result
			best.AllCards = copyCards(cards)
			hasBest = true
		}
	}
	return &best
}

// JudgeHand 判断手牌牌型（兼容旧接口）
func JudgeHand(cards []Card) HandRank {
	return EvaluateHand(cards).Rank
}

// BestHandHighlightCards 返回用于高亮展示的牌（只包含牌型主体，不含边牌）
func BestHandHighlightCards(res HandResult) []Card {
	if len(res.AllCards) == 0 {
		return nil
	}

	switch res.Rank {
	case RoyalFlush, StraightFlush, Straight, Flush, FullHouse:
		return copyCards(res.AllCards)
	case FourOfAKind:
		return pickCardsByRanks(res.AllCards, topRanksByCount(res.AllCards, 4, 1))
	case ThreeOfAKind:
		return pickCardsByRanks(res.AllCards, topRanksByCount(res.AllCards, 3, 1))
	case TwoPair:
		return pickCardsByRanks(res.AllCards, topRanksByCount(res.AllCards, 2, 2))
	case OnePair:
		return pickCardsByRanks(res.AllCards, topRanksByCount(res.AllCards, 2, 1))
	case HighCard:
		sorted := copyCards(res.AllCards)
		sortByRankDesc(sorted)
		if len(sorted) > 0 {
			return []Card{sorted[0]}
		}
		return nil
	default:
		return copyCards(res.AllCards)
	}
}

func copyCards(cards []Card) []Card {
	if len(cards) == 0 {
		return nil
	}
	dup := make([]Card, len(cards))
	copy(dup, cards)
	return dup
}

func straightRankCards(sorted []Card, straightHighCard Rank) []Card {
	if len(sorted) == 0 {
		return nil
	}
	if straightHighCard != Five {
		return copyCards(sorted)
	}
	order := []Rank{Five, Four, Three, Deuce, Ace}
	result := make([]Card, 0, len(order))
	for _, r := range order {
		for _, c := range sorted {
			if c.Rank == r {
				result = append(result, c)
				break
			}
		}
	}
	return result
}

func topRanksByCount(cards []Card, count int, limit int) []Rank {
	counts := make(map[Rank]int)
	for _, c := range cards {
		counts[c.Rank]++
	}

	ranks := make([]Rank, 0, len(counts))
	for r, c := range counts {
		if c == count {
			ranks = append(ranks, r)
		}
	}
	if len(ranks) == 0 {
		return nil
	}
	if len(ranks) > 1 {
		sort.Slice(ranks, func(i, j int) bool {
			return ranks[i] > ranks[j]
		})
	}
	if limit > 0 && len(ranks) > limit {
		ranks = ranks[:limit]
	}
	return ranks
}

func pickCardsByRanks(cards []Card, ranks []Rank) []Card {
	if len(cards) == 0 || len(ranks) == 0 {
		return nil
	}
	set := make(map[Rank]struct{}, len(ranks))
	for _, r := range ranks {
		set[r] = struct{}{}
	}
	result := make([]Card, 0, len(cards))
	for _, c := range cards {
		if _, ok := set[c.Rank]; ok {
			result = append(result, c)
		}
	}
	return result
}

func isFlush(cards []Card) bool {
	suit := cards[0].Suit
	for _, c := range cards[1:5] {
		if c.Suit != suit {
			return false
		}
	}
	return true
}

func isStraight(cards []Card) (bool, Rank) {
	// 提取不重复的点数并排序（从大到小）
	uniqueRanks := make([]Rank, 0, 5)
	seen := make(map[Rank]bool)
	for _, c := range cards {
		if !seen[c.Rank] {
			uniqueRanks = append(uniqueRanks, c.Rank)
			seen[c.Rank] = true
		}
	}
	if len(uniqueRanks) < 5 {
		return false, 0
	}
	sort.Slice(uniqueRanks, func(i, j int) bool {
		return uniqueRanks[i] > uniqueRanks[j]
	})

	// 检查 A-2-3-4-5 顺子（wheel）
	if uniqueRanks[0] == Ace && uniqueRanks[1] == Five && uniqueRanks[2] == Four &&
		uniqueRanks[3] == Three && uniqueRanks[4] == Deuce {
		return true, Five
	}

	// 检查普通顺子：5张牌必须恰好相差4（即连续）
	if uniqueRanks[0]-uniqueRanks[4] == 4 {
		// 进一步验证确实是连续的
		isConsecutive := true
		for i := 0; i < 4; i++ {
			if uniqueRanks[i]-uniqueRanks[i+1] != 1 {
				isConsecutive = false
				break
			}
		}
		if isConsecutive {
			return true, uniqueRanks[0]
		}
	}

	return false, 0
}

func sortByRankDesc(cards []Card) {
	for i := 0; i < len(cards); i++ {
		for j := i + 1; j < len(cards); j++ {
			if cards[i].Rank < cards[j].Rank {
				cards[i], cards[j] = cards[j], cards[i]
			}
		}
	}
}

func bubbleSortDesc(arr []int) {
	for i := 0; i < len(arr); i++ {
		for j := i + 1; j < len(arr); j++ {
			if arr[i] < arr[j] {
				arr[i], arr[j] = arr[j], arr[i]
			}
		}
	}
}

func bubbleSortDescWithRanks(counts []int, ranks []Rank) {
	for i := 0; i < len(counts); i++ {
		for j := i + 1; j < len(counts); j++ {
			if counts[i] < counts[j] || (counts[i] == counts[j] && ranks[i] < ranks[j]) {
				counts[i], counts[j] = counts[j], counts[i]
				ranks[i], ranks[j] = ranks[j], ranks[i]
			}
		}
	}
}
