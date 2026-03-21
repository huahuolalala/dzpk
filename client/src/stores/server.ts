import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

const STORAGE_KEY = 'poker_server_address'

export const useServerStore = defineStore('server', () => {
  const serverAddress = ref(localStorage.getItem(STORAGE_KEY) || 'localhost:8080')

  watch(serverAddress, (val) => {
    localStorage.setItem(STORAGE_KEY, val)
  })

  function getWsUrl(): string {
    const addr = serverAddress.value.trim()
    if (!addr) return 'ws://localhost:8080/ws'

    // 如果已经包含 ws:// 或 wss:// 前缀，直接使用
    if (addr.startsWith('ws://') || addr.startsWith('wss://')) {
      return addr.endsWith('/ws') ? addr : `${addr}/ws`
    }

    // 添加 ws:// 前缀和 /ws 后缀
    return `ws://${addr}/ws`
  }

  function getHttpUrl(): string {
    const addr = serverAddress.value.trim()
    if (!addr) return 'http://localhost:8080'

    // 如果已经包含 http:// 或 https:// 前缀，直接使用
    if (addr.startsWith('http://') || addr.startsWith('https://')) {
      return addr
    }

    // 添加 http:// 前缀
    return `http://${addr}`
  }

  function setAddress(addr: string) {
    serverAddress.value = addr
  }

  return {
    serverAddress,
    getWsUrl,
    getHttpUrl,
    setAddress,
  }
})
