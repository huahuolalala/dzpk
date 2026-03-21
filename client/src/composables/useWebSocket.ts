import { ref, watch, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useRoomStore } from '../stores/room'
import { useGameStore } from '../stores/game'
import { useServerStore } from '../stores/server'

// 单例模式 - 整个应用共享一个 WebSocket 实例
let ws: WebSocket | null = null
let instanceCount = 0

export function useWebSocket() {
  const connected = ref(false)
  const error = ref('')
  const roomStore = useRoomStore()
  const gameStore = useGameStore()
  const serverStore = useServerStore()
  const router = useRouter()

  instanceCount++

  // 监听服务器地址变化，重连
  watch(() => serverStore.serverAddress, () => {
    if (connected.value) {
      disconnect()
      connect()
    }
  })

  function connect() {
    if (ws && ws.readyState === WebSocket.OPEN) {
      connected.value = true
      return
    }

    if (ws) {
      ws.close()
      ws = null
    }

    const wsUrl = serverStore.getWsUrl()
    console.log('[WS] Connecting to', wsUrl)

    try {
      ws = new WebSocket(wsUrl)

      ws.onopen = () => {
        connected.value = true
        error.value = ''
        console.log('[WS] Connected')
      }

      ws.onclose = () => {
        connected.value = false
        ws = null
        console.log('[WS] Disconnected')
      }

      ws.onerror = (e) => {
        error.value = '连接失败，请检查服务器地址'
        console.error('WebSocket error:', e)
      }

      ws.onmessage = (event) => {
        handleMessage(event.data)
      }
    } catch (e) {
      error.value = '连接失败，请检查服务器地址'
    }
  }

  function handleMessage(data: string) {
    try {
      const msg = JSON.parse(data)
      console.log('[WS] Received:', msg.type, msg.data)

      switch (msg.type) {
        case 'room_update':
          handleRoomUpdate(msg.data)
          break
        case 'game_state':
          gameStore.updateState(msg.data)
          if (roomStore.status === 'waiting') {
            router.push('/game')
          }
          roomStore.setStatus('playing')
          break
        case 'game_result':
          console.log('game_result:', msg.data)
          gameStore.setResult(msg.data)
          if (msg.data?.final_players) {
            const chipMap: Map<string, number> = new Map()
            for (const p of msg.data.final_players) {
              chipMap.set(p.id, p.chips)
            }
            roomStore.updatePlayers(
              roomStore.players.map(p => ({
                ...p,
                chips: chipMap.get(p.id) ?? p.chips,
              }))
            )
          }
          roomStore.setStatus('waiting')
          break
        case 'room_dismissed':
          roomStore.reset()
          gameStore.reset()
          break
        case 'error':
          error.value = msg.data.message
          break
      }
    } catch (e) {
      console.error('Failed to parse message:', e)
    }
  }

  function handleRoomUpdate(data: any) {
    console.log('[WS] room_update:', data)
    if (data.room_code) {
      // 新建房间或加入房间时
      roomStore.setRoom({
        roomCode: data.room_code,
        playerId: data.player_id,
        hostId: data.host_id,
        players: data.players,
      })
    } else if (data.action === 'player_joined' && data.players) {
      // 有人加入房间 - 更新完整玩家列表
      roomStore.updatePlayers(data.players)
      console.log('[WS] Player joined, total players:', data.players.length)
    } else if (data.action === 'player_left' && data.player_id) {
      roomStore.removePlayer(data.player_id)
      console.log('[WS] Player left:', data.player_id)
    }
  }

  function send(type: string, data: object) {
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      error.value = '未连接到服务器'
      throw new Error('WebSocket not connected')
    }
    const message = JSON.stringify({ type, data })
    ws.send(message)
    console.log('[WS] Sent:', type, data)
  }

  async function createRoom(playerName: string) {
    send('create_room', { player_name: playerName })
  }

  async function joinRoom(roomCode: string, playerName: string) {
    send('join_room', { room_code: roomCode, player_name: playerName })
  }

  async function leaveRoom() {
    try {
      send('leave_room', {})
    } catch (e) {
      // ignore
    }
    roomStore.reset()
    gameStore.reset()
  }

  async function startGame() {
    send('start_game', { room_code: roomStore.roomCode })
  }

  async function playerAction(action: string, amount?: number) {
    send('player_action', { action, amount })
  }

  function disconnect() {
    if (ws) {
      ws.close()
      ws = null
    }
  }

  onUnmounted(() => {
    instanceCount--
    // 只有当所有实例都卸载时才关闭连接
    if (instanceCount === 0) {
      disconnect()
    }
  })

  return {
    connected,
    error,
    connect,
    disconnect,
    createRoom,
    joinRoom,
    leaveRoom,
    startGame,
    playerAction,
  }
}
