<template>
  <div class="game-container">
    <!-- 操作历史面板 -->
    <ActionHistory />

    <!-- 顶部状态栏 -->
    <header class="game-header glass-panel">
      <div class="header-left">
        <div class="room-info">
          <span class="label">房间码</span>
          <span class="value">{{ roomStore.roomCode }}</span>
        </div>
      </div>

      <div class="header-center">
        <div class="phase-badge" :class="gameStore.phase">
          <span class="pulse-dot"></span>
          {{ phaseText }}
        </div>
      </div>

      <div class="header-right">
        <!-- 出牌顺序显示 -->
        <div v-if="gameStore.phase !== 'showdown' && turnOrder.length > 0" class="turn-order">
          <span
            v-for="item in turnOrder"
            :key="item.id"
            class="turn-order-item"
            :class="{ 'is-current': item.isCurrent }"
          >
            <span class="turn-num">{{ item.index }}</span>
            <span class="turn-name">{{ item.name }}</span>
          </span>
        </div>
        <button class="btn btn-secondary btn-sm" style="margin-right: 1rem;" @click="showHandRankHelp = true">牌型大小</button>
        <button class="btn btn-secondary btn-sm" @click="handleLeave">退出房间</button>
      </div>
    </header>

    <!-- All-in 通知横幅 -->
    <Transition name="slide-down">
      <div v-if="lastAllInAction && !gameStore.result" class="allin-notification">
        <span class="allin-icon">⚡</span>
        <span class="allin-text"><strong>{{ lastAllInAction.player_name }}</strong> ALL-IN 了！</span>
        <span class="allin-amount">🪙 {{ lastAllInAction.amount }}</span>
      </div>
    </Transition>

    <!-- 主视角对战桌 -->
    <main class="table-area">
      <div class="poker-table">
        
        <!-- 对手环绕区域 (上半部分) -->
        <div class="opponents-container">
          <div 
            v-for="player in opponents" 
            :key="player.id" 
            class="opponent-seat glass-panel"
            :class="{ 
              'is-active': gameStore.currentPlayer === player.seat && gameStore.phase !== 'showdown',
              'is-folded': player.status === 'fold',
              'is-allin': player.status === 'all-in'
            }"
          >
            <div class="active-ring" v-if="gameStore.currentPlayer === player.seat && gameStore.phase !== 'showdown'"></div>
            
            <div class="seat-top">
              <div class="avatar">{{ player.name.charAt(0).toUpperCase() }}</div>
              <div class="info">
                <div class="name">{{ player.name }}</div>
                <div class="chips">💰 {{ player.chips }}</div>
              </div>
            </div>
            
            <div class="seat-bottom">
              <div class="status-text" :class="player.status">
                {{ getStatusText(player.status) }}
              </div>
              <div class="blind-role" v-if="getBlindRole(player.seat)">
                {{ getBlindRole(player.seat) }}
              </div>
              <div class="bet-amount" v-if="player.bet > 0">
                🪙 {{ player.bet }}
              </div>
            </div>

            <!-- 对手手牌背面 -->
            <div class="opponent-cards" v-if="player.status !== 'fold' && player.cards && player.cards.length > 0">
              <div class="mini-card hidden"></div>
              <div class="mini-card hidden"></div>
            </div>
          </div>
        </div>

        <!-- 牌桌中央 (公共牌与底池) -->
        <div class="table-center">
          <div class="pot-display glass-panel">
            <span class="pot-label">总底池</span>
            <span class="pot-amount" :key="gameStore.pot">{{ displayPot }}</span>
          </div>
          
          <div class="community-cards">
            <div
              v-for="i in 5"
              :key="'cc-'+i"
              class="playing-card"
              :class="{
                'revealed': gameStore.communityCards[i-1],
                'flipping': flippingCards[i-1],
                'reveal-bounce': revealedCards[i-1] && !flippingCards[i-1],
                'is-red': isRedCard(gameStore.communityCards[i-1]),
                'is-highlight': isBestHandCard(gameStore.communityCards[i-1])
              }"
              :style="{ animationDelay: (i - 1) * 0.1 + 's' }"
            >
              <div class="card-inner" v-if="gameStore.communityCards[i-1]">
                {{ gameStore.communityCards[i-1] }}
              </div>
            </div>
          </div>
        </div>

        <!-- 弹幕层 -->
        <DanmakuOverlay />

      </div>
    </main>

    <!-- 底部：我的操作区与手牌 -->
    <footer class="bottom-area">
      <div class="my-zone">
        
        <!-- 我的状态信息 -->
        <div class="my-status glass-panel" v-if="myGamePlayer" :class="{ 'is-active': isMyTurn && gameStore.phase !== 'showdown' }">
          <div class="active-ring" v-if="isMyTurn && gameStore.phase !== 'showdown'"></div>
          <div class="info-row">
            <span class="my-name">{{ myGamePlayer.name }}</span>
            <span class="my-blind" v-if="getBlindRole(myGamePlayer.seat)">{{ getBlindRole(myGamePlayer.seat) }}</span>
            <span class="my-chips">💰 {{ myGamePlayer.chips }}</span>
          </div>
          <div class="info-row" v-if="myGamePlayer.bet > 0">
            <span class="my-bet">本轮下注: 🪙 {{ myGamePlayer.bet }}</span>
          </div>
          <div class="info-row" v-if="myGamePlayer.best_hand_rank">
            <span class="my-best-hand-badge">{{ translateHandRank(myGamePlayer.best_hand_rank) }}</span>
          </div>
        </div>

        <!-- 我的手牌 -->
        <div class="my-cards-section" v-if="myGamePlayer && myGamePlayer.cards && myGamePlayer.cards.length > 0">
          <div class="playing-card large revealed hole-card" :class="{
            'is-red': isRedCard(myGamePlayer.cards[0]),
            'is-highlight': isBestHandCard(myGamePlayer.cards[0])
          }">
             <div class="card-inner">{{ myGamePlayer.cards[0] || '?' }}</div>
          </div>
          <div class="playing-card large revealed hole-card" :class="{
            'is-red': isRedCard(myGamePlayer.cards[1]),
            'is-highlight': isBestHandCard(myGamePlayer.cards[1])
          }">
             <div class="card-inner">{{ myGamePlayer.cards[1] || '?' }}</div>
          </div>
        </div>

        <!-- 操作面板 -->
        <div class="action-controller glass-panel">
          <Transition name="fade" mode="out-in">
            <div v-if="gameStore.phase === 'showdown'" class="waiting-panel">
              <p>结算中...</p>
            </div>
            <div v-else-if="isMyTurn" class="my-turn-panel">
              <h3 class="turn-title">
                <span v-if="toCall === 0">轮到你了，你可以过牌或加注</span>
                <span v-else-if="cantAffordToCall">轮到你了，<span class="text-danger">筹码不够</span>，只能全下或弃牌</span>
                <span v-else>轮到你了，需跟注 <strong class="text-gold">{{ toCall }}</strong></span>
              </h3>
              <div class="action-buttons">
                <button
                  v-if="availableButtons.includes('fold')"
                  class="btn btn-danger-outline btn-lg"
                  @click="doAction('fold')"
                >弃牌</button>
                <button
                  v-if="availableButtons.includes('check')"
                  class="btn btn-primary btn-lg"
                  @click="doAction('check')"
                >过牌 (Check)</button>
                <button
                  v-if="availableButtons.includes('call')"
                  class="btn btn-primary btn-lg"
                  @click="doAction('call')"
                >跟注 {{ actualCallAmount }}</button>
                <button
                  v-if="availableButtons.includes('raise')"
                  class="btn btn-warning btn-lg"
                  @click="showRaiseModal = true"
                >加注 (Raise)</button>
                <button
                  v-if="availableButtons.includes('allin')"
                  class="btn btn-danger btn-lg"
                  @click="doAction('allin')"
                >All-in</button>
              </div>
            </div>
            
            <div v-else class="waiting-panel">
              <div class="spinner"></div>
              <p>等待 <span class="highlight">{{ currentPlayerName }}</span> 行动...</p>
            </div>
          </Transition>
        </div>
        
      </div>
    </footer>

    <!-- 加注弹窗 -->
    <Transition name="fade">
      <div v-if="showRaiseModal" class="modal-overlay">
        <div class="modal-content glass-panel">
          <h3>加注</h3>
          <p class="modal-desc">当前需跟注: {{ toCall }}，最低加注额: {{ gameStore.minRaise }}</p>

          <div class="raise-input-group">
            <input
              v-model.number="raiseAmount"
              type="number"
              class="input text-center"
              :min="gameStore.minRaise"
              :max="myGamePlayer?.chips || 0"
            />
          </div>

          <div class="modal-actions">
            <button class="btn btn-secondary" @click="showRaiseModal = false">取消</button>
            <button class="btn btn-warning" @click="confirmRaise" :disabled="raiseAmount < gameStore.minRaise">
              确认加注
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- All-in 确认弹窗 -->
    <Transition name="fade">
      <div v-if="showAllInModal" class="modal-overlay">
        <div class="modal-content glass-panel">
          <h3 class="text-danger">⚠️ 确认 All-in</h3>
          <p class="modal-desc">你确定要 all-in 吗？</p>
          <p class="all-in-amount">将投入 {{ myGamePlayer?.chips || 0 }} 筹码</p>

          <div class="modal-actions">
            <button class="btn btn-secondary" @click="showAllInModal = false">取消</button>
            <button class="btn btn-danger btn-lg" @click="confirmAllIn">
              确定 All-in
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- 错误提示弹窗 -->
    <Transition name="fade">
      <div v-if="errorMessage" class="modal-overlay">
        <div class="modal-content glass-panel error-modal">
          <h3 class="text-danger">❌ 操作失败</h3>
          <p class="modal-desc">{{ errorMessage }}</p>
          <div class="modal-actions">
            <button class="btn btn-primary" @click="errorMessage = ''">知道了</button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- 摊牌演出动画 -->
    <Transition name="showdown">
      <div v-if="showShowdown" class="showdown-overlay">
        <div class="showdown-content">
          <!-- 标题 -->
          <div class="showdown-title" :class="{ 'animate': showdownTitleAnimate }">
            <span class="title-text">{{ showdownPhaseText }}</span>
          </div>

          <!-- 公共牌展示 -->
          <div class="community-cards-showcase">
            <div class="community-label">公共牌</div>
            <div class="community-cards-row">
              <div
                v-for="(_, idx) in 5"
                :key="'cc-' + idx"
                class="playing-card large showcase-card"
                :class="{
                  'is-revealed': gameStore.communityCards[idx],
                  'is-red': isRedCard(gameStore.communityCards[idx])
                }"
              >
                <div class="card-inner" v-if="gameStore.communityCards[idx]">
                  {{ gameStore.communityCards[idx] }}
                </div>
              </div>
            </div>
          </div>

          <!-- 当前正在亮牌的用户 -->
          <div v-if="currentRevealPlayer" class="current-reveal">
            <div class="reveal-player-name">{{ currentRevealPlayer.name }}</div>
            <div class="reveal-cards">
              <div
                v-for="(_, idx) in 2"
                :key="'reveal-' + idx"
                class="playing-card large reveal-card"
                :class="{
                  'is-flipped': currentRevealCardIdx > idx,
                  'is-red': currentRevealCardIdx > idx && isRedCard(currentRevealPlayer.cards[idx])
                }"
              >
                <div class="card-inner" v-if="currentRevealCardIdx > idx">
                  {{ currentRevealPlayer.cards[idx] }}
                </div>
              </div>
            </div>
            <div v-if="currentRevealCardIdx > 1" class="reveal-rank animate-in">
              <span class="rank-label">{{ translateHandRank(currentRevealPlayer.best_hand_rank) }}</span>
            </div>
          </div>

          <!-- 所有玩家排名展示 -->
          <div class="rankings-container" :class="{ 'show': showRankings }">
            <div class="rankings-title">牌力排名</div>
            <div class="rankings-list">
              <div
                v-for="(player, idx) in rankedPlayers"
                :key="player.id"
                class="ranking-item"
                :class="{
                  'is-winner': isWinner(player.id),
                  'is-current': currentRevealPlayer?.id === player.id
                }"
                :style="{ animationDelay: idx * 0.1 + 's' }"
              >
                <div class="ranking-position">
                  <span v-if="idx === 0" class="crown">👑</span>
                  <span v-else class="position-num">#{{ idx + 1 }}</span>
                </div>
                <div class="ranking-info">
                  <div class="ranking-name">{{ player.name }}</div>
                  <div class="ranking-cards">{{ player.cards.join(' ') }}</div>
                </div>
                <div class="ranking-hand">
                  <span class="hand-name">{{ translateHandRank(player.best_hand_rank) }}</span>
                  <span class="hand-cards">{{ getBestHandDisplay(player) }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- 继续按钮 -->
          <div v-if="showContinueBtn" class="showdown-actions">
            <button class="btn btn-primary btn-lg" @click="showResultModal">查看结算</button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- 结算弹窗 -->
    <Transition name="result-modal">
      <div v-if="gameStore.result" class="modal-overlay result-overlay">
        <div class="result-modal glass-panel">
          <div class="result-header">
            <h2 class="result-title">🎉 本局结束 🎉</h2>
          </div>

          <div class="total-pot-display">
            <span class="total-pot-label">总底池</span>
            <span class="total-pot-amount">🪙 {{ animatedTotalPot }}</span>
          </div>

          <div class="result-scroll-area">
            <div class="players-result" :class="{ 'is-compact': gameStore.players.length > 5, 'is-sparse': gameStore.players.length <= 3 }">
              <div
                v-for="(player, idx) in gameStore.players"
                :key="player.id"
                class="player-result-item"
                :class="{ 'is-winner': isWinner(player.id) }"
                :style="{ animationDelay: idx * 0.15 + 's' }"
              >
                <div class="pr-info">
                  <span class="pr-name">{{ player.name }}</span>
                  <span class="pr-cards" v-if="player.cards.length">{{ player.cards.join(' ') }}</span>
                </div>
                <div class="pr-change" :class="getNetChange(player.id) >= 0 ? 'is-win' : 'is-loss'">
                  <span class="pr-initial">({{ player.initialChips }})</span>
                  <span class="pr-current">{{ animatedChips[player.id] || player.initialChips }}</span>
                  <span class="pr-net" :class="{ 'is-animating': isNetAnimating[player.id] }">
                    {{ getNetChange(player.id) >= 0 ? '+' : '' }}{{ animatedNetChange[player.id] || 0 }}
                  </span>
                </div>
              </div>
            </div>

            <div class="winners-section" v-if="gameStore.result.winners?.length">
              <h3 class="winners-title">🏆 赢家</h3>
              <div class="winners-list">
                <div v-for="w in gameStore.result.winners" :key="w.player_id" class="winner-item">
                  <span class="w-name">{{ w.player_name }}</span>
                  <span class="w-rank">{{ translateHandRank(w.hand_rank) }}</span>
                  <span class="w-amount">+{{ animatedWinAmount[w.player_id] || 0 }}</span>
                </div>
              </div>
            </div>
          </div>

          <button class="btn btn-primary btn-lg" style="width: 100%; margin-top: 1.5rem; flex-shrink: 0;" @click="closeResultAndReset">
            返回等待房间
          </button>
        </div>
      </div>
    </Transition>

    <!-- 牌型帮助弹窗 -->
    <Transition name="fade">
      <div v-if="showHandRankHelp" class="modal-overlay" @click.self="showHandRankHelp = false">
        <div class="rank-help-modal glass-panel">
          <button class="close-btn" @click="showHandRankHelp = false">✕</button>
          <HandRankShowcase />
        </div>
      </div>
    </Transition>

    <!-- 聊天输入 -->
    <ChatInput />

  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useRoomStore } from '../stores/room'
