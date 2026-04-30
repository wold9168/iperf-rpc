<template>
  <div class="result-card">
    <h3>测速结果</h3>

    <div v-if="!result" class="placeholder">
      <p>尚未执行测速，请在左侧构建指令并提交。</p>
    </div>

    <div v-else class="result-content">
      <div :class="['status-badge', result.status]">
        {{ statusText(result.status) }}
      </div>

      <div class="detail-row">
        <span class="key">指令</span>
        <code class="command">{{ result.command }}</code>
      </div>

      <div v-if="result.summary" class="summary-grid">
        <div class="summary-item">
          <span class="summary-value">{{ formatBitrate(result.summary.bitrate_bps) }}</span>
          <span class="summary-label">吞吐量</span>
        </div>
        <div class="summary-item">
          <span class="summary-value">{{ result.summary.jitter_ms?.toFixed(2) ?? '-' }} ms</span>
          <span class="summary-label">抖动</span>
        </div>
        <div class="summary-item">
          <span class="summary-value">{{ result.summary.lost_percent ?? '-' }}%</span>
          <span class="summary-label">丢包率</span>
        </div>
        <div class="summary-item">
          <span class="summary-value">{{ result.summary.duration_sec?.toFixed(1) ?? '-' }}s</span>
          <span class="summary-label">耗时</span>
        </div>
      </div>

      <details class="raw-output">
        <summary>原始输出</summary>
        <pre>{{ result.output || '(无输出)' }}</pre>
      </details>

      <div class="timestamps">
        <span>开始: {{ formatTime(result.started_at) }}</span>
        <span v-if="result.finished_at">结束: {{ formatTime(result.finished_at) }}</span>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'ResultView',
  props: {
    result: { type: Object, default: null },
  },
  methods: {
    statusText(s) {
      return { running: '运行中', completed: '已完成', error: '执行失败' }[s] || s
    },
    formatBitrate(bps) {
      if (!bps || bps === 0) return '-'
      if (bps >= 1e9) return (bps / 1e9).toFixed(2) + ' Gbps'
      if (bps >= 1e6) return (bps / 1e6).toFixed(2) + ' Mbps'
      if (bps >= 1e3) return (bps / 1e3).toFixed(2) + ' Kbps'
      return bps.toFixed(0) + ' bps'
    },
    formatTime(t) {
      if (!t) return '-'
      return new Date(t).toLocaleTimeString()
    },
  },
}
</script>

<style scoped>
.result-card {
  background: #fff;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.08);
  min-height: 300px;
}

.result-card h3 {
  font-size: 16px;
  margin-bottom: 16px;
  color: #1a1a2e;
}

.placeholder {
  text-align: center;
  padding: 60px 20px;
  color: #999;
  font-size: 14px;
}

.status-badge {
  display: inline-block;
  padding: 4px 14px;
  border-radius: 10px;
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 12px;
}

.status-badge.completed { background: #d4edda; color: #155724; }
.status-badge.running { background: #fff3cd; color: #856404; }
.status-badge.error { background: #f8d7da; color: #721c24; }

.detail-row { margin-bottom: 12px; }

.key { font-size: 13px; color: #888; display: block; margin-bottom: 4px; }

.command {
  display: block;
  background: #f7f7f9;
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 13px;
  word-break: break-all;
  border: 1px solid #e9ecef;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 10px;
  margin: 16px 0;
}

.summary-item {
  text-align: center;
  background: #f8f9fa;
  padding: 12px 8px;
  border-radius: 8px;
}

.summary-value {
  display: block;
  font-size: 20px;
  font-weight: 700;
  color: #1a1a2e;
}

.summary-label {
  display: block;
  font-size: 12px;
  color: #888;
  margin-top: 4px;
}

.raw-output { margin-top: 12px; }

.raw-output summary {
  cursor: pointer;
  font-size: 13px;
  color: #4361ee;
  margin-bottom: 8px;
}

.raw-output pre {
  background: #1a1a2e;
  color: #a8b2d1;
  padding: 12px;
  border-radius: 6px;
  font-size: 12px;
  max-height: 300px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
}

.timestamps {
  margin-top: 12px;
  display: flex;
  gap: 20px;
  font-size: 12px;
  color: #999;
}
</style>
