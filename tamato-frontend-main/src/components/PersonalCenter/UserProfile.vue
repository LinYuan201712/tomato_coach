<template>
  <div class="user-profile-component">
    <div class="profile-header">
      <div class="header-left">
        <div 
          :class="['avatar-wrapper', { 'is-editing': isEditing }]" 
          @click="isEditing ? changeAvatar() : null"
        >
          <img :src="displayAvatar || require('@/assets/images/avatar.png')" alt="头像" class="avatar-img" @error="handleAvatarError" />
          <div v-if="isEditing" class="avatar-edit-mask">
            点击更换
          </div>
        </div>
        
        <div v-if="!isEditing" class="header-info">
          <div class="header-name">{{ userInfo.username }}</div>
          <div class="header-id">ID: {{ userInfo.userId }}</div>
        </div>
      </div>

      <div class="header-right">
        <div class="tomato-stat">
          <img src="@/assets/logo.png" alt="番茄" class="tomato-icon" />
          <span class="tomato-count">{{ userInfo.tomatoCount }}</span>
        </div>

        <div class="action-buttons">
          <button 
            v-if="!isEditing" 
            class="btn-primary" 
            @click="startEdit"
          >
            修改资料
          </button>
          <div v-else class="edit-actions">
            <div v-if="overallError" class="overall-error">{{ overallError }}</div>
            <div v-if="successMessage" class="success-message">{{ successMessage }}</div>
            
            <button class="btn-save" @click="saveEdit" :disabled="isSaving">
              {{ isSaving ? '保存中...' : '保存' }}
            </button>
            <button class="btn-cancel" @click="cancelEdit" :disabled="isSaving">取消</button>
          </div>
        </div>
      </div>
    </div>

    <hr class="divider" />

    <div class="profile-details">
      <div class="detail-item">
        <label>用户名</label>
        <div v-if="!isEditing" class="read-only-text">{{ userInfo.username }}</div>
        <div v-else>
          <input v-model="editForm.username" type="text" class="form-input" />
          <div v-if="editErrors.username" class="error-message">{{ editErrors.username }}</div>
        </div>
      </div>

      <div class="detail-item">
        <label>邮箱</label>
        <div v-if="!isEditing" class="read-only-text">{{ userInfo.email }}</div>
        <div v-else>
          <input v-model="editForm.email" type="email" class="form-input" />
          <div v-if="editErrors.email" class="error-message">{{ editErrors.email }}</div>
        </div>
      </div>
      
      <div class="detail-item">
        <label>手机号</label>
        <div v-if="!isEditing" class="read-only-text">{{ userInfo.phone }}</div>
        <input v-else v-model="editForm.phone" type="tel" class="form-input" />
      </div>

      <div class="detail-item">
        <label>性别</label>
        <div v-if="!isEditing" class="read-only-text">{{ userInfo.gender }}</div>
        <select v-else v-model="editForm.gender" class="form-input">
          <option value="男">男</option>
          <option value="女">女</option>
          <option value="保密">保密</option>
        </select>
      </div>

      <div class="detail-item">
        <label>生日</label>
        <div v-if="!isEditing" class="read-only-text">{{ userInfo.birthday }}</div>
        <div v-else class="birthday-selector">
          <select v-model="birthdayYear" class="form-input birthday-select" @change="updateBirthday">
            <option value="">年</option>
            <option v-for="year in years" :key="year" :value="year">{{ year }}</option>
          </select>
          <span class="birthday-separator">年</span>
          <select v-model="birthdayMonth" class="form-input birthday-select" @change="updateBirthday">
            <option value="">月</option>
            <option v-for="month in months" :key="month" :value="String(month)">{{ month }}</option>
          </select>
          <span class="birthday-separator">月</span>
          <select v-model="birthdayDay" class="form-input birthday-select" @change="updateBirthday">
            <option value="">日</option>
            <option v-for="day in availableDays" :key="day" :value="String(day)">{{ day }}</option>
          </select>
          <span class="birthday-separator">日</span>
        </div>
      </div>

      <div class="detail-item">
        <label>地区</label>
        <div v-if="!isEditing" class="read-only-text">{{ userInfo.region }}</div>
        <select v-else v-model="editForm.region" class="form-input region-select">
          <option value="">请选择地区</option>
          <option v-for="province in provinces" :key="province" :value="province">
            {{ province }}
          </option>
        </select>
      </div>

    </div>
  </div>
