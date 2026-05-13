<template>
  <div class="join-room-container">
    <!-- 顶部导航栏 -->
    <nav class="navbar">
      <div class="nav-brand">Tomato</div>
      <div class="nav-links">
        <a class="nav-link" @click="goToHome">返回首页</a>
      </div>
    </nav>

    <!-- 主要内容区域 -->
    <main class="main-content">
      <div class="join-room-header">
        <h1 class="page-title">加入自习室</h1>
        <p class="page-subtitle">选择你喜欢的方式加入自习室</p>
      </div>

      <div class="join-methods-grid">
        <!-- 左侧：输入加入码 -->
        <div class="join-method-card code-join-section">
          <div class="method-header">
            <div class="method-icon">🔑</div>
            <h2 class="method-title">输入加入码加入</h2>
          </div>
          
          <div class="method-content">
            <div class="form-group">
              <label class="form-label">自习室加入码</label>
              <input 
                type="text" 
                v-model="joinCode" 
                placeholder="请输入6位房间号"
                class="form-input"
                :disabled="loading"
                maxlength="6"
                @keyup.enter="validateAndJoin"
              >
              <div v-if="errorMessage" class="error-message">
                {{ errorMessage }}
              </div>
              <div v-if="successMessage" class="success-message">
                {{ successMessage }}
              </div>
            </div>

            <div class="button-group">
              <button 
                class="btn-primary" 
                @click="validateAndJoin"
                :disabled="loading || !joinCode.trim()"
              >
                <span v-if="loading">
                  <span class="loading-spinner"></span> 处理中...
                </span>
                <span v-else>确认加入</span>
              </button>
              <button 
                class="btn-secondary" 
                @click="goToHome"
                :disabled="loading"
              >
                返回首页
              </button>
            </div>

            <div class="create-link">
              <span>还没有自习室？</span>
              <a @click="goToCreateRoom">立即创建</a>
            </div>
          </div>
        </div>

        <!-- 右侧：快速加入 -->
        <div class="join-method-card quick-join-section">
          <div class="method-header">
            <div class="method-icon">⚡</div>
            <h2 class="method-title">快速加入</h2>
          </div>
          
          <div class="method-content">
            <div class="dynamic-content-placeholder">
              <div class="placeholder-icon">📚</div>
              <p class="placeholder-text">热门自习室将在这里显示</p>
              <p class="placeholder-subtext">内容动态加载中...</p>
            </div>
          </div>
        </div>
      </div>

      <!-- 房间不存在提示弹窗 -->
      <div v-if="showRoomNotFound" class="room-not-found-modal">
        <div class="modal-overlay" @click="closeRoomNotFound">
          <div class="modal-content" @click.stop>
            <div class="modal-header">
              <h3>房间不存在</h3>
            </div>
            <div class="modal-body">
              <p>房间号 <strong>{{ joinCode }}</strong> 不存在，请检查后重试。</p>
            </div>
            <div class="modal-footer">
              <button class="modal-btn" @click="closeRoomNotFound">确定</button>
            </div>
          </div>
        </div>
      </div>

      <!-- 房间已满提示弹窗 -->
      <div v-if="showRoomFull" class="room-not-found-modal">
        <div class="modal-overlay" @click="closeRoomFull">
          <div class="modal-content" @click.stop>
            <div class="modal-header">
              <h3>房间已满</h3>
            </div>
            <div class="modal-body">
              <p>房间号 <strong>{{ joinCode }}</strong> 已满员，无法加入。</p>
              <p class="tip-text">建议创建新的自习室或稍后再试。</p>
            </div>
            <div class="modal-footer">
              <button class="modal-btn" @click="closeRoomFull">确定</button>
              <button class="modal-btn secondary" @click="goToCreateRoom">创建新房间</button>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script>
import { getRoomDetail, joinRoom } from '@/api/studyRooms'
import { getCurrentUser } from '@/api/user'

