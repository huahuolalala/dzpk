<template>
  <div class="chat-input-container">
    <!-- 聊天图标按钮 -->
    <button
      v-if="!showInput"
      class="chat-toggle-btn glass-panel"
      @click="toggleInput"
      title="发送弹幕"
    >
      <span class="chat-icon">💬</span>
    </button>

    <!-- 输入框 -->
    <Transition name="slide-up">
      <div v-if="showInput" class="chat-input-wrapper glass-panel">
        <input
          ref="inputRef"
          v-model="message"
          type="text"
          class="chat-input"
          placeholder="发送弹幕..."
          maxlength="100"
          @keydown.enter="sendMessage"
          @keydown.escape="closeInput"
        />
        <span class="char-count">{{ message.length }}/100</span>
        <button class="send-btn" @click="sendMessage" :disabled="!message.trim()">
          发送
        </button>
        <button class="close-btn" @click="closeInput">✕</button>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick } from 'vue'
import { useWebSocket } from '../composables/useWebSocket'

const { sendChat } = useWebSocket()

const showInput = ref(false)
const message = ref('')
const inputRef = ref<HTMLInputElement | null>(null)

function toggleInput() {
  showInput.value = !showInput.value
  if (showInput.value) {
    nextTick(() => {
      inputRef.value?.focus()
    })
  }
}

function closeInput() {
  showInput.value = false
  message.value = ''
}

function sendMessage() {
  const content = message.value.trim()
  if (content) {
    sendChat(content)
    message.value = ''
    closeInput()
  }
}
</script>

<style scoped>
.chat-input-container {
  position: fixed;
  bottom: 2rem;
  right: 2rem;
  z-index: 200;
}

.chat-toggle-btn {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s ease;
  background: rgba(0, 0, 0, 0.6);
}

.chat-toggle-btn:hover {
  transform: scale(1.1);
  background: rgba(0, 0, 0, 0.8);
}

.chat-icon {
  font-size: 1.5rem;
}

.chat-input-wrapper {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  border-radius: 25px;
  background: rgba(0, 0, 0, 0.85);
  min-width: 320px;
}

.chat-input {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  color: white;
  font-size: 0.95rem;
  padding: 0.5rem 0;
}

.chat-input::placeholder {
  color: rgba(255, 255, 255, 0.5);
}

.char-count {
  font-size: 0.75rem;
  color: rgba(255, 255, 255, 0.4);
  white-space: nowrap;
}

.send-btn {
  background: var(--primary, #10b981);
  color: white;
  border: none;
  padding: 0.5rem 1rem;
  border-radius: 15px;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.send-btn:hover:not(:disabled) {
  background: #059669;
}

.send-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.close-btn {
  background: transparent;
  border: none;
  color: rgba(255, 255, 255, 0.5);
  cursor: pointer;
  font-size: 1rem;
  padding: 0.25rem;
  line-height: 1;
}

.close-btn:hover {
  color: white;
}

/* 过渡动画 */
.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.3s ease;
}

.slide-up-enter-from,
.slide-up-leave-to {
  opacity: 0;
  transform: translateY(20px) scale(0.95);
}

/* 响应式 */
@media (max-width: 480px) {
  .chat-input-wrapper {
    min-width: 280px;
    padding: 0.5rem 0.75rem;
  }

  .char-count {
    display: none;
  }
}
</style>