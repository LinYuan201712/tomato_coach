<template>
  <div class="privacy-settings">
    <div v-if="loading" class="loading-container">
      <div class="loading-spinner"></div>
      <p>加载中...</p>
    </div>

    <div v-else class="settings-content">
      <div v-if="successMessage" class="success-message">{{ successMessage }}</div>
      <div v-if="errorMessage" class="error-message">{{ errorMessage }}</div>

      <!-- 信息可见性设置 -->
      <div class="settings-section">
        <h3 class="section-title">信息可见性</h3>
        <p class="section-description">控制哪些信息对他人可见</p>

        <div class="setting-item">
          <div class="setting-label">
            <label>生日可见性</label>
            <span class="setting-desc">控制谁可以看到您的生日</span>
          </div>
          <select v-model="privacyForm.show_birthday" class="form-select">
            <option value="public">所有人可见</option>
            <option value="friends">仅好友可见</option>
            <option value="private">仅自己可见</option>
          </select>
        </div>

        <div class="setting-item">
          <div class="setting-label">
            <label>地区可见性</label>
            <span class="setting-desc">控制谁可以看到您的地区信息</span>
          </div>
          <select v-model="privacyForm.show_location" class="form-select">
            <option value="public">所有人可见</option>
            <option value="friends">仅好友可见</option>
            <option value="private">仅自己可见</option>
          </select>
        </div>
      </div>

      <!-- 社交设置 -->
      <div class="settings-section">
        <h3 class="section-title">社交设置</h3>
        <p class="section-description">控制您的社交功能</p>

        <div class="setting-item toggle-item">
          <div class="setting-label">
            <label>允许好友申请</label>
            <span class="setting-desc">允许其他用户向您发送好友申请</span>
          </div>
          <label class="toggle-switch">
            <input 
              type="checkbox" 
              v-model="privacyForm.allow_friend_request"
            />
            <span class="toggle-slider"></span>
          </label>
        </div>

        <div class="setting-item toggle-item">
          <div class="setting-label">
            <label>允许被搜索</label>
            <span class="setting-desc">允许其他用户通过搜索找到您</span>
          </div>
          <label class="toggle-switch">
            <input 
              type="checkbox" 
              v-model="privacyForm.searchable"
            />
            <span class="toggle-slider"></span>
          </label>
        </div>
      </div>

      <!-- 操作按钮 -->
      <div class="settings-actions">
        <button 
          class="btn-save" 
          @click="saveSettings" 
          :disabled="isSaving"
        >
          {{ isSaving ? '保存中...' : '保存设置' }}
        </button>
        <button 
          class="btn-reset" 
          @click="resetSettings" 
          :disabled="isSaving"
        >
          重置
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import { getCurrentUserPrivacy, updateCurrentUserPrivacy } from '@/api/user'

export default {
  name: 'UserSettings',
  data() {
    return {
      loading: true,
      isSaving: false,
      successMessage: '',
      errorMessage: '',
      privacyForm: {
        show_birthday: 'public',
        show_location: 'public',
        allow_friend_request: true,
        searchable: true
      },
      originalForm: null // 保存原始数据用于重置
    }
  },
  created() {
    this.fetchPrivacySettings()
  },
  methods: {
    async fetchPrivacySettings() {
      this.loading = true
      this.errorMessage = ''
      
      try {
        const result = await getCurrentUserPrivacy()
        
        // 后端可能返回 {code, message, data} 或 {success, data} 格式
        let data = null
        if (result.code === 200 && result.data) {
          data = result.data
        } else if (result.success && result.data) {
          data = result.data
        } else if (result.data) {
          data = result.data
        }
        
        if (data) {
          this.privacyForm = {
            show_birthday: data.show_birthday || 'public',
            show_location: data.show_location || 'public',
            allow_friend_request: data.allow_friend_request !== false, // 默认为true
            searchable: data.searchable !== false // 默认为true
          }
          // 保存原始数据
          this.originalForm = JSON.parse(JSON.stringify(this.privacyForm))
        } else {
          // 如果获取失败，使用默认值
          this.originalForm = JSON.parse(JSON.stringify(this.privacyForm))
        }
      } catch (error) {
        console.error('获取隐私设置失败:', error)
        this.errorMessage = '获取隐私设置失败，请刷新重试'
        // 使用默认值
        this.originalForm = JSON.parse(JSON.stringify(this.privacyForm))
      } finally {
        this.loading = false
      }
    },

    async saveSettings() {
      this.isSaving = true
      this.successMessage = ''
      this.errorMessage = ''

      try {
        const result = await updateCurrentUserPrivacy(this.privacyForm)

        // 后端可能返回 {code, message, data} 或 {success, data} 格式
        if ((result.code === 200) || (result.success === true)) {
          this.successMessage = '隐私设置已保存'
          // 更新原始数据
          this.originalForm = JSON.parse(JSON.stringify(this.privacyForm))
          // 3秒后清除成功消息
          setTimeout(() => {
            this.successMessage = ''
          }, 3000)
        } else {
          this.errorMessage = result.message || '保存失败，请重试'
        }
      } catch (error) {
        console.error('保存隐私设置失败:', error)
        this.errorMessage = '保存失败，请检查网络连接后重试'
      } finally {
        this.isSaving = false
      }
    },

    resetSettings() {
      if (this.originalForm) {
        this.privacyForm = JSON.parse(JSON.stringify(this.originalForm))
        this.successMessage = ''
        this.errorMessage = ''
      }
    }
  }
}
</script>