export default {
  name: 'JoinRoomView',
  data() {
    return {
      joinCode: '',
      loading: false,
      errorMessage: '',
      successMessage: '',
      showRoomNotFound: false,
      showRoomFull: false,
      currentUser: null,
      currentUserId: null
    }
  },
  async mounted() {
    await this.loadCurrentUser()
    
    this.$nextTick(() => {
      const input = this.$el.querySelector('.form-input')
      if (input) input.focus()
    })
  },
  methods: {
    async loadCurrentUser() {
      try {
        const response = await getCurrentUser()
        
        if (response.success && response.data) {
          this.currentUser = response.data
          this.currentUserId = response.data.id || 
                               response.data.userId || 
                               response.data.user_id
          
          console.log('当前用户信息:', this.currentUser)
          console.log('用户ID:', this.currentUserId)
          
          if (!this.currentUserId) {
            console.warn('无法获取用户ID')
          }
        }
      } catch (error) {
        console.error('获取用户信息失败:', error)
      }
    },
    
    async validateAndJoin() {
      const roomId = this.joinCode.trim()
      
      if (!roomId) {
        this.errorMessage = '请输入房间号'
        return
      }
      
      if (!/^[A-Za-z0-9]{6}$/.test(roomId)) {
        this.errorMessage = '房间号应为6位数字或字母'
        return
      }
      
      this.loading = true
      this.errorMessage = ''
      this.successMessage = ''
      
      try {
        console.log(`=== 开始加入房间流程 ===`)
        console.log(`房间ID: ${roomId}`)
        
        // 确保有用户ID
        if (!this.currentUserId) {
          await this.loadCurrentUser()
        }
        
        if (!this.currentUserId) {
          this.errorMessage = '用户身份验证失败，请重新登录'
          return
        }
        
        // 将用户ID转换为数字（后端需要Long类型）
        const userId = Number(this.currentUserId)
        if (isNaN(userId)) {
          console.error('用户ID不是有效的数字:', this.currentUserId)
          this.errorMessage = '用户ID格式错误'
          return
        }
        
        console.log(`使用的参数 - 房间ID: ${roomId}, 用户ID: ${userId}`)
        
        // 1. 验证房间是否存在
        console.log(`步骤1: 验证房间 ${roomId} 是否存在...`)
        const roomDetail = await getRoomDetail(roomId)
        console.log('房间详情响应:', roomDetail)
        
        if (!roomDetail || !(roomDetail.code === 200 || roomDetail.success === true)) {
          console.warn('房间验证失败')
          this.showRoomNotFound = true
          return
        }
        
        console.log('✅ 房间验证成功')
        
        // 2. 加入房间
        console.log(`步骤2: 调用加入房间API...`)
        console.log(`请求路径: /api/rooms/${roomId}/join?userId=${userId}`)
        
        const joinResult = await joinRoom(roomId, userId)
        console.log('加入房间响应:', joinResult)
        
        if (joinResult && (joinResult.code === 200 || joinResult.success === true)) {
          console.log('✅ 加入房间成功')
          this.successMessage = '加入成功！正在跳转...'
          
          setTimeout(() => {
            this.$router.push({
              name: 'study-room',
              params: { roomId: roomId }
            })
          }, 500)
          
        } else {
          const errorMsg = joinResult?.message || '加入失败，请重试'
          console.warn('加入房间失败:', errorMsg)
          
          if (errorMsg.includes('满') || errorMsg.includes('full')) {
            this.showRoomFull = true
          } else {
            this.errorMessage = errorMsg
          }
        }
        
      } catch (error) {
        console.error('加入房间过程中出错:', error)
        
        if (error.response) {
          const { status, data } = error.response
          console.error(`HTTP ${status}:`, data)
          
          switch (status) {
            case 400:
              this.errorMessage = data?.message || '请求参数错误'
              break
            case 404:
              this.showRoomNotFound = true
              break
            case 403:
              this.errorMessage = '权限不足，无法加入该自习室'
              break
            case 409:
              this.successMessage = '您已在房间中，正在跳转...'
              setTimeout(() => {
                this.$router.push({
                  name: 'study-room',
                  params: { roomId: roomId }
                })
              }, 500)
              break
            case 422:
              this.showRoomFull = true
              break
            case 500:
              if (data?.message?.includes('MissingServletRequestParameterException')) {
                this.errorMessage = '服务器参数错误：缺少用户ID参数'
              } else {
                this.errorMessage = '服务器内部错误，请稍后重试'
              }
              break
            default:
              this.errorMessage = data?.message || `服务器错误: ${status}`
          }
        } else if (error.request) {
          this.errorMessage = '网络连接失败，请检查网络设置'
        } else {
          this.errorMessage = error.message || '未知错误'
        }
      } finally {
        this.loading = false
        console.log('=== 加入流程结束 ===')
      }
    },
    
    closeRoomNotFound() {
      this.showRoomNotFound = false
      this.joinCode = ''
    },
    
    closeRoomFull() {
      this.showRoomFull = false
      this.joinCode = ''
    },
    
    goToCreateRoom() {
      this.$router.push('/create-room')
    },
    
    goToHome() {
      this.$router.push('/home')
    }
  }
}
</script>


