import { defineStore } from 'pinia'
import { ref } from 'vue'

// 开发环境使用完整 URL，生产环境使用相对路径（nginx 代理）
const isDev = import.meta.env.DEV
const DEFAULT_SERVER_HOST = '8.145.38.16:8080'

export const useServerStore = defineStore('server', () => {
  const serverAddress = ref(DEFAULT_SERVER_HOST)

  function getWsUrl(): string {
    // 生产环境：使用相对路径，nginx 代理
    if (!isDev) {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      return `${protocol}//${window.location.host}/ws`
    }

    // 开发环境：使用完整 URL
    const addr = serverAddress.value.trim()
    if (!addr) return `ws://${DEFAULT_SERVER_HOST}/ws`

    if (addr.startsWith('ws://') || addr.startsWith('wss://')) {
      return addr.endsWith('/ws') ? addr : `${addr}/ws`
    }

    return `ws://${addr}/ws`
  }

  function getHttpUrl(): string {
    // 生产环境：使用相对路径，nginx 代理
    if (!isDev) {
      return '' // 空字符串表示相对路径
    }

    // 开发环境：使用完整 URL
    const addr = serverAddress.value.trim()
    if (!addr) return `http://${DEFAULT_SERVER_HOST}`

    if (addr.startsWith('http://') || addr.startsWith('https://')) {
      return addr
    }

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