<style scoped>
.privacy-settings {
  padding: 20px;
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 0;
  color: #666;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 4px solid #f3f3f3;
  border-top: 4px solid #cc2a1f;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 20px;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.settings-content {
  max-width: 800px;
}

.success-message {
  background: linear-gradient(135deg, #4caf50, #66bb6a);
  color: white;
  padding: 12px 20px;
  border-radius: 8px;
  margin-bottom: 20px;
  box-shadow: 0 2px 8px rgba(76, 175, 80, 0.3);
  animation: slideIn 0.3s ease;
}

.error-message {
  background: linear-gradient(135deg, #f44336, #ef5350);
  color: white;
  padding: 12px 20px;
  border-radius: 8px;
  margin-bottom: 20px;
  box-shadow: 0 2px 8px rgba(244, 67, 54, 0.3);
  animation: slideIn 0.3s ease;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.settings-section {
  background: white;
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  border: 1px solid #f0f0f0;
}

.section-title {
  font-size: 1.3em;
  font-weight: bold;
  color: #333;
  margin: 0 0 8px 0;
}

.section-description {
  color: #666;
  font-size: 0.9em;
  margin: 0 0 20px 0;
}

.setting-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 0;
  border-bottom: 1px solid #f0f0f0;
}

.setting-item:last-child {
  border-bottom: none;
}

.setting-label {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.setting-label label {
  font-weight: 500;
  color: #333;
  font-size: 1em;
}

.setting-desc {
  color: #888;
  font-size: 0.85em;
}

.form-select {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 1em;
  min-width: 150px;
  height: 40px;
  outline: none;
  transition: border-color 0.2s;
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%23333' d='M6 9L1 4h10z'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
  padding-right: 35px;
}

.form-select:focus {
  border-color: #cc2a1f;
}

/* 开关样式 */
.toggle-item {
  padding: 20px 0;
}

.toggle-switch {
  position: relative;
  display: inline-block;
  width: 50px;
  height: 26px;
  cursor: pointer;
}

.toggle-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.toggle-slider {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #ccc;
  transition: 0.3s;
  border-radius: 26px;
}

.toggle-slider:before {
  position: absolute;
  content: "";
  height: 20px;
  width: 20px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: 0.3s;
  border-radius: 50%;
}

input:checked + .toggle-slider {
  background-color: #cc2a1f;
}

input:checked + .toggle-slider:before {
  transform: translateX(24px);
}

.toggle-switch:hover .toggle-slider {
  box-shadow: 0 0 8px rgba(204, 42, 31, 0.3);
}

/* 操作按钮 */
.settings-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 30px;
  padding-top: 20px;
  border-top: 1px solid #f0f0f0;
}

.btn-save,
.btn-reset {
  padding: 10px 24px;
  border: none;
  border-radius: 6px;
  font-size: 1em;
  cursor: pointer;
  transition: all 0.3s;
  font-weight: 500;
}

.btn-save {
  background: #cc2a1f;
  color: white;
}

.btn-save:hover:not(:disabled) {
  background: #b52217;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(204, 42, 31, 0.3);
}

.btn-reset {
  background: #f5f5f5;
  color: #666;
  border: 1px solid #ddd;
}

.btn-reset:hover:not(:disabled) {
  background: #e0e0e0;
}

button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none !important;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .setting-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .form-select {
    width: 100%;
  }

  .settings-actions {
    flex-direction: column;
  }

  .btn-save,
  .btn-reset {
    width: 100%;
  }
}
</style>
