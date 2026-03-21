<template>
  <div class="home-container">
    <!-- 未登录状态 -->
    <template v-if="!isAuthenticated">
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

        <!-- 最近登录账号 -->
        <div v-if="recentAccounts.length > 0" class="recent-section">
          <div class="section-title">最近账号</div>
          <div class="recent-list">
            <div
              v-for="account in recentAccounts"
              :key="account.user_id"
              class="recent-item"
              @click="switchAccount(account)"
            >
              <span class="recent-avatar">{{ account.avatar }}</span>
              <span class="recent-name">{{ account.nickname }}</span>
              <span class="recent-chips">{{ formatChips(account.chips) }}</span>
              <button class="btn-remove" @click.stop="removeAccount(account.user_id)">×</button>
            </div>
          </div>
        </div>

        <div class="auth-actions">
          <button
            class="btn btn-primary btn-lg"
            @click="goToRegister"
            :disabled="!connected"
          >
            注册账号
          </button>
          <div class="divider">
            <span>或</span>
          </div>
          <button
            class="btn btn-secondary btn-lg"
            @click="goToLogin"
            :disabled="!connected"
          >
            登录
          </button>
        </div>
      </div>
    </template>

    <!-- 已登录状态 -->
    <template v-else>
      <div class="home-card glass-panel">
        <div class="user-header">
          <div class="user-avatar">{{ user?.avatar }}</div>
          <div class="user-info">
            <span class="user-name">{{ user?.nickname }}</span>
            <span class="user-chips">{{ formatChips(user?.chips || 0) }} 筹码</span>
          </div>
          <div class="user-actions">
            <button v-if="isAdmin" class="btn btn-sm btn-admin" @click="showAdminPanel = true">管理</button>
            <button class="btn btn-sm btn-ghost" @click="goToProfile">个人</button>
            <button class="btn btn-sm btn-ghost" @click="handleLogout">切换</button>
          </div>
        </div>

        <div class="connection-status" :class="connected ? 'is-connected' : 'is-disconnected'">
          <span class="status-dot"></span>
          <span>{{ connected ? '已连接' : '未连接' }}</span>
        </div>

        <Transition name="fade">
          <div v-if="serverError" class="error-toast">{{ serverError }}</div>
        </Transition>

        <div class="divider">
          <span>创建或加入房间</span>
        </div>

        <div class="actions" :class="{ 'is-disabled': !connected }">
          <button
            class="btn btn-primary btn-lg"
            @click="handleCreate"
            :disabled="!connected"
          >
            创建新房间
          </button>

          <button
            class="btn btn-secondary btn-lg ai-room-btn"
            @click="showAIRoomDialog = true"
            :disabled="!connected"
          >
            🤖 AI对战
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
              :disabled="roomCode.length !== 6 || !connected"
            >
              加入房间
            </button>
          </div>
        </div>
      </div>

      <!-- 管理员面板 -->
      <Transition name="fade">
        <div v-if="showAdminPanel" class="modal-overlay" @click.self="showAdminPanel = false">
          <div class="admin-panel glass-panel">
            <div class="admin-header">
              <h3>🎛️ 管理员面板</h3>
              <button class="btn-close" @click="showAdminPanel = false">×</button>
            </div>

            <div class="admin-content">
              <div class="admin-section">
                <h4>所有玩家</h4>
                <div class="user-list">
                  <div v-for="u in adminUsers" :key="u.id" class="user-item">
                    <div class="user-item-info">
                      <span class="user-item-avatar">{{ u.avatar }}</span>
                      <span class="user-item-name">{{ u.nickname }}</span>
                      <span class="user-item-chips">{{ formatChips(u.chips) }}</span>
                    </div>
                    <div class="user-item-actions">
                      <input
                        v-model.number="editChips[u.id]"
                        type="number"
                        class="input chips-input"
                        min="0"
                      />
                      <button class="btn btn-sm btn-warning" @click="updateUserChips(u.id)">修改</button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </Transition>

      <!-- AI房间对话框 -->
      <Transition name="fade">
        <div v-if="showAIRoomDialog" class="modal-overlay" @click.self="showAIRoomDialog = false">
          <div class="ai-dialog glass-panel">
            <div class="dialog-header">
              <h3>🤖 AI对战</h3>
              <button class="btn-close" @click="showAIRoomDialog = false">×</button>
            </div>

            <div class="dialog-content">
              <p class="dialog-desc">与AI玩家对战，练习德州扑克技巧</p>

              <div class="form-group">
                <label>AI难度</label>
                <div class="ai-level-options">
                  <button
                    class="ai-level-btn"
                    :class="{ 'is-selected': selectedAILevel === 'easy' }"
                    @click="selectedAILevel = 'easy'"
                  >
                    <span class="level-icon">😊</span>
                    <span class="level-name">简单</span>
                    <span class="level-desc">适合新手</span>
                  </button>
                  <button
                    class="ai-level-btn"
                    :class="{ 'is-selected': selectedAILevel === 'normal' }"
                    @click="selectedAILevel = 'normal'"
                  >
                    <span class="level-icon">🤔</span>
                    <span class="level-name">普通</span>
                    <span class="level-desc">标准难度</span>
                  </button>
                  <button
                    class="ai-level-btn"
                    :class="{ 'is-selected': selectedAILevel === 'hard' }"
                    @click="selectedAILevel = 'hard'"
                  >
                    <span class="level-icon">😈</span>
                    <span class="level-name">困难</span>
                    <span class="level-desc">高手挑战</span>
                  </button>
                </div>
              </div>

              <div class="ai-info">
                <span>你将与 5 位 AI 玩家进行 6 人短桌对战</span>
              </div>

              <button class="btn btn-primary btn-lg" @click="handleCreateAIRoom">
                开始游戏
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useWebSocket } from '../composables/useWebSocket'
import { useRoomStore } from '../stores/room'
import { useServerStore } from '../stores/server'
import { useAuthStore, type RecentAccount } from '../stores/auth'