import { useGameStore } from '../stores/game'
import { useWebSocket } from '../composables/useWebSocket'
import { useSound, playSound } from '../composables/useSound'
import HandRankShowcase from '../components/HandRankShowcase.vue'
import ActionHistory from '../components/ActionHistory.vue'
import DanmakuOverlay from '../components/DanmakuOverlay.vue'
import ChatInput from '../components/ChatInput.vue'

const router = useRouter()
const roomStore = useRoomStore()
const gameStore = useGameStore()
const { leaveRoom, playerAction, error: wsError } = useWebSocket()
const { activateAudio } = useSound()

// 用户首次交互时激活音频
onMounted(() => {
  const handleFirstInteraction = () => {
    activateAudio()
    document.removeEventListener('click', handleFirstInteraction)
  }
  document.addEventListener('click', handleFirstInteraction)
})

// 监听 WebSocket 错误
watch(wsError, (err) => {
  if (err) {
    errorMessage.value = err
  }
})

const showRaiseModal = ref(false)
const showHandRankHelp = ref(false)
const showAllInModal = ref(false)
const raiseAmount = ref(0)
const errorMessage = ref('')

// 摊牌演出动画状态
const showShowdown = ref(false)
const showdownPlayers = ref<any[]>([])
const currentRevealPlayer = ref<any>(null)
const currentRevealCardIdx = ref(0)
const showdownTitleAnimate = ref(false)
const showRankings = ref(false)
const showContinueBtn = ref(false)
const rankedPlayers = ref<any[]>([])

