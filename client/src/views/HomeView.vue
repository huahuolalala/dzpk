<template>
  <div class="home-container">
    <div class="home-card glass-panel">
      <div class="hero-section">
        <h1 class="title">德州扑克</h1>
        <p class="subtitle">TEXAS HOLD'EM</p>
      </div>

      <div class="form-group">
        <label>服务器地址</label>
        <div class="server-input-group">
          <input
            v-model="serverAddress"
            type="text"
            class="input"
            placeholder="例: 1.1.1.1:8080"
          />
          <button class="btn btn-sm btn-secondary" @click="reconnect" :disabled="!serverAddress.trim()">
            连接
          </button>
        </div>
        <span class="input-hint">输入服务器 IP 和端口，如 192.168.1.100:8080</span>
      </div>

      <div class="connection-status" :class="connected ? 'is-connected' : 'is-disconnected'">
        <span class="status-dot"></span>
        <span>{{ connected ? '已连接' : '未连接' }}</span>
      </div>

      <Transition name="fade">
        <div v-if="serverError" class="error-toast">{{ serverError }}</div>
      </Transition>

      <div class="form-group">
        <label>你的昵称</label>
        <input
          v-model="playerName"
          type="text"
          class="input"
          placeholder="请输入你的牌桌代号"
          maxlength="20"
          :disabled="!connected"
        />
      </div>

      <div class="actions" :class="{ 'is-disabled': !connected }">
        <button
          class="btn btn-primary btn-lg"
          @click="handleCreate"
          :disabled="!playerName.trim() || !connected"
        >
          创建新房间
        </button>

        <div class="divider">
          <span>或加入已有房间</span>
        </div>

        <div class="join-group">
          <input
            v-model="roomCode"
            type="text"
            class="input text-center"
            placeholder="6位房间码"
            maxlength="6"
            :disabled="!connected"
          />
          <button
            class="btn btn-secondary btn-lg"
            @click="handleJoin"
            :disabled="!playerName.trim() || roomCode.length !== 6 || !connected"
          >
            加入房间
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useWebSocket } from '../composables/useWebSocket'
import { useRoomStore } from '../stores/room'
import { useServerStore } from '../stores/server'

const router = useRouter()
const roomStore = useRoomStore()
const serverStore = useServerStore()
const { connect, createRoom, joinRoom, error, connected } = useWebSocket()

const playerName = ref('')
const roomCode = ref('')
const serverError = ref('')
const serverAddress = ref(serverStore.serverAddress)

onMounted(() => {
  connect()
})

watch(serverAddress, (val) => {
  serverStore.setAddress(val)
})

watch(error, (err) => {
  if (err) {
    serverError.value = err
    setTimeout(() => {
      serverError.value = ''
    }, 3000)
  }
})

watch(connected, (val) => {
  if (val) {
    serverError.value = ''
  }
})

function reconnect() {
  serverError.value = ''
  connect()
}

watch(() => roomStore.roomCode, (newCode) => {
  if (newCode) {
    router.push('/room')
  }
})

async function handleCreate() {
  createRoom(playerName.value.trim())
}

async function handleJoin() {
  joinRoom(roomCode.value.trim(), playerName.value.trim())
}
</script>

<style scoped>
.home-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  background: var(--bg-table);
}

.home-card {
  width: 100%;
  max-width: 420px;
  padding: 2.5rem 2rem;
  text-align: center;
}

.hero-section {
  margin-bottom: 2rem;
}

.title {
  font-size: 2.5rem;
  font-weight: 700;
  color: var(--gold);
  margin-bottom: 0.25rem;
  letter-spacing: 2px;
  text-shadow: 0 2px 10px rgba(251, 191, 36, 0.3);
}

.subtitle {
  color: var(--text-muted);
  font-size: 0.875rem;
  letter-spacing: 4px;
}

.form-group {
  text-align: left;
  margin-bottom: 1.5rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-size: 0.875rem;
  color: var(--text-muted);
  font-weight: 500;
}

.server-input-group {
  display: flex;
  gap: 0.5rem;
}

.server-input-group .input {
  flex: 1;
}

.input-hint {
  display: block;
  margin-top: 0.4rem;
  font-size: 0.7rem;
  color: var(--text-muted);
}

.connection-status {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  margin-bottom: 1.5rem;
  padding: 0.5rem;
  border-radius: 20px;
  font-size: 0.8rem;
}

.connection-status.is-connected {
  background: rgba(16, 185, 129, 0.1);
  color: var(--success, #10b981);
}

.connection-status.is-disconnected {
  background: rgba(239, 68, 68, 0.1);
  color: var(--danger, #ef4444);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.is-connected .status-dot {
  background: var(--success, #10b981);
  box-shadow: 0 0 6px var(--success, #10b981);
}

.is-disconnected .status-dot {
  background: var(--danger, #ef4444);
  box-shadow: 0 0 6px var(--danger, #ef4444);
}

.actions {
  transition: opacity 0.2s;
}

.actions.is-disabled {
  opacity: 0.5;
  pointer-events: none;
}

.btn-lg {
  width: 100%;
}

.divider {
  display: flex;
  align-items: center;
  text-align: center;
  color: var(--text-muted);
  font-size: 0.875rem;
  margin: 1rem 0;
}

.divider::before,
.divider::after {
  content: '';
  flex: 1;
  border-bottom: 1px solid var(--glass-border);
}

.divider span {
  padding: 0 1rem;
}

.join-group {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.text-center {
  text-align: center;
  letter-spacing: 2px;
  font-size: 1.25rem;
}

.error-toast {
  margin-bottom: 1rem;
  padding: 0.75rem;
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.2);
  color: var(--danger);
  border-radius: 8px;
  font-size: 0.875rem;
  text-align: center;
}

.btn-sm {
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
}
</style>
