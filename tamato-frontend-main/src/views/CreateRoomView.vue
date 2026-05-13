<template>
  <div class="create-room-view">
    <!-- 顶部导航栏 -->
    <nav class="navbar">
      <div class="nav-brand">Tomato</div>
      <div class="nav-links">
        <a class="nav-link" @click="goToHome">返回首页</a>
      </div>
    </nav>

    <!-- 主要内容区域 -->
    <main class="main-content">
      <div class="create-room-header">
        <h1 class="page-title">创建自习室</h1>
      </div>

      <div class="create-room-content">
        <!-- 左侧表单区域 -->
        <div class="form-section">
          <h2 class="section-title">基本信息：</h2>
          <br>
          <br>
          
          <div class="form-group">
            <label class="form-label">自习室名称：</label>
            <input 
              type="text" 
              v-model="roomForm.room_name" 
              placeholder="请输入自习室名称"
              class="form-input"
              :disabled="loading"
            >
          </div>

          <div class="form-group">
            <label class="form-label">自习室简介：</label>
            <textarea 
              v-model="roomForm.description" 
              placeholder="请输入自习室简介"
              class="form-textarea"
              rows="3"
              :disabled="loading"
            ></textarea>
          </div>

          <div class="form-group">
            <label class="form-label">最大人数：</label>
            <div class="select-wrapper">
              <select v-model="roomForm.max_members" class="form-select" :disabled="loading">
                <option value="1">1人</option>
                <option value="2">2人</option>
                <option value="3">3人</option>
                <option value="4">4人</option>
                <option value="5">5人</option>
                <option value="6">6人</option>
              </select>
              <span class="select-arrow">↓</span>
            </div>
          </div>

          <!-- 新增音乐选择字段 -->
          <div class="form-group">
            <label class="form-label">背景音乐：</label>
            <div class="select-wrapper">
              <select v-model="roomForm.music_name" class="form-select" :disabled="loading">
                <option value="无">无背景音乐</option>
                <option value="轻音乐">轻音乐</option>
                <option value="白噪音">白噪音</option>
                <option value="自然声">自然声</option>
                <option value="古典音乐">古典音乐</option>
              </select>
              <span class="select-arrow">↓</span>
            </div>
          </div>
        </div>

        <!-- 右侧操作按钮区域 -->
        <div class="action-section">
          <div class="button-group">
            <button @click="createRoom" class="confirm-btn" :disabled="loading">
              {{ loading ? '创建中...' : '确认创建' }}
            </button>
            <button @click="goToHome" class="cancel-btn" :disabled="loading">
              返回首页
            </button>
          </div>
          
          <!-- 创建提示 -->
          <div class="action-tips">
            <p>创建后可以邀请好友一起学习</p>
            <p>自习室信息随时可以修改</p>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script>
import { createRoom } from '@/api/studyRooms'
import { getCurrentUser } from '@/api/user'  // 导入获取用户信息的API

