import request from './request'
import { API_BASE_URL } from './config'

// 搜索用户（通过用户名）
// 只有ID存在的用户名才可以被搜索到
export const searchUser = async (username) => {
  return request.get(`${API_BASE_URL}/users/${encodeURIComponent(username)}`)
}

// 发送好友申请
export const sendFriendRequest = async (data) => {
  return request.post(`${API_BASE_URL}/friends/requests`, data)
}

// 查看好友申请
export const getFriendRequests = async () => {
  return request.get(`${API_BASE_URL}/friends/requests`)
}

// 处理好友申请（接受/拒绝）
export const processFriendRequest = async (data) => {
  return request.put(`${API_BASE_URL}/friends/requests`, data)
}

// 获取好友列表
export const getFriends = async () => {
  return request.get(`${API_BASE_URL}/friends`)
}

// 删除好友
export const deleteFriend = async (data) => {
  return request.post(`${API_BASE_URL}/friends/delete`, data)
}

