/** 解包后端统一响应 { success, data, message } */
export function unwrapApiData(response) {
  if (response && response.success === false) {
    throw new Error(response.message || '请求失败')
  }
  if (response && Object.prototype.hasOwnProperty.call(response, 'data')) {
    return response.data
  }
  return response
}
