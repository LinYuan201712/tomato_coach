<template>
  <div class="home-container">
    <!-- 顶部导航栏 -->
    <nav class="navbar">
      <div class="nav-brand">Tomato</div>
      <div class="nav-links">
        <a class="nav-link" @click="goToFriends">好友</a>
        <a class="nav-link" @click="goToKnowledgeBase">知识库</a>
        <a class="nav-link" @click="goToStudyReport">学习报告</a>
        <a class="nav-link" @click="goToTaskManagement">任务管理</a>
        <a class="nav-link" @click="goToProfile">个人中心</a>
      </div>
      <div class="user-avatar-container">
        <div class="user-name">{{ userInfo.username }}</div>
        <div 
          class="user-avatar" 
          @mouseenter="showDropdown = true"
          @mouseleave="handleAvatarLeave"
        >
          <img 
            :src="displayAvatar" 
            alt="用户头像" 
            @error="handleAvatarError"
            @load="handleAvatarLoad"
          />
        </div>
        <!-- 下拉菜单 - 独立元素 -->
        <div 
          v-show="showDropdown" 
          class="dropdown-menu"
          @mouseenter="handleDropdownEnter"
          @mouseleave="handleDropdownLeave"
        >
          <div class="dropdown-item" @click="handleThemeClick">
            <span class="dropdown-icon">🎨</span>
            主题设置
          </div>
          <div class="dropdown-item" @click="logout">
            <span class="dropdown-icon">🚪</span>
            退出登录
          </div>
        </div>
      </div>
    </nav>

    <main class="main-grid">
      <aside class="friends-list-area">
        <div class="friends-list-wrapper">
          <FriendList />
        </div>
        <div class="task-sidebar-wrapper">
          <TaskSidebar @task-status-changed="handleTaskStatusChange" />
        </div>
      </aside>

      <section class="content-area">
        <div class="poster-carousel">
          <div class="poster-slide">
            <img :src="currentPoster" alt="宣传海报" class="poster-image" />
          </div>
          <button class="carousel-arrow left-arrow" @click="prevPoster">‹</button>
          <button class="carousel-arrow right-arrow" @click="nextPoster">›</button>
          
          <div class="carousel-indicators">
            <span 
              v-for="(poster, index) in posters" 
              :key="index"
              :class="['indicator', { active: currentPosterIndex === index }]"
              @click="switchPoster(index)"
            ></span>
          </div>
        </div>

        <!-- 快速加入区域 - 直接嵌入 QuickJoin 组件 -->
        <div class="quick-join-container">
          <QuickJoin 
            :rooms-per-page="4"
            :auto-refresh="false"
            @join-room="handleJoinRoom"
          />
        </div>
      </section>

      <aside class="right-sidebar">
        <!-- 每日签到组件 -->
        <DailyCheckIn @checkin-success="handleCheckInSuccess" />
        
        <div class="sticky-buttons">
          <!-- 如果在房间中，显示返回房间按钮 -->
          <button 
            v-if="userInfo.current_room_id" 
            class="btn-primary return-room-btn" 
            @click="returnToRoom"
          >
            <span class="btn-icon">🏠</span>
            返回当前自习室
          </button>

          <button class="btn-primary" @click="createRoom">
            <span class="btn-icon"></span>
            创建自习室
          </button>

          <button class="btn-primary personal-room-btn" @click="goToPersonalRoom">
            <span class="btn-icon">🎯</span>
            个人自习室
          </button>
          <button class="btn-secondary" @click="joinRoom">
            <span class="btn-icon"></span>
            加入自习室
          </button>
        </div>
      </aside>
    </main>
  </div>
  <!-- 在文件的最后，</template>标签之前添加 -->
  <ThemeModal
    :visible="showThemeModal"
    :current-theme="currentTheme"
    @theme-select="handleThemeSelect"
    @update:visible="showThemeModal = $event"
  />
</template>

<script>
// 只导入头像，海报改为动态导入
import avatarImage from '@/assets/images/avatar.png'
// 添加任务侧边栏组件
import TaskSidebar from '@/components/TaskSidebar/TaskSidebar.vue'
// 导入快速加入组件
import QuickJoin from '@/components/QuickJoin/QuickJoin.vue'
// 【新增】导入 FriendList 组件
import FriendList from '@/components/FriendList/FriendList.vue'
// 导入每日签到组件
import DailyCheckIn from '@/components/DailyCheckIn/DailyCheckIn.vue'
//导入主题设置组件
import ThemeModal from '@/components/ThemeModal/ThemeModal.vue';

