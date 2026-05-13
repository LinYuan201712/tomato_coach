<template>
  <div class="personal-center">
    <nav class="top-navbar">
      <div class="nav-back" @click="goBack">
        <span class="back-arrow">←</span>
        返回首页
      </div>
      <div class="nav-title">个人中心</div>
      <div class="nav-placeholder"></div>
    </nav>

    <div class="main-content">
      <aside class="sidebar">
        <div class="user-info">
          <div class="avatar-container">
            <img 
              :src="getAvatarUrl(userProfileData.avatar)" 
              alt="用户头像" 
              class="avatar" 
              @error="handleSidebarAvatarError"
            />
          </div>
          <div class="user-id">ID: {{ userProfileData.userId || '加载中...' }}</div>
        </div>

        <nav class="side-nav">
          <div
            v-for="item in navItems"
            :key="item.key"
            :class="['nav-item', { active: activeNav === item.key }]"
            @click="switchNav(item.key)"
          >
            <span class="nav-text">{{ item.text }}</span>
          </div>
        </nav>
      </aside>

      <main class="content-area">
        <div class="content-header">
          <h2>{{ currentNavText }}</h2>
          <div v-if="loading" class="loading-indicator">数据加载中...</div>
          <div v-if="error" class="error-indicator">{{ error }}</div>
        </div>
        
        <div class="content-body">
          <keep-alive>
            <component 
              :is="activeNav === 'profile' ? 'UserProfile' : 
                   activeNav === 'statistics' ? 'UserStatistics' : 
                   activeNav === 'knowledge' ? 'KnowledgeBase' :
                   activeNav === 'channels' ? 'ChannelSettings' :
                   'UserSettings'"
              :user-info="userProfileData"
              @update-user-info="updateSidebarInfo"
              class="tab-content" 
            />
          </keep-alive>
        </div>
      </main>
    </div>
  </div>
</template>

<script>
import avatarImage from '@/assets/images/avatar.png'
import UserProfile from '@/components/PersonalCenter/UserProfile.vue';
import UserSettings from '@/components/PersonalCenter/UserSettings.vue';
import UserStatistics from '@/components/PersonalCenter/UserStatistics.vue';
import KnowledgeBase from '@/components/PersonalCenter/KnowledgeBase.vue';
import ChannelSettings from '@/components/PersonalCenter/ChannelSettings.vue';
import { getCurrentUser } from '@/api/user';

export default {
  name: 'PersonalCenter',

  components: {
    UserProfile,
    UserStatistics,
    UserSettings,
    KnowledgeBase,
    ChannelSettings
  },

  data() {
    return {
      avatarImage: avatarImage,
      activeNav: 'profile', // 默认选中个人资料
      navItems: [
        { key: 'profile', text: '个人资料' },
        { key: 'statistics', text: '专注统计' },
        { key: 'knowledge', text: '我的知识库' },
        { key: 'channels', text: '渠道集成' },
        { key: 'settings', text: '隐私设置' }
      ],
      loading: false,
      error: '',
      userProfileData: {
        userId: '加载中...', 
        username: '加载中...',
        email: '',
        phone: '',
        gender: '', // 对应 API 的 sex
        birthday: '',
        region: '', // 对应 API 的 province
        tomatoCount: 0, // 对应 API 的 tomato
        avatar: avatarImage, 
        password_hash: '', 
      }
    }
  },
  computed: {
    currentNavText() {
      const item = this.navItems.find(item => item.key === this.activeNav)
      return item ? item.text : '个人中心'
    }
  },
  created() {
    this.fetchUserInfo();
  },
  methods: {
    // 获取头像URL（处理相对路径和默认头像）
    getAvatarUrl(avatar) {
      if (!avatar) {
        return this.avatarImage
      }
      
      // 如果是data URL，直接返回
      if (avatar.startsWith('data:')) {
        return avatar
      }
      
      // 如果已经是完整URL，直接返回
      if (avatar.startsWith('http://') || avatar.startsWith('https://')) {
        return avatar
      }
      
      // 如果是相对路径，转换为完整URL
      if (avatar.startsWith('/')) {
        return `http://localhost:8090${avatar}`
      }
      
      // 其他情况返回默认头像
      return this.avatarImage
    },
    
    // 处理左侧头像加载错误
    handleSidebarAvatarError(event) {
      // 如果加载失败，使用默认头像
      if (!event.target.src.includes('avatar.png')) {
        event.target.src = this.avatarImage
      }
    },
    
    // 格式化生日：将时间戳（毫秒）转换为 YYYY-MM-DD 格式
    formatBirthday(timestamp) {
      if (!timestamp) {
        return ''
      }
      const date = new Date(timestamp)
      // 检查日期是否有效
      if (isNaN(date.getTime())) {
        return ''
      }
      const year = date.getFullYear()
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      return `${year}-${month}-${day}`
    },
    
    // 格式化用户信息数据
    formatUserData(data) {
      return {
        userId: data.userId ? String(data.userId) : (data.user_id ? String(data.user_id) : 'N/A'),
        username: data.username || 'N/A',
        email: data.email || '',
        phone: data.phone || '',
        gender: data.sex || '',
        // 生日：后端返回的是时间戳（毫秒），需要转换为 YYYY-MM-DD 格式
        birthday: data.birthday ? (typeof data.birthday === 'number' 
          ? this.formatBirthday(data.birthday) 
          : data.birthday) : '',
        region: data.province || '',
        tomatoCount: data.tomato || 0,
        // 优先使用服务器返回的头像URL
        // 如果服务器返回了头像URL，使用它；否则使用默认头像
        avatar: data.avatar || null, // 存储原始URL，让组件处理URL转换
        password_hash: data.passwordHash || data.password_hash || '',
      }
    },
    
    // 从 /user/me 获取用户信息
    async fetchUserInfo() {
      // 只有在没有正在进行加载时才显示 loading，或强制显示 loading
      if (!this.loading) {
        this.loading = true;
      }
      this.error = ''
      
      try {
        const result = await getCurrentUser()
        
        if (result.success && result.data) {
          // 确保整个 userProfileData 被最新的服务器数据覆盖
          this.userProfileData = this.formatUserData(result.data)
        } else {
          this.error = result.message || '加载用户信息失败。'
          console.error('API Error:', result)
        }
      } catch (err) {
        this.error = err.message || '网络请求失败，请检查服务。'
        console.error('Fetch Error:', err)
      } finally {
        this.loading = false
      }
    },

    // 接收来自 UserProfile 组件的更新事件。
    // 在子组件保存成功后被调用，用于重新获取最新数据。
    updateSidebarInfo() {
      // 重新调用 fetchUserInfo 来获取最新的用户信息，实现数据刷新
      this.fetchUserInfo();
    },
    
    goBack() {
      this.$router.push('/home')
    },
    
    switchNav(navKey) {
      if (navKey === 'knowledge') {
        this.$router.push('/knowledge-base')
        return
      }
      this.activeNav = navKey
    }
  }
}
</script>

