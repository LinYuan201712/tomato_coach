<template>
  <div id="app">
    <router-view/>
    <!-- 悬浮桌宠 - 只在登录后显示 -->
    <DesktopPet v-if="isAuthenticated && !isLoginPage" />
  </div>
</template>

<script>
import DesktopPet from '@/components/DesktopPet/DesktopPet.vue'
import { getToken } from '@/api/config'
import { setUserOffline, getCurrentUser } from '@/api/user'
import { getUserIdFromToken } from '@/utils/jwt'

export default {
  name: 'App',
  components: {
    DesktopPet
  },
  data() {
    return {
      currentUserId: null,
      // 用于 beforeunload 事件的闭包变量，避免 sessionStorage 被覆盖的问题
      windowUserId: null,
      // 标记是否已经发送过离线请求，防止重复发送
      offlineRequestSent: false
    }
  },
  computed: {
    // 检查是否已登录（有token）
    isAuthenticated() {
      return !!getToken()
    },
    // 检查当前是否在登录页面
    isLoginPage() {
      return this.$route.path === '/login'
    }
  },
  async mounted() {
    // 初始化当前用户ID，然后再绑定页面关闭事件
    await this.initCurrentUser()
    this.setupPageUnloadHandler()
    
    // 监听路由变化，确保在登录后重新获取用户ID
    this.$watch('$route', async (to, from) => {
      // 如果从非登录页跳转到登录页，清除用户ID
      if (to.path === '/login') {
        this.currentUserId = null
      }
      // 如果从登录页跳转到其他页面，重新获取用户ID
      else if (from.path === '/login' && this.isAuthenticated && !this.currentUserId) {
        await this.initCurrentUser()
      }
    })
  },
  methods: {
    async initCurrentUser() {
      if (!this.isAuthenticated) {
        this.currentUserId = null
        this.windowUserId = null
        return
      }
      try {
        const response = await getCurrentUser()
        const data = response?.data || response
        // 兼容多种字段名
        this.currentUserId =
          data?.userId || data?.user_id || data?.id || null
        // 同时保存到闭包变量，避免 sessionStorage 被其他标签页覆盖
        this.windowUserId = this.currentUserId
        console.log('App 初始化当前用户ID:', this.currentUserId)
      } catch (error) {
        console.error('初始化当前用户信息失败:', error)
        this.currentUserId = null
        this.windowUserId = null
      }
    },

    setupPageUnloadHandler() {
      // 保存组件实例引用
      const self = this
      
      // 统一的处理函数：更新用户状态为离线
      const handlePageUnload = (eventName) => {
        // 防止重复发送请求
        if (self.offlineRequestSent) {
          console.log(`⚠️ ${eventName}: 离线请求已发送，跳过重复请求`)
          return
        }

        // 获取当前窗口的 token（每个窗口的 localStorage 是独立的）
        const token = getToken()
        if (!token) {
          console.log(`⚠️ ${eventName}: 没有 token，跳过离线更新`)
          return
        }

        // 直接从 token 解析 userId（这是最可靠的方式，不依赖任何共享状态）
        const userId = getUserIdFromToken(token)
        
        console.log(`🔔 ${eventName} 事件触发 [窗口ID: ${window.name || 'unnamed'}]:`, {
          userId: userId,
          tokenPreview: token ? token.substring(0, 20) + '...' : null,
          hasToken: !!token,
          windowName: window.name,
          windowLocation: window.location.href
        })
        
        // 如果成功解析到 userId，更新状态为离线
        if (userId) {
          // 标记已发送，防止重复
          self.offlineRequestSent = true
          
          console.log(`✅ ${eventName}: 准备更新离线状态`, { 
            userId,
            windowId: window.name || 'unnamed'
          })
          
          // 调用离线接口（使用 fetch with keepalive，确保请求在页面关闭后也能完成）
          setUserOffline(userId)
        } else {
          console.log(`⚠️ ${eventName}: 无法从 token 解析 userId，跳过离线更新`)
        }
      }
      
      // 使用 beforeunload 事件监听页面关闭（桌面浏览器）
      window.addEventListener('beforeunload', () => {
        handlePageUnload('beforeunload')
      })

      // 使用 pagehide 事件监听页面隐藏（移动浏览器更可靠，也适用于桌面浏览器）
      // pagehide 事件在 beforeunload 之后触发，但在某些情况下更可靠
      window.addEventListener('pagehide', () => {
        // 只有当页面被持久化缓存时（persisted=true），说明页面被关闭
        // 如果 persisted=false，页面只是被添加到bfcache，用户可能还会回来
        // 但为了保险起见，我们也在 pagehide 时更新状态
        handlePageUnload('pagehide')
      })
    }
  }
}
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: 'Avenir', Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

#app {
  min-height: 100vh;
}

/* 主题滤镜 */

/* 排除所有图片 */
:root[data-theme] img {
  filter: none !important;
}

:root[data-theme="warm"] {
  filter: hue-rotate(-30deg) saturate(1.1);
}

:root[data-theme="dark"] {
  filter: hue-rotate(180deg) invert(1) saturate(0.8);
}


</style>