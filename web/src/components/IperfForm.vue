<template>
  <div class="form-card">
    <h3>iperf3 指令构造</h3>

    <div class="form-group">
      <label>模式</label>
      <select v-model="form.mode">
        <option value="client">客户端 (client)</option>
        <option value="server">服务端 (server)</option>
      </select>
    </div>

    <div class="form-group" v-if="form.mode === 'client'">
      <label>目标地址 <span class="required">*</span></label>
      <input v-model="form.args.target" placeholder="例如: 10.0.0.2" />
    </div>

    <div class="form-row">
      <div class="form-group">
        <label>端口</label>
        <input v-model.number="form.args.port" type="number" placeholder="5201" />
      </div>
      <div v-if="form.mode === 'client'" class="form-group">
        <label>协议</label>
        <select v-model="form.args.protocol">
          <option value="tcp">TCP</option>
          <option value="udp">UDP</option>
        </select>
      </div>
    </div>

    <template v-if="form.mode === 'client'">
      <div class="form-row">
        <div class="form-group">
          <label>时长 (秒)</label>
          <input v-model.number="form.args.duration" type="number" placeholder="10" />
        </div>
        <div class="form-group">
          <label>并行流数</label>
          <input v-model.number="form.args.parallel" type="number" placeholder="1" />
        </div>
      </div>

      <div class="form-group">
        <label>带宽限制</label>
        <input v-model="form.args.bandwidth" placeholder="例如: 100M, 1G (留空不限制)" />
      </div>

      <div class="form-group">
        <label>
          <input type="checkbox" v-model="form.args.reverse" />
          反向测试 (服务端发送)
        </label>
      </div>

      <div class="form-group">
        <label>额外参数</label>
        <input v-model="form.args.extra" placeholder="例如: --omit 2 --window 256K" />
      </div>
    </template>

    <button v-if="form.mode === 'client'" class="btn btn-primary btn-block" @click="submit" :disabled="loading">
      {{ loading ? '执行中...' : '执行测速' }}
    </button>

    <div v-if="error" class="error-msg">{{ error }}</div>
  </div>

  <div class="form-card http-card">
    <h3>HTTP 测速 (支持 SOCKS5 代理)</h3>

    <div class="form-group">
      <label>目标实例 <span class="required">*</span></label>
      <div class="target-row">
        <select class="proto-select" v-model="http.tls">
          <option :value="false">HTTP</option>
          <option :value="true">HTTPS</option>
        </select>
        <input class="target-input" v-model="http.target" placeholder="10.0.0.2:8080" />
      </div>
      <span class="hint">{{ httpUrl }}</span>
    </div>

    <div class="form-group">
      <label>SOCKS5 代理 (可选)</label>
      <input v-model="http.proxy" placeholder="socks5://proxy:1080" />
    </div>

    <div class="form-row">
      <div class="form-group">
        <label>方向</label>
        <select v-model="http.direction">
          <option value="download">下载 (download)</option>
          <option value="upload">上传 (upload)</option>
        </select>
      </div>
      <div class="form-group" v-if="http.direction === 'download'">
        <label>数据量</label>
        <input v-model="http.dataSize" placeholder="100M" />
      </div>
      <div class="form-group" v-else>
        <label>时长 (秒)</label>
        <input v-model.number="http.duration" type="number" placeholder="10" />
      </div>
    </div>

    <div class="form-group" v-if="http.tls">
      <label>
        <input type="checkbox" v-model="http.insecure" />
        跳过 TLS 证书验证 (不安全)
      </label>
    </div>

    <button class="btn btn-primary btn-block" @click="submitHttp" :disabled="httpLoading">
      {{ httpLoading ? '执行中...' : '执行 HTTP 测速' }}
    </button>

    <div v-if="httpError" class="error-msg">{{ httpError }}</div>
  </div>
</template>

<script>
import { runIperf, runHttpTest } from '../api/index.js'