<style scoped>
/* 样式保持不变 */
.personal-center {
  min-height: 100vh;
  background-color: #f8f9fa;
}

/* 顶部导航栏 */
.top-navbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  background: white;
  border-bottom: 1px solid #e0e0e0;
  position: sticky;
  top: 0;
  z-index: 100;
}

.nav-back {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  color: #000000;
  font-weight: 500;
  padding: 8px 12px;
  border-radius: 6px;
  transition: background-color 0.2s;
}

.nav-back:hover {
  background-color: #f0f0f0;
}

.back-arrow {
  font-size: 1.2em;
}

.nav-title {
  font-size: 1.3em;
  font-weight: bold;
  color: #333;
}

.nav-placeholder {
  width: 80px;
}

/* 主要内容布局 */
.main-content {
  display: flex;
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
  gap: 20px;
  min-height: calc(100vh - 80px);
}

/* 左侧边栏 */
.sidebar {
  width: 280px;
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  height: fit-content;
  position: sticky;
  top: 100px;
}

/* 用户信息 */
.user-info {
  text-align: center;
  padding-bottom: 20px;
  border-bottom: 1px solid #f0f0f0;
  margin-bottom: 20px;
}

.avatar-container {
  margin-bottom: 12px;
}

.avatar {
  width: 100px;
  height: 100px;
  border-radius: 50%;
  border: 3px solid #cc2a1f;
  object-fit: cover;
  background: #f8f9fa;
}

.user-id {
  color: #666;
  font-size: 0.9em;
  font-weight: 500;
}

/* 侧边导航 */
.side-nav {
  display: flex;
  flex-direction: column;
  gap: 4px;
  align-items: center;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  border: none;
  background: none;
  text-align: center;
  justify-content: center;
  width: 100%;
  max-width: 200px;
}

.nav-item:hover {
  background: #f8f9ff;
  color: #cc2a1f;
}

.nav-item.active {
  background: #cc2a1f;
  color: white;
}

.nav-text {
  font-weight: 500;
  text-align: center;
  width: 100%;
}

/* 右侧内容区域 */
.content-area {
  flex: 1;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  overflow: hidden;
}

.content-header {
  padding: 20px 24px;
  border-bottom: 1px solid #f0f0f0;
  background: #fafafa;
}

.content-header h2 {
  margin: 0;
  color: #333;
  font-size: 1.5em;
}

.content-body {
  padding: 30px;
  min-height: 400px;
}

/* 标签内容 */
.tab-content {
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.loading-indicator, .error-indicator {
  padding: 10px 0;
  text-align: center;
  font-weight: 500;
}
.loading-indicator {
  color: #007bff;
}
.error-indicator {
  color: #cc2a1f;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .main-content {
    flex-direction: column;
    padding: 15px;
  }
  
  .sidebar {
    width: 100%;
    position: static;
  }
  
  .content-body {
    padding: 20px;
  }
  
  .avatar {
    width: 80px;
    height: 80px;
  }

  .fade-enter-active, .fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter, .fade-leave-to {
  opacity: 0;
}
}
</style>