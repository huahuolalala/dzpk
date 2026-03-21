import { ref } from 'vue'

// 音效类型
export type SoundType = 'deal' | 'bet' | 'allin' | 'fold' | 'win' | 'click' | 'check' | 'card-reveal'

// 全局音频上下文
let audioContext: AudioContext | null = null

// 音量设置
const volume = ref(0.5)
const muted = ref(false)

// 获取或创建音频上下文
function getAudioContext(): AudioContext {
  if (!audioContext) {
    audioContext = new (window.AudioContext || (window as any).webkitAudioContext)()
  }
  return audioContext
}

// 生成合成音效
function createSound(type: SoundType, ctx: AudioContext): void {
  if (muted.value) return

  const gainNode = ctx.createGain()
  gainNode.connect(ctx.destination)
  gainNode.gain.value = volume.value * 0.3

  switch (type) {
    case 'deal':
    case 'card-reveal':
      // 发牌音效：短促的滑动音
      playDealSound(ctx, gainNode)
      break
    case 'bet':
      // 下注音效：筹码碰撞声
      playBetSound(ctx, gainNode)
      break
    case 'allin':
      // All-in音效：更强烈的筹码声
      playAllInSound(ctx, gainNode)
      break
    case 'fold':
      // 弃牌音效：低沉的滑动声
      playFoldSound(ctx, gainNode)
      break
    case 'win':
      // 获胜音效：欢快的上升音
      playWinSound(ctx, gainNode)
      break
    case 'click':
      // 点击音效：清脆的点击声
      playClickSound(ctx, gainNode)
      break
    case 'check':
      // 过牌音效：轻柔的提示音
      playCheckSound(ctx, gainNode)
      break
  }
}

function playDealSound(ctx: AudioContext, gainNode: GainNode): void {
  // 真实翻牌音效：模拟扑克牌翻转的物理声
  const now = ctx.currentTime
  const baseVolume = volume.value * 0.25

  // 1. 低频"啪"声 - 扑克牌落在桌面的声音
  const lowPop = ctx.createOscillator()
  const lowGain = ctx.createGain()
  lowPop.connect(lowGain)
  lowGain.connect(gainNode)
  lowPop.type = 'sine'
  lowPop.frequency.setValueAtTime(80 + Math.random() * 20, now)
  lowPop.frequency.exponentialRampToValueAtTime(40, now + 0.05)
  lowGain.gain.setValueAtTime(baseVolume * 0.8, now)
  lowGain.gain.exponentialRampToValueAtTime(0.001, now + 0.08)
  lowPop.start(now)
  lowPop.stop(now + 0.08)

  // 2. 中频"嗒"声 - 纸牌的质感
  const midClick = ctx.createOscillator()
  const midGain = ctx.createGain()
  midClick.connect(midGain)
  midGain.connect(gainNode)
  midClick.type = 'triangle'
  midClick.frequency.setValueAtTime(400 + Math.random() * 100, now + 0.01)
  midClick.frequency.exponentialRampToValueAtTime(200, now + 0.04)
  midGain.gain.setValueAtTime(baseVolume * 0.4, now + 0.01)
  midGain.gain.exponentialRampToValueAtTime(0.001, now + 0.05)
  midClick.start(now + 0.01)
  midClick.stop(now + 0.05)

  // 3. 高频滑动声 - 用噪声模拟纸牌摩擦
  const bufferSize = ctx.sampleRate * 0.06
  const noiseBuffer = ctx.createBuffer(1, bufferSize, ctx.sampleRate)
  const output = noiseBuffer.getChannelData(0)
  for (let i = 0; i < bufferSize; i++) {
    output[i] = Math.random() * 2 - 1
  }

  const noise = ctx.createBufferSource()
  noise.buffer = noiseBuffer

  // 带通滤波器 - 只保留纸牌摩擦的频段
  const bandpass = ctx.createBiquadFilter()
  bandpass.type = 'bandpass'
  bandpass.frequency.value = 3000 + Math.random() * 1000
  bandpass.Q.value = 2

  const noiseGain = ctx.createGain()
  noise.connect(bandpass)
  bandpass.connect(noiseGain)
  noiseGain.connect(gainNode)

  noiseGain.gain.setValueAtTime(0, now)
  noiseGain.gain.linearRampToValueAtTime(baseVolume * 0.15, now + 0.01)
  noiseGain.gain.exponentialRampToValueAtTime(0.001, now + 0.04)

  noise.start(now)
  noise.stop(now + 0.04)

  // 4. 轻微的高频"嘶"声 - 纸牌边缘
  const highSizzle = ctx.createOscillator()
  const highGain = ctx.createGain()
  highSizzle.connect(highGain)
  highGain.connect(gainNode)
  highSizzle.type = 'sine'
  highSizzle.frequency.value = 2000 + Math.random() * 500
  highGain.gain.setValueAtTime(baseVolume * 0.1, now + 0.015)
  highGain.gain.exponentialRampToValueAtTime(0.001, now + 0.035)
  highSizzle.start(now + 0.015)
  highSizzle.stop(now + 0.035)
}

