import request from './request'

// 获取昨日学习报告
export function getDailyReport() {
  return request.get('/api/reports/daily')
}

// 重新生成昨日学习报告
export function regenerateDailyReport() {
  return request.post('/api/reports/daily/regenerate')
}
