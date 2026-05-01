import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 120000,
})

export function runIperf(data) {
  return api.post('/iperf/run', data)
}

export function startServer(port) {
  return api.post('/iperf/server/start', null, { params: { port } })
}

export function stopServer() {
  return api.post('/iperf/server/stop')
}

export function getServerStatus() {
  return api.get('/iperf/status')
}

export function healthCheck() {
  return api.get('/health')
}

export function runHttpTest(data) {
  return api.post('/http/run', data)
}

export default api