<style scoped>

/* 成功提示 */
.success-message {
  color: #2b8a3e;
  font-size: 0.9em;
  margin-top: 8px;
  padding: 8px 12px;
  background: #e7f5e9;
  border-radius: 6px;
  border-left: 3px solid #2b8a3e;
}

/* 加载动画 */
.loading-spinner {
  display: inline-block;
  width: 12px;
  height: 12px;
  border: 2px solid #fff;
  border-top: 2px solid transparent;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-right: 6px;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}


/* 错误提示 */
.error-message {
  color: #ff6b6b;
  font-size: 0.9em;
  margin-top: 8px;
  padding: 8px 12px;
  background: #fff5f5;
  border-radius: 6px;
  border-left: 3px solid #ff6b6b;
}

/* 房间不存在弹窗 */
.room-not-found-modal .modal-overlay {
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

.room-not-found-modal .modal-content {
  background: white;
  border-radius: 16px;
  padding: 30px;
  max-width: 400px;
  width: 90%;
  text-align: center;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
  border: 1px solid #ffe4cc;
}

.room-not-found-modal .modal-header h3 {
  margin: 0 0 16px 0;
  color: #333;
  font-size: 1.5em;
  font-weight: 600;
}

.room-not-found-modal .modal-body p {
  margin: 0 0 20px 0;
  color: #666;
  line-height: 1.6;
}

.room-not-found-modal .modal-body strong {
  color: #eeaa67;
  font-weight: 600;
}

.room-not-found-modal .modal-footer {
  margin-top: 20px;
}

.room-not-found-modal .modal-btn {
  padding: 12px 24px;
  background: #eeaa67;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 1em;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.room-not-found-modal .modal-btn:hover {
  background: #e69c55;
  transform: translateY(-1px);
}

/* 按钮禁用状态 */
.btn-primary:disabled,
.btn-secondary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none !important;
}

/* 输入框禁用状态 */
.form-input:disabled {
  background: #f8f9fa;
  cursor: not-allowed;
  opacity: 0.7;
}

/* 输入框焦点状态 */
.form-input:focus {
  outline: none;
  border-color: #eeaa67;
  box-shadow: 0 0 0 3px rgba(238, 170, 103, 0.1);
}

