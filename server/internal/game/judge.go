package game

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
		return HandResult{Rank: HighCard, AllCards: cards}
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
		result := HandResult{AllCards: cards[:5]}
		if straightHighCard == Ace {
			result.Rank = RoyalFlush
			result.RankCards = cards[:5]
		} else {
			result.Rank = StraightFlush
			result.RankCards = cards[:5]
		}
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

	result := HandResult{AllCards: cards[:5]}

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
		result.RankCards = cards[:5]
		return result
	}
	if isStraight {
		result.Rank = Straight
		result.RankCards = cards[:5]
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
	result.RankCards = cards[:5]
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
	combinations := getCombinations(len(allCards), 5)
	for _, combo := range combinations {
		cards := make([]Card, 5)
		for i, idx := range combo {
			cards[i] = allCards[idx]
		}
		result := EvaluateHand(cards)
		if result.Rank > best.Rank {
			best = result
			best.AllCards = cards
		} else if result.Rank == best.Rank && result.Beats(best) {
			best = result
			best.AllCards = cards
		}
	}
	return &best
}

// JudgeHand 判断手牌牌型（兼容旧接口）
func JudgeHand(cards []Card) HandRank {
	return EvaluateHand(cards).Rank
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
	// 检查普通顺子 A-K-Q-J-10 到 5-4-3-2-A
	if cards[0].Rank-cards[4].Rank == 4 {
		return true, cards[0].Rank
	}
	// 检查 A-2-3-4-5 顺子
	if cards[0].Rank == Ace && cards[1].Rank == Five && cards[2].Rank == Four &&
		cards[3].Rank == Three && cards[4].Rank == Deuce {
		return true, Five // A-2-3-4-5 的高牌是5
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
			if counts[i] < counts[j] {
				counts[i], counts[j] = counts[j], counts[i]
				ranks[i], ranks[j] = ranks[j], ranks[i]
			}
		}
	}
}
