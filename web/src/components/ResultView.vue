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

      <template v-if="parsed && parsed.end">
        <div class="summary-grid">
          <div class="summary-item">
            <span class="summary-value">{{ formatBitrate(sender.bitrate) }}</span>
            <span class="summary-label">{{ sender.label }}吞吐量</span>
          </div>
          <div class="summary-item">
            <span class="summary-value">{{ formatBytes(sender.bytes) }}</span>
            <span class="summary-label">传输总量</span>
          </div>
          <div class="summary-item">
            <span class="summary-value">{{ sender.retransmits ?? 0 }}</span>
            <span class="summary-label">重传次数</span>
          </div>
          <div class="summary-item">
            <span class="summary-value">{{ parsed.end.sum?.jitter_ms?.toFixed(2) ?? '-' }} ms</span>
            <span class="summary-label">抖动</span>
          </div>
          <div class="summary-item">
            <span class="summary-value">{{ parsed.end.sum?.lost_percent ?? 0 }}%</span>
            <span class="summary-label">丢包率</span>
          </div>
          <div class="summary-item">
            <span class="summary-value">{{ sender.duration?.toFixed(1) ?? '-' }}s</span>
            <span class="summary-label">耗时</span>
          </div>
          <div class="summary-item" v-if="parsed.end.cpu_utilization_percent">
            <span class="summary-value">{{ cpuLocal }}% / {{ cpuRemote }}%</span>
            <span class="summary-label">CPU (本机/远端)</span>
          </div>
          <div class="summary-item">
            <span class="summary-value">{{ mtu }}</span>
            <span class="summary-label">MTU</span>
          </div>
        </div>

        <div v-if="chartData" class="chart-container">
          <Line :data="chartData" :options="chartOptions" />
        </div>
      </template>

      <div v-else-if="result.status === 'error'" class="error-block">
        <pre>{{ result.output }}</pre>
      </div>

      <details class="raw-output">
        <summary>JSON 原始输出</summary>
        <pre>{{ formattedOutput }}</pre>
      </details>

      <div class="timestamps">
        <span>开始: {{ formatTime(result.started_at) }}</span>
        <span v-if="result.finished_at">结束: {{ formatTime(result.finished_at) }}</span>
      </div>
    </div>
  </div>
</template>

<script>
import { Line } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Filler,
} from 'chart.js'

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Filler)

export default {
  name: 'ResultView',
  components: { Line },
  props: {
    result: { type: Object, default: null },
  },
  computed: {
    parsed() {
      if (!this.result?.output) return null
      try {
        return JSON.parse(this.result.output)
      } catch {
        return null
      }
    },
    sender() {
      if (!this.parsed?.end) return {}
      const sent = this.parsed.end.sum_sent
      const recv = this.parsed.end.sum_received
      if (sent) return { ...sent, label: '发送', bitrate: sent.bits_per_second, duration: sent.seconds }
      if (recv) return { ...recv, label: '接收', bitrate: recv.bits_per_second, duration: recv.seconds }
      return {}
    },
    cpuLocal() {
      return this.parsed?.end?.cpu_utilization_percent?.host_total?.toFixed(1) ?? '-'
    },
    cpuRemote() {
      return this.parsed?.end?.cpu_utilization_percent?.remote_total?.toFixed(1) ?? '-'
    },
    mtu() {
      const conn = this.parsed?.start?.connected?.[0]
      return conn?.mtu ?? '-'
    },
    formattedOutput() {
      if (!this.result?.output) return ''
      try {
        return JSON.stringify(JSON.parse(this.result.output), null, 2)
      } catch {
        return this.result.output
      }
    },
    chartData() {
      const intervals = this.parsed?.intervals
      if (!intervals?.length) return null

      const labels = intervals.map((_, i) => `${i + 1}s`)
      const bitrates = intervals.map((iv) => {
        const s = iv.sum
        return s ? (s.bits_per_second / 1e6) : 0
      })

      return {
        labels,
        datasets: [
          {
            label: '吞吐量 (Mbps)',
            data: bitrates,
            borderColor: '#4361ee',
            backgroundColor: 'rgba(67,97,238,0.1)',
            fill: true,
            tension: 0.3,
            pointRadius: 3,
            pointHoverRadius: 5,
          },
        ],
      }
    },
    chartOptions() {
      return {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { display: false },
          tooltip: {
            callbacks: {
              label: (ctx) => `${ctx.parsed.y.toFixed(1)} Mbps`,
            },
          },
        },
        scales: {
          x: {
            title: { display: true, text: '时间' },
            grid: { display: false },
          },
          y: {
            title: { display: true, text: 'Mbps' },
            beginAtZero: true,
          },
        },
      }
    },
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
    formatBytes(bytes) {
      if (!bytes || bytes === 0) return '-'
      if (bytes >= 1e9) return (bytes / 1e9).toFixed(2) + ' GB'
      if (bytes >= 1e6) return (bytes / 1e6).toFixed(2) + ' MB'
      if (bytes >= 1e3) return (bytes / 1e3).toFixed(2) + ' KB'
      return bytes + ' B'
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
  font-size: 18px;
  font-weight: 700;
  color: #1a1a2e;
}

.summary-label {
  display: block;
  font-size: 12px;
  color: #888;
  margin-top: 4px;
}

.chart-container {
  height: 260px;
  margin: 20px 0;
}

.error-block {
  background: #fff5f5;
  border: 1px solid #fecaca;
  padding: 12px;
  border-radius: 6px;
  margin: 12px 0;
}

.error-block pre {
  margin: 0;
  font-size: 13px;
  color: #991b1b;
  white-space: pre-wrap;
  word-break: break-all;
}

.raw-output { margin-top: 16px; }

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
