import request from './request'
import { API_BASE_URL } from './config'

// 获取当前用户的所有任务
export function getTasks(userId) {
  return request.get(`${API_BASE_URL}/me/tasks?userId=${userId}`)
}

// 创建新任务
export function createTask(data) {
  return request.post(`${API_BASE_URL}/tasks`, data)
}

// 更新任务信息
export function updateTask(taskId, data) {
  // 后端使用 /tasks/edit 端点
  const updateData = {
    ...data,
    task_id: taskId,
    taskId: taskId
  }
  console.log('📤 updateTask 发送的完整数据:', JSON.stringify(updateData, null, 2))
  return request.put(`${API_BASE_URL}/tasks/edit`, updateData)
}

// 完成任务（使用后端专门的完成接口）
export function completeTaskApi(taskId) {
  console.log(`🚀 调用专门的完成接口: /tasks/${taskId}/complete`)
  return request.put(`${API_BASE_URL}/tasks/${taskId}/complete`)
}

// 删除任务
export function deleteTask(taskId) {
  // 后端 DeleteTaskRequest 只需要 task_id，用户身份从 token 中解析
  const deleteBody = { task_id: taskId }
  console.log('📤 deleteTask 发送的请求体:', deleteBody)
  return request.delete(`${API_BASE_URL}/tasks/delete`, deleteBody)
}