</template>

<script>
import { updateCurrentUser } from '@/api/user';

export default {
  name: 'UserProfile',
  props: {
    userInfo: {
      type: Object,
      required: true,
      default: () => ({
        userId: 'N/A', 
        username: 'N/A',
        avatar: '',
        gender: '保密',
        birthday: '',
        region: '',
        phone: '',
        email: '',
        tomatoCount: 0, 
        password_hash: '', 
      })
    },
  },
  data() {
    return {
      isEditing: false, 
      isSaving: false,
      overallError: '',
      successMessage: '',
      avatarPreview: null, // 用于存储头像预览（不直接修改prop）
      editErrors: {
        username: '',
        email: ''
      },
      editForm: {},
      // 生日选择器的年月日（用于三个下拉框）
      birthdayYear: '',
      birthdayMonth: '',
      birthdayDay: '',
      // 中国省份列表
      provinces: [
        '北京市', '天津市', '河北省', '山西省', '内蒙古自治区',
        '辽宁省', '吉林省', '黑龙江省', '上海市', '江苏省',
        '浙江省', '安徽省', '福建省', '江西省', '山东省',
        '河南省', '湖北省', '湖南省', '广东省', '广西壮族自治区',
        '海南省', '重庆市', '四川省', '贵州省', '云南省',
        '西藏自治区', '陕西省', '甘肃省', '青海省', '宁夏回族自治区',
        '新疆维吾尔自治区', '香港特别行政区', '澳门特别行政区', '台湾省'
      ]
    }
  },
  computed: {
    // 生成年份列表（从当前年份往前100年）
    years() {
      const currentYear = new Date().getFullYear()
      const years = []
      for (let i = currentYear; i >= currentYear - 100; i--) {
        years.push(i)
      }
      return years
    },
    // 生成月份列表（1-12）
    months() {
      return Array.from({ length: 12 }, (_, i) => i + 1)
    },
    // 根据选择的年月计算可用的日期
    availableDays() {
      if (!this.birthdayYear || !this.birthdayMonth) {
        return []
      }
      const year = parseInt(this.birthdayYear)
      const month = parseInt(this.birthdayMonth)
      // 获取该月的天数
      const daysInMonth = new Date(year, month, 0).getDate()
      return Array.from({ length: daysInMonth }, (_, i) => i + 1)
    },
    // 计算属性：优先使用预览头像，否则使用prop中的头像
    displayAvatar() {
      // 使用prop中的头像（服务器返回的最新数据）
      const avatar = this.userInfo.avatar
      
      // 如果有预览头像，优先使用预览（上传后立即显示）
      // 当父组件更新prop后，avatarPreview会被清除，然后使用prop中的avatar
      if (this.avatarPreview) {
        console.log('displayAvatar: 使用预览头像（等待prop更新）', this.avatarPreview)
        return this.avatarPreview
      }
      
      // 使用prop中的头像
      console.log('displayAvatar: 使用prop中的头像', avatar)
      
      if (!avatar) {
        // 如果没有头像，返回默认头像
        console.log('displayAvatar: 没有头像，使用默认头像')
        return require('@/assets/images/avatar.png')
      }
      
      // 如果是data URL（base64），直接返回
      if (avatar.startsWith('data:')) {
        return avatar
      }
      
      // 如果已经是完整URL，直接返回
      if (avatar.startsWith('http://') || avatar.startsWith('https://')) {
        return avatar
      }
      
      // 如果是相对路径，需要转换为完整URL
      if (avatar.startsWith('/')) {
        // 无论是开发环境还是生产环境，都使用完整URL
        // 因为图片资源需要通过完整的URL访问
        // 注意：如果路径已经是 /api/ 开头，直接拼接；否则需要添加 /api
        let fullUrl
        if (avatar.startsWith('/api/')) {
          // 路径已经是 /api/ 开头，直接拼接基础URL
          fullUrl = `http://localhost:8090${avatar}`
        } else {
          // 路径不是 /api/ 开头，需要添加
          fullUrl = `http://localhost:8090/api${avatar}`
        }
        console.log('构建头像完整URL:', fullUrl, '原始路径:', avatar)
        return fullUrl
      }
      
      // 其他情况直接返回
      return avatar
    }
  },
  methods: {
    isValidEmail(email) {
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
      return emailRegex.test(email)
    },

    validateForm() {
      this.editErrors = { username: '', email: '' }
      this.overallError = ''
      let isValid = true
      
      const { username, email } = this.editForm

      // 1. 用户名验证 (3-10位)
      if (!username) {
        this.editErrors.username = '请输入用户名'
        isValid = false
      } else if (username.length < 3 || username.length > 10) {
        this.editErrors.username = '用户名长度应为3-10位'
        isValid = false
      }

      // 2. 邮箱验证
      if (!email) {
        this.editErrors.email = '请输入邮箱地址'
        isValid = false
      } else if (!this.isValidEmail(email)) {
        this.editErrors.email = '请输入有效的邮箱地址'
        isValid = false
      }

      if (!isValid) {
        this.overallError = '请检查并修正表单中的错误。'
      }

      return isValid
    },
    
    handleAvatarError(event) {
      // 头像加载失败时，使用默认头像
      console.error('头像加载失败，URL:', event.target.src)
      const defaultAvatar = require('@/assets/images/avatar.png')
      // 避免无限循环：如果已经是默认头像，不再设置
      if (!event.target.src.includes('avatar.png')) {
        event.target.src = defaultAvatar
      }
    },
    
    changeAvatar() {
      // 创建文件输入元素
      const input = document.createElement('input')
      input.type = 'file'
      input.accept = 'image/*' // 只接受图片文件
      input.style.display = 'none'
      
      // 监听文件选择 - 使用 async 箭头函数以支持 await
      input.onchange = async (event) => {
        const file = event.target.files[0]
        if (!file) {
          return
        }
        
        // 检查文件大小（限制为5MB）
        if (file.size > 5 * 1024 * 1024) {
          this.overallError = '图片大小不能超过5MB'
          setTimeout(() => { this.overallError = '' }, 3000)
          return
        }
        
        // 检查文件类型
        if (!file.type.startsWith('image/')) {
          this.overallError = '请选择图片文件'
          setTimeout(() => { this.overallError = '' }, 3000)
          return
        }
        
        // 创建FileReader来预览图片
        const reader = new FileReader()
        reader.onload = (e) => {
          // 使用本地数据属性存储预览，不直接修改prop
          // 先显示base64预览，上传成功后会替换为服务器URL
          this.avatarPreview = e.target.result
          console.log('设置base64预览头像')
        }
        reader.readAsDataURL(file)
        
        // 上传到服务器
        try {
          const { uploadAvatar } = await import('@/api/user')
          const result = await uploadAvatar(file)
          
          if (result.success) {
            console.log('头像上传成功，返回数据:', result)
            
            // 获取新头像URL
            let newAvatarUrl = null
            if (result.data) {
              // 后端返回的UserResponse中包含avatar字段
              newAvatarUrl = result.data.avatar || result.data.avatarUrl
              console.log('新头像URL:', newAvatarUrl)
            }
            
            if (newAvatarUrl && newAvatarUrl.trim() !== '') {
              // 上传成功，使用服务器返回的新头像URL作为预览
              // 这样在父组件更新数据之前，也能显示新头像
              // 如果是相对路径，转换为完整URL
              let previewUrl = null
              if (newAvatarUrl.startsWith('/')) {
                previewUrl = `http://localhost:8090${newAvatarUrl}`
              } else if (newAvatarUrl.startsWith('http://') || newAvatarUrl.startsWith('https://')) {
                previewUrl = newAvatarUrl
              } else {
                // 如果不是相对路径也不是完整URL，可能是其他格式，尝试构建完整URL
                previewUrl = `http://localhost:8090/api/uploads/avatars/${newAvatarUrl}`
              }
              
              // 直接设置，Vue会自动处理响应式更新
              this.avatarPreview = previewUrl
              console.log('✅ 设置头像预览URL:', this.avatarPreview)
              console.log('✅ avatarPreview当前值:', this.avatarPreview)
              
              // 强制触发视图更新
              this.$forceUpdate()
            } else {
              console.error('❌ 服务器返回的数据中没有有效的头像URL')
              console.error('newAvatarUrl值:', newAvatarUrl)
              console.error('result.data:', result.data)
            }
            
            this.successMessage = '头像上传成功！'
            setTimeout(() => { this.successMessage = '' }, 3000)
            
            // 通知父组件更新数据（父组件会重新获取用户信息，包含新的头像URL）
            this.$emit('update-user-info')
            
            // 等待父组件更新后，清除预览，让计算属性使用prop中的头像
            // 延迟1.5秒，确保父组件已经完成数据更新
            setTimeout(() => {
              // 检查prop中的头像是否已经更新
              const currentAvatar = this.userInfo.avatar
              if (currentAvatar && currentAvatar === newAvatarUrl) {
                // prop已经更新，清除预览，使用prop中的头像
                console.log('✅ prop已更新，清除预览，使用prop中的头像')
                this.avatarPreview = null
              } else {
                // prop还没更新，再等一会儿
                console.log('⏳ prop还未更新，继续使用预览')
                setTimeout(() => {
                  this.avatarPreview = null
                }, 1000)
              }
            }, 1500)
          } else {
            // 上传失败，清除预览
            this.avatarPreview = null
            this.overallError = result.message || '头像上传失败'
            setTimeout(() => { this.overallError = '' }, 3000)
          }
        } catch (error) {
          console.error('头像上传失败:', error)
          // 上传失败，清除预览
          this.avatarPreview = null
          this.overallError = '头像上传失败，请重试'
          setTimeout(() => { this.overallError = '' }, 3000)
        }
      }
      
      // 触发文件选择对话框
      document.body.appendChild(input)
      input.click()
      document.body.removeChild(input)
    },

    startEdit() {
      // 从 prop 复制数据到编辑表单
      this.editForm = JSON.parse(JSON.stringify(this.userInfo))
      this.editErrors = { username: '', email: '' }
      this.overallError = ''
      this.isEditing = true
      
      // 解析生日，填充到年月日下拉框
      if (this.editForm.birthday) {
        // 如果生日是 YYYY-MM-DD 格式
        if (typeof this.editForm.birthday === 'string' && this.editForm.birthday.includes('-')) {
          const parts = this.editForm.birthday.split('-')
          if (parts.length === 3) {
            this.birthdayYear = parts[0]
            this.birthdayMonth = parseInt(parts[1]).toString()
            this.birthdayDay = parseInt(parts[2]).toString()
          }
        } else if (typeof this.editForm.birthday === 'number') {
          // 如果是时间戳，转换为日期
          const date = new Date(this.editForm.birthday)
          this.birthdayYear = date.getFullYear().toString()
          this.birthdayMonth = (date.getMonth() + 1).toString()
          this.birthdayDay = date.getDate().toString()
        }
      } else {
        // 清空年月日
        this.birthdayYear = ''
        this.birthdayMonth = ''
        this.birthdayDay = ''
      }
    },
    
    cancelEdit() {
      this.isEditing = false
      this.editForm = {}
      this.editErrors = { username: '', email: '' }
      this.overallError = ''
      // 清空年月日
      this.birthdayYear = ''
      this.birthdayMonth = ''
      this.birthdayDay = ''
    },
    
    // 更新生日（当年月日改变时调用）
    updateBirthday() {
      // 确保所有值都存在且有效
      const year = this.birthdayYear
      const month = this.birthdayMonth
      const day = this.birthdayDay
      
      if (year && month && day) {
        // 格式化为 YYYY-MM-DD
        // 确保所有值都是字符串类型，然后使用 padStart
        const yearStr = String(year).trim()
        const monthStr = String(month).trim()
        const dayStr = String(day).trim()
        
        if (yearStr && monthStr && dayStr) {
          const monthPadded = monthStr.padStart(2, '0')
          const dayPadded = dayStr.padStart(2, '0')
          this.editForm.birthday = `${yearStr}-${monthPadded}-${dayPadded}`
        } else {
          this.editForm.birthday = ''
        }
      } else {
        this.editForm.birthday = ''
      }
    },
    
    async saveEdit() {
      if (!this.validateForm()) {
        return
      }

      this.isSaving = true
      this.overallError = '正在保存...'
      
      try {
        // 构造 PUT 请求体，只包含后端 UpdateUserRequest 支持的字段
        const updatePayload = {
          username: this.editForm.username,
          email: this.editForm.email,
          phone: this.editForm.phone || null,
          sex: this.editForm.gender || null,
          province: this.editForm.region || null
        }
        
        // 生日需要转换为时间戳（毫秒）
        // 前端格式：YYYY-MM-DD，需要转换为时间戳
        if (this.editForm.birthday) {
          const date = new Date(this.editForm.birthday + 'T00:00:00') // 设置为当天的0点
          updatePayload.birthday = date.getTime() // 时间戳（毫秒）
        }
        
        // 调用更新 API (PUT /me)
        const result = await updateCurrentUser(updatePayload);

        if (result.success) {
          // 更新成功，触发事件通知父组件重新获取最新数据
          this.$emit('update-user-info'); 
          
          this.isEditing = false;
          this.overallError = '';
          this.successMessage = '个人资料已更新成功！';
          
          // 3秒后清除成功提示
          setTimeout(() => {
            this.successMessage = '';
          }, 3000)
        } else {
          // 更新失败
          this.successMessage = '';
          this.overallError = result.message || '资料保存失败，请重试';
        }
      } catch (error) {
        this.overallError = '网络请求失败，请检查服务连接。'
        console.error('Save Error:', error);
      } finally {
        this.isSaving = false;
      }
    }
  }
}
</script>

