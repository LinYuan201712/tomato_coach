// 番茄通知工具函数
import { createApp, h } from 'vue'
import TomatoNotification from '@/components/TomatoNotification/TomatoNotification.vue'

let notificationInstance = null
let notificationContainer = null

/**
 * 显示番茄通知
 * @param {Object} options - 通知选项
 * @param {String} options.type - 类型: 'earn' 或 'spend'
 * @param {String} options.title - 标题
 * @param {String} options.message - 消息内容
 * @param {Number} options.duration - 持续时间（毫秒）
 */
export function showTomatoNotification(options) {
  console.log('📢 显示番茄通知:', options)
  
  // 如果已有通知，先关闭
  if (notificationInstance) {
    hideTomatoNotification()
  }

  // 创建容器
  if (!notificationContainer) {
    notificationContainer = document.createElement('div')
    document.body.appendChild(notificationContainer)
    console.log('✅ 通知容器已创建并添加到body')
  }

  // 创建Vue应用实例，使用render函数而不是template
  const app = createApp({
    render() {
      return h(TomatoNotification, {
        visible: true,
        type: options.type || 'earn',
        title: options.title || '',
        message: options.message || '',
        duration: options.duration || 3000,
        autoClose: options.autoClose !== false,
        onClose: () => {
          setTimeout(() => {
            hideTomatoNotification()
          }, 400) // 等待动画完成
        }
      })
    }
  })

  notificationInstance = app
  notificationInstance.mount(notificationContainer)
}

/**
 * 隐藏番茄通知
 */
export function hideTomatoNotification() {
  if (notificationInstance) {
    notificationInstance.unmount()
    notificationInstance = null
  }
  if (notificationContainer) {
    document.body.removeChild(notificationContainer)
    notificationContainer = null
  }
}

/**
 * 显示获得番茄通知
 * @param {Number} count - 番茄数量
 * @param {String} reason - 原因
 */
export function showEarnTomatoNotification(count, reason) {
  showTomatoNotification({
    type: 'earn',
    title: `获得 ${count} 个番茄 🍅`,
    message: reason || '继续加油哦！',
    duration: 3000
  })
}

/**
 * 显示消耗番茄通知
 * @param {Number} count - 番茄数量
 * @param {String} reason - 原因
 */
export function showSpendTomatoNotification(count, reason) {
  console.log('🍅 显示消耗番茄通知:', count, reason)
  showTomatoNotification({
    type: 'spend',
    title: `消耗 ${count} 个番茄 🍅`,
    message: reason || '',
    duration: 3000
  })
}

