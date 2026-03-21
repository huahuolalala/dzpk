<template>
  <div class="login-container">
    <div class="login-card glass-panel">
      <h1 class="title">登录</h1>

      <form @submit.prevent="handleLogin">
        <div class="form-group">
          <label>账号</label>
          <input
            v-model="username"
            type="text"
            class="input"
            placeholder="请输入账号"
            maxlength="20"
            required
          />
        </div>

        <div class="form-group">
          <label>密码</label>
          <input
            v-model="password"
            type="password"
            class="input"
            placeholder="请输入密码"
            maxlength="20"
            required
          />
        </div>

        <Transition name="fade">
          <div v-if="error" class="error-toast">{{ error }}</div>
        </Transition>

        <button type="submit" class="btn btn-primary btn-lg" :disabled="isLoading">
          {{ isLoading ? '登录中...' : '登录' }}
        </button>
      </form>

      <div class="footer">
        <router-link to="/">返回首页</router-link>
        <span class="divider">|</span>
        还没有账号？<router-link to="/register">立即注册</router-link>
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

const username = ref('')
const password = ref('')
const error = ref('')
const isLoading = ref(false)

async function handleLogin() {
  error.value = ''
  isLoading.value = true

  const success = await authStore.login(username.value, password.value)

  if (success) {
    router.push('/')
  } else {
    error.value = authStore.error || '登录失败'
  }

  isLoading.value = false
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
  background: var(--bg-table);
}

.login-card {
  width: 100%;
  max-width: 360px;
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

.footer .divider {
  margin: 0 0.5rem;
  color: var(--text-muted);
}
</style>