const showdownPhaseText = computed(() => {
  if (!showRankings.value) return 'SHOWDOWN'
  return '最终排名'
})

// 监听 showdown 阶段开始摊牌演出
watch(() => gameStore.phase, (phase) => {
  if (phase === 'showdown' && gameStore.players.length > 0) {
    startShowdownAnimation()
  }
})

function startShowdownAnimation() {
  // 过滤出未 fold 的玩家，按座位顺序排列
  showdownPlayers.value = gameStore.players
    .filter(p => p.status !== 'fold')
    .sort((a, b) => a.seat - b.seat)

  rankedPlayers.value = [...showdownPlayers.value].sort((a, b) => {
    return compareHands(b.best_hand_rank, a.best_hand_rank, b.best_hand_cards, a.best_hand_cards)
  })

  currentRevealPlayer.value = null
  currentRevealCardIdx.value = 0
  showdownTitleAnimate.value = false
  showRankings.value = false
  showContinueBtn.value = false
  showShowdown.value = true

  // 开始演出 - 减慢整体节奏
  setTimeout(() => {
    showdownTitleAnimate.value = true
    setTimeout(() => {
      revealNextPlayer()
    }, 1200) // 加长标题展示时间
  }, 500) // 加长初始等待
}

function revealNextPlayer() {
  const currentIdx = showdownPlayers.value.findIndex(p => p.id === currentRevealPlayer.value?.id)
  const nextIdx = currentIdx + 1

  if (nextIdx < showdownPlayers.value.length) {
    currentRevealPlayer.value = showdownPlayers.value[nextIdx]
    currentRevealCardIdx.value = 0

    // 逐张翻牌 - 减慢速度
    setTimeout(() => {
      currentRevealCardIdx.value = 1
      setTimeout(() => {
        currentRevealCardIdx.value = 2
        // 翻完两张后等待更长时间，营造悬念感
        setTimeout(() => {
          revealNextPlayer()
        }, 2000) // 加长等待时间
      }, 800) // 加长翻牌间隔
    }, 600) // 加长初始等待
  } else {
    // 所有玩家都翻完了，显示排名
    setTimeout(() => {
      showRankings.value = true
      setTimeout(() => {
        showContinueBtn.value = true
      }, rankedPlayers.value.length * 150 + 800) // 加长排名动画
    }, 1200)
  }
}

// 比较两手牌大小（降序排列：返回负数表示A排在B前面）
function compareHands(rankA: string, rankB: string, cardsA: string[], cardsB: string[]): number {
  const rankOrder: Record<string, number> = {
    'High Card': 0, 'One Pair': 1, 'Two Pair': 2, 'Three of a Kind': 3,
    'Straight': 4, 'Flush': 5, 'Full House': 6, 'Four of a Kind': 7,
    'Straight Flush': 8, 'Royal Flush': 9
  }
  const orderA = rankOrder[rankA] ?? 0
  const orderB = rankOrder[rankB] ?? 0
  if (orderA !== orderB) return orderB - orderA  // 降序：牌力大的排前面

  // 同牌型时，比较手牌大小
  return compareCardValues(cardsA, cardsB)
}

// 比较手牌列表的大小（降序）
function compareCardValues(cardsA: string[], cardsB: string[]): number {
  const cardOrder: Record<string, number> = {
    '2': 2, '3': 3, '4': 4, '5': 5, '6': 6, '7': 7, '8': 8, '9': 9, '10': 10,
    'J': 11, 'Q': 12, 'K': 13, 'A': 14
  }

  // 提取每张牌的数值（去掉花色）
  const valuesA = cardsA.map(c => cardOrder[c.slice(0, -1)] || 0).sort((a, b) => b - a)
  const valuesB = cardsB.map(c => cardOrder[c.slice(0, -1)] || 0).sort((a, b) => b - a)

  // 从大到小比较
  for (let i = 0; i < Math.min(valuesA.length, valuesB.length); i++) {
    if (valuesA[i] !== valuesB[i]) {
      return valuesA[i] - valuesB[i]  // 降序
    }
  }
  return 0
}

