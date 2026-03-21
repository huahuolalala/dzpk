<template>
  <div class="chat-messages">
    <TransitionGroup name="message">
      <div
        v-for="msg in displayMessages"
        :key="msg.id"
        class="chat-message"
      >
        <div class="message-avatar">{{ msg.playerName.charAt(0).toUpperCase() }}</div>
        <div class="message-content">
          <span class="message-name">{{ msg.playerName }}</span>
          <span class="message-text">{{ msg.content }}</span>
        </div>
      </div>
    </TransitionGroup>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useGameStore, type DanmakuItem } from '../stores/game'

const gameStore = useGameStore()

const displayMessages = ref<DanmakuItem[]>([])
const messageQueue = ref<DanmakuItem[]>([])
const isProcessing = ref(false)

const DISPLAY_DURATION = 3000 // 显示3秒

// 监听弹幕列表变化
watch(() => gameStore.danmakuList, (newList) => {
  if (newList && newList.length > 0) {
    // 找出还没有处理的消息
    const lastMessage = newList[newList.length - 1]
    const alreadyQueued = messageQueue.value.some(m => m.id === lastMessage.id) ||
                          displayMessages.value.some(m => m.id === lastMessage.id)
    if (!alreadyQueued) {
      messageQueue.value.push(lastMessage)
      processQueue()
    }
  }
}, { deep: true })

function processQueue() {
  if (isProcessing.value || messageQueue.value.length === 0) return

  isProcessing.value = true
  const msg = messageQueue.value.shift()!

  // 添加到显示列表
  displayMessages.value.push(msg)

  // 限制同时显示最多3条
  if (displayMessages.value.length > 3) {
    displayMessages.value.shift()
  }

  // 3秒后移除
  setTimeout(() => {
    const index = displayMessages.value.findIndex(m => m.id === msg.id)
    if (index !== -1) {
      displayMessages.value.splice(index, 1)
    }
    gameStore.removeDanmaku(msg.id)
    isProcessing.value = false
    // 处理下一条
    if (messageQueue.value.length > 0) {
      setTimeout(processQueue, 200)
    }
  }, DISPLAY_DURATION)
}
</script>

<style scoped>
.chat-messages {
  position: fixed;
  bottom: 6rem;
  left: 2rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  z-index: 200;
  max-width: 320px;
}

.chat-message {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  background: rgba(0, 0, 0, 0.8);
  border-radius: 12px;
  backdrop-filter: blur(8px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.message-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.875rem;
  font-weight: 600;
  color: white;
  flex-shrink: 0;
}

.message-content {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  min-width: 0;
}

.message-name {
  color: #fbbf24;
  font-weight: 600;
  font-size: 0.875rem;
}

.message-text {
  color: white;
  font-size: 0.875rem;
  word-break: break-word;
}

/* 过渡动画 */
.message-enter-active {
  animation: slide-in 0.3s ease-out;
}

.message-leave-active {
  animation: fade-out 0.5s ease-in;
}

@keyframes slide-in {
  from {
    opacity: 0;
    transform: translateX(-20px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

@keyframes fade-out {
  from {
    opacity: 1;
  }
  to {
    opacity: 0;
  }
}
</style>