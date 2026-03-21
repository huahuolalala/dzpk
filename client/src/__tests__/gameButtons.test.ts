import { describe, it, expect } from 'vitest'

// 按钮选项类型
type ActionButton = 'raise' | 'check' | 'call' | 'fold' | 'allin'

interface PlayerState {
  chips: number      // 玩家剩余筹码
  bet: number        // 本轮已下注
  status: 'active' | 'all-in' | 'folded'
}

interface GameButtonsState {
  currentBet: number   // 场上当前最大下注额
  myPlayer: PlayerState
  hasAllInPlayer: boolean  // 是否有其他玩家 all-in
}

/**
 * 计算当前玩家可以选择的按钮选项
 * 规则：
 * 1. 正常回合：加注、过牌/跟注、弃牌、All-in
 * 2. 当 currentBet > myPlayer.chips + myPlayer.bet 时，只能选 All-in 或 Fold
 */
function getAvailableButtons(state: GameButtonsState): ActionButton[] {
  const { currentBet, myPlayer } = state

  // 已 fold 的玩家不能操作
  if (myPlayer.status === 'folded') {
    return []
  }

  // 已 all-in 的玩家不能操作
  if (myPlayer.status === 'all-in') {
    return []
  }

  const toCall = Math.max(0, currentBet - myPlayer.bet)

  // 当最大下注额大于玩家手上筹码+已下筹码时，只能 All-in 或 Fold
  if (currentBet > myPlayer.chips + myPlayer.bet) {
    return ['fold', 'allin']
  }

  const buttons: ActionButton[] = []

  // 弃牌始终可选
  buttons.push('fold')

  // All-in 始终可选
  buttons.push('allin')

  // 过牌/跟注
  if (toCall === 0) {
    buttons.push('check')
  } else {
    buttons.push('call')
  }

  // 加注始终可选（但在模板中是通过弹窗实现）
  buttons.push('raise')

  return buttons
}

describe('Game Buttons Logic', () => {
  describe('正常回合 - 可以跟注', () => {
    it('toCall=0 时显示 过牌、加注、弃牌、All-in', () => {
      const state: GameButtonsState = {
        currentBet: 100,
        myPlayer: { chips: 900, bet: 100, status: 'active' },
        hasAllInPlayer: false,
      }
      const buttons = getAvailableButtons(state)
      expect(buttons).toContain('check')
      expect(buttons).toContain('raise')
      expect(buttons).toContain('fold')
      expect(buttons).toContain('allin')
      expect(buttons).not.toContain('call')
    })

    it('toCall>0 时显示 跟注、加注、弃牌、All-in', () => {
      const state: GameButtonsState = {
        currentBet: 300,
        myPlayer: { chips: 700, bet: 100, status: 'active' },
        hasAllInPlayer: false,
      }
      const buttons = getAvailableButtons(state)
      expect(buttons).toContain('call')
      expect(buttons).toContain('raise')
      expect(buttons).toContain('fold')
      expect(buttons).toContain('allin')
      expect(buttons).not.toContain('check')
    })

    it('有其他玩家 All-in 时也应正常显示所有选项', () => {
      const state: GameButtonsState = {
        currentBet: 300,
        myPlayer: { chips: 700, bet: 100, status: 'active' },
        hasAllInPlayer: true,  // 有其他玩家 all-in
      }
      const buttons = getAvailableButtons(state)
      // 有其他玩家 all-in 不影响自己的选项
      expect(buttons).toContain('call')
      expect(buttons).toContain('raise')
      expect(buttons).toContain('fold')
      expect(buttons).toContain('allin')
    })
  })

  describe('只能 All-in 或 Fold 的情况', () => {
    it('currentBet > chips + bet 时只能 fold 或 allin', () => {
      const state: GameButtonsState = {
        currentBet: 1000,
        myPlayer: { chips: 400, bet: 100, status: 'active' },  // chips+bet=500 < currentBet=1000
        hasAllInPlayer: false,
      }
      const buttons = getAvailableButtons(state)
      expect(buttons).toEqual(['fold', 'allin'])
      expect(buttons).not.toContain('check')
      expect(buttons).not.toContain('call')
      expect(buttons).not.toContain('raise')
    })

    it('正好等于时应该可以正常跟注', () => {
      const state: GameButtonsState = {
        currentBet: 500,
        myPlayer: { chips: 400, bet: 100, status: 'active' },  // chips+bet=500 == currentBet=500
        hasAllInPlayer: false,
      }
      const buttons = getAvailableButtons(state)
      // 能刚好跟注，应该显示正常选项
      expect(buttons).toContain('call')
      expect(buttons).toContain('raise')
      expect(buttons).toContain('fold')
      expect(buttons).toContain('allin')
    })
  })

  describe('玩家状态', () => {
    it('已 fold 的玩家返回空数组', () => {
      const state: GameButtonsState = {
        currentBet: 300,
        myPlayer: { chips: 500, bet: 0, status: 'folded' },
        hasAllInPlayer: false,
      }
      const buttons = getAvailableButtons(state)
      expect(buttons).toEqual([])
    })

    it('已 all-in 的玩家返回空数组', () => {
      const state: GameButtonsState = {
        currentBet: 300,
        myPlayer: { chips: 0, bet: 500, status: 'all-in' },
        hasAllInPlayer: false,
      }
      const buttons = getAvailableButtons(state)
      expect(buttons).toEqual([])
    })
  })
})
