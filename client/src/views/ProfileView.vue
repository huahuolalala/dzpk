<template>
  <div class="profile-container">
    <div class="profile-card glass-panel">
      <div class="profile-header">
        <div class="avatar-large">{{ user?.avatar }}</div>
        <div class="user-info">
          <h2 class="nickname">{{ user?.nickname }}</h2>
          <p class="username">@{{ user?.username }}</p>
        </div>
        <button class="btn btn-secondary btn-sm" @click="handleLogout">退出</button>
      </div>

      <div class="chips-display">
        <span class="chips-label">剩余筹码</span>
        <span class="chips-value">{{ formatChips(user?.chips || 0) }}</span>
      </div>

      <div class="stats-grid">
        <div class="stat-item">
          <span class="stat-value">{{ stats?.total_games || 0 }}</span>
          <span class="stat-label">总局数</span>
        </div>
        <div class="stat-item">
          <span class="stat-value">{{ stats?.wins || 0 }}</span>
          <span class="stat-label">胜利次数</span>
        </div>
        <div class="stat-item">
          <span class="stat-value">{{ stats?.win_rate || '0%' }}</span>
          <span class="stat-label">胜率</span>
        </div>
        <div class="stat-item">
          <span class="stat-value" :class="{ profit: (stats?.total_profit ?? 0) >= 0, loss: (stats?.total_profit ?? 0) < 0 }">
            {{ (stats?.total_profit ?? 0) >= 0 ? '+' : '' }}{{ formatChips(stats?.total_profit ?? 0) }}
          </span>
          <span class="stat-label">总盈亏</span>
        </div>
      </div>
    </div>

    <button class="btn btn-secondary" @click="goBack">返回大厅</button>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const user = authStore.user
const stats = authStore.stats

function formatChips(chips: number): string {
  if (chips >= 1000000) {
    return (chips / 1000000).toFixed(1) + 'M'
  }
  if (chips >= 1000) {
    return (chips / 1000).toFixed(1) + 'K'
  }
  return chips.toString()
}

function handleLogout() {
  authStore.logout()
  router.push('/login')
}

function goBack() {
  router.push('/')
}

onMounted(async () => {
  await authStore.fetchStats()
})
</script>

<style scoped>
.profile-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  background: var(--bg-table);
}

.profile-card {
  width: 100%;
  max-width: 420px;
  padding: 1.5rem;
}

.profile-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.avatar-large {
  font-size: 3rem;
  width: 64px;
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--glass-bg);
  border-radius: 16px;
}

.user-info {
  flex: 1;
}

.nickname {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text);
  margin: 0;
}

.username {
  font-size: 0.875rem;
  color: var(--text-muted);
  margin: 0;
}

.chips-display {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 1.5rem;
  background: rgba(251, 191, 36, 0.1);
  border-radius: 12px;
  margin-bottom: 1.5rem;
}

.chips-label {
  font-size: 0.875rem;
  color: var(--text-muted);
  margin-bottom: 0.25rem;
}

.chips-value {
  font-size: 2rem;
  font-weight: 700;
  color: var(--gold);
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 1rem;
  background: var(--glass-bg);
  border-radius: 12px;
}

.stat-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text);
}

.stat-value.profit {
  color: var(--success, #10b981);
}

.stat-value.loss {
  color: var(--danger, #ef4444);
}

.stat-label {
  font-size: 0.75rem;
  color: var(--text-muted);
  margin-top: 0.25rem;
}

.btn-sm {
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
}

.btn-secondary {
  margin-top: 1rem;
  width: 100%;
  max-width: 420px;
}
</style>