function getBestHandDisplay(player: any): string {
  if (!player.best_hand_cards || player.best_hand_cards.length === 0) return ''
  return player.best_hand_cards.slice(0, 5).join(' ')
}

function showResultModal() {
  showShowdown.value = false
}

// 翻牌动画状态
const flippingCards = ref<boolean[]>([false, false, false, false, false])
const revealedCards = ref<boolean[]>([false, false, false, false, false])

// 监听公共牌变化，触发翻牌动画
watch(() => gameStore.communityCards, (newCards, oldCards) => {
  if (!newCards || !oldCards) return

  newCards.forEach((card, index) => {
    if (card && !oldCards[index]) {
      // 新牌触发动画
      flippingCards.value[index] = true
      // 播放发牌音效
      playSound('deal')

      // 翻转动画结束后显示bounce效果
      setTimeout(() => {
        flippingCards.value[index] = false
        revealedCards.value[index] = true
        // 翻牌完成音效
        playSound('card-reveal')
      }, 600 + index * 100)
    }
  })
}, { deep: true })

// 底池数字滚动动画
const displayPot = ref(0)
let potAnimationFrame: number | null = null

watch(() => gameStore.pot, (newPot, oldPot) => {
  if (oldPot === undefined || oldPot === null) {
    displayPot.value = newPot
    return
  }

  const start = oldPot
  const end = newPot
  const duration = 500
  const startTime = performance.now()

  const animate = (currentTime: number) => {
    const elapsed = currentTime - startTime
    const progress = Math.min(elapsed / duration, 1)

    // easeOutCubic
    const eased = 1 - Math.pow(1 - progress, 3)
    displayPot.value = Math.round(start + (end - start) * eased)

    if (progress < 1) {
      potAnimationFrame = requestAnimationFrame(animate)
    }
  }

  if (potAnimationFrame) {
    cancelAnimationFrame(potAnimationFrame)
  }
  potAnimationFrame = requestAnimationFrame(animate)
}, { immediate: true })

// 结算动画状态
const animatedTotalPot = ref(0)
const animatedChips = ref<Record<string, number>>({})
const animatedNetChange = ref<Record<string, number>>({})
const animatedWinAmount = ref<Record<string, number>>({})
const isNetAnimating = ref<Record<string, boolean>>({})

// 结算动画
watch(() => gameStore.result, (result) => {
  if (!result) return

  // 播放获胜音效
  playSound('win')

  // 重置动画状态
  animatedChips.value = {}
  animatedNetChange.value = {}
  animatedWinAmount.value = {}
  isNetAnimating.value = {}

  const totalDuration = 1500
  const startTime = performance.now()

  // 底池动画
  const animatePot = () => {
    const elapsed = performance.now() - startTime
    const progress = Math.min(elapsed / totalDuration, 1)
    const eased = 1 - Math.pow(1 - progress, 3)
    animatedTotalPot.value = Math.round(gameStore.pot * eased)

    if (progress < 1) {
      requestAnimationFrame(animatePot)
    }
  }
  requestAnimationFrame(animatePot)

  // 玩家结算动画
  gameStore.players.forEach((player: any, idx: number) => {
    const playerDelay = 300 + idx * 150
    const netChange = getNetChange(player.id)

    setTimeout(() => {
      isNetAnimating.value[player.id] = true

      const duration = 800
      const playerStart = performance.now()

      const animatePlayer = () => {
        const elapsed = performance.now() - playerStart
        const progress = Math.min(elapsed / duration, 1)
        const eased = 1 - Math.pow(1 - progress, 3)

        animatedNetChange.value[player.id] = Math.round(netChange * eased)
        animatedChips.value[player.id] = Math.round(player.initialChips + netChange * eased)

        if (progress < 1) {
          requestAnimationFrame(animatePlayer)
        } else {
          isNetAnimating.value[player.id] = false
        }
      }
      requestAnimationFrame(animatePlayer)
    }, playerDelay)
  })

  // 赢家金额动画
  result.winners?.forEach((w: any, idx: number) => {
    setTimeout(() => {
      const duration = 600
      const winStart = performance.now()

      const animateWin = () => {
        const elapsed = performance.now() - winStart
        const progress = Math.min(elapsed / duration, 1)
        const eased = 1 - Math.pow(1 - progress, 3)
        animatedWinAmount.value[w.player_id] = Math.round(w.win_amount * eased)

        if (progress < 1) {
          requestAnimationFrame(animateWin)
        }
      }
      requestAnimationFrame(animateWin)
    }, 600 + idx * 200)
  })
})

function isWinner(playerId: string): boolean {
  return gameStore.result?.winners?.some((w: any) => w.player_id === playerId) || false
}

const myGamePlayer = computed(() => {
  return gameStore.players.find(p => p.id === roomStore.playerId)
})

const opponents = computed(() => {
  return gameStore.players.filter(p => p.id !== roomStore.playerId)
})

const isMyTurn = computed(() => {
  return gameStore.players[gameStore.currentPlayer]?.id === roomStore.playerId
})

const currentPlayerName = computed(() => {
  return gameStore.players[gameStore.currentPlayer]?.name || ''
})

// 出牌顺序（从庄家位置开始，按顺序循环，跳过已弃牌玩家）
const turnOrder = computed(() => {
  const players = gameStore.players
  if (!players || players.length === 0) return []

  // 获取当前出牌玩家的 ID（currentPlayer 是数组索引）
  const currentPlayerId = players[gameStore.currentPlayer]?.id

  const result: Array<{ index: number; id: string; name: string; isCurrent: boolean }> = []
  const dealerIndex = gameStore.dealerIndex

  for (let i = 0; i < players.length; i++) {
    const pos = (dealerIndex + i) % players.length
    const player = players[pos]

    // 跳过已弃牌的玩家
    if (player.status === 'fold') {
      continue
    }

    result.push({
      index: result.length + 1,
      id: player.id,
      name: player.name,
      isCurrent: player.id === currentPlayerId,
    })
  }

  return result
})

const toCall = computed(() => {
  if (!myGamePlayer.value) return 0
  const required = gameStore.currentBet - myGamePlayer.value.bet
  return Math.max(0, required)
})

// 实际跟注额（不能超过玩家剩余筹码）
const actualCallAmount = computed(() => {
  if (!myGamePlayer.value) return 0
  return Math.min(toCall.value, myGamePlayer.value.chips)
})

// 我是否负担不起跟注（当前最大下注额 > 我手上有筹码 + 已下注）
const cantAffordToCall = computed(() => {
  return gameStore.currentBet > ((myGamePlayer.value?.chips || 0) + (myGamePlayer.value?.bet || 0))
})

// 场上是否有其他玩家已 All-in
const hasAllInPlayer = computed(() => {
  return gameStore.players.some(p => p.status === 'all-in' && p.id !== roomStore.playerId)
})

