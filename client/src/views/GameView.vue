<template>
  <div class="game-container">
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
        <button class="btn btn-secondary btn-sm" style="margin-right: 1rem;" @click="showHandRankHelp = true">牌型大小</button>
        <button class="btn btn-secondary btn-sm" @click="handleLeave">退出房间</button>
      </div>
    </header>

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
            <span class="pot-amount">{{ gameStore.pot }}</span>
          </div>
          
          <div class="community-cards">
            <div 
              v-for="i in 5" 
              :key="'cc-'+i" 
              class="playing-card" 
              :class="{ 
                'revealed': gameStore.communityCards[i-1], 
                'is-red': isRedCard(gameStore.communityCards[i-1]),
                'is-highlight': isBestHandCard(gameStore.communityCards[i-1])
              }"
            >
              <div class="card-inner" v-if="gameStore.communityCards[i-1]">
                {{ gameStore.communityCards[i-1] }}
              </div>
            </div>
          </div>
        </div>

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
          <div class="playing-card large revealed" :class="{
            'is-red': isRedCard(myGamePlayer.cards[0]),
            'is-highlight': isBestHandCard(myGamePlayer.cards[0])
          }">
             <div class="card-inner">{{ myGamePlayer.cards[0] || '?' }}</div>
          </div>
          <div class="playing-card large revealed" :class="{
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
                <span v-else>轮到你了，需跟注 <strong class="text-gold">{{ toCall }}</strong></span>
              </h3>
              <div class="action-buttons">
                <button class="btn btn-danger-outline btn-lg" @click="doAction('fold')">弃牌</button>
                <button v-if="toCall === 0" class="btn btn-primary btn-lg" @click="doAction('check')">过牌 (Check)</button>
                <button v-else class="btn btn-primary btn-lg" @click="doAction('call')">跟注 {{ toCall }}</button>
                <button class="btn btn-warning btn-lg" @click="openRaiseModal">加注</button>
                <button class="btn btn-danger btn-lg" @click="showAllInModal = true">All-in</button>
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

    <!-- 结算弹窗 -->
    <Transition name="fade">
      <div v-if="gameStore.result" class="modal-overlay result-overlay">
        <div class="result-modal glass-panel">
          <div class="result-header">
            <h2>🎉 本局结束 🎉</h2>
          </div>

          <div class="players-result">
            <div v-for="player in gameStore.players" :key="player.id" class="player-result-item">
              <div class="pr-info">
                <span class="pr-name">{{ player.name }}</span>
                <span class="pr-cards" v-if="player.cards.length">{{ player.cards.join(' ') }}</span>
              </div>
              <div class="pr-change" :class="getNetChange(player.id) >= 0 ? 'is-win' : 'is-loss'">
                <span class="pr-initial">({{ player.initialChips }})</span>
                <span class="pr-current">{{ player.chips }}</span>
                <span class="pr-net">{{ getNetChange(player.id) >= 0 ? '+' : '' }}{{ getNetChange(player.id) }}</span>
              </div>
            </div>
          </div>

          <div class="winners-section" v-if="gameStore.result.winners?.length">
            <h3>赢家</h3>
            <div class="winners-list">
              <div v-for="w in gameStore.result.winners" :key="w.player_id" class="winner-item">
                <span class="w-name">{{ w.player_name }}</span>
                <span class="w-rank">{{ translateHandRank(w.hand_rank) }}</span>
              </div>
            </div>
          </div>

          <button class="btn btn-primary btn-lg" style="width: 100%; margin-top: 1.5rem;" @click="closeResultAndReset">
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

  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useRoomStore } from '../stores/room'
import { useGameStore } from '../stores/game'
import { useWebSocket } from '../composables/useWebSocket'
import HandRankShowcase from '../components/HandRankShowcase.vue'

const router = useRouter()
const roomStore = useRoomStore()
const gameStore = useGameStore()
const { leaveRoom, playerAction, error: wsError } = useWebSocket()

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

const toCall = computed(() => {
  if (!myGamePlayer.value) return 0
  const required = gameStore.currentBet - myGamePlayer.value.bet
  return Math.max(0, required)
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
  await playerAction(action, amount)
  showRaiseModal.value = false
}

function openRaiseModal() {
  raiseAmount.value = gameStore.minRaise
  showRaiseModal.value = true
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
  padding: 2.5rem;
  border: 1px solid rgba(251, 191, 36, 0.3);
  box-shadow: 0 0 40px rgba(251, 191, 36, 0.15);
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
.players-result {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-bottom: 1.5rem;
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
</style>
