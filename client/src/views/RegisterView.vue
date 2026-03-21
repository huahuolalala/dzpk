<template>
  <div class="register-container">
    <div class="register-card glass-panel">
      <h1 class="title">注册账号</h1>

      <form @submit.prevent="handleRegister">
        <div class="form-group">
          <label>选择头像</label>
          <div class="avatar-grid">
            <button
              v-for="avatar in avatars"
              :key="avatar"
              type="button"
              class="avatar-btn"
              :class="{ selected: selectedAvatar === avatar }"
              @click="selectedAvatar = avatar"
            >
              {{ avatar }}
            </button>
          </div>
        </div>

        <div class="form-group">
          <label>昵称</label>
          <input
            v-model="nickname"
            type="text"
            class="input"
            placeholder="你的牌桌代号"
            maxlength="20"
            required
          />
        </div>

        <div class="form-group">
          <label>账号</label>
          <input
            v-model="username"
            type="text"
            class="input"
            placeholder="3-20位字母或数字"
            maxlength="20"
            pattern="[a-zA-Z0-9]{3,20}"
            required
          />
        </div>

        <div class="form-group">
          <label>密码</label>
          <input
            v-model="password"
            type="password"
            class="input"
            placeholder="6-20位密码"
            maxlength="20"
            minlength="6"
            required
          />
        </div>

        <div class="form-group">
          <label>确认密码</label>
          <input
            v-model="confirmPassword"
            type="password"
            class="input"
            placeholder="再次输入密码"
            maxlength="20"
            required
          />
        </div>

        <Transition name="fade">
          <div v-if="error" class="error-toast">{{ error }}</div>
        </Transition>

        <button type="submit" class="btn btn-primary btn-lg" :disabled="isLoading">
          {{ isLoading ? '注册中...' : '注册' }}
        </button>
      </form>

      <div class="footer">
        已有账号？<router-link to="/login">立即登录</router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const avatars = ['🐱', '🐶', '🐰', '🦊', '🐼', '🦁', '🐸', '🐵', '🐷']
const selectedAvatar = ref(avatars[0])
const username = ref('')
const nickname = ref('')
const password = ref('')
const confirmPassword = ref('')
const error = ref('')
const isLoading = ref(false)

async function handleRegister() {
  error.value = ''

  if (password.value !== confirmPassword.value) {
    error.value = '两次输入的密码不一致'
    return
  }

  if (!selectedAvatar.value) {
    error.value = '请选择一个头像'
    return
  }

  isLoading.value = true

  const success = await authStore.register(
    username.value,
    password.value,
    nickname.value || username.value,
    selectedAvatar.value
  )

  if (success) {
    router.push('/')
  } else {
    error.value = authStore.error || '注册失败'
  }

  isLoading.value = false
}
</script>

<style scoped>
.register-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  background: var(--bg-table);
}

.register-card {
  width: 100%;
  max-width: 400px;
  padding: 2rem;
}

.title {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--gold);
  text-align: center;
  margin-bottom: 1.5rem;
}

.form-group {
  margin-bottom: 1.25rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-size: 0.875rem;
  color: var(--text-muted);
}

.avatar-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 0.5rem;
}

.avatar-btn {
  aspect-ratio: 1;
  font-size: 1.75rem;
  border: 2px solid var(--glass-border);
  border-radius: 12px;
  background: var(--glass-bg);
  cursor: pointer;
  transition: all 0.2s;
}

.avatar-btn:hover {
  border-color: var(--gold);
  transform: scale(1.05);
}

.avatar-btn.selected {
  border-color: var(--gold);
  background: rgba(251, 191, 36, 0.15);
  box-shadow: 0 0 12px rgba(251, 191, 36, 0.3);
}

.input {
  width: 100%;
  padding: 0.75rem 1rem;
  font-size: 1rem;
  border: 1px solid var(--glass-border);
  border-radius: 8px;
  background: var(--glass-bg);
  color: var(--text);
  outline: none;
  transition: border-color 0.2s;
}

.input:focus {
  border-color: var(--gold);
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

.btn-lg {
  width: 100%;
  margin-top: 0.5rem;
}

.footer {
  margin-top: 1.5rem;
  text-align: center;
  font-size: 0.875rem;
  color: var(--text-muted);
}

.footer a {
  color: var(--gold);
  text-decoration: none;
}

.footer a:hover {
  text-decoration: underline;
}
</style>