<style scoped>
.user-profile-component {
  padding: 10px;
}

/* 顶部布局 */
.error-message {
  color: #cc2a1f; /* 红色 */
  font-size: 0.85em;
  margin-top: 4px;
}

.overall-error {
  color: #cc2a1f;
  background-color: #ffe0e0;
  border: 1px solid #cc2a1f;
  padding: 8px 12px;
  border-radius: 6px;
  margin-bottom: 10px;
  text-align: left;
}

.success-message {
  color: #22543d;
  background-color: #d4edda;
  border: 1px solid #c3e6cb;
  padding: 8px 12px;
  border-radius: 6px;
  margin-bottom: 10px;
  text-align: left;
}

/* 按钮样式和布局 */
.action-buttons {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.edit-actions {
  display: flex;
  flex-direction: column;
  gap: 10px;
  width: 100%;
}

.profile-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 20px;
  /* 确保在编辑模式下，左侧布局不会因为 header-info 隐藏而塌陷 */
  min-height: 80px; 
}

/* 头像区域：新增样式 */
.avatar-wrapper {
  position: relative;
  width: 80px;
  height: 80px;
  flex-shrink: 0; 
}

.avatar-wrapper.is-editing {
  cursor: pointer;
}

.avatar-img {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  object-fit: cover;
  border: 3px solid #f0f0f0;
  display: block;
}