export default {
  name: 'IperfForm',
  emits: ['result'],
  data() {
    return {
      form: {
        mode: 'client',
        args: {
          target: '',
          port: 5201,
          duration: 10,
          parallel: 1,
          bandwidth: '',
          protocol: 'tcp',
          reverse: false,
          extra: '',
        },
      },
      loading: false,
      error: '',

      http: {
        target: '',
        proxy: '',
        direction: 'download',
        duration: 10,
        dataSize: '100M',
        tls: false,
        insecure: false,
      },
      httpLoading: false,
      httpError: '',
    }
  },
  computed: {
    httpUrl() {
      if (!this.http.target) return ''
      const scheme = this.http.tls ? 'https' : 'http'
      const base = `${scheme}://${this.http.target}/api/v1/http`
      if (this.http.direction === 'download') {
        return `${base}/data?size=${this.http.dataSize || '100M'}`
      }
      return `${base}/upload`
    },
  },
  methods: {
    async submit() {
      this.error = ''

      if (this.form.mode === 'client' && !this.form.args.target) {
        this.error = '客户端模式必须填写目标地址'
        return
      }

      this.loading = true
      try {
        const { data } = await runIperf({
          mode: this.form.mode,
          args: {
            ...this.form.args,
            port: this.form.args.port || 5201,
            duration: this.form.args.duration || 10,
            parallel: this.form.args.parallel || 1,
          },
        })
        this.$emit('result', { type: 'iperf', data: data.data })
      } catch (e) {
        this.error = '请求失败: ' + (e.response?.data?.message || e.message)
      } finally {
        this.loading = false
      }
    },
    async submitHttp() {
      this.httpError = ''

      if (!this.http.target) {
        this.httpError = '必须填写目标实例地址'
        return
      }

      this.httpLoading = true
      try {
        const { data } = await runHttpTest({
          url: this.httpUrl,
          proxy: this.http.proxy || undefined,
          direction: this.http.direction,
          duration: this.http.duration || 10,
          insecure: this.http.insecure,
        })
        this.$emit('result', { type: 'http', data: data.data })
      } catch (e) {
        this.httpError = '请求失败: ' + (e.response?.data?.message || e.message)
      } finally {
        this.httpLoading = false
      }
    },
  },
}
</script>

<style scoped>
.form-card {
  background: #fff;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.08);
}

.http-card { margin-top: 16px; }

.form-card h3 {
  font-size: 16px;
  margin-bottom: 16px;
  color: #1a1a2e;
}

.form-group { margin-bottom: 12px; }

.form-group label {
  display: block;
  font-size: 13px;
  color: #555;
  margin-bottom: 4px;
}

.required { color: #e63946; }

.target-row {
  display: flex;
  gap: 6px;
}

.form-group .proto-select {
  width: 72px;
  flex-shrink: 0;
  padding: 8px 8px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  outline: none;
}

.form-group .proto-select:focus {
  border-color: #4361ee;
}

.target-input {
  flex: 1;
}

.hint {
  display: block;
  font-size: 11px;
  color: #aaa;
  margin-top: 4px;
  word-break: break-all;
}

.form-group input[type="text"],
.form-group input[type="number"],
.form-group select {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  outline: none;
  transition: border-color 0.2s;
}

.form-group input:focus,
.form-group select:focus {
  border-color: #4361ee;
}

.form-row {
  display: flex;
  gap: 12px;
}

.form-row .form-group { flex: 1; }

.btn-block { width: 100%; margin-top: 8px; }

.btn {
  border: none;
  padding: 10px 20px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
}

.btn:disabled { opacity: 0.6; cursor: not-allowed; }

.btn-primary { background: #4361ee; color: #fff; }
.btn-primary:hover:not(:disabled) { background: #3a56d4; }

.error-msg {
  margin-top: 10px;
  padding: 8px 12px;
  background: #f8d7da;
  color: #721c24;
  border-radius: 6px;
  font-size: 13px;
}
</style>
