import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useServerStore } from './server'

const API_BASE = '/api/auth'
const USER_API = '/api/user'
const ADMIN_API = '/api/admin'
const RECENT_ACCOUNTS_KEY = 'dz_recent_accounts'
const MAX_RECENT_ACCOUNTS = 5

export interface AdminUser {
  id: string
  username: string
  nickname: string
  avatar: string
  chips: number
}

export interface User {
  user_id: string
  username: string
  nickname: string
  avatar: string
  chips: number
}

export interface Stats {
  total_games: number
  wins: number
  win_rate: string
  total_profit: number
}

export interface GameHistory {
  id: string
  room_code: string
  user_id: string
  initial_chips: number
  final_chips: number
  profit: number
  final_rank: number
  best_hand: string
  created_at: string
}

export interface RecentAccount {
  user_id: string
  username: string
  nickname: string
  avatar: string
  token: string
  chips: number
}

function loadRecentAccounts(): RecentAccount[] {
  try {
    const data = localStorage.getItem(RECENT_ACCOUNTS_KEY)
    return data ? JSON.parse(data) : []
  } catch {
    return []
  }
}

function saveRecentAccounts(accounts: RecentAccount[]) {
  localStorage.setItem(RECENT_ACCOUNTS_KEY, JSON.stringify(accounts))
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('dz_token') || '')
  const user = ref<User | null>(null)
  const stats = ref<Stats | null>(null)
  const isLoading = ref(false)
  const error = ref('')
  const recentAccounts = ref<RecentAccount[]>(loadRecentAccounts())

  const isAuthenticated = computed(() => !!token.value && !!user.value)
  const serverStore = useServerStore()

  function setToken(newToken: string) {
    token.value = newToken
    if (newToken) {
      localStorage.setItem('dz_token', newToken)
    } else {
      localStorage.removeItem('dz_token')
    }
  }

  // 添加到最近账号
  function addToRecentAccounts(account: RecentAccount) {
    const filtered = recentAccounts.value.filter(a => a.user_id !== account.user_id)
    filtered.unshift(account)
    if (filtered.length > MAX_RECENT_ACCOUNTS) {
      filtered.pop()
    }
    recentAccounts.value = filtered
    saveRecentAccounts(filtered)
  }

  // 切换到指定账号
  async function switchToAccount(account: RecentAccount): Promise<boolean> {
    // 验证 token 是否有效
    const valid = await validateToken(account.token)
    if (valid) {
      setToken(account.token)
      user.value = {
        user_id: account.user_id,
        username: account.username,
        nickname: account.nickname,
        avatar: account.avatar,
        chips: account.chips,
      }
      await fetchUserInfo()
      return true
    } else {
      // token 失效，移除该账号
      removeFromRecentAccounts(account.user_id)
      return false
    }
  }

  // 移除指定账号
  function removeFromRecentAccounts(userId: string) {
    recentAccounts.value = recentAccounts.value.filter(a => a.user_id !== userId)
    saveRecentAccounts(recentAccounts.value)
  }

  // 验证 token
  async function validateToken(t: string): Promise<boolean> {
    try {
      const response = await fetch(`${serverStore.getHttpUrl()}${USER_API}/info`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${t}`,
        },
      })
      return response.ok
    } catch {
      return false
    }
  }

  async function register(username: string, password: string, nickname: string, avatar: string): Promise<boolean> {
    isLoading.value = true
    error.value = ''

    try {
      const response = await fetch(`${serverStore.getHttpUrl()}${API_BASE}/register`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username, password, nickname, avatar }),
      })

      const data = await response.json()

      if (!response.ok) {
        error.value = data.message || 'Registration failed'
        return false
      }

      setToken(data.token)
      user.value = {
        user_id: data.user_id,
        username: data.username,
        nickname: data.nickname,
        avatar: data.avatar,
        chips: data.chips,
      }

      // 添加到最近账号
      addToRecentAccounts({
        user_id: data.user_id,
        username: data.username,
        nickname: data.nickname,
        avatar: data.avatar,
        token: data.token,
        chips: data.chips,
      })

      return true
    } catch (e) {
      error.value = 'Network error'
      return false
    } finally {
      isLoading.value = false
    }
  }

  async function login(username: string, password: string): Promise<boolean> {
    isLoading.value = true
    error.value = ''

    try {
      const response = await fetch(`${serverStore.getHttpUrl()}${API_BASE}/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username, password }),
      })

      const data = await response.json()

      if (!response.ok) {
        error.value = data.message || 'Login failed'
        return false
      }

      setToken(data.token)
      user.value = {
        user_id: data.user_id,
        username: data.username,
        nickname: data.nickname,
        avatar: data.avatar,
        chips: data.chips,
      }

      // 添加到最近账号
      addToRecentAccounts({
        user_id: data.user_id,
        username: data.username,
        nickname: data.nickname,
        avatar: data.avatar,
        token: data.token,
        chips: data.chips,
      })

      return true
    } catch (e) {
      error.value = 'Network error'
      return false
    } finally {
      isLoading.value = false
    }
  }

  async function fetchUserInfo(): Promise<boolean> {
    if (!token.value) return false

    isLoading.value = true
    error.value = ''

    try {
      const response = await fetch(`${serverStore.getHttpUrl()}${USER_API}/info`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token.value}`,
        },
      })

      const data = await response.json()

      if (!response.ok) {
        error.value = data.message || 'Failed to fetch user info'
        logout()
        return false
      }

      user.value = {
        user_id: data.user_id,
        username: data.username,
        nickname: data.nickname,
        avatar: data.avatar,
        chips: data.chips,
      }
      if (data.stats) {
        stats.value = data.stats
      }

      // 更新最近账号的筹码
      updateRecentAccountChips(data.chips)

      return true
    } catch (e) {
      error.value = 'Network error'
      return false
    } finally {
      isLoading.value = false
    }
  }

  function updateRecentAccountChips(chips: number) {
    if (!user.value) return
    const idx = recentAccounts.value.findIndex(a => a.user_id === user.value!.user_id)
    if (idx !== -1) {
      recentAccounts.value[idx].chips = chips
      saveRecentAccounts(recentAccounts.value)
    }
  }

  async function fetchStats(): Promise<boolean> {
    if (!token.value) return false

    try {
      const response = await fetch(`${serverStore.getHttpUrl()}${USER_API}/info`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token.value}`,
        },
      })

      const data = await response.json()

      if (response.ok && data.stats) {
        stats.value = data.stats
        if (user.value) {
          user.value.chips = data.chips
        }
        return true
      }
      return false
    } catch (e) {
      return false
    }
  }

  async function fetchGameHistory(): Promise<GameHistory[]> {
    if (!token.value) return []

    try {
      const response = await fetch(`${serverStore.getHttpUrl()}${USER_API}/history`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token.value}`,
        },
      })

      const data = await response.json()

      if (response.ok && data.games) {
        return data.games
      }
      return []
    } catch (e) {
      return []
    }
  }

  function logout() {
    token.value = ''
    user.value = null
    stats.value = null
    localStorage.removeItem('dz_token')
  }

  function updateChips(newChips: number) {
    if (user.value) {
      user.value.chips = newChips
      updateRecentAccountChips(newChips)
    }
  }

  // 检查是否是管理员
  const isAdmin = computed(() => user.value?.username === 'huahuo')

  // 获取所有用户（管理员）
  async function fetchAllUsers(): Promise<AdminUser[]> {
    if (!token.value) return []

    try {
      const response = await fetch(`${serverStore.getHttpUrl()}${ADMIN_API}/users`, {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token.value}`,
        },
      })

      if (response.ok) {
        return await response.json()
      }
      return []
    } catch {
      return []
    }
  }

  // 更新用户筹码（管理员）
  async function adminUpdateChips(userId: string, chips: number): Promise<boolean> {
    if (!token.value) return false

    try {
      const response = await fetch(`${serverStore.getHttpUrl()}${ADMIN_API}/chips`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token.value}`,
        },
        body: JSON.stringify({ user_id: userId, chips }),
      })

      return response.ok
    } catch {
      return false
    }
  }

  return {
    token,
    user,
    stats,
    isLoading,
    error,
    isAuthenticated,
    recentAccounts,
    login,
    register,
    logout,
    fetchUserInfo,
    fetchStats,
    fetchGameHistory,
    updateChips,
    switchToAccount,
    removeFromRecentAccounts,
    isAdmin,
    fetchAllUsers,
    adminUpdateChips,
  }
})
