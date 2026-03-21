<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="isOpen" class="modal-overlay" @click.self="close">
        <div class="modal-container">
          <button class="close-btn" @click="close">✕</button>
          <div class="modal-content">
            <div class="tutorial-inner">
              <!-- 游戏简介 -->
              <section class="tutorial-section">
                <h2 class="section-title">
                  <span class="section-icon">🎴</span>
                  游戏简介
                </h2>
                <p class="section-text">
                  德州扑克是一种流行的扑克牌游戏，每位玩家获得两张私有手牌，然后通过五张公共牌的组合来决定胜负。游戏目标是用手中的两张手牌与五张公共牌组成最大的牌型。
                </p>
              </section>

              <!-- 游戏流程 -->
              <section class="tutorial-section">
                <h2 class="section-title">
                  <span class="section-icon">🔄</span>
                  游戏流程
                </h2>
                <div class="flow-steps">
                  <div class="flow-step">
                    <div class="step-num">1</div>
                    <div class="step-content">
                      <div class="step-title">发牌阶段 <span class="en">Pre-Flop</span></div>
                      <div class="step-desc">每位玩家获得2张手牌，从庄家左边开始下注</div>
                    </div>
                  </div>
                  <div class="flow-step">
                    <div class="step-num">2</div>
                    <div class="step-content">
                      <div class="step-title">翻牌圈 <span class="en">Flop</span></div>
                      <div class="step-desc">发出3张公共牌，再次进行下注</div>
                    </div>
                  </div>
                  <div class="flow-step">
                    <div class="step-num">3</div>
                    <div class="step-content">
                      <div class="step-title">转牌圈 <span class="en">Turn</span></div>
                      <div class="step-desc">发出第4张公共牌，继续下注</div>
                    </div>
                  </div>
                  <div class="flow-step">
                    <div class="step-num">4</div>
                    <div class="step-content">
                      <div class="step-title">河牌圈 <span class="en">River</span></div>
                      <div class="step-desc">发出第5张公共牌，最后一轮下注</div>
                    </div>
                  </div>
                  <div class="flow-step">
                    <div class="step-num">5</div>
                    <div class="step-content">
                      <div class="step-title">摊牌 <span class="en">Showdown</span></div>
                      <div class="step-desc">所有玩家亮牌，最大牌型获胜</div>
                    </div>
                  </div>
                </div>
              </section>

              <!-- 名词解释 -->
              <section class="tutorial-section">
                <h2 class="section-title">
                  <span class="section-icon">📚</span>
                  名词解释
                </h2>
                <div class="terms-grid">
                  <div class="term-item">
                    <div class="term-name">小盲 <span class="en">Small Blind</span></div>
                    <div class="term-desc">庄家左边第一位，强制下注小盲注（50）</div>
                  </div>
                  <div class="term-item">
                    <div class="term-name">大盲 <span class="en">Big Blind</span></div>
                    <div class="term-desc">小盲左边第一位，强制下注大盲注（100）</div>
                  </div>
                  <div class="term-item">
                    <div class="term-name">跟注 <span class="en">Call</span></div>
                    <div class="term-desc">匹配当前下注额，继续游戏</div>
                  </div>
                  <div class="term-item">
                    <div class="term-name">加注 <span class="en">Raise</span></div>
                    <div class="term-desc">高于当前下注额，增加筹码</div>
                  </div>
                  <div class="term-item">
                    <div class="term-name">过牌 <span class="en">Check</span></div>
                    <div class="term-desc">不下注，把行动权交给下一位</div>
                  </div>
                  <div class="term-item">
                    <div class="term-name">弃牌 <span class="en">Fold</span></div>
                    <div class="term-desc">放弃本局，失去所有已下注筹码</div>
                  </div>
                  <div class="term-item highlight">
                    <div class="term-name">All-in <span class="en">全下</span></div>
                    <div class="term-desc">投入全部筹码，可博弈超额奖金池</div>
                  </div>
                  <div class="term-item">
                    <div class="term-name">底池 <span class="en">Pot</span></div>
                    <div class="term-desc">本局所有玩家下注的筹码总和</div>
                  </div>
                </div>
              </section>

              <!-- 牌型大小 -->
              <section class="tutorial-section">
                <h2 class="section-title">
                  <span class="section-icon">🏆</span>
                  牌型大小（从大到小）
                </h2>
                <div class="hands-list">
                  <div class="hand-item">
                    <div class="hand-rank">1</div>
                    <div class="hand-info">
                      <div class="hand-name">皇家同花顺 <span class="en">Royal Flush</span></div>
                      <div class="hand-cards">A K Q J 10 同花色</div>
                    </div>
                    <div class="hand-visual royal">♠️❤️</div>
                  </div>
                  <div class="hand-item">
                    <div class="hand-rank">2</div>
                    <div class="hand-info">
                      <div class="hand-name">同花顺 <span class="en">Straight Flush</span></div>
                      <div class="hand-cards">5张同花色连续牌</div>
                    </div>
                    <div class="hand-visual straight-flush">♠️♠️♠️</div>
                  </div>
                  <div class="hand-item">
                    <div class="hand-rank">3</div>
                    <div class="hand-info">
                      <div class="hand-name">四条 <span class="en">Four of a Kind</span></div>
                      <div class="hand-cards">4张相同点数的牌</div>
                    </div>
                    <div class="hand-visual four">🃏🃏🃏🃏</div>
                  </div>
                  <div class="hand-item">
                    <div class="hand-rank">4</div>
                    <div class="hand-info">
                      <div class="hand-name">葫芦 <span class="en">Full House</span></div>
                      <div class="hand-cards">三条 + 一对</div>
                    </div>
                    <div class="hand-visual full">🃏🃏🃏 ♠️♠️</div>
                  </div>
                  <div class="hand-item">
                    <div class="hand-rank">5</div>
                    <div class="hand-info">
                      <div class="hand-name">同花 <span class="en">Flush</span></div>
                      <div class="hand-cards">5张同花色（非顺子）</div>
                    </div>
                    <div class="hand-visual flush">❤️❤️❤️❤️❤️</div>
                  </div>
                  <div class="hand-item">
                    <div class="hand-rank">6</div>
                    <div class="hand-info">
                      <div class="hand-name">顺子 <span class="en">Straight</span></div>
                      <div class="hand-cards">5张连续花色（不同花）</div>
                    </div>
                    <div class="hand-visual straight">♠️♥️♠️♦️♣️</div>
                  </div>
                  <div class="hand-item">
                    <div class="hand-rank">7</div>
                    <div class="hand-info">
                      <div class="hand-name">三条 <span class="en">Three of a Kind</span></div>
                      <div class="hand-cards">3张相同点数的牌</div>
                    </div>
                    <div class="hand-visual three">🃏🃏🃏</div>
                  </div>
                  <div class="hand-item">
                    <div class="hand-rank">8</div>
                    <div class="hand-info">
                      <div class="hand-name">两对 <span class="en">Two Pair</span></div>
                      <div class="hand-cards">2组对子</div>
                    </div>
                    <div class="hand-visual two-pair">♠️♠️♦️♦️</div>
                  </div>
                  <div class="hand-item">
                    <div class="hand-rank">9</div>
                    <div class="hand-info">
                      <div class="hand-name">一对 <span class="en">One Pair</span></div>
                      <div class="hand-cards">2张相同点数的牌</div>
                    </div>
                    <div class="hand-visual pair">♠️♠️</div>
                  </div>
                  <div class="hand-item">
                    <div class="hand-rank">10</div>
                    <div class="hand-info">
                      <div class="hand-name">高牌 <span class="en">High Card</span></div>
                      <div class="hand-cards">无以上牌型时，点数最大者胜</div>
                    </div>
                    <div class="hand-visual high">A</div>
                  </div>
                </div>
              </section>

              <!-- 游戏规则 -->
              <section class="tutorial-section">
                <h2 class="section-title">
                  <span class="section-icon">⚠️</span>
                  游戏规则
                </h2>
                <ul class="rules-list">
                  <li>每局开始前，所有玩家需要点击<strong>「准备」</strong>按钮</li>
                  <li>只有<strong>房主</strong>可以开始游戏，必须所有玩家都准备</li>
                  <li>每局初始筹码为 <strong>10000</strong>，大小盲分别为 <strong>50/100</strong></li>
                  <li>玩家回合需要在倒计时内做出决策，超时视为弃牌</li>
                  <li>当玩家筹码归零，将被移出游戏</li>
                  <li>当只剩一名玩家时，该玩家获胜</li>
                </ul>
              </section>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const isOpen = ref(false)

