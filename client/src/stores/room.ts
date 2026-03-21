import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface PlayerInfo {
  id: string
  name: string
  seat: number
  chips: number
  connected: boolean
  ready?: boolean
  status?: 'active' | 'fold' | 'all-in'
  cards?: string[]
  // AI 相关字段
  is_ai?: boolean
  ai_level?: string
}

export interface RoomState {
  roomCode: string
  playerId: string
  hostId: string
  players: PlayerInfo[]
  status: 'idle' | 'waiting' | 'playing'
}

export const useRoomStore = defineStore('room', () => {
  const roomCode = ref('')
  const playerId = ref('')
  const hostId = ref('')
  const players = ref<PlayerInfo[]>([])
  const status = ref<'idle' | 'waiting' | 'playing'>('idle')

  function setRoom(data: {
    roomCode: string
    playerId: string
    hostId?: string
    players?: PlayerInfo[]
  }) {
    roomCode.value = data.roomCode
    playerId.value = data.playerId
    if (data.hostId) hostId.value = data.hostId
    if (data.players) players.value = data.players
    status.value = 'waiting'
  }

  function updatePlayers(newPlayers: PlayerInfo[]) {
    players.value = newPlayers
  }

  function addPlayer(player: PlayerInfo) {
    players.value.push(player)
  }

  function removePlayer(playerId: string) {
    players.value = players.value.filter((p) => p.id !== playerId)
  }

  function setPlayerReady(playerId: string, ready: boolean) {
    const index = players.value.findIndex((p) => p.id === playerId)
    if (index !== -1) {
      const updatedPlayer = { ...players.value[index], ready }
      const newPlayers = [...players.value]
      newPlayers[index] = updatedPlayer
      players.value = newPlayers
    }
  }

  function setStatus(s: 'idle' | 'waiting' | 'playing') {
    status.value = s
  }

  function reset() {
    roomCode.value = ''
    playerId.value = ''
    hostId.value = ''
    players.value = []
    status.value = 'idle'
  }

  const isHost = () => playerId.value === hostId.value

  return {
    roomCode,
    playerId,
    hostId,
    players,
    status,
    setRoom,
    updatePlayers,
    addPlayer,
    removePlayer,
    setPlayerReady,
    setStatus,
    reset,
    isHost,
  }
})
