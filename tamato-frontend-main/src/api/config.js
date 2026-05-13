// API配置文件
// 后端服务地址
// 开发环境：使用代理，直接使用 /api
// 生产环境：使用完整URL http://localhost:8090/api
export const API_BASE_URL = process.env.NODE_ENV === 'production'
  ? 'http://localhost:8091/api'
  : '/api'

// 从localStorage获取token
export const getToken = () => {
  return localStorage.getItem('token') || ''
}

// 保存token到localStorage
export const setToken = (token) => {
  localStorage.setItem('token', token)
}

// 清除token
export const removeToken = () => {
  localStorage.removeItem('token')
}

