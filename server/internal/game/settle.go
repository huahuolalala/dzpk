package game

import "sort"

// WinnerInfo 赢家信息
type WinnerInfo struct {
	PlayerID   string
	PlayerName string
	HandRank   HandRank
	WinAmount  int64
}

type sidePot struct {
	Amount     int64
	Contenders []*Player
}

// DetermineWinners 判断赢家并分配底池
func DetermineWinners(game *GameState) []WinnerInfo {
	if len(game.Players) == 0 {
		return nil
	}

	activePlayers := collectActivePlayers(game)
	if len(activePlayers) == 0 {
		return nil
	}

	handResults := buildHandResults(activePlayers, game.CommunityCards)
	if len(handResults) == 0 {
		return nil
	}

	winAmounts := make(map[string]int64)

	pots := buildSidePots(game)
	if len(pots) == 0 && game.Pot > 0 {
		pots = []sidePot{{Amount: game.Pot, Contenders: activePlayers}}
	}

	for _, pot := range pots {
		if pot.Amount == 0 || len(pot.Contenders) == 0 {
			continue
		}
		winners := pickWinners(pot.Contenders, handResults)
		if len(winners) == 0 {
			continue
		}
		distributePot(game, pot.Amount, winners, winAmounts)
	}

	results := make([]WinnerInfo, 0, len(winAmounts))
	for _, p := range game.Players {
		amount, ok := winAmounts[p.ID]
		if !ok || amount == 0 {
			continue
		}
		res := handResults[p.ID]
		if res == nil {
			continue
		}
		results = append(results, WinnerInfo{
			PlayerID:   p.ID,
			PlayerName: p.Name,
			HandRank:   res.Rank,
			WinAmount:  amount,
		})
	}

	return results
}

func collectActivePlayers(game *GameState) []*Player {
	activePlayers := make([]*Player, 0, len(game.Players))
	for _, p := range game.Players {
		if p.Status == Active || p.Status == AllIn {
			activePlayers = append(activePlayers, p)
		}
	}
	return activePlayers
}

func buildHandResults(players []*Player, community []Card) map[string]*HandResult {
	results := make(map[string]*HandResult, len(players))
	for _, p := range players {
		res := EvaluateBestHand(p, community)
		if res != nil {
			results[p.ID] = res
		}
	}
	return results
}

func buildSidePots(game *GameState) []sidePot {
	levels := make([]int64, 0, len(game.Players))
	for _, p := range game.Players {
		if p.TotalBet > 0 {
			levels = append(levels, p.TotalBet)
		}
	}
	if len(levels) == 0 {
		return nil
	}

	sort.Slice(levels, func(i, j int) bool { return levels[i] < levels[j] })
	uniqueLevels := make([]int64, 0, len(levels))
	var last int64
	for i, level := range levels {
		if i == 0 || level != last {
			uniqueLevels = append(uniqueLevels, level)
			last = level
		}
	}

	prev := int64(0)
	pots := make([]sidePot, 0, len(uniqueLevels))
	for _, level := range uniqueLevels {
		if level <= prev {
			continue
		}
		count := 0
		contenders := make([]*Player, 0, len(game.Players))
		for _, p := range game.Players {
			if p.TotalBet >= level {
				count++
				if p.Status == Active || p.Status == AllIn {
					contenders = append(contenders, p)
				}
			}
		}
		if count == 0 {
			prev = level
			continue
		}
		amount := (level - prev) * int64(count)
		pots = append(pots, sidePot{Amount: amount, Contenders: contenders})
		prev = level
	}

	return pots
}

func pickWinners(contenders []*Player, handResults map[string]*HandResult) []*Player {
	var best *HandResult
	winners := make([]*Player, 0, len(contenders))
	for _, p := range contenders {
		res := handResults[p.ID]
		if res == nil {
			continue
		}
		if best == nil || res.Beats(*best) {
			best = res
			winners = []*Player{p}
			continue
		}
		if res.Rank == best.Rank && equalRanks(res, best) {
			winners = append(winners, p)
		}
	}
	return winners
}

func distributePot(game *GameState, potAmount int64, winners []*Player, winAmounts map[string]int64) {
	if potAmount == 0 || len(winners) == 0 {
		return
	}
	base := potAmount / int64(len(winners))
	remainder := potAmount % int64(len(winners))
	for _, p := range winners {
		winAmounts[p.ID] += base
		p.Chips += base
	}
	if remainder == 0 {
		return
	}
	winnerSet := make(map[string]struct{}, len(winners))
	for _, p := range winners {
		winnerSet[p.ID] = struct{}{}
	}
	ordered := orderWinnersBySeat(game.Players, winnerSet)
	for i := int64(0); i < remainder; i++ {
		p := ordered[i%int64(len(ordered))]
		winAmounts[p.ID] += 1
		p.Chips += 1
	}
}

func orderWinnersBySeat(players []*Player, winners map[string]struct{}) []*Player {
	ordered := make([]*Player, 0, len(winners))
	for _, p := range players {
		if _, ok := winners[p.ID]; ok {
			ordered = append(ordered, p)
		}
	}
	return ordered
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
