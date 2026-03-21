import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface GamePlayer {
  id: string
  name: string
  seat: number
  chips: number
  initialChips: number  // 游戏开始时的筹码
  bet: number
  status: 'active' | 'fold' | 'all-in'
  cards: string[]
  best_hand_rank?: string
  best_hand_cards?: string[]
}

export interface GameAction {
  player_name: string
  action: string
  amount: number
  phase: string
}

export interface DanmakuItem {
  id: string
  playerId: string
  playerName: string
  avatar: string
  content: string
  createdAt: number
}

export const useGameStore = defineStore('game', () => {
  const phase = ref('')
  const pot = ref(0)
  const communityCards = ref<string[]>([])
  const currentPlayer = ref(0)
  const currentBet = ref(0)
  const minRaise = ref(0)
  const smallBlind = ref(0)
  const bigBlind = ref(0)
  const dealerIndex = ref(0)
  const smallBlindIndex = ref(0)
  const bigBlindIndex = ref(0)
  const players = ref<GamePlayer[]>([])
  const actions = ref<GameAction[]>([])
  const result = ref<any>(null)
  const danmakuList = ref<DanmakuItem[]>([])

  function updateState(data: any) {
    phase.value = data.phase
    pot.value = data.pot
    communityCards.value = data.community_cards || []
    currentPlayer.value = data.current_player
    currentBet.value = data.current_bet || 0
    minRaise.value = data.min_raise || 0
    smallBlind.value = data.small_blind || 0
    bigBlind.value = data.big_blind || 0
    dealerIndex.value = data.dealer_index ?? 0
    smallBlindIndex.value = data.small_blind_index ?? 0
    bigBlindIndex.value = data.big_blind_index ?? 0

    if (data.players) {
      players.value = data.players.map((p: any) => {
        const existingPlayer = players.value.find(ep => ep.id === p.id)
        // 优先使用服务端下发的 initial_chips，其次使用已有值，最后才用当前 chips
        const initialChips = p.initial_chips ?? existingPlayer?.initialChips ?? p.chips

        return {
          id: p.id,
          name: p.name,
          seat: p.seat,
          chips: p.chips,
          initialChips: initialChips,
          bet: p.bet || 0,
          status: p.status,
          cards: p.cards || [],
          best_hand_rank: p.best_hand_rank,
          best_hand_cards: p.best_hand_cards || [],
        }
      })
    }

    // 更新操作历史
    if (data.actions) {
      actions.value = data.actions
    }
  }

  function reset() {
    phase.value = ''
    pot.value = 0
    communityCards.value = []
    currentPlayer.value = 0
    currentBet.value = 0
    minRaise.value = 0
    smallBlind.value = 0
    bigBlind.value = 0
    dealerIndex.value = 0
    smallBlindIndex.value = 0
    bigBlindIndex.value = 0
    players.value = []
    actions.value = []
    result.value = null
  }

  const isMyTurn = (playerId: string) => {
    return players.value[currentPlayer.value]?.id === playerId
  }

  function setResult(data: any) {
    result.value = data
  }

  function clearResult() {
    result.value = null
  }

  // 获取玩家的净盈亏
  function getNetChange(playerId: string): number {
    const player = players.value.find(p => p.id === playerId)
    if (!player) return 0
    return player.chips - player.initialChips
  }

  // 添加弹幕
  function addDanmaku(item: DanmakuItem) {
    danmakuList.value.push(item)
    // 限制最大弹幕数量
    if (danmakuList.value.length > 20) {
      danmakuList.value.shift()
    }
  }

  // 移除弹幕
  function removeDanmaku(id: string) {
    const index = danmakuList.value.findIndex(d => d.id === id)
    if (index !== -1) {
      danmakuList.value.splice(index, 1)
    }
  }

  return {
    phase,
    pot,
    communityCards,
    currentPlayer,
    currentBet,
    minRaise,
    smallBlind,
    bigBlind,
    dealerIndex,
    smallBlindIndex,
    bigBlindIndex,
    players,
    actions,
    result,
    danmakuList,
    updateState,
    reset,
    isMyTurn,
    setResult,
    clearResult,
    getNetChange,
    addDanmaku,
    removeDanmaku,
  }
})
