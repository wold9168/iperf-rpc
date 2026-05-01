<template>
  <div id="app">
    <header class="app-header">
      <h1>iperf-rpc</h1>
      <span class="subtitle">Kubernetes 网络测速工具</span>
    </header>
    <main class="app-main">
      <StatusBar />
      <div class="columns">
        <div class="column column-left">
          <IperfForm @result="handleResult" />
        </div>
        <div class="column column-right">
          <ResultView :result="currentResult" />
        </div>
      </div>
    </main>
  </div>
</template>

<script>
import IperfForm from './components/IperfForm.vue'
import ResultView from './components/ResultView.vue'
import StatusBar from './components/StatusBar.vue'

export default {
  name: 'App',
  components: { IperfForm, ResultView, StatusBar },
  data() {
    return {
      currentResult: null,
    }
  },
  methods: {
    handleResult(result) {
      this.currentResult = result.data
    },
  },
}
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background: #f0f2f5;
  color: #333;
}

.app-header {
  background: #1a1a2e;
  color: #fff;
  padding: 16px 24px;
  display: flex;
  align-items: baseline;
  gap: 12px;
}

.app-header h1 {
  font-size: 20px;
  font-weight: 600;
}

.app-header .subtitle {
  font-size: 13px;
  color: #8892b0;
}

.app-main {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

.columns {
  display: flex;
  gap: 20px;
  margin-top: 16px;
}

.column-left {
  flex: 0 0 400px;
}

.column-right {
  flex: 1;
  min-width: 0;
}

@media (max-width: 768px) {
  .columns {
    flex-direction: column;
  }
  .column-left {
    flex: 1;
  }
}
</style>
