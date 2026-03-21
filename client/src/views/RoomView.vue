<template>
  <div class="room-container">
    <header class="room-header glass-panel">
      <div class="room-info">
        <span class="label">当前房间</span>
        <span class="code">{{ roomStore.roomCode }}</span>
      </div>
      <div class="header-actions">
        <button class="btn btn-secondary btn-sm" @click="tutorial?.open()">📖 教程</button>
        <button class="btn btn-secondary btn-sm" @click="handleLeave">退出房间</button>
      </div>
    </header>

    <!-- 游戏教程 -->
    <GameTutorial ref="tutorial" />

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
            :class="{ 'is-host': player.id === roomStore.hostId, 'is-me': player.id === roomStore.playerId, 'is-ready': player.ready, 'is-ai': player.is_ai }"
          >
            <div class="host-badge" v-if="player.id === roomStore.hostId">房主</div>
            <div class="ai-badge" v-if="player.is_ai">🤖 AI</div>
            <div class="avatar">
              <span v-if="player.ready" class="avatar-check">✓</span>
              <span v-else-if="player.is_ai">🤖</span>
              <span v-else>{{ player.name.charAt(0).toUpperCase() }}</span>
            </div>
            <div class="player-details">
              <div class="name">
                {{ player.name }}
                <span v-if="player.id === roomStore.playerId" class="me-tag">(我)</span>
                <span v-if="player.is_ai" class="ai-tag">{{ getAILevelLabel(player.ai_level || '') }}</span>
                <span v-if="player.ready && !player.is_ai" class="ready-tag">已准备</span>
              </div>
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
        <!-- 非房主：准备/取消准备按钮 -->
        <div v-if="!roomStore.isHost()" class="guest-controls glass-panel">
          <button
            v-if="!isMyReady"
            class="btn btn-primary btn-lg ready-btn"
            @click="handleReady"
          >
            准备游戏
          </button>
          <button
            v-else
            class="btn btn-secondary btn-lg ready-btn"
            @click="handleReady"
          >
            取消准备
          </button>
        </div>

        <!-- 房主：开始游戏按钮 -->
        <div v-else class="host-controls glass-panel">
          <div class="ready-status host-ready">
            <span class="ready-icon">✓</span>
            <span>房主已准备</span>
          </div>
          <p class="hint" v-if="roomStore.players.length < 2">至少需要 2 名玩家才能开始游戏</p>
          <p class="hint" v-else-if="!allPlayersReady">等待所有玩家准备...</p>
          <p class="hint success" v-else>所有玩家已准备，可以开局</p>

          <button
            class="btn btn-primary btn-lg start-btn"
            @click="handleStart"
            :disabled="roomStore.players.length < 2 || !allPlayersReady"
          >
            开始游戏
          </button>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useRoomStore } from '../stores/room'
import { useWebSocket } from '../composables/useWebSocket'
import GameTutorial from '../components/GameTutorial.vue'

const router = useRouter()
const roomStore = useRoomStore()
const { leaveRoom, startGame, playerReady } = useWebSocket()
const tutorial = ref<InstanceType<typeof GameTutorial> | null>(null)

const isMyReady = computed(() => {
  const me = roomStore.players.find(p => p.id === roomStore.playerId)
  return me?.ready ?? false
})

const allPlayersReady = computed(() => {
  if (roomStore.players.length < 2) return false
  return roomStore.players.every(p => p.ready)
})

async function handleLeave() {
  await leaveRoom()
  router.push('/')
}

async function handleStart() {
  await startGame()
}

async function handleReady() {
  await playerReady()
}

function getAILevelLabel(level: string): string {
  switch (level) {
    case 'easy': return '简单'
    case 'hard': return '困难'
    default: return '普通'
  }
}
</script>

<style scoped>
.room-container {
  min-height: 100vh;
  padding: 1rem;
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
  padding: 0.75rem 1.25rem;
  margin-bottom: 1rem;
}

.room-info {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.header-actions {
  display: flex;
  gap: 0.75rem;
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
  gap: 1.5rem;
  flex: 1;
  overflow: hidden;
}

.players-section {
  max-height: 55vh;
  overflow-y: auto;
  overflow-x: hidden;
}

/* 自定义滚动条 */
.players-section::-webkit-scrollbar {
  width: 6px;
}

.players-section::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 3px;
}

