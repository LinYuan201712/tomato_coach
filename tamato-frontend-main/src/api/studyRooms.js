// api/studyRooms.js
import request from './request'
import { API_BASE_URL, getToken } from './config'

const BASE_URL = API_BASE_URL

// 获取自习室列表
export const getRoomsList = () => {
  return request.get(`${BASE_URL}/rooms`)
}

// 创建自习室
export const createRoom = (roomData) => {
  const token = getToken()
  const config = token ? {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    }
  } : {}
  
  console.log('创建自习室请求配置:', config)
  return request.post(`${BASE_URL}/rooms`, roomData, config)
}

// 获取自习室详情
export const getRoomDetail = (roomId, userId) => {
  const token = getToken()
  const config = token ? {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  } : {}
  const url = userId ? `${BASE_URL}/rooms/${roomId}?userId=${userId}` : `${BASE_URL}/rooms/${roomId}`
  
  return request.get(url, config)
}

// 更新自习室信息
export const updateRoom = (roomId, roomData) => {
  const token = getToken()
  const config = token ? {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    }
  } : {}
  
  return request.put(`${BASE_URL}/rooms/${roomId}`, roomData, config)
}

// 解散自习室
export const deleteRoom = (roomId, userId) => {
  const token = getToken()
  const config = token ? {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  } : {}
  const url = userId ? `${BASE_URL}/rooms/${roomId}?userId=${userId}` : `${BASE_URL}/rooms/${roomId}`
  
  return request.delete(url, null, config)
}

// 加入自习室
export const joinRoom = (roomId, userId) => {
  const token = getToken()
  const config = token ? {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  } : {}
  
  const url = `${BASE_URL}/rooms/${roomId}/join?userId=${userId}`
  
  console.log('加入房间请求URL:', url)
  console.log('请求配置:', config)
  
  return request.post(url, {}, config)
}

// 退出房间
export const leaveRoom = (roomId, userId) => {
  const url = `${BASE_URL}/rooms/${roomId}/leave?userId=${userId}`
  console.log('退出房间请求URL:', url)
  return request.post(url, {}, {})
}

// 房主退出自习室（房主身份转移给下一个成员）
export const leaveRoomAsHost = (roomId, userId) => {
  const url = `${BASE_URL}/rooms/${roomId}/leave-as-host?userId=${userId}`
  console.log('房主退出房间请求URL:', url)
  return request.post(url, {}, {})
}

// 获取自习室成员列表
export const getRoomMembers = (roomId, userId) => {
  const token = getToken()
  const config = token ? {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  } : {}
  const url = userId ? `${BASE_URL}/rooms/${roomId}/members?userId=${userId}` : `${BASE_URL}/rooms/${roomId}/members`
  
  return request.get(url, config)
}

// 踢出成员
export const kickMember = (roomId, userId) => {
  const token = getToken()
  const config = token ? {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  } : {}
  
  return request.delete(`${BASE_URL}/rooms/${roomId}/members/${userId}`, config)
}

// 更新用户状态（专注/休息）
export const updateUserStatus = async (roomId, statusData) => {
  const token = getToken()
  const config = token ? {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    }
  } : {}

  const url = `${BASE_URL}/rooms/${roomId}/status`
  
  console.log('🔄 更新用户状态', { roomId, statusData, url })
  
  try {
    const response = await request.put(url, statusData, config)
    console.log('✅ 用户状态更新成功', response)
    return response
  } catch (error) {
    console.error('❌ 用户状态更新失败', error)
    throw error
  }
}

// 获取或创建个人自习室
export const getOrCreatePersonalRoom = () => {
  const token = getToken()
  const config = token ? {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  } : {}
  return request.get(`${BASE_URL}/rooms/personal`, config)
}