export default {
  name: 'HomeView',
  components: {
    // 【新增】注册 FriendList 组件
    FriendList,
    TaskSidebar,
    QuickJoin,
    DailyCheckIn,
    ThemeModal
  },
  data() {
    return {
      // 使用导入的图片
      avatarImage: avatarImage,
      
      // 用户信息
      userInfo: {
        username: '用户',
        current_room_id: null
      },
      
      // 下拉菜单显示状态
      showDropdown: false,
      dropdownTimer: null,
      
      // 海报轮播数据 - 初始为空数组，将在created钩子中动态加载
      posters: [],
      currentPosterIndex: 0,
      carouselTimer: null, // 自动轮播定时器
      
      // 快速加入房间的假数据
      quickJoinRooms: [
        { id: 1, name: '考研数学冲刺', members: 15, status: '专注中' },
        { id: 2, name: '英语阅读小组', members: 8, status: '空闲' },
        { id: 3, name: '深夜代码角', members: 25, status: '专注中' },
        { id: 4, name: '物理学习室', members: 12, status: '空闲' },
        { id: 5, name: '历史讨论组', members: 6, status: '空闲' },
        { id: 6, name: '编程自习班', members: 18, status: '专注中' },
        { id: 7, name: '化学实验室', members: 9, status: '空闲' },
        { id: 8, name: '文学创作间', members: 11, status: '专注中' },
        { id: 9, name: '医学考研组', members: 20, status: '专注中' },
        { id: 10, name: '法律自习室', members: 7, status: '空闲' }
      ],

      // 主题设置相关
      currentTheme: 'default',
      showThemeModal: false,
      showLoginPrompt: false
    }
  },
  computed: {
    // 当前显示的海报
    currentPoster() {
      return this.posters.length > 0 ? this.posters[this.currentPosterIndex] : ''
    },
    // 显示头像（优先使用用户头像，否则使用默认头像）
    displayAvatar() {
      // 确保avatar存在且不为空字符串
      if (this.userInfo && 
          this.userInfo.avatar && 
          typeof this.userInfo.avatar === 'string' &&
          this.userInfo.avatar.trim() !== '' && 
          this.userInfo.avatar !== 'undefined' &&
          this.userInfo.avatar !== 'null' &&
          this.userInfo.avatar !== this.avatarImage) {
        
        let avatarUrl = this.userInfo.avatar.trim()
        
        // 如果是相对路径，添加完整URL
        if (avatarUrl.startsWith('/')) {
          const fullUrl = `http://localhost:8090${avatarUrl}`
          console.log('构建头像URL:', fullUrl)
          return fullUrl
        }
        // 如果已经是完整URL，直接返回
        if (avatarUrl.startsWith('http')) {
          console.log('使用完整头像URL:', avatarUrl)
          return avatarUrl
        }
        // 其他情况，尝试添加基础URL
        console.warn('未知的头像URL格式:', avatarUrl)
        return this.avatarImage
      }
      // 默认头像
      return this.avatarImage
    },
    // 判断用户是否已登录（使用token检查）
      isLoggedIn() {
        const token = localStorage.getItem('token') || sessionStorage.getItem('token');
        return !!token;
      }

  },
  created() {
    // 组件创建时动态加载海报
    this.loadPosters()
    // 获取用户信息（包括头像）
    this.fetchUserInfo()
  },
  
  mounted() {
    // 启动自动轮播
    this.startCarousel()

    // 初始化主题：优先使用本地存储的主题
    const savedTheme = localStorage.getItem('app-theme');
    if (savedTheme) {
      this.applyTheme(savedTheme);
    } else {
      // 检测系统偏好
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
      if (prefersDark) {
        this.applyTheme('dark');
      }
    }
  },
  
  beforeUnmount() {
    // 组件销毁前清除定时器
    this.stopCarousel()
  },
  
  // 当路由激活时，重新获取用户信息（从个人中心返回时刷新头像）
  activated() {
    this.fetchUserInfo()
    // 刷新任务列表（如果TaskSidebar已加载）
    if (this.$refs.taskSidebar) {
      this.$refs.taskSidebar.refreshTasks()
    }
    // 重新启动自动轮播
    this.startCarousel()
  },
  
  deactivated() {
    // 路由失活时停止轮播
    this.stopCarousel()
  },
  methods: {
    // 获取用户信息
    async fetchUserInfo() {
      try {
        const { getCurrentUser } = await import('@/api/user')
        const result = await getCurrentUser()
        
        console.log('获取用户信息响应:', result)
        
        if (result.success && result.data) {
          console.log('解析用户信息数据:', result.data)
          
          // 整合所有后端返回的数据
          this.userInfo = {
            ...this.userInfo,
            ...result.data,
            avatar: null, // 先置空，稍后精细化处理
            username: result.data.username || '用户',
            current_room_id: result.data.current_room_id || null
          }
          
          console.log('更新后的 userInfo:', JSON.stringify(this.userInfo))
          
          // 确保avatar是有效值，空字符串、undefined、null都设置为null
          const avatar = result.data.avatar
          console.log('原始avatar值:', avatar, '类型:', typeof avatar)
          
          if (avatar && typeof avatar === 'string' && avatar.trim() !== '' && avatar !== 'undefined' && avatar !== 'null') {
            this.userInfo.avatar = avatar.trim()
            console.log('设置头像URL:', this.userInfo.avatar)
          } else {
            this.userInfo.avatar = null
            console.log('头像为空或无效，使用默认头像')
          }
        } else {
          // 如果获取失败，使用默认值
          console.warn('获取用户信息失败，result.success为false')
          this.userInfo.avatar = null
          this.userInfo.username = '用户'
        }
      } catch (error) {
        // 如果是401未授权错误（token过期），request.js已经处理了跳转，这里静默处理
        if (error.isUnauthorized || error.status === 401) {
          console.warn('Token已过期，已自动跳转到登录页')
          // 清除用户信息，避免显示错误
          this.userInfo.avatar = null
          this.userInfo.username = '用户'
          return
        }
        
        // 其他错误也静默处理，不显示错误信息
        console.error('获取用户信息失败:', error)
        this.userInfo.avatar = null
        this.userInfo.username = '用户'
      }
    },
    
    // 处理头像加载成功
    handleAvatarLoad(event) {
      // 头像加载成功，清除默认标记
      event.target.dataset.isDefault = 'false'
    },
    
    // 处理头像加载错误
    handleAvatarError(event) {
      const failedSrc = event.target.src
      console.warn('头像加载失败:', failedSrc)
      
      // 防止无限循环：如果已经是默认头像还失败，就不再处理
      if (failedSrc === this.avatarImage || 
          failedSrc.includes('avatar.png') ||
          event.target.dataset.isDefault === 'true') {
        console.warn('默认头像也加载失败，停止处理')
        // 移除错误处理器，防止无限循环
        event.target.onerror = null
        return
      }
      
      // 标记为默认头像，防止再次触发错误
      event.target.dataset.isDefault = 'true'
      
      // 清除无效的头像URL，设置为null，让计算属性返回默认头像
      this.userInfo.avatar = null
      
      // 直接设置默认头像
      event.target.src = this.avatarImage
    },
    
    // 动态加载海报图片
    async loadPosters() {
      try {
        // 海报文件数量 - 根据你的文件列表，有6个海报
        const posterCount = 6
        
        // 使用 Promise.all 并行加载所有海报
        const posterPromises = []
        
        for (let i = 1; i <= posterCount; i++) {
          // 动态导入海报图片
          const posterPromise = import(`@/assets/images/poster${i}.jpg`)
            .then(module => module.default)
            .catch(error => {
              console.warn(`无法加载海报 poster${i}.jpg:`, error)
              return null
            })
          posterPromises.push(posterPromise)
        }
        
        // 等待所有图片加载完成
        const loadedPosters = await Promise.all(posterPromises)
        
        // 过滤掉加载失败的图片
        this.posters = loadedPosters.filter(poster => poster !== null)
        
        console.log(`成功加载 ${this.posters.length} 张海报`)
        
        // 如果海报加载完成且组件已挂载，启动自动轮播
        if (this.posters.length > 1 && this.$el) {
          this.startCarousel()
        }
        
      } catch (error) {
        console.error('加载海报时出错:', error)
        this.posters = [] // 确保posters始终是数组
      }
    },
    
    // 鼠标从头像移出
    handleAvatarLeave() {
      // 短暂延迟，让用户有时间移动到下拉菜单
      this.dropdownTimer = setTimeout(() => {
        this.showDropdown = false
      }, 150)
    },
    
    // 鼠标进入下拉菜单
    handleDropdownEnter() {
      // 取消隐藏计时器
      if (this.dropdownTimer) {
        clearTimeout(this.dropdownTimer)
      }
    },
    
    // 鼠标从下拉菜单移出
    handleDropdownLeave() {
      // 立即隐藏下拉菜单
      this.showDropdown = false
    },
    
    // 切换主题
    toggleTheme() {
      alert('主题设置功能待实现')
      this.showDropdown = false
    },
    
    // 退出登录 - 调用后端API并跳转到登录页面
    async logout() {
      if (!confirm('确定要退出登录吗？')) {
        return
      }
      
      try {
        // 调用退出登录API（logout函数内部会清除token）
        const { logout } = await import('@/api/auth')
        await logout()
      } catch (error) {
        console.error('退出登录失败:', error)
        // 即使API调用失败，也清除本地token
        const { removeToken } = await import('@/api/config')
        removeToken()
      } finally {
        // 跳转到登录页面
        this.$router.push('/login')
      }
    },
    
    // 跳转到好友界面（预留）
    goToFriends() {
      this.$router.push('/friends')
    },
    
    // 跳转到知识库界面
    goToKnowledgeBase() {
      this.$router.push('/knowledge-base')
    },
    
    // 跳转到任务管理界面
    goToTaskManagement() {
      this.$router.push('/task-management')
    },
    
    // 跳转到学习报告界面
    goToStudyReport() {
      this.$router.push('/study-report')
    },
    
    // 跳转到个人中心（预留）
    goToJoinRoom() {
      this.$router.push('/join-room')
    },
    goToPersonalRoom() {
      this.$router.push('/study-room/personal')
    },
    goToProfile() {
      //alert('个人中心功能正在开发中...')
      this.$router.push('/personal-center') // 预留跳转逻辑
    },
    
    // 海报轮播方法
    nextPoster() {
      if (this.posters.length === 0) return
      this.currentPosterIndex = (this.currentPosterIndex + 1) % this.posters.length
      // 重置自动轮播定时器
      this.resetCarousel()
    },
    prevPoster() {
      if (this.posters.length === 0) return
      this.currentPosterIndex = (this.currentPosterIndex - 1 + this.posters.length) % this.posters.length
      // 重置自动轮播定时器
      this.resetCarousel()
    },
    switchPoster(index) {
      if (this.posters.length === 0) return
      this.currentPosterIndex = index
      // 重置自动轮播定时器
      this.resetCarousel()
    },
    
    // 启动自动轮播
    startCarousel() {
      if (this.posters.length <= 1) return // 只有一张或没有海报时不启动
      this.stopCarousel() // 先清除可能存在的定时器
      this.carouselTimer = setInterval(() => {
        this.nextPoster()
      }, 4000) // 每4秒切换一次
    },
    
    // 停止自动轮播
    stopCarousel() {
      if (this.carouselTimer) {
        clearInterval(this.carouselTimer)
        this.carouselTimer = null
      }
    },
    
    // 重置自动轮播（手动切换后重新计时）
    resetCarousel() {
      this.stopCarousel()
      this.startCarousel()
    },
    
    // 创建自习室
    createRoom() {
      this.$router.push('/create-room')
    },
    
    // 加入自习室
    joinRoom() {
      this.$router.push('/join-room')
    },
    
    // 返回当前所在的自习室
    returnToRoom() {
      if (this.userInfo.current_room_id) {
        this.$router.push({
          name: 'study-room',
          params: { roomId: this.userInfo.current_room_id }
        })
      }
    },
    
    quickJoin(roomId) {
      alert(`快速加入房间 ${roomId} - 功能待实现`)
    },

    handleTaskStatusChange(data) {
      console.log('任务状态改变:', data)
      // 可以在这里处理任务状态改变的逻辑
      // 例如：显示提示、更新其他数据等
    },

    async handleJoinRoom(roomId) {
      console.log('快速加入房间:', roomId)
      
      try {
        // 先获取用户信息
        const { getCurrentUser } = await import('@/api/user')
        const userResponse = await getCurrentUser()
        
        if (!userResponse.success || !userResponse.data) {
          alert('获取用户信息失败，请重新登录')
          return
        }
        
        const currentUser = userResponse.data
        const userId = currentUser.id || currentUser.userId || currentUser.user_id
        
        if (!userId) {
          alert('用户身份验证失败，请重新登录')
          return
        }
        
        // 将用户ID转换为数字
        const userIdNumber = Number(userId)
        if (isNaN(userIdNumber)) {
          console.error('用户ID不是有效的数字:', userId)
          alert('用户ID格式错误')
          return
        }
        
        console.log(`调用加入房间API: roomId=${roomId}, userId=${userIdNumber}`)
        
        // 调用加入房间API
        const { joinRoom } = await import('@/api/studyRooms')
        const joinResult = await joinRoom(roomId, userIdNumber)
        
        console.log('加入房间响应:', joinResult)
        
        // 检查加入结果
        if (joinResult && (joinResult.code === 200 || joinResult.success === true)) {
          console.log('✅ 加入房间成功，跳转到自习室页面')
          // 加入成功，跳转到自习室页面
          this.$router.push({
            name: 'study-room',
            params: { roomId: roomId }
          })
        } else {
          // 如果已经在房间中（409），也允许跳转
          if (joinResult?.message?.includes('已在') || joinResult?.code === 409) {
            console.log('用户已在房间中，直接跳转')
            this.$router.push({
              name: 'study-room',
              params: { roomId: roomId }
            })
          } else {
            const errorMsg = joinResult?.message || '加入失败，请重试'
            alert(`加入失败: ${errorMsg}`)
          }
        }
      } catch (error) {
        console.error('快速加入房间失败:', error)
        
        // 如果是409（已在房间中），允许跳转
        if (error.status === 409 || error.message?.includes('已在')) {
          console.log('用户已在房间中，直接跳转')
          this.$router.push({
            name: 'study-room',
            params: { roomId: roomId }
          })
        } else {
          alert(`加入失败: ${error.message || '网络错误，请稍后重试'}`)
        }
      }
    },

    handleCheckInSuccess(data) {
      console.log('签到成功:', data)
      // 可以在这里显示成功提示或更新其他数据
      // 例如：更新用户资产信息、显示通知等
    },

    // 处理主题设置点击
  handleThemeClick() {
    // 关闭下拉菜单
    this.showDropdown = false;
    
    // 检查登录状态
    if (!this.isLoggedIn) {
      this.showLoginRequired();
      return;
    }
    
    // 已登录，打开主题弹窗
    this.openThemeModal();
  },
  
  // 打开主题设置弹窗
  openThemeModal() {
    this.showThemeModal = true;
  },
  
  // 关闭主题设置弹窗
  closeThemeModal() {
    this.showThemeModal = false;
  },
  
  // 显示登录提示
  showLoginRequired() {
    if (confirm('主题设置需要登录后才能使用\n\n是否立即前往登录？')) {
      this.$router.push('/login');
    }
  },
  
  // 应用主题
  applyTheme(themeId) {
    const html = document.documentElement;
    
    // 移除所有主题属性
    html.removeAttribute('data-theme');
    
    if (themeId !== 'default') {
      html.setAttribute('data-theme', themeId);
    }
    
    // 保存到本地存储
    localStorage.setItem('app-theme', themeId);
    this.currentTheme = themeId;
  },
  
  // 处理主题选择
  handleThemeSelect(themeId) {
    this.applyTheme(themeId);
  }
}
}
</script>

