<template>
  <div class="hand-rank-showcase">
    <div class="showcase-header">
      <h2>🎴 牌型大小</h2>
    </div>

    <div class="rank-grid">
      <div v-for="rank in handRanks" :key="rank.name" class="rank-row">
        <div class="cards-mini">
          <span
            v-for="card in rank.exampleCards"
            :key="card"
            class="mini-card"
            :class="getCardClass(card)"
          >{{ card }}</span>
        </div>
        <div class="rank-info">
          <span class="rank-name">{{ rank.name }}</span>
          <span class="rank-desc">{{ rank.description }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface HandRank {
  name: string
  score: number
  exampleCards: string[]
  description: string
}

const handRanks: HandRank[] = [
  { name: '皇家同花顺', score: 10, exampleCards: ['A♠', 'K♠', 'Q♠', 'J♠', '10♠'], description: '最大' },
  { name: '同花顺', score: 9, exampleCards: ['9♥', '8♥', '7♥', '6♥', '5♥'], description: '同花+顺' },
  { name: '四条', score: 8, exampleCards: ['A♠', 'A♥', 'A♦', 'A♣', 'K♠'], description: '四同点' },
  { name: '葫芦', score: 7, exampleCards: ['A♠', 'A♥', 'A♦', 'K♣', 'K♠'], description: '三条+对' },
  { name: '同花', score: 6, exampleCards: ['A♠', 'J♠', '8♠', '6♠', '3♠'], description: '同花色' },
  { name: '顺子', score: 5, exampleCards: ['9♠', '8♥', '7♦', '6♣', '5♠'], description: '五连张' },
  { name: '三条', score: 4, exampleCards: ['A♠', 'A♥', 'A♦', 'K♣', 'Q♠'], description: '三同点' },
  { name: '两对', score: 3, exampleCards: ['A♠', 'A♥', 'K♣', 'K♠', 'Q♠'], description: '两个对' },
  { name: '一对', score: 2, exampleCards: ['A♠', 'A♥', 'K♣', 'Q♠', 'J♠'], description: '一个对' },
  { name: '高牌', score: 1, exampleCards: ['A♠', 'J♥', '8♦', '6♣', '3♠'], description: '比单牌' },
]

function getCardClass(card: string): string {
  return card.includes('♥') || card.includes('♦') ? 'is-red' : 'is-black'
}
</script>

<style scoped>
.hand-rank-showcase {
  padding: 1rem;
  overflow-y: auto;
}

.showcase-header {
  text-align: center;
  margin-bottom: 0.75rem;
}

.showcase-header h2 {
  color: var(--gold);
  font-size: 1.1rem;
  margin: 0;
}

.rank-grid {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.rank-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.4rem 0.5rem;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.03);
  transition: background 0.15s;
}

.rank-row:hover {
  background: rgba(255, 255, 255, 0.08);
}

.cards-mini {
  display: flex;
  gap: 3px;
  flex-shrink: 0;
}

.mini-card {
  width: 32px;
  height: 38px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.7rem;
  font-weight: 700;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.3);
}

.mini-card.is-red {
  background: linear-gradient(145deg, #fff 0%, #ffe5e5 100%);
  color: #dc2626;
}

.mini-card.is-black {
  background: linear-gradient(145deg, #fff 0%, #f0f0f0 100%);
  color: #1e293b;
}

.rank-info {
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
  min-width: 0;
}

.rank-name {
  font-weight: 600;
  font-size: 0.8rem;
  color: var(--gold);
}

.rank-desc {
  font-size: 0.65rem;
  color: var(--text-muted);
}
</style>
