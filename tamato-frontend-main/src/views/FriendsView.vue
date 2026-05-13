<template>
  <div class="friends-view">
    <!-- 顶部导航栏 -->
    <nav class="navbar">
      <div class="nav-brand">Tomato</div>
      <div class="nav-links">
        <a class="nav-link" @click="goToHome">返回首页</a>
      </div>
    </nav>

    <main class="main-content">
      <div class="friends-layout">
        <!-- 左侧：导航栏和好友列表 -->
        <aside class="left-sidebar">
          <!-- 导航栏 -->
          <div class="sidebar-nav">
            <h2 class="nav-title">好友管理</h2>
            <div class="nav-menu">
              <button 
                :class="['nav-item', { active: activeTab === 'friends' }]"
                @click="activeTab = 'friends'"
              >
                <span class="nav-icon">👥</span>
                我的好友
              </button>
              <button 
                :class="['nav-item', { active: activeTab === 'search' }]"
                @click="activeTab = 'search'"
              >
                <span class="nav-icon">🔍</span>
                搜索用户
              </button>
              <button 
                :class="['nav-item', { active: activeTab === 'requests' }]"
                @click="activeTab = 'requests'"
              >
                <span class="nav-icon">📬</span>
                好友申请
                <span v-if="friendRequests.length > 0" class="nav-badge">{{ friendRequests.length }}</span>
              </button>
            </div>
          </div>

          <!-- 好友列表 -->
          <div class="friends-list-container">
            <div class="friends-list-header">
              <h3>我的好友 ({{ friendsList.length }})</h3>
              <button @click="refreshFriends" class="refresh-btn" :disabled="loadingFriends" title="刷新">
                <span v-if="loadingFriends">⏳</span>
                <span v-else>🔄</span>
              </button>
            </div>
            
            <div v-if="loadingFriends" class="loading-state">
              <div class="spinner"></div>
              <p>加载中...</p>
            </div>
            <div v-else-if="friendsList.length === 0" class="empty-state">
              <div class="empty-icon">👥</div>
              <p>暂无好友</p>
            </div>
            <div v-else class="friends-list-scroll">
              <div 
                v-for="friend in friendsList" 
                :key="friend.friend_id || friend.id"
                class="friend-card-expanded"
                :class="{ 'selected': selectedFriend && selectedFriend.friend_id === friend.friend_id }"
                @click="selectFriend(friend)"
              >
                <div class="friend-avatar-wrapper">
                  <div class="friend-avatar">
                    <img 
                      :src="getAvatarUrl(friend.friend_avatar || friend.avatar)" 
                      alt="好友头像"
                      @error="handleAvatarError"
                    />
                  </div>
                  <span :class="['online-indicator', friend.isOnline ? 'online' : 'offline']"></span>
                </div>
                
                <div class="friend-details">
                  <h4 class="friend-name">{{ friend.friend_name || friend.friend_username || friend.username }}</h4>
                  
                  <div class="friend-status-info">
                    <span :class="['status-badge', getStatusClass(friend.friend_status || friend.status)]">
                      {{ friend.friend_status || friend.status || '离线' }}
                    </span>
                    <span v-if="friend.tomatoStatus" :class="['tomato-status', friend.tomatoStatus]">
                      {{ getTomatoStatusText(friend.tomatoStatus) }}
                    </span>
                  </div>
                  
                  <div v-if="friend.currentTask" class="current-task-info">
                    <span class="task-icon">📚</span>
                    <span class="task-text">{{ friend.currentTask }}</span>
                  </div>
                </div>
                
                <div class="friend-actions">
                  <button 
                    @click.stop="showDeleteConfirm(friend)"
                    class="btn-delete-friend"
                    title="删除好友"
                  >
                    🗑️
                  </button>
                </div>
              </div>
            </div>
          </div>
        </aside>

        <!-- 右侧：主要内容区域 -->
        <div class="right-content">
          <!-- 搜索用户 -->
          <div v-if="activeTab === 'search'" class="content-section">
            <h2 class="section-title">搜索用户</h2>
            <div class="search-box">
              <input 
                type="text" 
                v-model="searchUsername" 
                placeholder="请输入用户名"
                class="search-input"
                @keyup.enter="handleSearch"
              >
              <button @click="handleSearch" class="search-btn" :disabled="searching">
                {{ searching ? '搜索中...' : '搜索' }}
              </button>
            </div>
            
            <!-- 搜索结果 -->
            <div v-if="searchResult" class="search-result">
              <div class="user-card">
                <div class="user-avatar">
                  <img 
                    :src="getAvatarUrl(searchResult.avatar)" 
                    alt="用户头像"
                    @error="handleAvatarError"
                  />
                </div>
                <div class="user-info">
                  <h3 class="user-name">{{ searchResult.username }}</h3>
                  <p class="user-id">ID: {{ searchResult.user_id }}</p>
                  <div class="user-status">
                    <span :class="['status-dot', getStatusClass(searchResult.status)]"></span>
                    <span>{{ searchResult.status || '未知' }}</span>
                  </div>
                </div>
                <div class="user-actions">
                  <button 
                    v-if="!isFriend(searchResult.user_id) && !hasPendingRequest(searchResult.user_id)"
                    @click="showSendRequestModal(searchResult)"
                    class="btn-add-friend"
                  >
                    添加好友
                  </button>
                  <span v-else-if="isFriend(searchResult.user_id)" class="friend-badge">已是好友</span>
                  <span v-else class="pending-badge">申请中</span>
                </div>
              </div>
            </div>
            
            <!-- 搜索错误提示 -->
            <div v-if="searchError" class="error-message">
              {{ searchError }}
            </div>
          </div>

          <!-- 好友申请 -->
          <div v-if="activeTab === 'requests'" class="content-section">
            <h2 class="section-title">好友申请</h2>
            <div v-if="loadingRequests" class="loading-state">
              <div class="spinner"></div>
              <p>加载中...</p>
            </div>
            <div v-else-if="friendRequests.length === 0" class="empty-state">
              <div class="empty-icon">📬</div>
              <p>暂无好友申请</p>
            </div>
            <div v-else class="requests-list">
              <div 
                v-for="request in friendRequests" 
                :key="request.id"
                class="request-item"
              >
                <div class="request-avatar">
                  <img 
                    :src="getAvatarUrl(request.from_user_avatar || request.avatar)" 
                    alt="头像"
                    @error="handleAvatarError"
                  />
                </div>
                <div class="request-info">
                  <h4 class="request-name">{{ request.from_user_name || request.from_username }}</h4>
                  <p class="request-message" v-if="request.message">{{ request.message }}</p>
                  <p class="request-time">{{ formatTime(request.created_at || request.createdAt) }}</p>
                </div>
                <div class="request-actions">
                  <button 
                    @click="handleRequest(request, 'accept')"
                    class="btn-accept"
                    :disabled="request.status !== '待处理'"
                  >
                    接受
                  </button>
                  <button 
                    @click="handleRequest(request, 'reject')"
                    class="btn-reject"
                    :disabled="request.status !== '待处理'"
                  >
                    拒绝
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- 好友主页 -->
          <div v-if="activeTab === 'friends' && selectedFriend" class="content-section">
            <!-- 好友主页头部 -->
            <div class="friend-profile-header">
              <div class="profile-avatar">
                <img 
                  :src="getAvatarUrl(selectedFriend.friend_avatar || selectedFriend.avatar)" 
                  alt="好友头像"
                  @error="handleAvatarError"
                />
                <span :class="['online-indicator-large', selectedFriend.isOnline ? 'online' : 'offline']"></span>
              </div>
              <div class="profile-info">
                <h2 class="profile-name">{{ selectedFriend.friend_name || selectedFriend.friend_username || selectedFriend.username }}</h2>
                <div class="profile-status-row">
                  <span :class="['status-badge-large', getStatusClass(selectedFriend.friend_status || selectedFriend.status)]">
                    {{ selectedFriend.friend_status || selectedFriend.status || '离线' }}
                  </span>
                  <span v-if="selectedFriend.tomatoStatus" :class="['tomato-status-large', selectedFriend.tomatoStatus]">
                    {{ getTomatoStatusText(selectedFriend.tomatoStatus) }}
                  </span>
                </div>
              </div>
            </div>

            <!-- 导航栏 -->
            <div class="profile-nav">
              <button 
                :class="['nav-tab', { active: friendSubTab === 'info' }]"
                @click="friendSubTab = 'info'"
              >
                <span class="tab-icon">👤</span>
                个人信息
              </button>
            </div>

            <!-- 个人信息标签页 -->
            <div v-if="friendSubTab === 'info'" class="profile-content">
              <div class="info-section">
                <h3 class="info-section-title">基本信息</h3>
                <div class="info-grid">
                  <div class="info-item">
                    <span class="info-label">用户ID</span>
                    <span class="info-value">{{ selectedFriend.friend_id || selectedFriend.user_id }}</span>
                  </div>
                  <div class="info-item">
                    <span class="info-label">用户名</span>
                    <span class="info-value">{{ selectedFriend.friend_name || selectedFriend.friend_username || selectedFriend.username }}</span>
                  </div>
                  <div class="info-item">
                    <span class="info-label">在线状态</span>
                    <span :class="['info-value', 'status-value', getStatusClass(selectedFriend.friend_status || selectedFriend.status)]">
                      {{ selectedFriend.friend_status || selectedFriend.status || '离线' }}
                    </span>
                  </div>
                  <div v-if="selectedFriend.birthday" class="info-item">
                    <span class="info-label">生日</span>
                    <span class="info-value">{{ formatBirthday(selectedFriend.birthday) }}</span>
                  </div>
                  <div v-if="selectedFriend.province" class="info-item">
                    <span class="info-label">地区</span>
                    <span class="info-value">{{ selectedFriend.province }}</span>
                  </div>
                  <div v-if="selectedFriend.tomatoStatus" class="info-item">
                    <span class="info-label">番茄钟状态</span>
                    <span :class="['info-value', 'tomato-value', selectedFriend.tomatoStatus]">
                      {{ getTomatoStatusText(selectedFriend.tomatoStatus) }}
                    </span>
                  </div>
                </div>
              </div>

              <div v-if="selectedFriend.currentTask" class="info-section">
                <h3 class="info-section-title">当前活动</h3>
                <div class="activity-card">
                  <div v-if="selectedFriend.currentTask" class="activity-item">
                    <span class="activity-icon">📚</span>
                    <div class="activity-content">
                      <span class="activity-label">正在学习</span>
                      <span class="activity-value">{{ selectedFriend.currentTask }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 默认提示 -->
          <div v-if="activeTab === 'friends' && !selectedFriend" class="content-section">
            <div class="welcome-message">
              <div class="welcome-icon">👋</div>
              <h2>欢迎使用好友管理</h2>
              <p>从左侧选择好友查看详细信息，或使用导航栏进行其他操作</p>
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- 用户不存在提示对话框 -->
    <div v-if="showUserNotFoundModal" class="modal-overlay" @click="closeUserNotFoundModal">
      <div class="modal-content user-not-found-modal" @click.stop>
        <div class="modal-icon">👤</div>
        <h3 class="modal-title">用户不存在</h3>
        <p class="modal-message">抱歉，未找到用户 "<strong>{{ searchUsername }}</strong>"</p>
        <p class="modal-hint">请检查用户名是否正确</p>
        <div class="form-actions">
          <button @click="closeUserNotFoundModal" class="confirm-btn">
            我知道了
          </button>
        </div>
      </div>
    </div>


    <!-- 删除好友确认对话框 -->
    <div v-if="showDeleteModal" class="modal-overlay" @click="closeDeleteModal">
      <div class="modal-content delete-confirm-modal" @click.stop>
        <div class="modal-icon">🗑️</div>
        <h3 class="modal-title">确认删除好友</h3>
        <p class="modal-message">
          确定要删除好友 <strong>{{ selectedFriend?.friend_name || selectedFriend?.friend_username || selectedFriend?.username }}</strong> 吗？
        </p>
        <p class="modal-hint">此操作不可恢复</p>
        <div class="form-actions">
          <button type="button" @click="closeDeleteModal" class="cancel-btn">
            取消
          </button>
          <button @click="confirmDelete" class="confirm-btn delete-confirm-btn" :disabled="deleting">
            {{ deleting ? '删除中...' : '确认删除' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 发送好友申请弹窗 -->
    <div v-if="showRequestModal" class="modal-overlay" @click="closeRequestModal">
      <div class="modal-content" @click.stop>
        <h3 class="modal-title">发送好友申请</h3>
        <form @submit.prevent="sendFriendRequest">
          <div class="form-group">
            <label class="form-label">用户名</label>
            <input 
              type="text" 
              :value="selectedUser?.username" 
              class="form-input"
              disabled
            >
          </div>
          <div class="form-group">
            <label class="form-label">申请留言（可选）</label>
            <textarea 
              v-model="requestMessage"
              placeholder="请输入申请留言"
              class="form-textarea"
              rows="3"
            ></textarea>
          </div>
          <div class="form-actions">
            <button type="button" @click="closeRequestModal" class="cancel-btn">
              取消
            </button>
            <button type="submit" class="confirm-btn" :disabled="sendingRequest">
              {{ sendingRequest ? '发送中...' : '发送申请' }}
            </button>
          </div>
        </form>
      </div>
    </div>

  </div>
</template>

<script>
import { searchUser, sendFriendRequest, getFriendRequests, processFriendRequest, getFriends, deleteFriend } from '@/api/friends'
import { getCurrentUser } from '@/api/user'
import avatarImage from '@/assets/images/avatar.png'
import { API_BASE_URL } from '@/api/config'

export default {
  name: 'FriendsView',
  data() {
    return {
      searchUsername: '',
      searching: false,
      searchResult: null,
      searchError: '',
      friendRequests: [],
      loadingRequests: false,
      showRequestModal: false,
      selectedUser: null,
      requestMessage: '',
      sendingRequest: false,
      currentUserId: null,
      showUserNotFoundModal: false,
      friendsList: [],
      loadingFriends: false,
      showDeleteModal: false,
      deleting: false,
      activeTab: 'friends',
      selectedFriend: null,
      friendSubTab: 'info',
      refreshTimer: null, // 好友列表自动刷新定时器
      refreshInterval: 10000 // 每10秒刷新一次好友状态
    }
  },
  async mounted() {
    await this.initUser()
    await this.loadFriendRequests()
    await this.loadFriends()
    // 启动自动刷新
    this.startAutoRefresh()
    // 监听页面可见性变化
    this.setupVisibilityHandler()
  },
  methods: {
    async initUser() {
      try {
        const userResult = await getCurrentUser()
        if (userResult.success && userResult.data) {
          this.currentUserId = userResult.data.user_id
        }
      } catch (error) {
        console.error('获取用户信息失败:', error)
      }
    },

    async handleSearch() {
      if (!this.searchUsername.trim()) {
        this.searchError = '请输入用户名'
        return
      }

      this.searching = true
      this.searchError = ''
      this.searchResult = null
      this.showUserNotFoundModal = false

      try {
        // 只有ID存在的用户名才可以被搜索到
        const response = await searchUser(this.searchUsername.trim())
        
        if (response.success && response.data) {
          // 确保返回的数据有user_id，只有存在的用户才会被搜索到
          if (response.data.user_id) {
            this.searchResult = response.data
          } else {
            // 用户不存在，显示对话框
            this.showUserNotFoundModal = true
          }
        } else {
          // 用户不存在，显示对话框
          this.showUserNotFoundModal = true
        }
      } catch (error) {
        console.error('搜索用户失败:', error)
        // 检查是否是404错误（用户不存在）或网络错误
        if (error.status === 404 || error.message.includes('404') || 
            error.message.includes('用户不存在') || 
            error.message.includes('not found')) {
          // 用户不存在，显示对话框
          this.showUserNotFoundModal = true
        } else if (error.isNetworkError || error.message.includes('ECONNREFUSED') || 
                   error.message.includes('无法连接到服务器')) {
          // 网络连接错误
          this.searchError = '无法连接到服务器，请确保后端服务正在运行 (http://localhost:8090)'
        } else {
          // 其他错误
          this.searchError = error.message || '搜索失败，请稍后重试'
        }
      } finally {
        this.searching = false
      }
    },

    closeUserNotFoundModal() {
      this.showUserNotFoundModal = false
      this.searchUsername = ''
    },

    showSendRequestModal(user) {
      this.selectedUser = user
      this.requestMessage = ''
      this.showRequestModal = true
    },

    closeRequestModal() {
      this.showRequestModal = false
      this.selectedUser = null
      this.requestMessage = ''
    },

    async sendFriendRequest() {
      if (!this.selectedUser) return

      this.sendingRequest = true
      try {
        const response = await sendFriendRequest({
          user_name: this.selectedUser.username || this.searchUsername.trim(),
          message: this.requestMessage.trim() || null
        })

        if (response.success === true || response.success === "true") {
          alert('好友申请发送成功！')
          this.closeRequestModal()
          this.searchResult = null
          this.searchUsername = ''
          await this.loadFriendRequests()
        } else {
          alert(response.message || '发送好友申请失败')
        }
      } catch (error) {
        console.error('发送好友申请失败:', error)
        alert('发送好友申请失败: ' + (error.message || '未知错误'))
      } finally {
        this.sendingRequest = false
      }
    },

    async loadFriendRequests() {
      this.loadingRequests = true
      try {
        const response = await getFriendRequests()
        
        if (response.success === true || response.success === "true" || response.code === 200) {
          // 只显示待处理的申请
          this.friendRequests = (response.data || []).filter(req => req.status === '待处理')
        } else {
          console.error('获取好友申请失败:', response.message)
          this.friendRequests = []
        }
      } catch (error) {
        console.error('加载好友申请失败:', error)
        this.friendRequests = []
      } finally {
        this.loadingRequests = false
      }
    },

    async handleRequest(request, action) {
      try {
        const response = await processFriendRequest({
          from_user_id: request.from_user_id,
          action: action
        })

        if (response.success === true || response.success === "true") {
          alert(action === 'accept' ? '已接受好友申请' : '已拒绝好友申请')
          await this.loadFriendRequests()
          if (action === 'accept') {
            // 接受好友申请后，刷新好友列表
            await this.loadFriends()
          }
        } else {
          alert(response.message || '处理失败')
        }
      } catch (error) {
        console.error('处理好友申请失败:', error)
        alert('处理失败: ' + (error.message || '未知错误'))
      }
    },



    async loadFriends() {
      this.loadingFriends = true
      try {
        const response = await getFriends()
        
        if (response.success === true || response.success === "true" || response.code === 200) {
          this.friendsList = (response.data || []).map(friend => ({
            ...friend,
            isOnline: (friend.friend_status || friend.status) === '在线' || (friend.friend_status || friend.status) === '专注中',
            tomatoStatus: this.getTomatoStatus(friend),
            currentTask: this.getCurrentTask(friend)
          }))
        } else {
          this.friendsList = []
        }
      } catch (error) {
        console.error('加载好友列表失败:', error)
        this.friendsList = []
      } finally {
        this.loadingFriends = false
      }
    },

    getTomatoStatus(friend) {
      // 根据好友状态判断番茄钟状态
      const status = friend.friend_status || friend.status
      if (status === '专注中') return 'studying'
      if (status === '在线') return 'resting'
      return 'idle'
    },

    getCurrentTask(friend) {
      // 这里可以从好友数据中获取当前任务，暂时返回模拟数据
      if ((friend.friend_status || friend.status) === '专注中') {
        return friend.current_task || '正在专注学习'
      }
      return null
    },

    getTomatoStatusText(status) {
      const statusMap = {
        'studying': '学习中',
        'resting': '休息中',
        'idle': '空闲'
      }
      return statusMap[status] || '空闲'
    },


    selectFriend(friend) {
      this.selectedFriend = friend
      this.activeTab = 'friends'
      this.friendSubTab = 'info' // 默认显示个人信息
    },


    refreshFriends() {
      this.loadFriends()
      this.loadFriendRequests()
    },

    showDeleteConfirm(friend) {
      this.selectedFriend = friend
      this.showDeleteModal = true
    },

    closeDeleteModal() {
      this.showDeleteModal = false
      this.selectedFriend = null
    },

    async confirmDelete() {
      if (!this.selectedFriend) return

      this.deleting = true
      try {
        const friendName = this.selectedFriend.friend_name || 
                          this.selectedFriend.friend_username || 
                          this.selectedFriend.username
        
        const response = await deleteFriend({
          friend_name: friendName
        })

        if (response.success === true || response.success === "true") {
          alert('删除好友成功')
          this.closeDeleteModal()
          await this.loadFriends()
        } else {
          alert(response.message || '删除好友失败')
        }
      } catch (error) {
        console.error('删除好友失败:', error)
        alert('删除好友失败: ' + (error.message || '未知错误'))
      } finally {
        this.deleting = false
      }
    },

    isFriend(userId) {
      if (!userId || !this.friendsList || this.friendsList.length === 0) {
        return false
      }
      // 检查好友列表中是否包含该用户ID
      // 好友对象可能包含 friend_id（好友的用户ID）字段
      return this.friendsList.some(friend => {
        // 检查 friend_id 是否匹配
        if (friend.friend_id === userId) {
          return true
        }
        // 如果好友对象已经包含完整的用户信息，也可能有 user_id 字段
        if (friend.user_id === userId) {
          return true
        }
        return false
      })
    },

    hasPendingRequest(userId) {
      return this.friendRequests.some(r => r.from_user_id === userId && r.status === '待处理')
    },

    getStatusClass(status) {
      if (status === '在线') return 'status-online'
      if (status === '专注中') return 'status-busy'
      return 'status-offline'
    },

    getAvatarUrl(avatar) {
      if (!avatar || avatar === avatarImage) {
        return avatarImage
      }
      if (avatar.startsWith('http')) {
        return avatar
      }
      if (avatar.startsWith('/')) {
        return `http://localhost:8090${avatar}`
      }
      return `${API_BASE_URL}${avatar}`
    },

    handleAvatarError(event) {
      event.target.src = avatarImage
    },

    formatTime(timeStr) {
      if (!timeStr) return '-'
      const date = new Date(timeStr)
      return date.toLocaleDateString() + ' ' + date.toLocaleTimeString().slice(0, 5)
    },

    formatBirthday(birthdayStr) {
      if (!birthdayStr) return '未知'
      // 如果是日期字符串格式 (YYYY-MM-DD)
      if (typeof birthdayStr === 'string' && birthdayStr.includes('-')) {
        const date = new Date(birthdayStr)
        if (!isNaN(date.getTime())) {
          const year = date.getFullYear()
          const month = date.getMonth() + 1
          const day = date.getDate()
          return `${year}年${month}月${day}日`
        }
      }
      // 如果是时间戳
      if (typeof birthdayStr === 'number') {
        const date = new Date(birthdayStr)
        if (!isNaN(date.getTime())) {
          const year = date.getFullYear()
          const month = date.getMonth() + 1
          const day = date.getDate()
          return `${year}年${month}月${day}日`
        }
      }
      return birthdayStr || '未知'
    },

    goToHome() {
      this.$router.push('/home')
    },

    // 启动自动刷新好友列表
    startAutoRefresh() {
      // 如果已经有定时器，先清除
      if (this.refreshTimer) {
        clearInterval(this.refreshTimer)
      }
      
      // 设置定时器，每10秒刷新一次好友列表
      this.refreshTimer = setInterval(() => {
        // 只有在好友标签页且不在加载中时才刷新
        if (this.activeTab === 'friends' && !this.loadingFriends) {
          this.loadFriends()
        }
      }, this.refreshInterval)
      
      console.log('✅ 好友列表自动刷新已启动，间隔:', this.refreshInterval / 1000, '秒')
    },

    // 停止自动刷新
    stopAutoRefresh() {
      if (this.refreshTimer) {
        clearInterval(this.refreshTimer)
        this.refreshTimer = null
        console.log('⏹️ 好友列表自动刷新已停止')
      }
    },

    // 设置页面可见性监听（页面隐藏时停止刷新，显示时恢复刷新）
    setupVisibilityHandler() {
      document.addEventListener('visibilitychange', () => {
        if (document.hidden) {
          // 页面隐藏时停止刷新
          this.stopAutoRefresh()
        } else {
          // 页面显示时恢复刷新
          this.startAutoRefresh()
          // 立即刷新一次，获取最新状态
          if (this.activeTab === 'friends' && !this.loadingFriends) {
            this.loadFriends()
          }
        }
      })
    }
  },
  beforeUnmount() {
    // 清理自动刷新定时器
    this.stopAutoRefresh()
  }
}
</script>

<style scoped>
/* 使用白色背景 */
.friends-view {
  min-height: 100vh;
  background: #ffffff;
}

/* 导航栏样式 */
.navbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 5%;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid #ffe4cc;
  position: sticky;
  top: 0;
  z-index: 100;
}

.nav-brand {
  font-size: 1.8em;
  font-weight: bold;
  color: #eeaa67;
}

.nav-links {
  display: flex;
  gap: 30px;
}

.nav-link {
  cursor: pointer;
  padding: 10px 20px;
  border-radius: 8px;
  transition: all 0.3s ease;
  color: #333;
  font-weight: 500;
}

.nav-link:hover {
  background-color: #fff5eb;
  color: #eeaa67;
  transform: translateY(-1px);
}

/* 主要内容区域 */
.main-content {
  width: 100%;
  padding: 20px;
}

.page-header {
  text-align: center;
  margin-bottom: 40px;
}

.page-title {
  font-size: 2.8em;
  color: #333;
  font-weight: 700;
  margin: 0 0 12px 0;
  background: linear-gradient(135deg, #eeaa67, #ff8c42);
  background-clip: text;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.page-subtitle {
  font-size: 1.2em;
  color: #666;
  margin: 0;
}

/* 好友布局 */
.friends-layout {
  display: grid;
  grid-template-columns: 320px 1fr;
  gap: 20px;
  align-items: start;
}

/* 左侧边栏 */
.left-sidebar {
  background: white;
  border-radius: 16px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  border: 1px solid #e9ecef;
  display: flex;
  flex-direction: column;
  height: calc(100vh - 100px);
  position: sticky;
  top: 100px;
}

.sidebar-nav {
  padding: 20px;
  border-bottom: 2px solid #ffe4cc;
}

.nav-title {
  font-size: 1.5em;
  color: #333;
  font-weight: 600;
  margin: 0 0 20px 0;
  text-align: center;
}

.nav-menu {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border: none;
  background: #f8f9fa;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 1em;
  color: #333;
  text-align: left;
  position: relative;
}

.nav-item:hover {
  background: #fff5eb;
  color: #eeaa67;
}

.nav-item.active {
  background: linear-gradient(135deg, #eeaa67, #ff8c42);
  color: white;
  font-weight: 600;
}

.nav-icon {
  font-size: 1.2em;
}

.nav-badge {
  margin-left: auto;
  background: rgba(255, 255, 255, 0.3);
  color: white;
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 0.85em;
  font-weight: 600;
}

.nav-item.active .nav-badge {
  background: rgba(255, 255, 255, 0.3);
}

/* 好友列表容器 */
.friends-list-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.friends-list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 20px;
  border-bottom: 1px solid #e9ecef;
}

.friends-list-header h3 {
  margin: 0;
  font-size: 1.1em;
  color: #333;
  font-weight: 600;
}

.refresh-btn {
  padding: 6px 10px;
  background: #f8f9fa;
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  cursor: pointer;
  font-size: 1em;
  transition: all 0.2s ease;
}

.refresh-btn:hover:not(:disabled) {
  background: #fff5eb;
  border-color: #eeaa67;
  color: #eeaa67;
  transform: rotate(90deg);
}

.refresh-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.friends-list-scroll {
  flex: 1;
  overflow-y: auto;
  padding: 10px;
}

.friends-list-scroll::-webkit-scrollbar {
  width: 6px;
}

.friends-list-scroll::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

.friends-list-scroll::-webkit-scrollbar-thumb {
  background: #eeaa67;
  border-radius: 3px;
}

.friends-list-scroll::-webkit-scrollbar-thumb:hover {
  background: #e69c55;
}

/* 展开的好友卡片 */
.friend-card-expanded {
  background: #f8f9fa;
  border-radius: 12px;
  padding: 15px;
  margin-bottom: 12px;
  border: 2px solid transparent;
  cursor: pointer;
  transition: all 0.3s ease;
}

.friend-card-expanded:hover {
  background: #fff5eb;
  border-color: #ffe4cc;
  transform: translateX(3px);
}

.friend-card-expanded.selected {
  background: linear-gradient(135deg, #fff5eb 0%, #ffe4cc 100%);
  border-color: #eeaa67;
  box-shadow: 0 4px 12px rgba(238, 170, 103, 0.2);
}

.friend-card-expanded .friend-avatar-wrapper {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.friend-card-expanded .friend-avatar {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  overflow: hidden;
  border: 2px solid #eeaa67;
  flex-shrink: 0;
}

.friend-card-expanded .friend-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.friend-card-expanded .friend-details {
  flex: 1;
  min-width: 0;
}

.friend-card-expanded .friend-name {
  margin: 0 0 8px 0;
  font-size: 1.05em;
  font-weight: 600;
  color: #333;
}

.friend-status-info {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 8px;
}

.friend-card-expanded .current-task-info,
.friend-card-expanded .countdown-info {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.85em;
  color: #666;
  margin-top: 6px;
}

.friend-card-expanded .task-text {
  color: #333;
  font-weight: 500;
}

.friend-card-expanded .countdown-text {
  color: #eeaa67;
  font-weight: 600;
  font-family: 'Courier New', monospace;
}

.friend-card-expanded .friend-actions {
  display: flex;
  gap: 8px;
  margin-top: 10px;
  justify-content: flex-end;
}

.btn-view-info {
  padding: 6px 12px;
  background: #e3f2fd;
  color: #1976d2;
  border: 1px solid #90caf9;
  border-radius: 6px;
  cursor: pointer;
  font-size: 1em;
  transition: all 0.2s ease;
}

.btn-view-info:hover {
  background: #bbdefb;
  transform: scale(1.05);
}

/* 右侧内容区域 */
.right-content {
  background: white;
  border-radius: 16px;
  padding: 30px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  border: 1px solid #e9ecef;
  min-height: calc(100vh - 100px);
}

.content-section {
  width: 100%;
}

.welcome-message {
  text-align: center;
  padding: 60px 20px;
  color: #666;
}

.welcome-icon {
  font-size: 4em;
  margin-bottom: 20px;
}

.welcome-message h2 {
  margin: 0 0 15px 0;
  color: #333;
  font-size: 1.8em;
}

.welcome-message p {
  margin: 0;
  font-size: 1.1em;
  line-height: 1.6;
}

.friend-detail-card {
  background: #f8f9fa;
  border-radius: 12px;
  padding: 25px;
  border: 1px solid #e9ecef;
}

.detail-header {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-bottom: 25px;
  padding-bottom: 20px;
  border-bottom: 2px solid #ffe4cc;
}

.detail-avatar {
  position: relative;
  width: 80px;
  height: 80px;
  border-radius: 50%;
  overflow: hidden;
  border: 4px solid #eeaa67;
  flex-shrink: 0;
}

.detail-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.detail-info {
  flex: 1;
}

.detail-info h3 {
  margin: 0 0 10px 0;
  font-size: 1.5em;
  color: #333;
  font-weight: 600;
}

.detail-status {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.btn-view-stats {
  padding: 10px 20px;
  background: linear-gradient(135deg, #eeaa67, #ff8c42);
  color: white;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-view-stats:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 15px rgba(238, 170, 103, 0.3);
}

.detail-content {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.detail-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px;
  background: white;
  border-radius: 8px;
  border: 1px solid #e9ecef;
}

.detail-label {
  font-weight: 600;
  color: #666;
  min-width: 100px;
}

.detail-value {
  color: #333;
  flex: 1;
}

.countdown-value {
  color: #eeaa67;
  font-weight: 600;
  font-family: 'Courier New', monospace;
  font-size: 1.1em;
}

/* 好友主页头部 */
.friend-profile-header {
  display: flex;
  align-items: center;
  gap: 25px;
  padding: 30px;
  background: linear-gradient(135deg, #fff5eb 0%, #ffe4cc 100%);
  border-radius: 16px;
  margin-bottom: 25px;
  border: 1px solid #ffe4cc;
}

.profile-avatar {
  position: relative;
  width: 100px;
  height: 100px;
  border-radius: 50%;
  overflow: hidden;
  border: 4px solid #eeaa67;
  flex-shrink: 0;
  box-shadow: 0 4px 15px rgba(238, 170, 103, 0.3);
}

.profile-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.online-indicator-large {
  position: absolute;
  bottom: 5px;
  right: 5px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  border: 4px solid white;
}

.online-indicator-large.online {
  background: #4CAF50;
  box-shadow: 0 0 0 2px #4CAF50;
}

.online-indicator-large.offline {
  background: #95a5a6;
  box-shadow: 0 0 0 2px #95a5a6;
}

.profile-info {
  flex: 1;
}

.profile-name {
  margin: 0 0 15px 0;
  font-size: 2em;
  color: #333;
  font-weight: 700;
}

.profile-status-row {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.status-badge-large {
  padding: 8px 20px;
  border-radius: 20px;
  font-size: 1em;
  font-weight: 600;
}

.status-badge-large.status-online {
  background: #e7f5e9;
  color: #2b8a3e;
}

.status-badge-large.status-busy {
  background: #fff9f2;
  color: #eeaa67;
}

.status-badge-large.status-offline {
  background: #f0f0f0;
  color: #666;
}

.tomato-status-large {
  padding: 8px 20px;
  border-radius: 20px;
  font-size: 1em;
  font-weight: 600;
}

.tomato-status-large.studying {
  background: #ffebee;
  color: #c92a2a;
}

.tomato-status-large.resting {
  background: #e3f2fd;
  color: #1976d2;
}

.tomato-status-large.idle {
  background: #f5f5f5;
  color: #666;
}

/* 导航栏 */
.profile-nav {
  display: flex;
  gap: 10px;
  margin-bottom: 25px;
  border-bottom: 2px solid #ffe4cc;
  padding-bottom: 0;
}

.nav-tab {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  border: none;
  background: transparent;
  border-bottom: 3px solid transparent;
  cursor: pointer;
  transition: all 0.3s ease;
  font-size: 1em;
  color: #666;
  font-weight: 500;
  margin-bottom: -2px;
}

.nav-tab:hover {
  color: #eeaa67;
  background: #fff5eb;
}

.nav-tab.active {
  color: #eeaa67;
  border-bottom-color: #eeaa67;
  font-weight: 600;
}

.tab-icon {
  font-size: 1.2em;
}

/* 个人信息内容 */
.profile-content {
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

.info-section {
  background: #f8f9fa;
  border-radius: 12px;
  padding: 25px;
  margin-bottom: 20px;
  border: 1px solid #e9ecef;
}

.info-section-title {
  margin: 0 0 20px 0;
  font-size: 1.3em;
  color: #333;
  font-weight: 600;
  padding-bottom: 15px;
  border-bottom: 2px solid #ffe4cc;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 15px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 15px;
  background: white;
  border-radius: 8px;
  border: 1px solid #e9ecef;
}

.info-label {
  font-size: 0.9em;
  color: #666;
  font-weight: 500;
}

.info-value {
  font-size: 1.1em;
  color: #333;
  font-weight: 600;
}

.status-value.status-online {
  color: #2b8a3e;
}

.status-value.status-busy {
  color: #eeaa67;
}

.status-value.status-offline {
  color: #666;
}

.tomato-value.studying {
  color: #c92a2a;
}

.tomato-value.resting {
  color: #1976d2;
}

.tomato-value.idle {
  color: #666;
}

.activity-card {
  background: white;
  border-radius: 12px;
  padding: 20px;
  border: 1px solid #e9ecef;
}

.activity-item {
  display: flex;
  align-items: center;
  gap: 15px;
  padding: 15px;
  background: #f8f9fa;
  border-radius: 8px;
  margin-bottom: 12px;
}

.activity-item:last-child {
  margin-bottom: 0;
}

.activity-icon {
  font-size: 2em;
  flex-shrink: 0;
}

.activity-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.activity-label {
  font-size: 0.85em;
  color: #666;
  font-weight: 500;
}

.activity-value {
  font-size: 1.1em;
  color: #333;
  font-weight: 600;
}

.countdown-display {
  color: #eeaa67;
  font-family: 'Courier New', monospace;
  font-size: 1.3em;
}

/* 好友列表区域 */
.friends-list-section {
  background: white;
  border-radius: 16px;
  padding: 30px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  border: 1px solid #e9ecef;
  margin-top: 30px;
}

.friends-list {
  display: flex;
  flex-direction: column;
  gap: 15px;
  margin-top: 20px;
}

.friend-card {
  background: #f8f9fa;
  border-radius: 12px;
  padding: 20px;
  border: 1px solid #e9ecef;
  transition: all 0.3s ease;
  cursor: pointer;
}

.friend-card:hover {
  border-color: #eeaa67;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(238, 170, 103, 0.2);
  background: #fff5eb;
}

.friend-header {
  display: flex;
  align-items: center;
  gap: 15px;
}

.friend-avatar-wrapper {
  position: relative;
  flex-shrink: 0;
}

.friend-avatar {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  overflow: hidden;
  border: 3px solid #eeaa67;
}

.friend-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.online-indicator {
  position: absolute;
  bottom: 2px;
  right: 2px;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  border: 3px solid white;
}

.online-indicator.online {
  background: #4CAF50;
}

.online-indicator.offline {
  background: #95a5a6;
}

.friend-main-info {
  flex: 1;
  min-width: 0;
}

.friend-name {
  margin: 0 0 8px 0;
  font-size: 1.1em;
  font-weight: 600;
  color: #333;
}

.friend-status-row {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.status-badge {
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 0.85em;
  font-weight: 500;
}

.status-badge.status-online {
  background: #e7f5e9;
  color: #2b8a3e;
}

.status-badge.status-busy {
  background: #fff9f2;
  color: #eeaa67;
}

.status-badge.status-offline {
  background: #f0f0f0;
  color: #666;
}

.tomato-status {
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 0.85em;
  font-weight: 500;
}

.tomato-status.studying {
  background: #ffebee;
  color: #c92a2a;
}

.tomato-status.resting {
  background: #e3f2fd;
  color: #1976d2;
}

.tomato-status.idle {
  background: #f5f5f5;
  color: #666;
}

.btn-delete-friend {
  padding: 8px 12px;
  background: #fff5f5;
  color: #c92a2a;
  border: 1px solid #ffcccc;
  border-radius: 8px;
  font-size: 1.2em;
  cursor: pointer;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.btn-delete-friend:hover {
  background: #ffebee;
  border-color: #ff9999;
  transform: scale(1.1);
}

.friend-activity {
  margin-top: 15px;
  padding-top: 15px;
  border-top: 1px solid #e9ecef;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.current-task, .countdown {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.9em;
  color: #666;
}

.task-icon, .countdown-icon {
  font-size: 1.2em;
}

.task-text {
  color: #333;
  font-weight: 500;
}

.countdown-text {
  color: #eeaa67;
  font-weight: 600;
  font-family: 'Courier New', monospace;
}

/* 内容区域样式 */
.content-section {
  width: 100%;
}

.search-section {
  background: white;
  border-radius: 16px;
  padding: 30px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  border: 1px solid #e9ecef;
}

.section-title {
  font-size: 1.6em;
  color: #333;
  font-weight: 600;
  margin: 0 0 20px 0;
}

.search-box {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
}

.search-input {
  flex: 1;
  padding: 12px 16px;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  font-size: 1em;
  transition: all 0.3s ease;
}

.search-input:focus {
  outline: none;
  border-color: #eeaa67;
  box-shadow: 0 0 0 3px rgba(238, 170, 103, 0.1);
}

.search-btn {
  padding: 12px 24px;
  background: #eeaa67;
  color: white;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
}

.search-btn:hover:not(:disabled) {
  background: #e69c55;
  transform: translateY(-1px);
  box-shadow: 0 4px 15px rgba(238, 170, 103, 0.3);
}

.search-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.search-result {
  margin-top: 20px;
}

.user-card {
  display: flex;
  align-items: center;
  gap: 15px;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 12px;
  border: 1px solid #e9ecef;
}

.user-avatar {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  overflow: hidden;
  flex-shrink: 0;
  background: #ddd;
}

.user-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.user-info {
  flex: 1;
  min-width: 0;
}

.user-name {
  margin: 0 0 4px 0;
  font-size: 1.1em;
  font-weight: 600;
  color: #333;
}

.user-id {
  margin: 0 0 8px 0;
  font-size: 0.9em;
  color: #666;
}

.user-status {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.85em;
  color: #666;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.status-dot.status-online {
  background-color: #4CAF50;
}

.status-dot.status-busy {
  background-color: #f1c40f;
}

.status-dot.status-offline {
  background-color: #95a5a6;
}

.user-actions {
  flex-shrink: 0;
}

.btn-add-friend {
  padding: 8px 16px;
  background: #eeaa67;
  color: white;
  border: none;
  border-radius: 6px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-add-friend:hover {
  background: #e69c55;
}

.friend-badge,
.pending-badge {
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 0.9em;
  font-weight: 500;
}

.friend-badge {
  background: #e7f5e9;
  color: #2b8a3e;
}

.pending-badge {
  background: #fff9f2;
  color: #eeaa67;
}

.error-message {
  margin-top: 12px;
  padding: 12px;
  background: #fff5f5;
  color: #c92a2a;
  border-radius: 8px;
  font-size: 0.9em;
}

/* 好友申请列表 */
.requests-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.request-item {
  display: flex;
  align-items: center;
  gap: 15px;
  padding: 16px;
  background: #f8f9fa;
  border-radius: 12px;
  border: 1px solid #e9ecef;
  transition: all 0.2s ease;
}

.request-item:hover {
  border-color: #eeaa67;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.request-avatar {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  overflow: hidden;
  flex-shrink: 0;
  background: #ddd;
}

.request-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.request-info {
  flex: 1;
  min-width: 0;
}

.request-name {
  margin: 0 0 4px 0;
  font-size: 1em;
  font-weight: 600;
  color: #333;
}

.request-message {
  margin: 0 0 4px 0;
  font-size: 0.9em;
  color: #666;
}

.request-time {
  margin: 0;
  font-size: 0.85em;
  color: #999;
}

.request-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.btn-accept,
.btn-reject {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-accept {
  background: #4CAF50;
  color: white;
}

.btn-accept:hover:not(:disabled) {
  background: #45a049;
}

.btn-reject {
  background: #f8f9fa;
  color: #666;
  border: 1px solid #e0e0e0;
}

.btn-reject:hover:not(:disabled) {
  background: #e9ecef;
  border-color: #ccc;
}

.btn-accept:disabled,
.btn-reject:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}


/* 空状态和加载状态 */
.empty-state,
.loading-state {
  text-align: center;
  padding: 40px 20px;
  color: #666;
}

.empty-icon {
  font-size: 4em;
  margin-bottom: 20px;
}

.empty-state h3 {
  margin: 0 0 12px 0;
  color: #333;
}

.empty-state p {
  margin: 0;
}

.spinner {
  border: 4px solid rgba(0, 0, 0, 0.1);
  border-top: 4px solid #eeaa67;
  border-radius: 50%;
  width: 30px;
  height: 30px;
  animation: spin 1s linear infinite;
  margin: 0 auto 15px;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

/* 弹窗样式 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 16px;
  padding: 30px;
  max-width: 500px;
  width: 90%;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
}

.modal-title {
  font-size: 1.5em;
  color: #333;
  margin: 0 0 25px 0;
  font-weight: 600;
}

.modal-message {
  margin: 0 0 25px 0;
  color: #666;
  line-height: 1.6;
}

.form-group {
  margin-bottom: 20px;
}

.form-label {
  display: block;
  font-weight: 500;
  color: #333;
  margin-bottom: 8px;
}

.form-input,
.form-textarea {
  width: 100%;
  padding: 12px 16px;
  border: 2px solid #e0e0e0;
  border-radius: 8px;
  font-size: 1em;
  transition: all 0.3s ease;
  font-family: inherit;
}

.form-input:focus,
.form-textarea:focus {
  outline: none;
  border-color: #eeaa67;
  box-shadow: 0 0 0 3px rgba(238, 170, 103, 0.1);
}

.form-input:disabled {
  background: #f8f9fa;
  cursor: not-allowed;
}

.form-textarea {
  resize: vertical;
  min-height: 80px;
}

.form-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 20px;
}

.cancel-btn,
.confirm-btn {
  padding: 12px 24px;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
}

.cancel-btn {
  background: #f8f9fa;
  color: #666;
  border: 2px solid #e0e0e0;
}

.cancel-btn:hover {
  background: #e9ecef;
}

.confirm-btn {
  background: #eeaa67;
  color: white;
}

.confirm-btn:hover:not(:disabled) {
  background: #e69c55;
  transform: translateY(-1px);
  box-shadow: 0 4px 15px rgba(238, 170, 103, 0.3);
}

.confirm-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.delete-confirm-btn {
  background: #c92a2a;
  color: white;
}

.delete-confirm-btn:hover:not(:disabled) {
  background: #b52217;
}

/* 删除确认对话框样式 */
.delete-confirm-modal {
  text-align: center;
  max-width: 450px;
}

.delete-confirm-modal .modal-icon {
  font-size: 4em;
  margin-bottom: 20px;
  animation: shake 0.5s ease;
}

@keyframes shake {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-10px); }
  75% { transform: translateX(10px); }
}

.delete-confirm-modal .modal-title {
  color: #c92a2a;
  margin-bottom: 15px;
}

.delete-confirm-modal .modal-message {
  font-size: 1.1em;
  color: #333;
  margin-bottom: 10px;
  line-height: 1.6;
}

.delete-confirm-modal .modal-message strong {
  color: #eeaa67;
  font-weight: 600;
}

.delete-confirm-modal .modal-hint {
  font-size: 0.9em;
  color: #999;
  margin: 15px 0 25px 0;
  font-style: italic;
}

/* 好友学习概览模态框 */
.friend-overview-modal {
  max-width: 800px;
  max-height: 90vh;
  overflow-y: auto;
}

.overview-header {
  display: flex;
  align-items: center;
  gap: 20px;
  padding-bottom: 20px;
  border-bottom: 2px solid #ffe4cc;
  margin-bottom: 25px;
  position: relative;
}

.overview-avatar {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  overflow: hidden;
  border: 4px solid #eeaa67;
  flex-shrink: 0;
}

.overview-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.overview-info {
  flex: 1;
}

.overview-name {
  margin: 0 0 8px 0;
  font-size: 1.5em;
  color: #333;
  font-weight: 600;
}

.overview-status {
  padding: 6px 16px;
  border-radius: 16px;
  font-size: 0.9em;
  font-weight: 500;
  display: inline-block;
}

.overview-status.status-online {
  background: #e7f5e9;
  color: #2b8a3e;
}

.overview-status.status-busy {
  background: #fff9f2;
  color: #eeaa67;
}

.overview-status.status-offline {
  background: #f0f0f0;
  color: #666;
}

.close-btn {
  position: absolute;
  top: 0;
  right: 0;
  width: 40px;
  height: 40px;
  border: none;
  background: #f8f9fa;
  border-radius: 50%;
  font-size: 1.5em;
  cursor: pointer;
  color: #666;
  transition: all 0.2s ease;
}

.close-btn:hover {
  background: #eeaa67;
  color: white;
  transform: rotate(90deg);
}

.overview-content {
  display: flex;
  flex-direction: column;
  gap: 25px;
}

.stat-card-overview {
  background: linear-gradient(135deg, #fff5eb 0%, #ffe4cc 100%);
  border-radius: 16px;
  padding: 25px;
  border: 1px solid #ffe4cc;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 20px;
}

.stat-item .stat-icon {
  font-size: 3em;
}

.stat-details {
  flex: 1;
}

.stat-item .stat-value {
  font-size: 2.5em;
  font-weight: 700;
  color: #eeaa67;
  margin-bottom: 5px;
}

.stat-item .stat-label {
  font-size: 1em;
  color: #666;
  font-weight: 500;
}

.chart-section {
  background: white;
  border-radius: 12px;
  padding: 20px;
  border: 1px solid #e9ecef;
}

.chart-title {
  margin: 0 0 20px 0;
  font-size: 1.2em;
  color: #333;
  font-weight: 600;
}

.bar-chart {
  display: flex;
  align-items: flex-end;
  justify-content: space-around;
  gap: 10px;
  height: 200px;
  padding: 10px 0;
}

.bar-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.bar {
  width: 100%;
  background: linear-gradient(180deg, #eeaa67 0%, #ff8c42 100%);
  border-radius: 4px 4px 0 0;
  min-height: 10px;
  transition: all 0.3s ease;
}

.bar-item:hover .bar {
  opacity: 0.8;
  transform: scaleY(1.05);
}

.bar-label {
  font-size: 0.85em;
  color: #666;
  font-weight: 500;
}

.bar-value {
  font-size: 0.9em;
  color: #eeaa67;
  font-weight: 600;
}

.line-chart {
  position: relative;
  height: 200px;
  padding: 20px 0;
}

.chart-svg {
  width: 100%;
  height: 100%;
}

.chart-labels {
  display: flex;
  justify-content: space-around;
  margin-top: 10px;
  font-size: 0.85em;
  color: #666;
}

.subjects-section {
  background: white;
  border-radius: 12px;
  padding: 20px;
  border: 1px solid #e9ecef;
}

.subject-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.subject-tag {
  padding: 8px 16px;
  background: linear-gradient(135deg, #fff5eb 0%, #ffe4cc 100%);
  border-radius: 20px;
  border: 1px solid #eeaa67;
  font-size: 0.9em;
  color: #333;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 8px;
}

.subject-count {
  background: #eeaa67;
  color: white;
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 0.85em;
  font-weight: 600;
}

/* 用户不存在对话框样式 */
.user-not-found-modal {
  text-align: center;
  max-width: 450px;
}

.modal-icon {
  font-size: 4em;
  margin-bottom: 20px;
  animation: bounce 0.6s ease;
}

@keyframes bounce {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-10px);
  }
}

.user-not-found-modal .modal-title {
  color: #eeaa67;
  margin-bottom: 15px;
}

.user-not-found-modal .modal-message {
  font-size: 1.1em;
  color: #333;
  margin-bottom: 10px;
  line-height: 1.6;
}

.user-not-found-modal .modal-message strong {
  color: #eeaa67;
  font-weight: 600;
}

.modal-hint {
  font-size: 0.9em;
  color: #666;
  margin: 15px 0 25px 0;
  line-height: 1.5;
}

.user-not-found-modal .form-actions {
  justify-content: center;
}

.user-not-found-modal .confirm-btn {
  min-width: 120px;
  padding: 12px 30px;
  background: linear-gradient(135deg, #eeaa67, #f5b877);
  box-shadow: 0 4px 15px rgba(238, 170, 103, 0.3);
}

.user-not-found-modal .confirm-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(238, 170, 103, 0.4);
}

/* 响应式设计 */
@media (max-width: 968px) {
  .friends-layout {
    grid-template-columns: 1fr;
  }
  
  .left-sidebar {
    position: static;
    height: auto;
    max-height: 60vh;
  }
  
  .right-content {
    min-height: auto;
  }
}

@media (max-width: 768px) {
  .main-content {
    padding: 20px;
  }
  
  .page-title {
    font-size: 2.2em;
  }
  
  .search-section,
  .requests-section,
  .right-section {
    padding: 20px;
  }
  
  .search-box {
    flex-direction: column;
  }
  
  .request-item,
  .friend-item {
    flex-wrap: wrap;
  }
  
  .request-actions,
  .user-actions {
    width: 100%;
    margin-top: 12px;
  }
}
</style>