<style scoped>
/* 整体容器 */
.home-container {
  min-height: 100vh;
  background-color: #fefaf5; /* 浅橘黄色背景 */
}

/* 顶部导航栏 */
.navbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  background: white;
  border-bottom: 1px solid #ffe4cc; /* 橘黄色边框 */
  position: sticky;
  top: 0;
  z-index: 100;
}

.nav-brand {
  font-size: 1.5em;
  font-weight: bold;
  color: #eeaa67; /* 橘黄色品牌色 */
}

.nav-links {
  display: flex;
  gap: 30px;
}

.nav-link {
  cursor: pointer;
  padding: 8px 12px;
  border-radius: 6px;
  transition: background-color 0.2s;
  color: #333;
}

.nav-link:hover {
  background-color: #fff5eb; /* 浅橘黄色悬停背景 */
  color: #eeaa67; /* 橘黄色文字 */
}

/* 用户头像 */
.user-avatar:hover img {
  border-color: #eeaa67; /* 橘黄色边框 */
  transform: scale(1.05);
}

/* 下拉菜单样式 */
.dropdown-item:hover {
  background-color: #fff5eb; /* 浅橘黄色悬停背景 */
}

/* ============================
   主要内容网格区域 (修改 grid-template-columns)
   ============================ */
.main-grid {
  display: grid;
  /* 调整列宽：左侧好友列表 280px，中央内容 1fr，右侧边栏 300px */
  grid-template-columns: 280px 1fr 300px; 
  gap: 20px;
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
}