export default {
  name: 'CreateRoomView',
  data() {
    return {
      loading: false,
      currentUser: null,  // 添加当前用户信息
      roomForm: {
        room_name: '',
        description: '',
        max_members: '4',
        music_name: '无'
      }
    }
  },
  async mounted() {
    // 页面加载时获取当前用户信息
    await this.loadCurrentUser()
  },
  methods: {
    // 加载当前用户信息
    async loadCurrentUser() {
      try {
        const response = await getCurrentUser()
        
        if (response.success && response.data) {
          this.currentUser = response.data
          console.log('获取到当前用户:', this.currentUser)
        } else {
          console.warn('获取用户信息失败:', response)
          // 如果获取失败，使用默认值
          this.setDefaultUser()
        }
      } catch (error) {
        console.error('获取用户信息时出错:', error)
        this.setDefaultUser()
      }
    },
    
    // 设置默认用户信息（备用）
    setDefaultUser() {
      this.currentUser = {
        id: 'user_unknown',
        username: '未知用户',
        email: 'unknown@example.com'
      }
    },
    
    async createRoom() {
      if (!this.roomForm.room_name.trim()) {
        alert('请输入自习室名称')
        return
      }
      
      // 检查是否有用户信息
      if (!this.currentUser) {
        alert('无法获取用户信息，请重新登录')
        this.$router.push('/login')
        return
      }

      this.loading = true

      try {
        // ✅ 修正：前端不生成roomId，让后端生成
        const requestData = {
          roomName: this.roomForm.room_name.trim(),
          createPerson: this.currentUser.id || this.currentUser.user_id, // 应该是数字ID
          maxMembers: parseInt(this.roomForm.max_members),
          // 添加其他可选字段
          description: this.roomForm.description.trim() || '',
          musicName: this.roomForm.music_name || '无'
          // 不传 roomId，让后端生成
        }

        console.log('发送创建请求:', requestData)
        
        const response = await createRoom(requestData)
        console.log('创建响应结果:', response)
        
        // 处理响应
        if (response.code === 200 || response.success) {
          // ✅ 关键：从后端响应中获取roomId
          // 尝试多种可能的字段名
          const roomId = response.data?.roomId || 
                        response.data?.room_id || 
                        response.data?.roomID
          
          if (!roomId) {
            console.error('后端未返回roomId，响应数据:', response.data)
            alert('创建成功，但未获取到房间ID')
            return
          }
          
          const roomName = response.data?.roomName || 
                          response.data?.room_name || 
                          requestData.roomName
          
          console.log('创建成功，房间信息:', { roomId, roomName })
          
          alert(`自习室创建成功！\n房间名称: ${roomName}\n房间ID: ${roomId}`)
          
          // ✅ 关键：使用后端返回的roomId进行跳转
          // 确保roomId是字符串
          const roomIdStr = String(roomId).trim()
          
          if (roomIdStr && roomIdStr !== 'undefined' && roomIdStr !== 'null') {
            console.log('准备跳转到自习室，roomId:', roomIdStr)
            
            // 尝试不同的跳转方式
            try {
              // 方式1：使用路由名称和params
              this.$router.push({
                name: 'study-room',
                params: { roomId: roomIdStr }
              })
            } catch (routeError) {
              console.error('路由跳转失败（方式1）:', routeError)
              
              // 方式2：使用完整路径
              this.$router.push(`/study-room/${roomIdStr}`)
            }
          } else {
            console.error('roomId无效:', roomId)
            alert('房间ID无效，无法跳转')
          }
        } else {
          const errorMessage = response.message || '创建失败，请稍后重试'
          alert(`创建失败: ${errorMessage}`)
        }
        
      } catch (error) {
        console.error('请求异常:', error)
        
        // 详细的错误诊断
        if (error.message) {
          console.error('错误信息:', error.message)
          
          if (error.message.includes('Missing required param')) {
            console.error('路由参数缺失，可能的原因：')
            console.error('1. 路由配置不正确')
            console.error('2. roomId参数传递有问题')
            console.error('3. 路由名称不正确')
            
            // 检查路由配置
            console.log('当前路由配置:', this.$router.options.routes)
            
            alert(`路由错误: ${error.message}\n请检查控制台查看详细信息`)
          } else if (error.message.includes('Cannot find module')) {
            alert('路由组件未找到，请检查路由配置')
          }
        }
        
        // 显示用户友好的错误信息
        if (error.response) {
          alert(`服务器错误: ${error.response.status} - ${error.response.data?.message || '未知错误'}`)
        } else if (error.request) {
          alert('网络错误，请检查网络连接')
        } else {
          alert(`创建失败: ${error.message || '未知错误'}`)
        }
        
      } finally {
        this.loading = false
      }
    },
    
    goToHome() {
      this.$router.push('/home')
    }
  }
}
</script>

<style scoped>
/* 保持你原有的样式不变，只添加loading相关样式 */
.form-input:disabled,
.form-textarea:disabled,
.form-select:disabled {
  background-color: #f8f9fa;
  cursor: not-allowed;
  opacity: 0.7;
}

.confirm-btn:disabled,
.cancel-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none !important;
}