// 可用操作按钮列表
const availableButtons = computed(() => {
  // 已 fold 或 all-in 的玩家不能操作
  if (myGamePlayer.value?.status === 'fold' || myGamePlayer.value?.status === 'all-in') {
    return []
  }

  // 当最大下注额大于玩家手上筹码+已下筹码时，只能 All-in 或 Fold
  if (cantAffordToCall.value) {
    return ['fold', 'allin']
  }

  const buttons: string[] = []

  // 弃牌始终可选
  buttons.push('fold')

  // 过牌/跟注
  if (toCall.value === 0) {
    buttons.push('check')
    // 没人 all-in 时才能加注
    if (!hasAllInPlayer.value) {
      buttons.push('raise')
    }
  } else {
    // toCall > 0 的情况
    // 有玩家 all-in 时，其他玩家的选项：
    // - 如果筹码刚好等于跟注额，只显示 call（等价于 all-in）
    // - 如果筹码 > 跟注额，可以 call（跟注）或 all-in（加注）
    if (hasAllInPlayer.value) {
      // 有人 all-in 了
      if (chips <= toCall.value) {
        // 筹码不够或刚好等于跟注额，只显示 call
        buttons.push('call')
      } else {
        // 筹码比跟注额多，可以跟注或 all-in 加注
        buttons.push('call', 'allin')
      }
    } else {
      // 没人 all-in
      if (chips <= toCall.value) {
        // 筹码不够跟注，All-in 和 Call 等价，只显示 Call
        buttons.push('call')
      } else {
        // 筹码够跟注，显示 Call、All-in 和 Raise
        buttons.push('call', 'allin', 'raise')
      }
    }
  }

  return buttons
})

// 最近的 all-in 动作（用于通知）
const lastAllInAction = computed(() => {
  const allInActions = gameStore.actions.filter(a => a.action === 'allin')
  return allInActions.length > 0 ? allInActions[allInActions.length - 1] : null
})

const phaseText = computed(() => {
  const p = gameStore.phase
  if (p === 'preflop') return '发牌阶段 (Pre-Flop)'
  if (p === 'flop') return '翻牌圈 (Flop)'
  if (p === 'turn') return '转牌圈 (Turn)'
  if (p === 'river') return '河牌圈 (River)'
  if (p === 'showdown') return '摊牌结算 (Showdown)'
  return p
})

const isRedCard = (cardStr?: string) => {
  if (!cardStr) return false
  return cardStr.includes('♥') || cardStr.includes('♦')
}

const isBestHandCard = (cardStr?: string) => {
  if (!cardStr || !myGamePlayer.value || !myGamePlayer.value.best_hand_cards) return false
  return myGamePlayer.value.best_hand_cards.includes(cardStr)
}

const getStatusText = (status: string) => {
  if (status === 'fold') return '已弃牌'
  if (status === 'all-in') return 'All-In'
  return '游戏中'
}

// 获取玩家的盲注角色（小盲/大盲）
const getBlindRole = (seatIndex: number): string => {
  if (gameStore.phase === 'showdown') return ''
  if (gameStore.smallBlindIndex === seatIndex) return '小盲'
  if (gameStore.bigBlindIndex === seatIndex) return '大盲'
  return ''
}

const translateHandRank = (rank: string) => {
  const map: Record<string, string> = {
    'High Card': '高牌',
    'One Pair': '一对',
    'Two Pair': '两对',
    'Three of a Kind': '三条',
    'Straight': '顺子',
    'Flush': '同花',
    'Full House': '葫芦',
    'Four of a Kind': '四条',
    'Straight Flush': '同花顺',
    'Royal Flush': '皇家同花顺'
  }
  return map[rank] || rank
}

async function doAction(action: string, amount?: number) {
  // 播放操作音效
  switch (action) {
    case 'fold':
      playSound('fold')
      break
    case 'check':
      playSound('check')
      break
    case 'call':
    case 'raise':
      playSound('bet')
      break
    case 'allin':
      playSound('allin')
      break
  }

  await playerAction(action, amount)
  showRaiseModal.value = false
}

async function confirmRaise() {
  if (raiseAmount.value >= gameStore.minRaise) {
    await doAction('raise', raiseAmount.value)
  }
}

async function confirmAllIn() {
  showAllInModal.value = false
  await doAction('allin')
}

async function handleLeave() {
  gameStore.reset()
  await leaveRoom()
  router.push('/')
}

function closeResultAndReset() {
  gameStore.clearResult()
  gameStore.reset()
  router.push('/room')
}

function getNetChange(playerId: string): number {
  return gameStore.getNetChange(playerId)
}
</script>

<style scoped>
.game-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--bg-table);
  position: relative;
  overflow: hidden;
}

/* Header */
.game-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 2rem;
  border-radius: 0 0 16px 16px;
  border-top: none;
  z-index: 10;
}

.header-left, .header-right {
  flex: 1;
}

.header-right {
  text-align: right;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.header-center {
  flex: 2;
  display: flex;
  justify-content: center;
}

.room-info {
  display: flex;
  flex-direction: column;
}

.room-info .label {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.room-info .value {
  font-weight: 700;
  font-size: 1.125rem;
  letter-spacing: 1px;
}

.phase-badge {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  background: rgba(0, 0, 0, 0.3);
  padding: 0.5rem 1.5rem;
  border-radius: 20px;
  font-weight: 600;
  letter-spacing: 1px;
  border: 1px solid var(--glass-border);
}

.pulse-dot {
  width: 8px;
  height: 8px;
  background: var(--primary);
  border-radius: 50%;
  box-shadow: 0 0 8px var(--primary);
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(16, 185, 129, 0.7); }
  70% { transform: scale(1); box-shadow: 0 0 0 6px rgba(16, 185, 129, 0); }
  100% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(16, 185, 129, 0); }
}

/* Table Area */
.table-area {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  position: relative;
}

