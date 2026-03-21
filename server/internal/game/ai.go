package game

import (
	"math/rand"
	"time"
)

// AILevel AI等级
type AILevel string

const (
	AILevelEasy   AILevel = "easy"
	AILevelNormal AILevel = "normal"
	AILevelHard   AILevel = "hard"
)

// AIDecision AI决策结果
type AIDecision struct {
	Action ActionType
	Amount int64
}

// AIConfig AI配置
type AIConfig struct {
	Level         AILevel
	WinRateWeight float64 // 胜率权重
	PotOddsWeight float64 // 底池赔率权重
	Aggression    float64 // 激进程度 0-1
	BluffFactor   float64 // 诈唬因子
}

// DefaultAIConfig 默认AI配置
var DefaultAIConfig = map[AILevel]AIConfig{
	AILevelEasy: {
		Level:         AILevelEasy,
		WinRateWeight: 0.8,
		PotOddsWeight: 0.2,
		Aggression:    0.3,
		BluffFactor:   0.1,
	},
	AILevelNormal: {
		Level:         AILevelNormal,
		WinRateWeight: 0.7,
		PotOddsWeight: 0.3,
		Aggression:    0.5,
		BluffFactor:   0.2,
	},
	AILevelHard: {
		Level:         AILevelHard,
		WinRateWeight: 0.6,
		PotOddsWeight: 0.4,
		Aggression:    0.6,
		BluffFactor:   0.3,
	},
}

// CalculateWinRate 蒙特卡洛模拟计算胜率
// myCards: AI的手牌, community: 公共牌, numOpponents: 剩余对手数量, simulations: 模拟次数
func CalculateWinRate(myCards []Card, community []Card, numOpponents int, simulations int) float64 {
	if len(myCards) < 2 {
		return 0.5 // 无法计算，返回中等胜率
	}

	// 创建已知牌集合（排除已知牌）
	knownCards := make(map[Card]bool)
	for _, c := range myCards {
		knownCards[c] = true
	}
	for _, c := range community {
		knownCards[c] = true
	}

	// 创建剩余牌堆
	remainingDeck := make([]Card, 0, 52-len(knownCards))
	for s := Spade; s <= Club; s++ {
		for r := Deuce; r <= Ace; r++ {
			c := NewCard(r, s)
			if !knownCards[c] {
				remainingDeck = append(remainingDeck, c)
			}
		}
	}

	wins := 0
	ties := 0

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < simulations; i++ {
		// 洗牌
		deck := make([]Card, len(remainingDeck))
		copy(deck, remainingDeck)
		rng.Shuffle(len(deck), func(j, k int) {
			deck[j], deck[k] = deck[k], deck[j]
		})

		// 分配剩余公共牌
		remainingCommunity := 5 - len(community)
		simCommunity := make([]Card, len(community), 5)
		copy(simCommunity, community)
		if remainingCommunity > 0 && len(deck) >= remainingCommunity {
			simCommunity = append(simCommunity, deck[:remainingCommunity]...)
			deck = deck[remainingCommunity:]
		}

		// 分配对手手牌
		allHands := make([][]Card, 0, numOpponents+1)
		allHands = append(allHands, myCards)

		for j := 0; j < numOpponents; j++ {
			if len(deck) >= 2 {
				allHands = append(allHands, deck[:2])
				deck = deck[2:]
			}
		}

		// 评估所有手牌
		myResult := EvaluateHand(append(myCards, simCommunity...))
		myBest := findBestHand(myCards, simCommunity)

		won := true
		tied := false

		for j := 1; j < len(allHands); j++ {
			oppResult := EvaluateHand(append(allHands[j], simCommunity...))
			oppBest := findBestHand(allHands[j], simCommunity)

			if oppBest.Rank > myBest.Rank || (oppBest.Rank == myBest.Rank && oppBest.Beats(myBest)) {
				won = false
				break
			}
			if oppResult.Rank == myResult.Rank {
				// 比较关键牌
				if !myBest.Beats(oppBest) && !oppBest.Beats(myBest) {
					tied = true
				}
			}
		}

		if won {
			if tied {
				ties++
			} else {
				wins++
			}
		}
	}

	return (float64(wins) + float64(ties)*0.5) / float64(simulations)
}

// findBestHand 找出最佳手牌组合
func findBestHand(holeCards []Card, community []Card) HandResult {
	allCards := append(holeCards, community...)
	if len(allCards) < 5 {
		return EvaluateHand(allCards)
	}

	// 找出所有5张牌组合中的最佳牌型
	combinations := getCombinations(len(allCards), 5)
	var best HandResult
	hasBest := false

	for _, combo := range combinations {
		cards := make([]Card, 5)
		for i, idx := range combo {
			cards[i] = allCards[idx]
		}
		result := EvaluateHand(cards)
		if !hasBest || result.Rank > best.Rank || (result.Rank == best.Rank && result.Beats(best)) {
			best = result
			hasBest = true
		}
	}

	return best
}

