import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import HomeView from './views/HomeView.vue'
import LoginView from './views/LoginView.vue'
import RegisterView from './views/RegisterView.vue'
import ProfileView from './views/ProfileView.vue'
import RoomView from './views/RoomView.vue'
import GameView from './views/GameView.vue'
import { useAuthStore } from './stores/auth'
import { playSound } from './composables/useSound'
import './styles/main.css'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'home', component: HomeView },
    { path: '/login', name: 'login', component: LoginView },
    { path: '/register', name: 'register', component: RegisterView },
    { path: '/profile', name: 'profile', component: ProfileView, meta: { requiresAuth: true } },
    { path: '/room', name: 'room', component: RoomView, meta: { requiresAuth: true } },
    { path: '/game', name: 'game', component: GameView, meta: { requiresAuth: true } },
  ],
})

// 路由守卫
router.beforeEach((to, _from, next) => {
  const authStore = useAuthStore()

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    // 需要认证但未登录，重定向到登录
    next({ name: 'login' })
  } else if ((to.name === 'login' || to.name === 'register') && authStore.isAuthenticated) {
    // 已登录访问登录/注册页，重定向到首页
    next({ name: 'home' })
  } else {
    next()
  }
})

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')

// 全局按钮点击音效
document.addEventListener('click', (e) => {
  const target = e.target as HTMLElement
  if (target && (target.classList.contains('btn') || target.closest('.btn'))) {
    playSound('click')
  }
})