.create-room-view {
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

.create-room-header {
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

/* 创建自习室内容布局 */
.create-room-content {
  display: grid;
  grid-template-columns: 1fr 400px;
  gap: 40px;
  align-items: start;
}

/* 左侧表单区域 */
.form-section {
  background: white;
  border-radius: 20px;
  padding: 40px;
  box-shadow: 0 10px 40px rgba(238, 170, 103, 0.1);
  border: 1px solid #ffe4cc;
  border-top: 4px solid #eeaa67;
  transition: all 0.3s ease;
}

.form-section:hover {
  transform: translateY(-2px);
  box-shadow: 0 15px 50px rgba(238, 170, 103, 0.15);
}

.form-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 32px;
}

.form-icon {
  font-size: 2em;
}

.section-title {
  font-size: 1.8em;
  color: #333;
  margin: 0;
  font-weight: 600;
}

.form-group {
  margin-bottom: 30px;
}

.form-label {
  display: block;
  margin-bottom: 12px;
  color: #333;
  font-weight: 500;
  font-size: 1.1em;
}

.form-input,
.form-textarea,
.form-select {
  width: 100%;
  padding: 16px 20px;
  border: 2px solid #e0e0e0;
  border-radius: 12px;
  font-size: 1em;
  transition: all 0.3s ease;
  box-sizing: border-box;
  background: white;
  color: #333;
  font-family: inherit;
}

.form-input:focus,
.form-textarea:focus,
.form-select:focus {
  outline: none;
  border-color: #eeaa67;
  box-shadow: 0 0 0 4px rgba(238, 170, 103, 0.1);
  transform: translateY(-1px);
}

.form-input::placeholder,
.form-textarea::placeholder {
  color: #999;
}

.form-textarea {
  resize: vertical;
  min-height: 100px;
  line-height: 1.5;
}

/* 下拉选择器样式 */
.select-wrapper {
  position: relative;
  display: inline-block;
  width: 200px;
}

.form-select {
  appearance: none;
  cursor: pointer;
  padding-right: 50px;
}

.select-arrow {
  position: absolute;
  right: 20px;
  top: 50%;
  transform: translateY(-50%);
  pointer-events: none;
  color: #666;
  font-size: 1.2em;
}

/* 右侧操作按钮区域 */
.action-section {
  position: sticky;
  top: 100px;
}

.action-card {
  background: white;
  border-radius: 20px;
  padding: 30px;
  box-shadow: 0 10px 40px rgba(238, 170, 103, 0.1);
  border: 1px solid #ffe4cc;
  border-top: 4px solid #66bb6a;
}

.action-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 25px;
}

.action-icon {
  font-size: 1.8em;
}

.action-title {
  font-size: 1.4em;
  color: #333;
  margin: 0;
  font-weight: 600;
}

.button-group {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-bottom: 25px;
}

.confirm-btn,
.cancel-btn {
  padding: 16px 24px;
  border: none;
  border-radius: 12px;
  font-size: 1.1em;
  cursor: pointer;
  transition: all 0.3s ease;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
}

.confirm-btn {
  background: linear-gradient(135deg, #eeaa67, #f5b877);
  color: white;
  box-shadow: 0 4px 15px rgba(238, 170, 103, 0.3);
}

.confirm-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(238, 170, 103, 0.4);
  background: linear-gradient(135deg, #e69c55, #f0ae65);
}

.cancel-btn {
  background: white;
  color: #666;
  border: 2px solid #e0e0e0;
}

.cancel-btn:hover {
  background: #f8f9fa;
  border-color: #ccc;
  color: #333;
  transform: translateY(-2px);
}

.btn-icon {
  font-size: 1.2em;
}

.action-tips {
  background: #f8f9fa;
  padding: 20px;
  border-radius: 12px;
  border-left: 4px solid #eeaa67;
}

.action-tips p {
  margin: 8px 0;
  color: #666;
  font-size: 0.9em;
  line-height: 1.4;
}

.action-tips p:last-child {
  margin-bottom: 0;
}

/* 响应式设计 */
@media (max-width: 968px) {
  .create-room-content {
    grid-template-columns: 1fr;
    gap: 30px;
  }
  
  .action-section {
    position: static;
  }
  
  .form-section,
  .action-card {
    padding: 30px;
  }
}

@media (max-width: 768px) {
  .main-content {
    padding: 40px 20px;
  }
  
  .page-title {
    font-size: 2.5em;
  }
  
  .form-section,
  .action-card {
    padding: 24px;
  }
  
  .navbar {
    padding: 12px 20px;
  }
  
  .select-wrapper {
    width: 100%;
  }
  
  .form-header {
    flex-direction: column;
    text-align: center;
    gap: 12px;
  }
  
  .action-header {
    flex-direction: column;
    text-align: center;
    gap: 8px;
  }
}
</style>
