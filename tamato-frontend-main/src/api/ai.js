// AI对话API
// 通过后端调用大模型API，避免CORS跨域问题

import request from './request'
import { API_BASE_URL, getToken } from './config'

const STREAM_API_BASE_URL = process.env.NODE_ENV === 'production'
  ? API_BASE_URL
  : 'http://localhost:8091/api'

/**
 * 与AI对话
 * @param {Array} messages - 消息历史
 * @param {Boolean} deductTomato - 是否扣除番茄
 * @param {Boolean} useKnowledge - 知识库增强
 * @param {String} sessionID - 会话ID
 * @returns {Promise} AI回复
 */
/**
 * 与AI对话 (普通模式)
 */
export const chatWithAI = async (messages, deductTomato = true, useKnowledge = false, sessionID = 'default', chatMode = 'thinking') => {
  try {
    const lastMessage = (messages && messages.length) ? messages[messages.length - 1].content : ''
    const response = await request.post(`${API_BASE_URL}/coach/chat`, {
      message: lastMessage,
      deductTomato: deductTomato !== false,
      use_knowledge: useKnowledge === true,
      session_id: sessionID,
      chat_mode: chatMode
    })
    
    if (response && typeof response === 'object') {
      if (response.content) {
        return { 
          content: response.content, 
          reasoning: response.reasoning,
          usage: response.usage 
        }
      }
    }
    return { content: '抱歉，后端响应异常' }
  } catch (error) {
    console.error('❌ 后端AI接口调用失败:', error)
    return { content: '抱歉，连接服务器失败' }
  }
}

/**
 * 流式与AI对话
 */
export const chatStreamWithAI = async (params, onMessage, onDone, onError) => {
  try {
    const response = await fetch(`${STREAM_API_BASE_URL}/coach/chat/stream`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'text/event-stream',
        'Authorization': `Bearer ${getToken()}`
      },
      body: JSON.stringify({
        message: params.message,
        use_knowledge: params.use_knowledge,
        session_id: params.session_id,
        chat_mode: params.chat_mode || 'thinking'
      })
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''

    const handleEventBlock = (block) => {
      let eventName = 'message'
      const dataLines = []

      for (const line of block.split(/\r?\n/)) {
        if (line.startsWith('event:')) {
          eventName = line.slice(6).trim()
        } else if (line.startsWith('data:')) {
          dataLines.push(line.slice(5).trimStart())
        }
      }

      if (dataLines.length === 0) return

      const dataContent = dataLines.join('\n')
      try {
        const data = JSON.parse(dataContent)
        if (eventName === 'done') {
          onDone(data)
        } else {
          onMessage(data)
        }
      } catch (e) {
        console.error('解析SSE数据失败:', e, dataContent)
      }
    }

    let reading = true
    while (reading) {
      const { value, done } = await reader.read()
      if (done) {
        if (buffer.trim()) {
          handleEventBlock(buffer)
        }
        reading = false
        break
      }
      
      buffer += decoder.decode(value, { stream: true })
      const blocks = buffer.split(/\r?\n\r?\n/)
      buffer = blocks.pop()

      for (const block of blocks) {
        if (block.trim()) {
          handleEventBlock(block)
        }
      }
    }
  } catch (error) {
    console.error('流式请求失败:', error)
    onError(error)
  }
}

/**
 * 获取会话列表
 */
export const getSessions = async () => {
  try {
    const response = await request.get(`${API_BASE_URL}/coach/sessions`)
    return response.data || []
  } catch (error) {
    console.error('获取会话列表失败:', error)
    return []
  }
}

/**
 * 获取指定会话的历史记录
 */
export const getHistory = async (sessionID) => {
  try {
    const response = await request.get(`${API_BASE_URL}/coach/sessions/${sessionID}/history`)
    return response.data || []
  } catch (error) {
    console.error('获取历史记录失败:', error)
    return []
  }
}

/**
 * 创建新会话
 */
export const createSession = async (sessionID, title) => {
  try {
    await request.post(`${API_BASE_URL}/coach/sessions`, {
      session_id: sessionID,
      title: title
    })
    return true
  } catch (error) {
    console.error('创建会话失败:', error)
    return false
  }
}

