package game

// WinnerInfo 赢家信息
type WinnerInfo struct {
	PlayerID   string
	PlayerName string
	HandRank   HandRank
	WinAmount  int64
}

// DetermineWinners 判断赢家并分配底池
func DetermineWinners(game *GameState) []WinnerInfo {
	if len(game.Players) == 0 {
		return nil
	}

	// 找出还有筹码的活跃玩家
	activePlayers := make([]*Player, 0)
	for _, p := range game.Players {
		if p.Status == Active || p.Status == AllIn {
			activePlayers = append(activePlayers, p)
		}
	}

	if len(activePlayers) == 0 {
		return nil
	}

	// 计算每人的最佳手牌
	bestResults := make([]*HandResult, len(activePlayers))
	for i, p := range activePlayers {
		bestResults[i] = EvaluateBestHand(p, game.CommunityCards)
	}

	// 找出最大牌型
	var bestResult *HandResult
	for _, r := range bestResults {
		if bestResult == nil || r.Beats(*bestResult) {
			bestResult = r
		}
	}

	// 统计赢家
	winners := make([]int, 0)
	for i, r := range bestResults {
		if r.Rank == bestResult.Rank && equalRanks(r, bestResult) {
			winners = append(winners, i)
		}
	}

	// 分配底池
	winAmount := game.Pot / int64(len(winners))
	results := make([]WinnerInfo, 0, len(winners))
	for _, idx := range winners {
		p := activePlayers[idx]
		results = append(results, WinnerInfo{
			PlayerID:   p.ID,
			PlayerName: p.Name,
			HandRank:   bestResults[idx].Rank,
			WinAmount:  winAmount,
		})
		p.Chips += winAmount
	}

	return results
}

func getCombinations(n, k int) [][]int {
	if k == 0 || k > n {
		return nil
	}
	if k == n {
		result := make([][]int, 1)
		result[0] = make([]int, n)
		for i := 0; i < n; i++ {
			result[0][i] = i
		}
		return result
	}

	var results [][]int
	// 递归生成组合
	var backtrack func(start int, combo []int)
	backtrack = func(start int, combo []int) {
		if len(combo) == k {
			tmp := make([]int, k)
			copy(tmp, combo)
			results = append(results, tmp)
			return
		}
		for i := start; i < n; i++ {
			combo = append(combo, i)
			backtrack(i+1, combo)
			combo = combo[:len(combo)-1]
		}
	}
	backtrack(0, nil)
	return results
}

func equalRanks(a, b *HandResult) bool {
	if len(a.RankCards) != len(b.RankCards) {
		return false
	}
	for i := range a.RankCards {
		if a.RankCards[i].Rank != b.RankCards[i].Rank {
			return false
		}
	}
	return true
}

// CalculatePots 计算主池和边池
func CalculatePots(game *GameState) (sidePots []int64, mainPot int64) {
	if len(game.Players) == 0 {
		return nil, 0
	}

	// 找出最小下注
	minBet := game.Players[0].Bet
	for _, p := range game.Players {
		if p.Bet < minBet {
			minBet = p.Bet
		}
	}

	// 主池 = 最小下注 × 玩家数
	mainPot = minBet * int64(len(game.Players))

	// 边池计算（简化版，暂不处理复杂边池）
	// 实际实现需要跟踪每个玩家的实际投入
	return nil, mainPot
}
