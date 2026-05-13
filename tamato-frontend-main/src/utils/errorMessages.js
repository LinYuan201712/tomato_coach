// 错误消息映射工具
// 将后端返回的技术性错误消息转换为用户友好的提示

export const getFriendlyErrorMessage = (errorMessage, type = 'login') => {
  if (!errorMessage) {
    return '操作失败，请重试'
  }

  const message = errorMessage.toLowerCase()

  // 登录相关错误
  if (type === 'login') {
    if (message.includes('用户不存在') || message.includes('用户不')) {
      return '用户名或密码错误，请检查后重新输入'
    }
    if (message.includes('密码错误') || message.includes('密码')) {
      return '密码输入错误，请重新输入'
    }
    if (message.includes('token') || message.includes('授权')) {
      return '登录已过期，请重新登录'
    }
    if (message.includes('网络') || message.includes('连接')) {
      return '网络连接失败，请检查网络后重试'
    }
  }

  // 注册相关错误
  if (type === 'register') {
    if (message.includes('用户名已存在')) {
      return '该用户名已被使用，请选择其他用户名'
    }
    if (message.includes('邮箱已被注册') || message.includes('邮箱')) {
      return '该邮箱已被注册，请使用其他邮箱或直接登录'
    }
    if (message.includes('手机号已被注册')) {
      return '该手机号已被注册，请使用其他手机号'
    }
    if (message.includes('网络') || message.includes('连接')) {
      return '网络连接失败，请检查网络后重试'
    }
  }

  // 通用错误
  if (message.includes('网络') || message.includes('连接') || message.includes('fetch')) {
    return '网络连接失败，请检查网络后重试'
  }

  // 如果无法匹配，返回原始消息（但可以稍作美化）
  return errorMessage
}