.poker-table {
  width: 100%;
  max-width: 1000px;
  height: 100%;
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

/* Opponents */
.opponents-container {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  display: flex;
  justify-content: center;
  flex-wrap: wrap;
  gap: 2rem;
  padding: 1rem;
}

.opponent-seat {
  position: relative;
  width: 160px;
  padding: 0.75rem;
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  transition: all 0.3s ease;
}

.opponent-seat.is-active {
  transform: translateY(-5px);
}

.opponent-seat.is-folded {
  opacity: 0.4;
  filter: grayscale(100%);
}

.active-ring {
  position: absolute;
  top: -2px; left: -2px; right: -2px; bottom: -2px;
  border: 2px solid var(--gold);
  border-radius: 14px;
  box-shadow: 0 0 15px rgba(251, 191, 36, 0.4);
  animation: ring-pulse 2s infinite;
  pointer-events: none;
}

@keyframes ring-pulse {
  0% { opacity: 0.6; }
  50% { opacity: 1; }
  100% { opacity: 0.6; }
}

.seat-top {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.opponent-seat .avatar {
  width: 36px; height: 36px;
  background: rgba(255,255,255,0.1);
  border-radius: 8px;
  display: flex; align-items: center; justify-content: center;
  font-weight: 600;
}

.opponent-seat .info {
  flex: 1;
  overflow: hidden;
}

.opponent-seat .name {
  font-size: 0.875rem;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.opponent-seat .chips {
  font-size: 0.75rem;
  color: var(--gold);
}

.seat-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.75rem;
}

.status-text {
  color: var(--primary);
}
.status-text.fold { color: var(--text-muted); }
.status-text.all-in { color: var(--danger); font-weight: bold; }

.blind-role {
  font-size: 0.625rem;
  color: #8b5cf6;
  font-weight: bold;
  background: rgba(139, 92, 246, 0.2);
  padding: 0.125rem 0.375rem;
  border-radius: 4px;
}

.bet-amount {
  background: rgba(0,0,0,0.3);
  padding: 0.2rem 0.5rem;
  border-radius: 4px;
  color: var(--text-main);
}

.opponent-cards {
  position: absolute;
  bottom: -20px;
  right: 10px;
  display: flex;
  gap: -10px;
  transform: scale(0.8);
}

.mini-card.hidden {
  width: 24px; height: 36px;
  background: repeating-linear-gradient(45deg, #1e293b, #1e293b 5px, #0f172a 5px, #0f172a 10px);
  border: 1px solid white;
  border-radius: 4px;
  margin-left: -10px;
  box-shadow: 2px 2px 5px rgba(0,0,0,0.5);
}

/* Table Center */
.table-center {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2rem;
  margin-top: 8rem; /* push down slightly to balance opponents */
}

.pot-display {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0.75rem 2.5rem;
  border-radius: 24px;
  background: rgba(0, 0, 0, 0.4);
  border: 1px solid rgba(251, 191, 36, 0.3);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5), inset 0 0 15px rgba(251, 191, 36, 0.1);
}

.pot-label {
  font-size: 0.75rem;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 2px;
}

.pot-amount {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--gold);
  text-shadow: 0 2px 4px rgba(0,0,0,0.5);
}

.community-cards {
  display: flex;
  gap: 0.75rem;
}

.playing-card.is-highlight {
  box-shadow: 0 0 15px rgba(251, 191, 36, 0.8), inset 0 0 0 2px rgba(251, 191, 36, 1);
  transform: translateY(-5px);
  z-index: 5;
}

/* Bottom Area */
.bottom-area {
  padding: 1rem 2rem 2rem;
  display: flex;
  justify-content: center;
  z-index: 10;
}

.my-zone {
  display: flex;
  align-items: flex-end;
  gap: 2rem;
  max-width: 1200px;
  width: 100%;
}

.my-status {
  position: relative;
  padding: 1rem 1.5rem;
  border-radius: 12px;
  min-width: 180px;
}

.my-status.is-active {
  background: rgba(20, 30, 25, 0.85);
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  margin-bottom: 0.5rem;
}
.info-row:last-child { margin-bottom: 0; }

.my-name { font-weight: 600; font-size: 1.125rem; }
.my-blind { font-size: 0.75rem; color: #8b5cf6; font-weight: bold; margin-left: 0.5rem; background: rgba(139, 92, 246, 0.2); padding: 0.125rem 0.5rem; border-radius: 4px; }
.my-chips { color: var(--gold); font-weight: 600; }
.my-bet { font-size: 0.875rem; color: var(--text-muted); background: rgba(0,0,0,0.3); padding: 0.25rem 0.5rem; border-radius: 4px; }
.my-best-hand-badge {
  display: inline-block;
  background: rgba(251, 191, 36, 0.2);
  color: var(--gold);
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-size: 0.875rem;
  font-weight: 600;
  border: 1px solid rgba(251, 191, 36, 0.4);
  box-shadow: 0 0 10px rgba(251, 191, 36, 0.2);
  margin-top: 0.25rem;
}

.my-cards-section {
  display: flex;
  gap: 0.5rem;
  transform: translateY(-10px);
}

.action-controller {
  flex: 1;
  padding: 1.5rem;
  border-radius: 16px;
  min-height: 120px;
  display: flex;
  align-items: center;
}

.my-turn-panel {
  width: 100%;
}

.turn-title {
  font-size: 1rem;
  font-weight: 400;
  color: var(--text-main);
  margin-bottom: 1rem;
  text-align: center;
}

.text-gold {
  color: var(--gold);
  font-weight: 700;
  font-size: 1.125rem;
}

.action-buttons {
  display: flex;
  justify-content: center;
  gap: 1rem;
}

.action-buttons .btn {
  min-width: 120px;
}

.waiting-panel {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  color: var(--text-muted);
  font-size: 1.125rem;
}

.highlight {
  color: var(--primary);
  font-weight: 600;
}

/* Modals */
.modal-overlay {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}

.modal-content {
  padding: 2rem;
  width: 100%;
  max-width: 400px;
  text-align: center;
}

.modal-content h3 { margin-bottom: 0.5rem; color: var(--gold); }
.modal-desc { color: var(--text-muted); font-size: 0.875rem; margin-bottom: 1.5rem; }

.raise-input-group {
  margin-bottom: 1.5rem;
}

.raise-input-group .input {
  font-size: 2rem;
  font-weight: 700;
  color: var(--gold);
  padding: 1rem;
}

.modal-actions {
  display: flex;
  gap: 1rem;
}
.modal-actions .btn { flex: 1; }

/* Result Modal */
.result-overlay {
  background: rgba(0, 0, 0, 0.85);
  backdrop-filter: blur(8px);
}

.result-modal {
  width: 100%;
  max-width: 500px;
  max-height: 90vh;
  padding: 2.5rem;
  border: 1px solid rgba(251, 191, 36, 0.3);
  box-shadow: 0 0 40px rgba(251, 191, 36, 0.15);
  display: flex;
  flex-direction: column;
}

.result-header h2 {
  color: var(--gold);
  font-size: 2rem;
  text-align: center;
  margin-bottom: 2rem;
  text-shadow: 0 2px 10px rgba(251, 191, 36, 0.4);
}

.winners-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.winner-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  background: rgba(251, 191, 36, 0.05);
  border-color: rgba(251, 191, 36, 0.2);
}

.w-info {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.w-name {
  font-weight: 600;
  font-size: 1.125rem;
}

.w-rank {
  font-size: 0.875rem;
  color: var(--text-muted);
}

.w-amount {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--gold);
}

/* 结算结果样式 */
.result-scroll-area {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  min-height: 0;
  padding-right: 0.5rem;
  margin-right: -0.5rem;
}

/* 自定义滚动条 */
.result-scroll-area::-webkit-scrollbar {
  width: 6px;
}
.result-scroll-area::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 4px;
}
.result-scroll-area::-webkit-scrollbar-thumb {
  background: rgba(251, 191, 36, 0.3);
  border-radius: 4px;
}
.result-scroll-area::-webkit-scrollbar-thumb:hover {
  background: rgba(251, 191, 36, 0.5);
}

.players-result {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-bottom: 1.5rem;
}

/* 多人紧凑模式 */
.players-result.is-compact {
  gap: 0.4rem;
}
.players-result.is-compact .player-result-item {
  padding: 0.5rem 0.75rem;
}
.players-result.is-compact .pr-name {
  font-size: 0.9rem;
}
.players-result.is-compact .pr-current {
  font-size: 0.9rem;
}
.players-result.is-compact .pr-net {
  font-size: 1.1rem;
}

/* 少人宽松模式 */
.players-result.is-sparse {
  gap: 1rem;
}
.players-result.is-sparse .player-result-item {
  padding: 1rem 1.25rem;
}
.players-result.is-sparse .pr-name {
  font-size: 1.1rem;
}

