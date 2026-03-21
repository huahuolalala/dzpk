import { describe, it, expect } from 'vitest'

interface Player {
  id: string
  name: string
  seat: number
  status: 'active' | 'fold' | 'all-in'
}

interface TurnOrderState {
  players: Player[]
  currentPlayer: number  // index into players array
  dealerIndex: number
}

/**
 * 计算出牌顺序
 * 跳过已弃牌的玩家
 * 返回带编号的顺序列表
 */
function getTurnOrder(state: TurnOrderState): Array<{ index: number, player: Player, isCurrent: boolean }> {
  const { players, currentPlayer, dealerIndex } = state
  const result: Array<{ index: number, player: Player, isCurrent: boolean }> = []

  // 从庄家位置开始，按顺序遍历所有玩家
  const totalPlayers = players.length

  for (let i = 0; i < totalPlayers; i++) {
    const pos = (dealerIndex + i) % totalPlayers
    const player = players[pos]

    // 跳过已弃牌的玩家
    if (player.status === 'fold') {
      continue
    }

    result.push({
      index: result.length + 1,
      player,
      isCurrent: pos === currentPlayer,
    })
  }

  return result
}

describe('Turn Order Logic', () => {
  it('正常显示顺序，跳过已弃牌', () => {
    const state: TurnOrderState = {
      players: [
        { id: 'p1', name: '嘻嘻', seat: 0, status: 'active' },
        { id: 'p2', name: '花火', seat: 1, status: 'fold' },
        { id: 'p3', name: '土豆', seat: 2, status: 'active' },
      ],
      currentPlayer: 0,
      dealerIndex: 0,
    }

    const order = getTurnOrder(state)

    // 应该只有嘻嘻和土豆，序号为1和2
    expect(order.length).toBe(2)
    expect(order[0].player.name).toBe('嘻嘻')
    expect(order[0].index).toBe(1)
    expect(order[0].isCurrent).toBe(true)

    expect(order[1].player.name).toBe('土豆')
    expect(order[1].index).toBe(2)
    expect(order[1].isCurrent).toBe(false)
  })

  it('正确高亮当前出牌玩家', () => {
    const state: TurnOrderState = {
      players: [
        { id: 'p1', name: '嘻嘻', seat: 0, status: 'active' },
        { id: 'p2', name: '花火', seat: 1, status: 'active' },
        { id: 'p3', name: '土豆', seat: 2, status: 'active' },
      ],
      currentPlayer: 2,  // 土豆是当前玩家
      dealerIndex: 0,
    }

    const order = getTurnOrder(state)

    // 当前玩家应该是土豆
    const current = order.find(o => o.isCurrent)
    expect(current?.player.name).toBe('土豆')
  })

  it('从庄家位置开始显示顺序', () => {
    const state: TurnOrderState = {
      players: [
        { id: 'p1', name: '嘻嘻', seat: 0, status: 'active' },
        { id: 'p2', name: '花火', seat: 1, status: 'active' },
        { id: 'p3', name: '土豆', seat: 2, status: 'active' },
      ],
      currentPlayer: 1,
      dealerIndex: 1,  // 花火是庄家
    }

    const order = getTurnOrder(state)

    // 从庄家开始，顺序应该是: 花火(1), 土豆(2), 嘻嘻(3)
    expect(order.length).toBe(3)
    expect(order[0].player.name).toBe('花火')
    expect(order[0].index).toBe(1)
    expect(order[1].player.name).toBe('土豆')
    expect(order[1].index).toBe(2)
    expect(order[2].player.name).toBe('嘻嘻')
    expect(order[2].index).toBe(3)
  })

  it('只剩一个玩家时显示单个', () => {
    const state: TurnOrderState = {
      players: [
        { id: 'p1', name: '嘻嘻', seat: 0, status: 'fold' },
        { id: 'p2', name: '花火', seat: 1, status: 'fold' },
        { id: 'p3', name: '土豆', seat: 2, status: 'active' },
      ],
      currentPlayer: 2,
      dealerIndex: 0,
    }

    const order = getTurnOrder(state)

    expect(order.length).toBe(1)
    expect(order[0].player.name).toBe('土豆')
  })

  it('all-in玩家仍然显示在顺序中', () => {
    const state: TurnOrderState = {
      players: [
        { id: 'p1', name: '嘻嘻', seat: 0, status: 'all-in' },
        { id: 'p2', name: '花火', seat: 1, status: 'active' },
        { id: 'p3', name: '土豆', seat: 2, status: 'active' },
      ],
      currentPlayer: 1,
      dealerIndex: 0,
    }

    const order = getTurnOrder(state)

    // all-in 玩家仍然显示，但已弃牌的不显示
    expect(order.length).toBe(3)  // 嘻嘻仍然在（all-in不是fold）
  })
})