.join-room-container {
  min-height: 100vh;
  background: linear-gradient(135deg, #fefaf5 0%, #fff5eb 100%);
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
  max-width: 1200px;
  margin: 0 auto;
  padding: 60px 5%;
}

.join-room-header {
  text-align: center;
  margin-bottom: 60px;
}

.page-title {
  font-size: 3em;
  color: #333;
  font-weight: 700;
  margin: 0 0 16px 0;
  background: linear-gradient(135deg, #eeaa67, #ff8c42);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.page-subtitle {
  font-size: 1.2em;
  color: #666;
  margin: 0;
}

/* 加入方式网格布局 */
.join-methods-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 40px;
  align-items: start;
}

.join-method-card {
  background: white;
  border-radius: 20px;
  padding: 40px;
  box-shadow: 0 10px 40px rgba(238, 170, 103, 0.1);
  border: 1px solid #ffe4cc;
  transition: all 0.3s ease;
  height: fit-content;
}

.join-method-card:hover {
  transform: translateY(-5px);
  box-shadow: 0 20px 60px rgba(238, 170, 103, 0.15);
}

/* 左侧：输入加入码样式 */
.code-join-section {
  border-top: 4px solid #eeaa67;
}

/* 右侧：快速加入样式 */
.quick-join-section {
  border-top: 4px solid #66bb6a;
}

.method-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 32px;
}

.method-icon {
  font-size: 2.5em;
}

.method-title {
  font-size: 1.8em;
  color: #333;
  margin: 0;
  font-weight: 600;
}

.method-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

/* 表单样式 */
.form-group {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.form-label {
  font-size: 1.1em;
  color: #333;
  font-weight: 500;
}

.form-input {
  padding: 16px 20px;
  border: 2px solid #e0e0e0;
  border-radius: 12px;
  font-size: 1.1em;
  transition: all 0.3s ease;
  background: white;
}

.form-input:focus {
  outline: none;
  border-color: #eeaa67;
  box-shadow: 0 0 0 4px rgba(238, 170, 103, 0.1);
  transform: translateY(-1px);
}

/* 按钮组 */
.button-group {
  display: flex;
  gap: 16px;
}

.btn-primary, .btn-secondary {
  flex: 1;
  padding: 16px 24px;
  border: none;
  border-radius: 12px;
  font-size: 1.1em;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  text-align: center;
}

.btn-primary {
  background: linear-gradient(135deg, #eeaa67, #f5b877);
  color: white;
  box-shadow: 0 4px 15px rgba(238, 170, 103, 0.3);
}

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(238, 170, 103, 0.4);
}

.btn-secondary {
  background: white;
  color: #666;
  border: 2px solid #e0e0e0;
}

.btn-secondary:hover {
  background: #f8f9fa;
  border-color: #ccc;
  transform: translateY(-2px);
}

/* 创建链接 */
.create-link {
  text-align: center;
  padding-top: 20px;
  border-top: 1px solid #f0f0f0;
  color: #666;
}

.create-link a {
  color: #eeaa67;
  text-decoration: none;
  font-weight: 600;
  margin-left: 8px;
  cursor: pointer;
  transition: color 0.3s ease;
}

.create-link a:hover {
  color: #e69c55;
  text-decoration: underline;
}

/* 动态内容占位符 */
.dynamic-content-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
  background: #f8f9fa;
  border-radius: 12px;
  border: 2px dashed #e0e0e0;
}

.placeholder-icon {
  font-size: 3em;
  margin-bottom: 20px;
  opacity: 0.5;
}

.placeholder-text {
  font-size: 1.2em;
  color: #666;
  margin: 0 0 8px 0;
  font-weight: 500;
}

.placeholder-subtext {
  font-size: 0.9em;
  color: #999;
  margin: 0;
}

/* 响应式设计 */
@media (max-width: 968px) {
  .join-methods-grid {
    grid-template-columns: 1fr;
    gap: 30px;
  }
  
  .join-method-card {
    padding: 30px;
  }
  
  .method-header {
    flex-direction: column;
    text-align: center;
    gap: 12px;
  }
  
  .button-group {
    flex-direction: column;
  }
}

@media (max-width: 768px) {
  .main-content {
    padding: 40px 20px;
  }
  
  .page-title {
    font-size: 2.5em;
  }
  
  .join-method-card {
    padding: 24px;
  }
  
  .navbar {
    padding: 12px 20px;
  }
}
</style>