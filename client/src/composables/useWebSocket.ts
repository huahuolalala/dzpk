import { ref, watch, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useRoomStore } from '../stores/room'
import { useGameStore } from '../stores/game'
import { useServerStore } from '../stores/server'
import { useAuthStore } from '../stores/auth'

// 单例模式 - 整个应用共享一个 WebSocket 实例
let ws: WebSocket | null = null
let instanceCount = 0
let currentAuthToken = ''
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
let reconnectAttempts = 0
const MAX_RECONNECT_ATTEMPTS = 5

export function useWebSocket() {
  const connected = ref(false)
  const error = ref('')
  const roomStore = useRoomStore()
  const gameStore = useGameStore()
  const serverStore = useServerStore()
  const authStore = useAuthStore()
  const router = useRouter()

  instanceCount++

  // 监听服务器地址变化，重连
  watch(() => serverStore.serverAddress, () => {
    if (connected.value) {
      disconnect()
      connect()
    }
  })

  // 监听 token 变化，重新认证 WebSocket
  watch(() => authStore.token, (newToken) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      if (newToken && currentAuthToken !== newToken) {
        // Token 变了，发送 ws_login 重新认证
        console.log('[WS] Token changed, re-authenticating...')
        currentAuthToken = newToken
        send('ws_login', { token: newToken })
      } else if (!newToken && currentAuthToken) {
        // Token 被清除，标记为未认证
        console.log('[WS] Token cleared, disconnecting...')
        disconnect()
      }
    }
  })

  function connect() {
    // 如果 token 变了，断开旧连接重新建立
    if (ws && ws.readyState === WebSocket.OPEN && currentAuthToken !== authStore.token) {
      console.log('[WS] Token changed, reconnecting...')
      ws.close()
      ws = null
      currentAuthToken = ''
    }

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
        currentAuthToken = authStore.token
        reconnectAttempts = 0 // 重置重连计数
        console.log('[WS] Connected')

        // 如果已登录，发送 ws_login
        if (authStore.isAuthenticated) {
          send('ws_login', { token: authStore.token })
        }
      }

      ws.onclose = () => {
        connected.value = false
        console.log('[WS] Disconnected')
        // 自动重连（除非是用户主动断开）
        if (reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
          reconnectAttempts++
          console.log(`[WS] Reconnecting in 2s... (attempt ${reconnectAttempts}/${MAX_RECONNECT_ATTEMPTS})`)
          reconnectTimer = setTimeout(() => {
            ws = null
            connect()
          }, 2000)
        } else {
          console.log('[WS] Max reconnect attempts reached')
          error.value = '连接已断开，请刷新页面重试'
        }
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
        case 'ws_login_resp':
          handleWSLoginResp(msg.data)
          break
        case 'auth_required':
          error.value = '请先登录'
          break
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
          // 设置 phase 为 showdown 以触发摊牌动画
          gameStore.phase = 'showdown'
          gameStore.setResult(msg.data)
          if (msg.data?.final_players) {
            const chipMap: Map<string, number> = new Map()
            const initialChipsMap: Map<string, number> = new Map()
            for (const p of msg.data.final_players) {
              chipMap.set(p.id, p.chips)
            }
            if (msg.data?.initial_chips) {
              for (const [id, chips] of Object.entries(msg.data.initial_chips)) {
                initialChipsMap.set(id, chips as number)
              }
            }
            gameStore.players = gameStore.players.map(p => ({
              ...p,
              chips: chipMap.get(p.id) ?? p.chips,
              initialChips: initialChipsMap.get(p.id) ?? p.initialChips,
            }))
            roomStore.updatePlayers(
              roomStore.players.map(p => ({
                ...p,
                chips: chipMap.get(p.id) ?? p.chips,
              }))
            )
          }
          // 刷新用户筹码
          authStore.fetchStats()
          break
        case 'room_dismissed':
          roomStore.reset()
          gameStore.reset()
          router.push('/')
          break
        case 'host_left':
          // 房主离开，房间解散
          error.value = msg.data.message || '房主已离开，房间已解散'
          roomStore.reset()
          gameStore.reset()
          setTimeout(() => {
            router.push('/')
          }, 1500)
          break
        case 'chips_update':
          // 管理员修改了筹码，更新本地状态
          console.log('[WS] Chips updated:', msg.data)
          if (msg.data.user_id === authStore.user?.user_id) {
            authStore.updateChips(msg.data.chips)
          }
          break
        case 'player_ready':
          if (msg.data.player_id && msg.data.ready !== undefined) {
            roomStore.setPlayerReady(msg.data.player_id, msg.data.ready)
          }
          break
        case 'error':
          if (msg.data.code === 'not_authenticated') {
            error.value = '请先登录'
            authStore.logout()
            router.push('/login')
          } else {
            error.value = msg.data.message
          }
          break
        case 'chat':
          // 收到聊天消息，添加到弹幕列表
          if (msg.data) {
            gameStore.addDanmaku({
              id: Date.now().toString() + Math.random().toString(36).substr(2, 9),
              playerId: msg.data.player_id,
              playerName: msg.data.player_name,
              avatar: msg.data.avatar,
              content: msg.data.content,
              createdAt: Date.now(),
            })
          }
          break
      }
    } catch (e) {
      console.error('Failed to parse message:', e)
    }
  }

  function handleWSLoginResp(data: any) {
    console.log('[WS] Login resp:', data)
    // 更新用户信息
    if (data.user_id && data.chips !== undefined) {
      authStore.user = {
        user_id: data.user_id,
        username: data.username ?? authStore.user?.username ?? '',
        nickname: data.nickname,
        avatar: data.avatar,
        chips: data.chips,
      }
      if (data.stats) {
        authStore.stats = data.stats
      }
    }
  }

  function handleRoomUpdate(data: any) {
    console.log('[WS] room_update:', data)
    if (data.action === 'game_ended' && data.players) {
      // 游戏结束 - 只更新玩家状态，不跳转
      // 让 GameView 处理结算流程（showdown 动画 + 结果弹窗）
      const updatedPlayers = data.players.map((p: any) => ({
        ...p,
        ready: p.ready ?? false
      }))
      roomStore.updatePlayers(updatedPlayers)
      roomStore.setStatus('waiting')
      // 注意：不在这里跳转，让 GameView 的 showdown 流程处理
      console.log('[WS] Game ended, waiting for GameView to show results')
    } else if (data.room_code) {
      // 新建房间或加入房间时
      roomStore.setRoom({
        roomCode: data.room_code,
        playerId: data.player_id,
        hostId: data.host_id,
        players: data.players,
      })
    } else if (data.action === 'sync_players' && data.players) {
      roomStore.updatePlayers(data.players)
      console.log('[WS] Players synced, total players:', data.players.length)
    } else if (data.action === 'player_joined' && data.players) {
      // 有人加入房间 - 更新完整玩家列表
      roomStore.updatePlayers(data.players)
      console.log('[WS] Player joined, total players:', data.players.length)
    } else if (data.action === 'player_left' && data.player_id) {
      // 如果有完整的 players 列表，更新整个列表；否则只移除离开的玩家
      if (data.players) {
        roomStore.updatePlayers(data.players)
      } else {
        roomStore.removePlayer(data.player_id)
      }
      console.log('[WS] Player left:', data.player_id)
    }
  }

  function send(type: string, data: object) {
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      error.value = '未连接到服务器，正在重连...'
      // 触发重连
      if (reconnectTimer) {
        clearTimeout(reconnectTimer)
        reconnectTimer = null
      }
      reconnectAttempts = 0
      ws = null
      connect()
      return
    }
    const message = JSON.stringify({ type, data })
    ws.send(message)
    console.log('[WS] Sent:', type, data)
  }

  async function createRoom(_playerName: string) {
    // 已认证用户不需要传 playerName
    send('create_room', {})
  }

  async function joinRoom(roomCode: string, _playerName: string) {
    // 已认证用户不需要传 playerName
    send('join_room', { room_code: roomCode })
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

  async function playerReady() {
    send('ready', {})
  }

  async function playerAction(action: string, amount?: number) {
    send('player_action', { action, amount })
  }

  async function createAIRoom(aiLevel: string) {
    send('create_ai_room', { ai_level: aiLevel })
  }

  async function sendChat(content: string) {
    send('chat', { content })
  }

  function disconnect() {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    reconnectAttempts = MAX_RECONNECT_ATTEMPTS // 阻止自动重连
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
    playerReady,
    playerAction,
    createAIRoom,
    sendChat,
  }
}