/* ============================
   左侧好友列表区域 (原 widgets-area)
   ============================ */
/* 统一类名并去除原有占位符样式 */
.friends-list-area { 
  /* 继承 FriendList.vue 中的样式 */
  display: flex;
  flex-direction: column;
  gap: 15px;
  height: fit-content;
  position: sticky;
  top: 100px;
  max-height: calc(100vh - 120px);
  overflow-y: auto;
}

.friends-list-wrapper {
  flex: 0 0 auto;
  max-height: 40vh; /* 限制好友列表最大高度为视口高度的40% */
  min-height: 200px;
  overflow-y: auto;
}

.task-sidebar-wrapper {
  flex: 1 1 auto; /* 待办任务占据剩余空间 */
  min-height: 300px;
  overflow-y: auto;
}

/* 移除原有的 widget-placeholder 样式，现在由 FriendList 组件内部处理 */

/* 海报轮播 */
.poster-carousel {
  position: relative;
  background: white;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 4px 8px rgba(238, 170, 103, 0.1); /* 橘黄色阴影 */
  border: 1px solid #ffe4cc; /* 橘黄色边框 */
  height: 400px; /* 固定高度 */
}

.poster-slide {
  width: 100%;
  height: 100%;
  position: relative;
  overflow: hidden;
}