const router = useRouter()
const roomStore = useRoomStore()
const serverStore = useServerStore()
const authStore = useAuthStore()
const { connect, createRoom, joinRoom, createAIRoom, error, connected } = useWebSocket()

const roomCode = ref('')
const serverError = ref('')
const serverAddress = ref(serverStore.serverAddress)
const showAIRoomDialog = ref(false)
const selectedAILevel = ref('normal')

const isAuthenticated = computed(() => authStore.isAuthenticated)
const user = computed(() => authStore.user)
const recentAccounts = computed(() => authStore.recentAccounts)
const isAdmin = computed(() => authStore.isAdmin)

// 管理员面板
const showAdminPanel = ref(false)
const adminUsers = ref<any[]>([])
const editChips = ref<Record<string, number>>({})

async function loadAdminUsers() {
  adminUsers.value = await authStore.fetchAllUsers()
  // 初始化编辑值
  editChips.value = {}
  for (const u of adminUsers.value) {
    editChips.value[u.id] = u.chips
  }
}

async function updateUserChips(userId: string) {
  const chips = editChips.value[userId]
  if (chips === undefined) return
  const success = await authStore.adminUpdateChips(userId, chips)
  if (success) {
    // 更新本地数据
    const user = adminUsers.value.find(u => u.id === userId)
    if (user) {
      user.chips = chips
    }
  }
}

// 打开管理员面板时加载数据
watch(showAdminPanel, async (val) => {
  if (val) {
    await loadAdminUsers()
  }
})

onMounted(async () => {
  // 如果有 token，尝试获取用户信息
  if (authStore.token) {
    await authStore.fetchUserInfo()
  }
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

function formatChips(chips: number): string {
  if (chips >= 1000000) {
    return (chips / 1000000).toFixed(1) + 'M'
  }
  if (chips >= 1000) {
    return (chips / 1000).toFixed(1) + 'K'
  }
  return chips.toString()
}

function reconnect() {
  serverError.value = ''
  connect()
}

function goToLogin() {
  router.push('/login')
}

function goToRegister() {
  router.push('/register')
}

function goToProfile() {
  router.push('/profile')
}

async function switchAccount(account: RecentAccount) {
  const success = await authStore.switchToAccount(account)
  if (!success) {
    serverError.value = '账号已失效，请重新登录'
  }
}

function removeAccount(userId: string) {
  authStore.removeFromRecentAccounts(userId)
}

function handleLogout() {
  authStore.logout()
}

watch(() => roomStore.roomCode, (newCode) => {
  if (newCode) {
    router.push('/room')
  }
})

async function handleCreate() {
  if (!isAuthenticated.value) {
    router.push('/login')
    return
  }
  if ((user.value?.chips || 0) <= 0) {
    serverError.value = '筹码不足，无法创建房间'
    return
  }
  createRoom('')
}

async function handleJoin() {
  if (!isAuthenticated.value) {
    router.push('/login')
    return
  }
  if ((user.value?.chips || 0) <= 0) {
    serverError.value = '筹码不足，无法加入房间'
    return
  }
  joinRoom(roomCode.value.trim(), '')
}

async function handleCreateAIRoom() {
  if (!isAuthenticated.value) {
    router.push('/login')
    return
  }
  if ((user.value?.chips || 0) <= 0) {
    serverError.value = '筹码不足，无法创建房间'
    showAIRoomDialog.value = false
    return
  }
  showAIRoomDialog.value = false
  createAIRoom(selectedAILevel.value)
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
  padding: 2rem;
  text-align: center;
}

.hero-section {
  margin-bottom: 1.5rem;
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

.user-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1.5rem;
  padding: 1rem;
  background: rgba(251, 191, 36, 0.1);
  border-radius: 12px;
}

.user-avatar {
  font-size: 2.5rem;
  width: 56px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--glass-bg);
  border-radius: 12px;
}

.user-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
}