// DecideAction AI决策
func DecideAction(gs *GameState, playerIndex int, level AILevel) AIDecision {
	player := gs.Players[playerIndex]
	if player.Status != Active {
		return AIDecision{Action: FoldAction}
	}

	// 获取AI配置
	config, ok := DefaultAIConfig[level]
	if !ok {
		config = DefaultAIConfig[AILevelNormal]
	}

	// 计算当前需要跟注的金额
	requiredToCall := gs.CurrentBet - player.Bet
	if requiredToCall < 0 {
		requiredToCall = 0
	}

	// 计算剩余活跃对手数量
	activeOpponents := 0
	for i, p := range gs.Players {
		if i != playerIndex && (p.Status == Active || p.Status == AllIn) {
			activeOpponents++
		}
	}

	// 蒙特卡洛模拟计算胜率
	simulations := 1000
	if level == AILevelHard {
		simulations = 3000
	}
	winRate := CalculateWinRate(player.HoleCards, gs.CommunityCards, activeOpponents, simulations)

	// 计算底池赔率
	potOdds := 0.0
	if requiredToCall > 0 {
		potOdds = float64(requiredToCall) / float64(gs.Pot+requiredToCall)
	}

	// 位置调整因子
	positionFactor := getPositionFactor(gs, playerIndex)

	// 综合评分
	score := calculateDecisionScore(winRate, potOdds, positionFactor, config)

	// 根据评分决策
	return makeDecision(score, requiredToCall, player.Chips, gs.CurrentBet, gs.MinRaise, gs.Pot, config)
}

// getPositionFactor 获取位置因子
func getPositionFactor(gs *GameState, playerIndex int) float64 {
	totalPlayers := len(gs.Players)
	if totalPlayers <= 2 {
		return 1.0
	}

	// 计算相对于庄家的位置
	distanceFromDealer := (playerIndex - gs.DealerIndex + totalPlayers) % totalPlayers

	// 庄家位置最激进
	if distanceFromDealer == 0 {
		return 1.2
	}

	// 盲位稍微保守
	if playerIndex == gs.SmallBlindIndex || playerIndex == gs.BigBlindIndex {
		return 0.9
	}

	// 中间位置
	return 1.0
}

// calculateDecisionScore 计算决策评分
func calculateDecisionScore(winRate, potOdds, positionFactor float64, config AIConfig) float64 {
	// 胜率调整
	adjustedWinRate := winRate * positionFactor

	// 底池赔率友好度
	potOddsScore := 1.0 - potOdds

	// 综合评分
	score := adjustedWinRate*config.WinRateWeight + potOddsScore*config.PotOddsWeight

	// 加入随机扰动（模拟不确定性）
	score += (rand.Float64() - 0.5) * 0.1 * config.BluffFactor

	return score
}

// makeDecision 根据评分做出决策
func makeDecision(score float64, requiredToCall, chips, currentBet, minRaise, pot int64, config AIConfig) AIDecision {
	// 如果可以免费看牌
	if requiredToCall == 0 {
		if score > 0.6 && rand.Float64() < config.Aggression {
			// 高胜率且激进，尝试加注
			raiseAmount := currentBet + minRaise
			if raiseAmount > chips+currentBet {
				raiseAmount = chips + currentBet // All-in
			}
			return AIDecision{Action: RaiseAction, Amount: raiseAmount}
		}
		return AIDecision{Action: CheckAction}
	}

	// 需要跟注的情况
	toCallRatio := float64(requiredToCall) / float64(chips)

	switch {
	case score > 0.7:
		// 高胜率
		if rand.Float64() < config.Aggression {
			// 激进加注
			raiseAmount := currentBet + minRaise + int64(float64(pot)*rand.Float64()*config.Aggression)
			if raiseAmount > chips+currentBet {
				raiseAmount = chips + currentBet
			}
			return AIDecision{Action: RaiseAction, Amount: raiseAmount}
		}
		return AIDecision{Action: CallAction}

	case score > 0.5:
		// 中等胜率
		if toCallRatio < 0.3 {
			// 跟注成本较低
			if rand.Float64() < config.Aggression*0.5 {
				raiseAmount := currentBet + minRaise
				if raiseAmount <= chips+currentBet {
					return AIDecision{Action: RaiseAction, Amount: raiseAmount}
				}
			}
			return AIDecision{Action: CallAction}
		}
		// 跟注成本较高
		if rand.Float64() < score {
			return AIDecision{Action: CallAction}
		}
		return AIDecision{Action: FoldAction}

	case score > 0.3:
		// 低胜率
		if toCallRatio < 0.15 && rand.Float64() < score {
			return AIDecision{Action: CallAction}
		}
		return AIDecision{Action: FoldAction}

	default:
		// 极低胜率
		if toCallRatio < 0.05 && rand.Float64() < 0.3 {
			return AIDecision{Action: CallAction}
		}
		return AIDecision{Action: FoldAction}
	}
}

// GetAIThinkDelay 获取AI思考延迟（毫秒）
func GetAIThinkDelay(level AILevel) time.Duration {
	switch level {
	case AILevelEasy:
		return time.Duration(500+rand.Intn(500)) * time.Millisecond
	case AILevelHard:
		return time.Duration(1000+rand.Intn(1000)) * time.Millisecond
	default:
		return time.Duration(800+rand.Intn(700)) * time.Millisecond
	}
}