.poster-image {
  width: 100%;
  height: 100%;
  object-fit: cover; /* 填满容器，保持比例 */
  display: block;
}

/* 快速加入区域 */
.quick-join {
  background: white;
  padding: 20px;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(238, 170, 103, 0.1); /* 橘黄色阴影 */
  border: 1px solid #ffe4cc; /* 橘黄色边框 */
}

.quick-join h3 {
  margin: 0 0 15px 0;
  color: #333;
}

/* 右侧边栏按钮 */
.btn-primary {
  background: linear-gradient(135deg, #eeaa67, #f5b877); /* 橘黄色渐变 */
  color: white;
}

.btn-secondary {
  background: white;
  color: #eeaa67; /* 橘黄色文字 */
  border: 2px solid #eeaa67; /* 橘黄色边框 */
}

.return-room-btn {
  background: linear-gradient(135deg, #66bb6a, #81c784) !important;
  box-shadow: 0 4px 15px rgba(102, 187, 106, 0.3) !important;
  margin-bottom: 10px;
}

.return-room-btn:hover {
  box-shadow: 0 8px 25px rgba(102, 187, 106, 0.4) !important;
}

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(238, 170, 103, 0.4); /* 橘黄色阴影 */
}

.btn-secondary:hover {
  background: #eeaa67; /* 橘黄色背景 */
  color: white;
}

.sidebar-placeholder {
  background: white;
  padding: 20px;
  border-radius: 10px;
  box-shadow: 0 2px 4px rgba(238, 170, 103, 0.1); /* 橘黄色阴影 */
  text-align: center;
  border: 1px solid #ffe4cc; /* 橘黄色边框 */
}

/* 其他样式保持不变 (确保没有多余的旧样式残留) */
/* ... (原 HomeView.vue 中的其他样式保持不变) ... */

.user-avatar-container {
  position: relative;
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-name {
  font-size: 14px;
  color: #333;
  font-weight: 500;
  cursor: default;
}

.user-avatar {
  cursor: pointer;
}

.user-avatar img {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  border: 2px solid #e0e0e0;
  object-fit: cover;
  transition: all 0.3s ease;
}

.dropdown-menu {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: 8px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  min-width: 150px;
  z-index: 1000;
  border: 1px solid #e0e0e0;
  overflow: hidden;
  animation: dropdownFade 0.2s ease;
}

@keyframes dropdownFade {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.dropdown-item {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  cursor: pointer;
  transition: background-color 0.2s;
  border-bottom: 1px solid #f0f0f0;
}

.dropdown-item:last-child {
  border-bottom: none;
}

.dropdown-icon {
  margin-right: 10px;
  font-size: 1.1em;
}

.main-grid {
  display: grid;
  grid-template-columns: 250px 1fr 300px;
  gap: 20px;
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
}

.widgets-area {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.content-area {
  display: flex;
  flex-direction: column;
  gap: 20px;
}


.carousel-arrow {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  background: rgba(255, 255, 255, 0.6); /* 半透明白色背景 */
  backdrop-filter: blur(4px); /* 毛玻璃效果 */
  color: #333;
  border: none;
  width: 45px;
  height: 45px;
  border-radius: 50%;
  cursor: pointer;
  font-size: 1.8em;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s ease;
  z-index: 10;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.left-arrow {
  left: 20px;
}

.right-arrow {
  right: 20px;
}

.carousel-arrow:hover {
  background: rgba(255, 255, 255, 0.9); /* 悬停时更不透明 */
  transform: translateY(-50%) scale(1.1);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.25);
}

.carousel-arrow:active {
  transform: translateY(-50%) scale(0.95);
}

.carousel-indicators {
  position: absolute;
  bottom: 15px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  gap: 8px;
}

.indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.5);
  cursor: pointer;
  transition: all 0.3s ease;
}

.indicator.active {
  background: white;
  transform: scale(1.2);
}

.indicator:hover {
  background: rgba(255, 255, 255, 0.8);
}

.quick-join h3 {
  margin: 0 0 15px 0;
  color: #333;
}

.quick-join-list {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
  max-height: 400px;
  overflow-y: auto;
  padding-right: 8px;
}

.quick-join-list::-webkit-scrollbar {
  width: 6px;
}

.quick-join-list::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

.quick-join-list::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.quick-join-list::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

.quick-join-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  transition: all 0.2s;
}

.quick-join-item:hover {
  border-color: #eeaa67; /* 橘黄色边框 */
  background: #fffaf5; /* 浅橘黄色背景 */
}

.room-avatar {
  font-size: 1.5em;
}

.room-info {
  flex-grow: 1;
}

.room-name {
  font-weight: bold;
  margin-bottom: 4px;
}

.room-stats {
  font-size: 0.8em;
  color: #666;
}

.join-btn {
  padding: 6px 12px;
  background: #eeaa67; /* 橘黄色按钮 */
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9em;
  transition: background-color 0.2s;
}

.join-btn:hover {
  background: #e69c55; /* 深橘黄色 */
}

.right-sidebar {
  position: sticky;
  top: 100px;
  display: flex;
  flex-direction: column;
  gap: 20px;
  z-index: 90;
  height: fit-content;
}

.sticky-buttons {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.sticky-buttons button {
  padding: 15px;
  border: none;
  border-radius: 10px;
  font-size: 1.1em;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

/* 响应式设计 (更新 grid-template-columns) */
@media (max-width: 1024px) {
  /* 调整为左侧 250px，中央 1fr */
  .main-grid {
    grid-template-columns: 250px 1fr;
  }
  .right-sidebar {
    display: none;
  }
}

@media (max-width: 768px) {
  .main-grid {
    grid-template-columns: 1fr;
    gap: 15px;
  }
  .friends-list-area {
    display: none; /* 在小屏上隐藏左侧好友列表 */
  }
  .widgets-area {
    display: none;
  }
}

.personal-room-btn { background: linear-gradient(135deg, #6366f1, #8b5cf6) !important; color: white !important; }
.personal-room-btn:hover { background: linear-gradient(135deg, #4f46e5, #7c3aed) !important; box-shadow: 0 4px 15px rgba(99, 102, 241, 0.4) !important; }
</style>
.personal-room-btn { background: linear-gradient(135deg, #6366f1, #8b5cf6) !important; color: white !important; }
.personal-room-btn:hover { background: linear-gradient(135deg, #4f46e5, #7c3aed) !important; box-shadow: 0 4px 15px rgba(99, 102, 241, 0.4) !important; }
