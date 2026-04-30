<template>
  <div class="status-bar">
    <div class="status-item">
      <span class="label">后端状态</span>
      <span :class="['badge', backendOnline ? 'badge-ok' : 'badge-err']">
        {{ backendOnline ? '在线' : '离线' }}
      </span>
    </div>
    <div class="status-item">
      <span class="label">iperf3 Server</span>
      <span :class="['badge', serverRunning ? 'badge-ok' : 'badge-off']">
        {{ serverRunning ? `运行中 (:${serverPort})` : '未启动' }}
      </span>
    </div>
    <div class="status-actions">
      <button v-if="!serverRunning" class="btn btn-sm btn-primary" @click="doStartServer" :disabled="!backendOnline">
        启动 Server
      </button>
      <button v-else class="btn btn-sm btn-danger" @click="doStopServer">
        停止 Server
      </button>
    </div>
  </div>
</template>

<script>
import { healthCheck, getServerStatus, startServer, stopServer } from '../api/index.js'

export default {
  name: 'StatusBar',
  data() {
    return {
      backendOnline: false,
      serverRunning: false,
      serverPort: 0,
    }
  },
  mounted() {
    this.refresh()
    this._timer = setInterval(() => this.refresh(), 5000)
  },
  beforeUnmount() {
    clearInterval(this._timer)
  },
  methods: {
    async refresh() {
      try {
        await healthCheck()
        this.backendOnline = true
        const { data: statusData } = await getServerStatus()
        if (statusData.code === 0 && statusData.data) {
          this.serverRunning = statusData.data.running
          this.serverPort = statusData.data.port
        }
      } catch {
        this.backendOnline = false
      }
    },
    async doStartServer() {
      try {
        await startServer(5201)
        await this.refresh()
      } catch (e) {
        alert('启动失败: ' + (e.response?.data?.message || e.message))
      }
    },
    async doStopServer() {
      try {
        await stopServer()
        await this.refresh()
      } catch (e) {
        alert('停止失败: ' + (e.response?.data?.message || e.message))
      }
    },
  },
}
</script>

<style scoped>
.status-bar {
  display: flex;
  align-items: center;
  gap: 24px;
  background: #fff;
  padding: 12px 20px;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.08);
}

.status-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.label {
  font-size: 13px;
  color: #666;
}

.badge {
  padding: 2px 10px;
  border-radius: 10px;
  font-size: 12px;
  font-weight: 500;
}

.badge-ok { background: #d4edda; color: #155724; }
.badge-err { background: #f8d7da; color: #721c24; }
.badge-off { background: #e2e3e5; color: #383d41; }

.status-actions { margin-left: auto; }

.btn {
  border: none;
  padding: 6px 14px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
}

.btn:disabled { opacity: 0.5; cursor: not-allowed; }

.btn-sm { padding: 4px 12px; font-size: 12px; }

.btn-primary { background: #4361ee; color: #fff; }
.btn-primary:hover:not(:disabled) { background: #3a56d4; }

.btn-danger { background: #e63946; color: #fff; }
.btn-danger:hover:not(:disabled) { background: #c1121f; }
</style>
