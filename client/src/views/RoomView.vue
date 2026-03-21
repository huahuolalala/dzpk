<template>
  <div class="room-container">
    <header class="room-header glass-panel">
      <div class="room-info">
        <span class="label">当前房间</span>
        <span class="code">{{ roomStore.roomCode }}</span>
      </div>
      <button class="btn btn-secondary btn-sm" @click="handleLeave">退出房间</button>
    </header>

    <main class="room-content">
      <div class="players-section">
        <div class="section-header">
          <h2>入座玩家 <span class="count">({{ roomStore.players.length }}/9)</span></h2>
        </div>
        
        <div class="players-grid">
          <div
            v-for="player in roomStore.players"
            :key="player.id"
            class="player-card glass-panel"
            :class="{ 'is-host': player.id === roomStore.hostId, 'is-me': player.id === roomStore.playerId }"
          >
            <div class="host-badge" v-if="player.id === roomStore.hostId">房主</div>
            <div class="avatar">{{ player.name.charAt(0).toUpperCase() }}</div>
            <div class="player-details">
              <div class="name">{{ player.name }} <span v-if="player.id === roomStore.playerId" class="me-tag">(我)</span></div>
              <div class="chips">💰 {{ player.chips }}</div>
            </div>
          </div>
          
          <!-- 空座位占位 -->
          <div 
            v-for="i in Math.max(0, 9 - roomStore.players.length)" 
            :key="'empty-'+i" 
            class="player-card empty glass-panel"
          >
            <div class="avatar empty-avatar">?</div>
            <div class="player-details">
              <div class="name text-muted">虚位以待</div>
            </div>
          </div>
        </div>
      </div>

      <div class="action-section">
        <div v-if="roomStore.isHost()" class="host-controls glass-panel">
          <p class="hint" v-if="roomStore.players.length < 2">至少需要 2 名玩家才能开始游戏</p>
          <p class="hint success" v-else>人员已就位，可以开局</p>
          
          <button
            class="btn btn-primary btn-lg start-btn"
            @click="handleStart"
            :disabled="roomStore.players.length < 2"
          >
            开始游戏
          </button>
        </div>
        
        <div v-else class="guest-controls glass-panel">
          <div class="spinner"></div>
          <p>正在等待房主开始游戏...</p>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useRoomStore } from '../stores/room'
import { useWebSocket } from '../composables/useWebSocket'

const router = useRouter()
const roomStore = useRoomStore()
const { leaveRoom, startGame } = useWebSocket()

async function handleLeave() {
  await leaveRoom()
  router.push('/')
}

async function handleStart() {
  await startGame()
  router.push('/game')
}
</script>

<style scoped>
.room-container {
  min-height: 100vh;
  padding: 2rem 1.5rem;
  background: var(--bg-table);
  display: flex;
  flex-direction: column;
  align-items: center;
}

.room-header {
  width: 100%;
  max-width: 900px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  margin-bottom: 2.5rem;
}

.room-info {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.room-info .label {
  color: var(--text-muted);
  font-size: 0.875rem;
}

.room-info .code {
  font-size: 1.5rem;
  font-weight: 700;
  letter-spacing: 2px;
  color: var(--gold);
  text-shadow: 0 0 10px rgba(251, 191, 36, 0.3);
}

.room-content {
  width: 100%;
  max-width: 900px;
  display: flex;
  flex-direction: column;
  gap: 3rem;
}

.section-header h2 {
  font-size: 1.25rem;
  font-weight: 600;
  margin-bottom: 1.5rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.section-header .count {
  color: var(--text-muted);
  font-weight: 400;
  font-size: 1rem;
}

.players-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 1.25rem;
}

.player-card {
  position: relative;
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 1.25rem;
  border-radius: 12px;
  transition: transform 0.2s ease;
}

.player-card.is-me {
  border-color: rgba(16, 185, 129, 0.5);
  box-shadow: 0 0 15px rgba(16, 185, 129, 0.1);
}

.player-card.is-host {
  border-color: rgba(251, 191, 36, 0.4);
}

.player-card.empty {
  opacity: 0.5;
  border-style: dashed;
}

.avatar {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: linear-gradient(135deg, var(--primary) 0%, #047857 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  font-weight: 700;
  box-shadow: 0 4px 10px rgba(0,0,0,0.2);
}

.empty-avatar {
  background: rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.3);
  box-shadow: none;
}

.player-details {
  flex: 1;
}

.player-details .name {
  font-weight: 600;
  font-size: 1rem;
  margin-bottom: 0.25rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.me-tag {
  font-size: 0.75rem;
  color: var(--primary);
  font-weight: 500;
}

.text-muted {
  color: var(--text-muted) !important;
}

.player-details .chips {
  font-size: 0.875rem;
  color: var(--gold);
}

.host-badge {
  position: absolute;
  top: -8px;
  right: -8px;
  background: var(--gold);
  color: #451a03;
  font-size: 0.75rem;
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-weight: 700;
  box-shadow: 0 2px 8px rgba(251, 191, 36, 0.4);
}

.action-section {
  display: flex;
  justify-content: center;
}

.host-controls, .guest-controls {
  padding: 2rem 3rem;
  text-align: center;
  border-radius: 16px;
  width: 100%;
  max-width: 400px;
}

.guest-controls {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  color: var(--text-muted);
}

.host-controls .hint {
  margin-bottom: 1.5rem;
  color: var(--text-muted);
  font-size: 0.875rem;
}

.host-controls .hint.success {
  color: var(--primary);
}

.start-btn {
  width: 100%;
  font-size: 1.125rem;
  letter-spacing: 2px;
}
</style>