/* 遮罩：在编辑模式下立即显示半透明灰色底，文本更明显 */
.avatar-edit-mask {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  border-radius: 50%;
  /* 半透明灰色底 */
  background: rgba(0, 0, 0, 0.5); 
  color: white;
  display: flex;
  justify-content: center;
  align-items: center;
  font-size: 0.85em; 
  opacity: 1; /* 确保在 v-if 存在时立即显示 */
  transition: background 0.2s;
  text-align: center;
  line-height: 1.2;
}

/* 优化 hover 效果 */
.avatar-wrapper.is-editing:hover .avatar-edit-mask {
  background: rgba(0, 0, 0, 0.65); 
}

/* 用户名和ID区域的样式 (仅在非编辑模式下显示) */
.header-name {
  font-size: 1.4em;
  font-weight: bold;
  color: #333;
}

.header-id {
  color: #888;
  font-size: 0.9em;
  margin-top: 4px;
}


.header-right {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 15px;
}

/* 番茄统计样式 */
.tomato-stat {
  display: flex;
  align-items: center;
  gap: 8px;
  background-color: #fff0f0; /* 淡淡的红色背景 */
  padding: 8px 16px;
  border-radius: 20px;
  border: 1px solid #ffcccc;
}

.tomato-icon {
  width: 24px;
  height: 24px;
  object-fit: contain;
}