.user-name {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--text);
}

.user-chips {
  font-size: 0.875rem;
  color: var(--gold);
}

.user-actions {
  display: flex;
  gap: 0.5rem;
}

.btn-ghost {
  background: transparent;
  border: 1px solid var(--glass-border);
  color: var(--text-muted);
}

.form-group {
  text-align: left;
  margin-bottom: 1rem;
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
  margin-bottom: 1rem;
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

.recent-section {
  margin-bottom: 1rem;
}

.section-title {
  font-size: 0.75rem;
  color: var(--text-muted);
  margin-bottom: 0.5rem;
  text-align: left;
}

.recent-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.recent-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  background: var(--glass-bg);
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.2s;
}

.recent-item:hover {
  background: rgba(251, 191, 36, 0.1);
}

.recent-avatar {
  font-size: 1.5rem;
}

.recent-name {
  flex: 1;
  font-size: 0.875rem;
  color: var(--text);
  text-align: left;
}

.recent-chips {
  font-size: 0.75rem;
  color: var(--gold);
}

.btn-remove {
  width: 24px;
  height: 24px;
  border: none;
  background: transparent;
  color: var(--text-muted);
  font-size: 1.25rem;
  cursor: pointer;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-remove:hover {
  background: rgba(239, 68, 68, 0.2);
  color: var(--danger);
}

.auth-actions {
  margin-top: 1rem;
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

.btn-admin {
  background: rgba(139, 92, 246, 0.2);
  border: 1px solid rgba(139, 92, 246, 0.4);
  color: #8b5cf6;
}

.btn-admin:hover {
  background: rgba(139, 92, 246, 0.3);
}

.btn-warning {
  background: var(--gold);
  color: #451a03;
}

.btn-close {
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  color: var(--text-muted);
  font-size: 1.5rem;
  cursor: pointer;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-close:hover {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text);
}

.admin-panel {
  width: 100%;
  max-width: 500px;
  max-height: 80vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.admin-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.admin-header h3 {
  color: var(--gold);
  margin: 0;
}

.admin-content {
  overflow-y: auto;
  flex: 1;
}

.admin-section h4 {
  color: var(--text);
  margin-bottom: 0.75rem;
  font-size: 0.9rem;
}

.user-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.user-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem;
  background: rgba(0, 0, 0, 0.2);
  border-radius: 8px;
  gap: 1rem;
}

.user-item-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.user-item-avatar {
  font-size: 1.25rem;
}

.user-item-name {
  font-size: 0.875rem;
  color: var(--text);
}

.user-item-chips {
  font-size: 0.75rem;
  color: var(--gold);
}

.user-item-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.chips-input {
  width: 100px;
  padding: 0.5rem;
  font-size: 0.875rem;
}

.ai-room-btn {
  margin-top: 0.75rem;
  background: linear-gradient(135deg, #6366f1 0%, #4f46e5 100%);
  border: none;
}

.ai-room-btn:hover {
  background: linear-gradient(135deg, #4f46e5 0%, #4338ca 100%);
}

.ai-dialog {
  width: 100%;
  max-width: 420px;
  padding: 1.5rem;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.dialog-header h3 {
  color: var(--gold);
  margin: 0;
  font-size: 1.25rem;
}

.dialog-content {
  text-align: center;
}

.dialog-desc {
  color: var(--text-muted);
  margin-bottom: 1.5rem;
  font-size: 0.9rem;
}

.ai-level-options {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.75rem;
  margin-bottom: 1.5rem;
}

.ai-level-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.25rem;
  padding: 1rem 0.5rem;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid var(--glass-border);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.ai-level-btn:hover {
  background: rgba(255, 255, 255, 0.1);
}

.ai-level-btn.is-selected {
  background: rgba(99, 102, 241, 0.2);
  border-color: #6366f1;
}

.level-icon {
  font-size: 1.5rem;
}

.level-name {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text);
}

.level-desc {
  font-size: 0.7rem;
  color: var(--text-muted);
}

.ai-info {
  padding: 0.75rem;
  background: rgba(99, 102, 241, 0.1);
  border-radius: 8px;
  margin-bottom: 1rem;
  font-size: 0.8rem;
  color: var(--text-muted);
}
</style>
