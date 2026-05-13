// JWT 工具函数
// 注意：这里只解析 JWT payload，不验证签名（前端无法验证）

/**
 * 从 JWT token 中解析 userId（不验证签名）
 * JWT 格式：header.payload.signature
 * payload 是 base64url 编码的 JSON
 */
export function getUserIdFromToken(token) {
  if (!token) {
    return null
  }

  try {
    // JWT 由三部分组成，用 . 分隔
    const parts = token.split('.')
    if (parts.length !== 3) {
      console.warn('JWT token 格式错误')
      return null
    }

    // 解析 payload（第二部分）
    const payload = parts[1]
    
    // Base64URL 解码
    // Base64URL 使用 - 和 _ 代替 + 和 /，并且没有填充 =
    let base64 = payload.replace(/-/g, '+').replace(/_/g, '/')
    
    // 添加填充
    while (base64.length % 4) {
      base64 += '='
    }

    // 解码
    const decodedPayload = JSON.parse(atob(base64))
    
    // JWT 的 subject (sub) 字段就是 userId
    const userId = decodedPayload.sub
    
    if (userId) {
      return Number(userId)
    }
    
    return null
  } catch (error) {
    console.error('解析 JWT token 失败:', error)
    return null
  }
}