function playBetSound(ctx: AudioContext, gainNode: GainNode): void {
  // 筹码碰撞声 - 多个频率叠加
  const frequencies = [200, 400, 600]
  frequencies.forEach((freq, i) => {
    const osc = ctx.createOscillator()
    const gain = ctx.createGain()
    osc.connect(gain)
    gain.connect(gainNode)
    osc.type = 'triangle'
    osc.frequency.value = freq + Math.random() * 50
    gain.gain.setValueAtTime(volume.value * 0.15, ctx.currentTime + i * 0.02)
    gain.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + 0.15 + i * 0.02)
    osc.start(ctx.currentTime + i * 0.02)
    osc.stop(ctx.currentTime + 0.15 + i * 0.02)
  })
}

function playAllInSound(ctx: AudioContext, gainNode: GainNode): void {
  // 更强烈的筹码声
  for (let i = 0; i < 5; i++) {
    const osc = ctx.createOscillator()
    const gain = ctx.createGain()
    osc.connect(gain)
    gain.connect(gainNode)
    osc.type = 'triangle'
    osc.frequency.value = 150 + i * 100 + Math.random() * 50
    gain.gain.setValueAtTime(volume.value * 0.2, ctx.currentTime + i * 0.03)
    gain.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + 0.2 + i * 0.03)
    osc.start(ctx.currentTime + i * 0.03)
    osc.stop(ctx.currentTime + 0.2 + i * 0.03)
  }
}

function playFoldSound(ctx: AudioContext, gainNode: GainNode): void {
  const osc = ctx.createOscillator()
  osc.connect(gainNode)
  osc.type = 'sine'
  osc.frequency.setValueAtTime(300, ctx.currentTime)
  osc.frequency.exponentialRampToValueAtTime(150, ctx.currentTime + 0.2)
  gainNode.gain.setValueAtTime(volume.value * 0.15, ctx.currentTime)
  gainNode.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + 0.2)
  osc.start(ctx.currentTime)
  osc.stop(ctx.currentTime + 0.2)
}

function playWinSound(ctx: AudioContext, gainNode: GainNode): void {
  // 获胜音效：上升的欢快音调
  const notes = [523.25, 659.25, 783.99, 1046.50] // C5, E5, G5, C6
  notes.forEach((freq, i) => {
    const osc = ctx.createOscillator()
    const gain = ctx.createGain()
    osc.connect(gain)
    gain.connect(gainNode)
    osc.type = 'sine'
    osc.frequency.value = freq
    gain.gain.setValueAtTime(0, ctx.currentTime + i * 0.1)
    gain.gain.linearRampToValueAtTime(volume.value * 0.15, ctx.currentTime + i * 0.1 + 0.02)
    gain.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + i * 0.1 + 0.3)
    osc.start(ctx.currentTime + i * 0.1)
    osc.stop(ctx.currentTime + i * 0.1 + 0.3)
  })
}

function playClickSound(ctx: AudioContext, gainNode: GainNode): void {
  const osc = ctx.createOscillator()
  osc.connect(gainNode)
  osc.type = 'sine'
  osc.frequency.value = 1000
  gainNode.gain.setValueAtTime(volume.value * 0.1, ctx.currentTime)
  gainNode.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + 0.05)
  osc.start(ctx.currentTime)
  osc.stop(ctx.currentTime + 0.05)
}

function playCheckSound(ctx: AudioContext, gainNode: GainNode): void {
  const osc = ctx.createOscillator()
  osc.connect(gainNode)
  osc.type = 'sine'
  osc.frequency.value = 600
  gainNode.gain.setValueAtTime(volume.value * 0.1, ctx.currentTime)
  gainNode.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + 0.08)
  osc.start(ctx.currentTime)
  osc.stop(ctx.currentTime + 0.08)
}

// 播放音效
export function playSound(type: SoundType): void {
  try {
    const ctx = getAudioContext()
    // 确保音频上下文处于运行状态（用户交互后激活）
    if (ctx.state === 'suspended') {
      ctx.resume()
    }
    createSound(type, ctx)
  } catch (e) {
    console.warn('Failed to play sound:', e)
  }
}

// Composable
export function useSound() {
  // 用户首次交互时激活音频上下文
  const activateAudio = () => {
    const ctx = getAudioContext()
    if (ctx.state === 'suspended') {
      ctx.resume()
    }
  }

  const setVolume = (v: number) => {
    volume.value = Math.max(0, Math.min(1, v))
  }

  const toggleMute = () => {
    muted.value = !muted.value
  }

  const mute = () => {
    muted.value = true
  }

  const unmute = () => {
    muted.value = false
  }

  return {
    volume,
    muted,
    playSound,
    setVolume,
    toggleMute,
    mute,
    unmute,
    activateAudio,
  }
}