<template>
  <div class="action-history" :class="{ collapsed: isCollapsed }">
    <button class="toggle-btn" @click="toggleCollapse" :title="isCollapsed ? '展开操作历史' : '收起操作历史'">
      <span class="toggle-icon">{{ isCollapsed ? '>' : '<' }}</span>
    </button>

    <div class="history-content" v-show="!isCollapsed">
      <h3 class="history-title">操作历史</h3>

      <div class="history-list" v-if="gameStore.actions.length > 0">
        <div
          v-for="(action, index) in gameStore.actions"
          :key="index"
          class="history-item"
          :class="getActionClass(action.action)"
        >
          <div class="action-header">
            <span class="player-name">{{ action.player_name }}</span>
            <span class="phase-badge" :class="action.phase">{{ translatePhase(action.phase) }}</span>
          </div>
          <div class="action-detail">
            <span class="action-text">{{ translateAction(action.action) }}</span>
            <span class="action-amount" v-if="action.amount > 0">{{ action.amount }}</span>
          </div>
        </div>
      </div>

      <div class="empty-state" v-else>
        <p>暂无操作记录</p>
      </div>
    </div>

    <!-- 折叠时显示的小点 -->
    <div class="collapsed-indicator" v-show="isCollapsed">
      <span class="indicator-dot" v-if="gameStore.actions.length > 0"></span>
      <span class="indicator-count" v-if="gameStore.actions.length > 0">{{ gameStore.actions.length }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useGameStore } from '../stores/game'

const gameStore = useGameStore()
const isCollapsed = ref(false)

function toggleCollapse() {
  isCollapsed.value = !isCollapsed.value
}

function translateAction(action: string): string {
  const map: Record<string, string> = {
    'check': '过牌',
    'call': '跟注',
    'raise': '加注',
    'fold': '弃牌',
    'allin': 'All-in',
    'small_blind': '小盲',
    'big_blind': '大盲',
  }
  return map[action] || action
}

function translatePhase(phase: string): string {
  const map: Record<string, string> = {
    'preflop': '翻牌前',
    'flop': '翻牌',
    'turn': '转牌',
    'river': '河牌',
    'showdown': '摊牌',
  }
  return map[phase] || phase
}

function getActionClass(action: string): string {
  const map: Record<string, string> = {
    'check': 'action-check',
    'call': 'action-call',
    'raise': 'action-raise',
    'fold': 'action-fold',
    'allin': 'action-allin',
    'small_blind': 'action-small_blind',
    'big_blind': 'action-big_blind',
  }
  return map[action] || ''
}
</script>

<style scoped>
.action-history {
  position: fixed;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  z-index: 50;
  display: flex;
  align-items: stretch;
  transition: all 0.3s ease;
}

.action-history.collapsed {
  left: 0;
}

.toggle-btn {
  width: 24px;
  min-height: 60px;
  background: var(--glass-bg);
  border: 1px solid var(--glass-border);
  border-left: none;
  border-radius: 0 8px 8px 0;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-muted);
  transition: all 0.2s ease;
}

.toggle-btn:hover {
  background: rgba(255, 255, 255, 0.15);
  color: var(--text-main);
}

.toggle-icon {
  font-size: 12px;
  font-weight: bold;
}

.history-content {
  width: 200px;
  background: var(--glass-bg);
  border: 1px solid var(--glass-border);
  border-left: none;
  border-radius: 0 12px 12px 0;
  padding: 1rem;
  max-height: 400px;
  display: flex;
  flex-direction: column;
  backdrop-filter: blur(10px);
}

.history-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-main);
  margin: 0 0 0.75rem 0;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  text-align: center;
}

.history-list {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.history-list::-webkit-scrollbar {
  width: 4px;
}

.history-list::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 2px;
}

.history-list::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 2px;
}

.history-item {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
  padding: 0.5rem 0.75rem;
  border-left: 3px solid var(--glass-border);
}

.history-item.action-check {
  border-left-color: #94a3b8;
}

.history-item.action-call {
  border-left-color: #3b82f6;
}

.history-item.action-raise {
  border-left-color: #f59e0b;
}

.history-item.action-fold {
  border-left-color: #6b7280;
}

.history-item.action-allin {
  border-left-color: #ef4444;
  background: rgba(239, 68, 68, 0.1);
}

.history-item.action-small_blind,
.history-item.action-big_blind {
  border-left-color: #8b5cf6;
  background: rgba(139, 92, 246, 0.1);
}

.action-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.25rem;
}

.player-name {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-main);
}

.phase-badge {
  font-size: 0.625rem;
  padding: 0.125rem 0.375rem;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-muted);
}

.action-detail {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.action-text {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.action-amount {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--gold);
}

.empty-state {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-muted);
  font-size: 0.75rem;
}

.collapsed-indicator {
  width: 24px;
  min-height: 60px;
  background: var(--glass-bg);
  border: 1px solid var(--glass-border);
  border-left: none;
  border-radius: 0 8px 8px 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.25rem;
  padding: 0.5rem 0;
}

.indicator-dot {
  width: 6px;
  height: 6px;
  background: var(--primary);
  border-radius: 50%;
  animation: pulse-dot 1.5s infinite;
}

.indicator-count {
  font-size: 0.625rem;
  font-weight: 600;
  color: var(--text-muted);
}

@keyframes pulse-dot {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}
</style>