.players-section::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 3px;
}

.players-section::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
}

.section-header h2 {
  font-size: 1.125rem;
  font-weight: 600;
  margin-bottom: 1rem;
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
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 0.75rem;
}

.player-card {
  position: relative;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.6rem 0.75rem;
  border-radius: 10px;
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
  opacity: 0.35;
  border-style: dashed;
  padding: 0.5rem 0.75rem;
}

.avatar {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: linear-gradient(135deg, var(--primary) 0%, #047857 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.125rem;
  font-weight: 700;
  box-shadow: 0 4px 10px rgba(0,0,0,0.2);
  flex-shrink: 0;
}

.avatar-check {
  font-size: 1.5rem;
  font-weight: 700;
  color: white;
}

.player-card.is-ready .avatar {
  background: linear-gradient(135deg, var(--success, #10b981) 0%, #059669 100%);
  box-shadow: 0 0 15px rgba(16, 185, 129, 0.4);
}

.player-card.is-ready {
  border-color: rgba(16, 185, 129, 0.5);
}

.empty-avatar {
  background: rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.3);
  box-shadow: none;
  width: 28px;
  height: 28px;
  font-size: 0.875rem;
}

.player-details {
  flex: 1;
}

.player-details .name {
  font-weight: 600;
  font-size: 0.875rem;
  margin-bottom: 0.125rem;
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.me-tag {
  font-size: 0.65rem;
  color: var(--primary);
  font-weight: 500;
}

.ready-tag {
  font-size: 0.6rem;
  color: white;
  background: var(--success, #10b981);
  padding: 0.1rem 0.35rem;
  border-radius: 8px;
  font-weight: 600;
  margin-left: 0.25rem;
}

.ai-tag {
  font-size: 0.6rem;
  color: white;
  background: #6366f1;
  padding: 0.1rem 0.35rem;
  border-radius: 8px;
  font-weight: 600;
  margin-left: 0.25rem;
}

.ai-badge {
  position: absolute;
  top: -6px;
  right: -6px;
  background: #6366f1;
  color: white;
  font-size: 0.6rem;
  padding: 0.15rem 0.4rem;
  border-radius: 8px;
  font-weight: 700;
  box-shadow: 0 2px 6px rgba(99, 102, 241, 0.4);
}

.player-card.is-ai {
  border-color: rgba(99, 102, 241, 0.4);
}

.player-card.is-ai .avatar {
  background: linear-gradient(135deg, #6366f1 0%, #4f46e5 100%);
}

.text-muted {
  color: var(--text-muted) !important;
}

.player-details .chips {
  font-size: 0.75rem;
  color: var(--gold);
}

.host-badge {
  position: absolute;
  top: -6px;
  right: -6px;
  background: var(--gold);
  color: #451a03;
  font-size: 0.65rem;
  padding: 0.15rem 0.5rem;
  border-radius: 8px;
  font-weight: 700;
  box-shadow: 0 2px 6px rgba(251, 191, 36, 0.4);
}

.ready-badge {
  position: absolute;
  top: -8px;
  left: -8px;
  background: var(--primary);
  color: white;
  font-size: 0.7rem;
  padding: 0.2rem 0.6rem;
  border-radius: 10px;
  font-weight: 600;
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.4);
}

.player-card.is-ready {
  border-color: rgba(16, 185, 129, 0.5);
  box-shadow: 0 0 15px rgba(16, 185, 129, 0.15);
}

.action-section {
  display: flex;
  justify-content: center;
}

.host-controls, .guest-controls {
  padding: 1.25rem 2rem;
  text-align: center;
  border-radius: 12px;
  width: 100%;
  max-width: 360px;
}

.guest-controls {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  color: var(--text-muted);
}

.ready-btn {
  width: 100%;
  font-size: 1.125rem;
  letter-spacing: 2px;
  background: linear-gradient(135deg, var(--primary) 0%, #047857 100%);
}

.ready-status {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  color: var(--primary);
  font-size: 1.125rem;
  font-weight: 600;
}

.host-ready {
  color: var(--gold);
  margin-bottom: 0.75rem;
}

.ready-icon {
  width: 36px;
  height: 36px;
  background: var(--primary);
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
  font-weight: 700;
}

.host-controls .hint {
  margin-bottom: 1rem;
  color: var(--text-muted);
  font-size: 0.8rem;
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