/**
 * 更新会话标题
 */
export const updateSessionTitle = async (sessionID, title) => {
  try {
    await request.put(`${API_BASE_URL}/coach/sessions/${sessionID}`, {
      title: title
    })
    return true
  } catch (error) {
    console.error('更新会话标题失败:', error)
    return false
  }
}

/**
 * 获取知识库文件列表
 */
export const listKnowledge = async (folderID = 0) => {
  try {
    const response = await request.get(`${API_BASE_URL}/coach/knowledge/list?folderID=${folderID}`)
    return response.data || []
  } catch (error) {
    console.error('获取知识库列表失败:', error)
    return []
  }
}

/**
 * 上传知识库文件
 */
export const uploadKnowledge = async (formData) => {
  try {
    const response = await request.post(`${API_BASE_URL}/coach/knowledge/upload`, formData)
    return response
  } catch (error) {
    console.error('上传知识库失败:', error)
    throw error
  }
}

/**
 * 删除知识库文件
 */
export const deleteKnowledge = async (id) => {
  try {
    await request.delete(`${API_BASE_URL}/coach/knowledge/${id}`)
    return true
  } catch (error) {
    console.error('删除知识库失败:', error)
    return false
  }
}

/**
 * 重命名知识库文件
 */
export const renameKnowledge = async (id, newName) => {
  try {
    await request.put(`${API_BASE_URL}/coach/knowledge/${id}/rename`, {
      newName: newName
    })
    return true
  } catch (error) {
    console.error('重命名文件失败:', error)
    return false
  }
}

/**
 * 创建文件夹
 */
export const createFolder = async (name, parentID = 0) => {
  try {
    const response = await request.post(`${API_BASE_URL}/coach/folders`, {
      name: name,
      parentId: parentID
    })
    return response.data
  } catch (error) {
    console.error('创建文件夹失败:', error)
    return null
  }
}

/**
 * 获取文件夹列表
 */
export const listFolders = async (parentID = 0) => {
  try {
    const response = await request.get(`${API_BASE_URL}/coach/folders?parentId=${parentID}`)
    return response.data || []
  } catch (error) {
    console.error('获取文件夹列表失败:', error)
    return []
  }
}

/**
 * 删除文件夹
 */
export const deleteFolder = async (id) => {
  try {
    await request.delete(`${API_BASE_URL}/coach/folders/${id}`)
    return true
  } catch (error) {
    console.error('删除文件夹失败:', error)
    return false
  }
}

/**
 * 移动知识库文件
 */
export const moveKnowledge = async (fileID, targetFolderID) => {
  try {
    await request.post(`${API_BASE_URL}/coach/knowledge/move`, {
      fileId: fileID,
      targetFolderId: targetFolderID
    })
    return true
  } catch (error) {
    console.error('移动文件失败:', error)
    return false
  }
}

/**
 * 获取知识库文件预览内容
 */
export const getKnowledgePreview = async (fileName) => {
  try {
    const response = await request.get(`${API_BASE_URL}/coach/knowledge/preview?fileName=${encodeURIComponent(fileName)}`)
    return response.data?.content || ''
  } catch (error) {
    console.error('获取文件预览失败:', error)
    return '无法加载文件预览内容。'
  }
}

/**
 * 获取用户画像及锁定状态
 */
export const getUserProfile = async () => {
  try {
    const response = await request.get(`${API_BASE_URL}/coach/profile`)
    return response.data
  } catch (error) {
    console.error('获取用户画像失败:', error)
    return null
  }
}

/**
 * 更新用户画像及锁定状态
 */
export const updateUserProfile = async (data) => {
  try {
    await request.put(`${API_BASE_URL}/coach/profile`, data)
    return true
  } catch (error) {
    console.error('更新用户画像失败:', error)
    return false
  }
}

export default {
  getSessions,
  createSession,
  getHistory,
  updateSessionTitle,
  getKnowledgePreview,
  listKnowledge,
  uploadKnowledge,
  deleteKnowledge,
  renameKnowledge,
  moveKnowledge,
  createFolder,
  listFolders,
  deleteFolder,
  getUserProfile,
  updateUserProfile
}