.tomato-count {
  font-weight: bold;
  color: #cc2a1f;
  font-size: 1.1em;
}

/* 按钮样式 */
.action-buttons {
  display: flex;
  gap: 10px;
}

button {
  cursor: pointer;
  border: none;
  border-radius: 6px;
  padding: 8px 20px;
  font-size: 0.95em;
  transition: all 0.3s;
}

.btn-primary {
  background: #cc2a1f;
  color: white;
}
.btn-primary:hover {
  background: #b52217;
}

.btn-save {
  background: #4caf50;
  color: white;
}

.btn-save:hover {
  background: #43a047;
}
/* 禁用按钮样式 */
button:disabled {
  background-color: #ccc;
  cursor: not-allowed;
}

.btn-cancel {
  background: #f5f5f5;
  color: #666;
  border: 1px solid #ddd;
}
.btn-cancel:hover {
  background: #e0e0e0;
}

.divider {
  border: none;
  border-top: 1px solid #f0f0f0;
  margin: 20px 0 30px 0;
}

/* 详细资料网格布局 */
.profile-details {
  display: grid;
  grid-template-columns: repeat(2, 1fr); /* 两列显示 */
  gap: 30px 50px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.detail-item label {
  color: #888;
  font-size: 0.9em;
}

.read-only-text {
  padding: 8px 0;
  font-size: 1.05em;
  color: #333;
  border-bottom: 1px solid transparent;
  min-height: 40px;
  display: flex;
  align-items: center;
}

.form-input {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 1em;
  height: 40px;
  outline: none;
  transition: border-color 0.2s;
}

.form-input:focus {
  border-color: #cc2a1f;
}

/* 生日选择器样式（三个下拉框） */
.birthday-selector {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.birthday-select {
  flex: 1;
  min-width: 80px;
  cursor: pointer;
}

.birthday-separator {
  color: #666;
  font-size: 0.9em;
  white-space: nowrap;
}

/* 地区选择下拉框样式 */
.region-select {
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%23333' d='M6 9L1 4h10z'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
  padding-right: 35px;
}

.region-select:focus {
  border-color: #cc2a1f;
  outline: none;
}

/* 响应式 */
@media (max-width: 600px) {
  .profile-details {
    grid-template-columns: 1fr; /* 小屏幕变单列 */
    gap: 20px;
  }
  
  .profile-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 20px;
  }
  
  .header-right {
    align-items: flex-start;
    width: 100%;
    flex-direction: row;
    justify-content: space-between;
  }
}
</style>