.player-result-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 10px;
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.pr-info {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.pr-name {
  font-weight: 600;
  font-size: 1rem;
}

.pr-cards {
  font-size: 0.8rem;
  color: var(--text-muted);
  letter-spacing: 2px;
}

.pr-change {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.1rem;
}

.pr-initial {
  font-size: 0.7rem;
  color: var(--text-muted);
}

.pr-current {
  font-size: 1rem;
  font-weight: 600;
}

.pr-net {
  font-size: 1.25rem;
  font-weight: 700;
}

.pr-change.is-win .pr-net {
  color: var(--success, #10b981);
}

.pr-change.is-win .pr-current {
  color: var(--gold);
}

.pr-change.is-loss .pr-net {
  color: var(--danger, #ef4444);
}

.pr-change.is-loss .pr-current {
  color: var(--text-muted);
}

/* 摊牌演出动画 */
.showdown-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.95);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
  flex-direction: column;
}

.showdown-content {
  text-align: center;
  width: 100%;
  max-width: 900px;
  padding: 1rem;
  position: relative;
}

.showdown-title {
  font-size: 2rem;
  font-weight: 900;
  color: #fbbf24;
  text-shadow: 0 0 20px rgba(251, 191, 36, 0.8), 0 0 40px rgba(251, 191, 36, 0.4);
  letter-spacing: 0.1em;
  margin-bottom: 2rem;
  opacity: 0;
  transform: scale(0.8);
  transition: all 0.6s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.showdown-title.animate {
  opacity: 1;
  transform: scale(1);
}

.title-text {
  background: linear-gradient(135deg, #fbbf24 0%, #f59e0b 50%, #fbbf24 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  animation: title-shimmer 2s ease-in-out infinite;
}

@keyframes title-shimmer {
  0%, 100% { filter: brightness(1); }
  50% { filter: brightness(1.3); }
}

/* 公共牌展示 */
.community-cards-showcase {
  margin: 0.5rem 0;
  padding: 0.75rem;
  background: rgba(0, 0, 0, 0.3);
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.community-label {
  font-size: 0.875rem;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 2px;
  margin-bottom: 1rem;
}

.community-cards-row {
  display: flex;
  justify-content: center;
  gap: 0.75rem;
}

.showcase-card {
  width: 56px;
  height: 78px;
  border-radius: 8px;
  background: var(--card-bg);
  border: none;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
}

.showcase-card .card-inner {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--card-black);
}

.showcase-card.is-red .card-inner {
  color: var(--card-red);
}

/* 当前亮牌用户 */
.current-reveal {
  margin: 0.5rem 0;
  padding: 1rem;
  background: radial-gradient(ellipse at center, rgba(251, 191, 36, 0.15) 0%, transparent 70%);
  border-radius: 16px;
  min-height: 160px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1.5rem;
}

.reveal-player-name {
  font-size: 1.25rem;
  font-weight: 700;
  color: #f8fafc;
  text-shadow: 0 2px 10px rgba(0,0,0,0.5);
}

.reveal-cards {
  display: flex;
  gap: 0.5rem;
}

.reveal-card {
  width: 72px;
  height: 100px;
  border-radius: 10px;
  position: relative;
  transform-style: preserve-3d;
  transition: all 0.8s cubic-bezier(0.34, 1.56, 0.64, 1); /* 减慢翻转动画 */
}

.reveal-card::before {
  content: '🂠';
  position: absolute;
  font-size: 3rem;
  color: rgba(255, 255, 255, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  background: repeating-linear-gradient(
    45deg,
    #1e293b,
    #1e293b 10px,
    #0f172a 10px,
    #0f172a 20px
  );
  border-radius: 12px;
  border: 2px solid rgba(255, 255, 255, 0.2);
  backface-visibility: hidden;
}

.reveal-card.is-flipped {
  transform: rotateY(180deg);
  transition: transform 0.8s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.reveal-card.is-flipped::before {
  content: '';
}

.reveal-card .card-inner {
  position: absolute;
  width: 100%;
  height: 100%;
  background: var(--card-bg);
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 2rem;
  font-weight: 700;
  color: var(--card-black);
  transform: rotateY(180deg);
  backface-visibility: hidden;
  box-shadow: 0 8px 20px rgba(0,0,0,0.4);
}

.reveal-card.is-red .card-inner {
  color: var(--card-red);
}

.reveal-rank {
  opacity: 0;
  transform: translateY(20px) scale(0.8);
  transition: all 0.8s cubic-bezier(0.34, 1.56, 0.64, 1); /* 减慢牌型显示动画 */
}

.reveal-rank.animate-in {
  opacity: 1;
  transform: translateY(0) scale(1);
}

.rank-label {
  font-size: 1rem;
  font-weight: 700;
  color: #fbbf24;
  background: rgba(251, 191, 36, 0.2);
  padding: 0.3rem 1rem;
  border-radius: 8px;
  border: 1px solid rgba(251, 191, 36, 0.4);
  box-shadow: 0 0 15px rgba(251, 191, 36, 0.3);
}

/* 排名展示 */
.rankings-container {
  margin-top: 1rem;
  margin-bottom: 1rem;
  opacity: 0;
  transform: translateY(30px);
  transition: all 0.6s ease-out;
  max-height: 250px;
  overflow-y: auto;
}

.rankings-container.show {
  opacity: 1;
  transform: translateY(0);
}

.rankings-title {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text-muted);
  margin-bottom: 1.5rem;
  text-transform: uppercase;
  letter-spacing: 2px;
}

.rankings-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.ranking-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  opacity: 0;
  transform: translateX(-20px);
  animation: ranking-slide-in 0.4s ease-out forwards;
}

.ranking-item.is-winner {
  background: linear-gradient(135deg, rgba(251, 191, 36, 0.2) 0%, rgba(251, 191, 36, 0.05) 100%);
  border-color: rgba(251, 191, 36, 0.5);
  box-shadow: 0 0 20px rgba(251, 191, 36, 0.2);
}

.ranking-item.is-current {
  border-color: rgba(16, 185, 129, 0.6);
  box-shadow: 0 0 15px rgba(16, 185, 129, 0.2);
}

@keyframes ranking-slide-in {
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

.ranking-position {
  width: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.crown {
  font-size: 1.25rem;
  animation: crown-bounce 1s ease-in-out infinite;
}

@keyframes crown-bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-5px); }
}

.position-num {
  font-size: 0.875rem;
  font-weight: 700;
  color: var(--text-muted);
}

.ranking-item.is-winner .position-num {
  color: #fbbf24;
}

.ranking-info {
  flex: 1;
  text-align: left;
}

.ranking-name {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-main);
}

.ranking-cards {
  font-size: 0.875rem;
  color: var(--text-muted);
  letter-spacing: 2px;
  margin-top: 0.25rem;
}

.ranking-hand {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.25rem;
}

.hand-name {
  font-size: 0.75rem;
  font-weight: 700;
  color: #fbbf24;
  background: rgba(251, 191, 36, 0.2);
  padding: 0.35rem 0.85rem;
  border-radius: 8px;
  border: 1px solid rgba(251, 191, 36, 0.4);
}

.hand-cards {
  font-size: 0.7rem;
  color: var(--text-muted);
  letter-spacing: 1px;
}

/* 继续按钮 */
.showdown-actions {
  margin-top: 2rem;
  animation: fade-in 0.5s ease-out;
}

@keyframes fade-in {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

/* 摊牌过渡动画 */
.showdown-enter-active {
  animation: showdown-in 0.6s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.showdown-leave-active {
  animation: showdown-out 0.4s ease-in;
}

@keyframes showdown-in {
  from {
    opacity: 0;
    transform: scale(1.1);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

@keyframes showdown-out {
  from {
    opacity: 1;
    transform: scale(1);
  }
  to {
    opacity: 0;
    transform: scale(0.95);
  }
}

/* 结算弹窗动画 */
.result-modal-enter-active {
  animation: result-pop-in 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.result-modal-leave-active {
  animation: result-pop-out 0.3s ease-in;
}

@keyframes result-pop-in {
  0% {
    opacity: 0;
    transform: scale(0.8) translateY(30px);
  }
  100% {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

@keyframes result-pop-out {
  0% {
    opacity: 1;
    transform: scale(1);
  }
  100% {
    opacity: 0;
    transform: scale(0.9);
  }
}

.result-title {
  animation: title-glow 2s ease-in-out infinite;
}

@keyframes title-glow {
  0%, 100% {
    text-shadow: 0 2px 10px rgba(251, 191, 36, 0.4);
  }
  50% {
    text-shadow: 0 2px 30px rgba(251, 191, 36, 0.8), 0 0 50px rgba(251, 191, 36, 0.4);
  }
}

.total-pot-display {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-bottom: 1.5rem;
  padding: 1.5rem;
  background: rgba(251, 191, 36, 0.1);
  border-radius: 16px;
  border: 1px solid rgba(251, 191, 36, 0.3);
  flex-shrink: 0;
}

.total-pot-label {
  font-size: 0.875rem;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 2px;
  margin-bottom: 0.5rem;
}

.total-pot-amount {
  font-size: 2.5rem;
  font-weight: 700;
  color: var(--gold);
  text-shadow: 0 2px 10px rgba(251, 191, 36, 0.5);
}

.player-result-item {
  animation: result-item-slide 0.5s ease-out backwards;
  transition: all 0.3s ease;
}

.player-result-item.is-winner {
  background: rgba(251, 191, 36, 0.15);
  border-color: rgba(251, 191, 36, 0.4);
  box-shadow: 0 0 20px rgba(251, 191, 36, 0.2);
}

@keyframes result-item-slide {
  0% {
    opacity: 0;
    transform: translateX(-30px);
  }
  100% {
    opacity: 1;
    transform: translateX(0);
  }
}

.winners-section {
  margin-bottom: 1.5rem;
}

.winners-title {
  color: var(--gold);
  font-size: 1.25rem;
  margin-bottom: 1rem;
  text-align: center;
}

.winners-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.winner-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  background: linear-gradient(135deg, rgba(251, 191, 36, 0.2) 0%, rgba(251, 191, 36, 0.05) 100%);
  border-radius: 12px;
  border: 1px solid rgba(251, 191, 36, 0.4);
  animation: winner-glow 1.5s ease-in-out infinite;
}

@keyframes winner-glow {
  0%, 100% {
    box-shadow: 0 0 15px rgba(251, 191, 36, 0.3);
  }
  50% {
    box-shadow: 0 0 30px rgba(251, 191, 36, 0.6), 0 0 50px rgba(251, 191, 36, 0.3);
  }
}

.w-amount {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--gold);
  animation: amount-pop 0.5s ease-out;
}

@keyframes amount-pop {
  0% {
    transform: scale(0);
    opacity: 0;
  }
  50% {
    transform: scale(1.3);
  }
  100% {
    transform: scale(1);
    opacity: 1;
  }
}

.pr-net.is-animating {
  animation: net-change-pulse 0.3s ease-out;
}

@keyframes net-change-pulse {
  0% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.2);
  }
  100% {
    transform: scale(1);
  }
}

.winners-section {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.winners-section h3 {
  font-size: 0.9rem;
  color: var(--gold);
  margin-bottom: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.winners-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.winner-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 0.75rem;
  background: rgba(251, 191, 36, 0.08);
  border-radius: 6px;
}
/* 牌型帮助样式 */
.rank-help-modal {
  position: relative;
  max-width: 480px;
  max-height: 80vh;
  padding: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.close-btn {
  position: absolute;
  top: 1rem;
  right: 1rem;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.1);
  border: none;
  color: var(--text-muted);
  font-size: 1.25rem;
  cursor: pointer;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.2);
  color: var(--text-main);
}

/* All-in 样式 */
.all-in-amount {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--danger);
  text-align: center;
  margin: 1rem 0;
}

.error-modal {
  border-color: rgba(239, 68, 68, 0.3);
}

.text-danger {
  color: var(--danger);
}

/* All-in 通知横幅 */
.allin-notification {
  position: fixed;
  top: 80px;
  left: 50%;
  transform: translateX(-50%);
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.95), rgba(185, 28, 28, 0.95));
  color: white;
  padding: 0.75rem 2rem;
  border-radius: 50px;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  box-shadow: 0 4px 20px rgba(239, 68, 68, 0.4), 0 0 30px rgba(239, 68, 68, 0.2);
  z-index: 50;
  font-weight: 600;
  animation: allin-pulse 1.5s ease-in-out infinite;
}

.allin-icon {
  font-size: 1.5rem;
  animation: flash 0.5s ease-in-out infinite alternate;
}

.allin-text {
  font-size: 1.125rem;
}

.allin-amount {
  background: rgba(0, 0, 0, 0.3);
  padding: 0.25rem 0.75rem;
  border-radius: 20px;
  font-weight: 700;
}

@keyframes allin-pulse {
  0%, 100% { transform: translateX(-50%) scale(1); }
  50% { transform: translateX(-50%) scale(1.02); }
}

@keyframes flash {
  from { opacity: 1; }
  to { opacity: 0.5; }
}

.slide-down-enter-active,
.slide-down-leave-active {
  transition: all 0.3s ease;
}

.slide-down-enter-from,
.slide-down-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(-20px);
}

/* 出牌顺序 */
.turn-order {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  background: rgba(0, 0, 0, 0.3);
  padding: 0.35rem 0.75rem;
  border-radius: 20px;
  font-size: 0.8rem;
  max-width: 300px;
  overflow-x: auto;
}

.turn-order-item {
  display: flex;
  align-items: center;
  gap: 0.2rem;
  padding: 0.2rem 0.5rem;
  border-radius: 12px;
  white-space: nowrap;
  transition: all 0.2s ease;
}

.turn-order-item.is-current {
  background: rgba(251, 191, 36, 0.3);
  border: 1px solid rgba(251, 191, 36, 0.6);
  box-shadow: 0 0 10px rgba(251, 191, 36, 0.3);
}

.turn-order-item.is-current .turn-name {
  color: var(--gold);
  font-weight: 600;
}

.turn-num {
  color: var(--text-muted);
  font-size: 0.7rem;
  min-width: 12px;
}

.turn-name {
  color: var(--text);
}
</style>
