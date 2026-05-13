import request from './request'
import { API_BASE_URL } from './config'

/**
 * 上传知识库文档
 * @param {FormData} formData - 包含文件的FormData
 * @returns {Promise}
 */
export const uploadDocument = (formData) => {
  return request.post(`${API_BASE_URL}/coach/knowledge/upload`, formData, {
    headers: {
      // 当发送 FormData 时，不要手动设置 Content-Type，fetch 会自动设置并包含 boundary
    }
  })
}

/**
 * 获取知识库文档列表
 * @returns {Promise}
 */
export const getDocumentList = () => {
  return request.get(`${API_BASE_URL}/coach/knowledge/list`)
}

/**
 * 删除知识库文档
 * @param {String} id - 文档ID
 * @returns {Promise}
 */
export const deleteDocument = (id) => {
  return request.delete(`${API_BASE_URL}/coach/knowledge/${id}`)
}