function open() {
  isOpen.value = true
}

function close() {
  isOpen.value = false
}

defineExpose({ open, close })
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal-container {
  background: linear-gradient(145deg, #1a2a3a 0%, #0f1a24 100%);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 20px;
  max-width: 700px;
  width: 100%;
  max-height: 85vh;
  display: flex;
  flex-direction: column;
  position: relative;
  box-shadow: 0 25px 80px rgba(0, 0, 0, 0.5);
}

.close-btn {
  position: absolute;
  top: 1rem;
  right: 1rem;
  width: 36px;
  height: 36px;
  border: none;
  background: rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.7);
  border-radius: 50%;
  cursor: pointer;
  font-size: 1.2rem;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
  z-index: 10;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.2);
  color: white;
}

.modal-content {
  padding: 2rem;
  overflow-y: auto;
  flex: 1;
}

.tutorial-inner {
  padding-right: 0.5rem;
}

.tutorial-section {
  margin-bottom: 2rem;
}

.tutorial-section:last-child {
  margin-bottom: 0;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 1.25rem;
  font-weight: 700;
  color: #fbbf24;
  margin-bottom: 1rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.section-icon {
  font-size: 1.5rem;
}

.section-text {
  color: #e5e7eb;
  line-height: 1.7;
  font-size: 0.95rem;
}

/* Flow Steps */
.flow-steps {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.flow-step {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  padding: 0.75rem;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 10px;
  transition: all 0.2s ease;
}

.flow-step:hover {
  background: rgba(255, 255, 255, 0.08);
  transform: translateX(4px);
}

.step-num {
  width: 28px;
  height: 28px;
  background: linear-gradient(135deg, #10b981 0%, #047857 100%);
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 0.875rem;
  flex-shrink: 0;
}

.step-content {
  flex: 1;
}

.step-title {
  font-weight: 600;
  margin-bottom: 0.25rem;
  color: #e5e7eb;
}

.step-title .en {
  color: #9ca3af;
  font-size: 0.85rem;
  font-weight: 400;
  margin-left: 0.5rem;
}

.step-desc {
  font-size: 0.85rem;
  color: #9ca3af;
  line-height: 1.4;
}

/* Terms Grid */
.terms-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 0.75rem;
}

.term-item {
  padding: 0.75rem;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 10px;
  transition: all 0.2s ease;
}

.term-item:hover {
  background: rgba(255, 255, 255, 0.08);
}

.term-item.highlight {
  background: rgba(239, 68, 68, 0.15);
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.term-name {
  font-weight: 600;
  font-size: 0.9rem;
  color: #e5e7eb;
  margin-bottom: 0.25rem;
}

.term-name .en {
  color: #9ca3af;
  font-size: 0.8rem;
  font-weight: 400;
  margin-left: 0.5rem;
}

.term-desc {
  font-size: 0.8rem;
  color: #9ca3af;
  line-height: 1.4;
}

/* Hands List */
.hands-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.hand-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.6rem 0.75rem;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 10px;
  transition: all 0.2s ease;
}

.hand-item:hover {
  background: rgba(255, 255, 255, 0.08);
  transform: translateX(4px);
}

.hand-rank {
  width: 24px;
  height: 24px;
  background: rgba(251, 191, 36, 0.2);
  color: #fbbf24;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 0.8rem;
  flex-shrink: 0;
}

.hand-info {
  flex: 1;
}

.hand-name {
  font-weight: 600;
  font-size: 0.9rem;
  color: #e5e7eb;
}

.hand-name .en {
  color: #9ca3af;
  font-size: 0.8rem;
  font-weight: 400;
  margin-left: 0.5rem;
}

.hand-cards {
  font-size: 0.75rem;
  color: #9ca3af;
  margin-top: 0.15rem;
}

.hand-visual {
  font-size: 1.25rem;
  letter-spacing: 2px;
}

/* Rules List */
.rules-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

.rules-list li {
  position: relative;
  padding-left: 1.5rem;
  color: #e5e7eb;
  font-size: 0.9rem;
  line-height: 1.5;
}

.rules-list li::before {
  content: "•";
  position: absolute;
  left: 0;
  color: #10b981;
  font-weight: 700;
}

.rules-list strong {
  color: #fbbf24;
}

/* Transitions */
.modal-enter-active,
.modal-leave-active {
  transition: all 0.3s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .modal-container,
.modal-leave-to .modal-container {
  transform: scale(0.95) translateY(20px);
}

.modal-enter-active .modal-container,
.modal-leave-active .modal-container {
  transition: transform 0.3s ease;
}
</style>
