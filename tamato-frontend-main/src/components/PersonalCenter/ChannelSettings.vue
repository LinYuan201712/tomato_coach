<template>
  <div class="channel-settings">
    <div v-if="loading" class="loading-container">
      <div class="loading-spinner"></div>
      <p>加载中...</p>
    </div>

    <div v-else class="settings-content">
      <div v-if="successMessage" class="success-message">{{ successMessage }}</div>
      <div v-if="errorMessage" class="error-message">{{ errorMessage }}</div>

      <!-- 飞书集成设置 -->
      <div class="settings-section">
        <div class="section-header">
          <div class="title-wrap">
            <h3 class="section-title">飞书机器人集成</h3>
            <p class="section-description">连接您的飞书机器人，实现多端同步消息和提醒</p>
          </div>
          <div class="status-badge" :class="{ active: feishuForm.enabled }">
            {{ feishuForm.enabled ? '已启用' : '已禁用' }}
          </div>
        </div>

        <div class="setting-item toggle-item">
          <div class="setting-label">
            <label>启用 WebSocket 连接</label>
            <span class="setting-desc">开启后，后端将尝试连接飞书服务器，以便实时接收和发送消息</span>
          </div>
          <label class="toggle-switch">
            <input 
              type="checkbox" 
              v-model="feishuForm.enabled"
            />
            <span class="toggle-slider"></span>
          </label>
        </div>

        <div class="form-grid">
          <div class="form-item">
            <label>App ID</label>
            <input type="text" v-model="feishuForm.app_id" placeholder="cli_xxxxxxxx" class="form-input" />
          </div>
          <div class="form-item">
            <label>App Secret</label>
            <input type="password" v-model="feishuForm.app_secret" placeholder="请输入密钥" class="form-input" />
          </div>
          <div class="form-item">
            <label>Verification Token</label>
            <input type="text" v-model="feishuForm.verification_token" placeholder="校验令牌 (可选)" class="form-input" />
          </div>
          <div class="form-item">
            <label>Encrypt Key</label>
            <input type="text" v-model="feishuForm.encrypt_key" placeholder="加密密钥 (可选)" class="form-input" />
          </div>
        </div>

        <div class="help-box">
          <p><strong>💡 如何获取？</strong></p>
          <ol>
            <li>前往 <a href="https://open.feishu.cn/" target="_blank">飞书开放平台</a> 创建自建应用。</li>
            <li>在“凭据与基础信息”中获取 App ID 和 App Secret。</li>
            <li>在“机器人”功能中开启机器人能力。</li>
            <li><strong>注意：</strong> 确保已开启“接收消息”权限并设置为 WebSocket 模式。</li>
          </ol>
        </div>
      </div>

      <!-- 操作按钮 -->
      <div class="settings-actions">
        <button 
          class="btn-save" 
          @click="saveSettings" 
          :disabled="isSaving"
        >
          {{ isSaving ? '保存并应用中...' : '保存并应用' }}
        </button>
        <button 
          class="btn-reset" 
          @click="fetchSettings" 
          :disabled="isSaving"
        >
          取消修改
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import { getFeishuConfig, updateFeishuConfig } from '@/api/user'

export default {
  name: 'ChannelSettings',
  data() {
    return {
      loading: true,
      isSaving: false,
      successMessage: '',
      errorMessage: '',
      feishuForm: {
        enabled: false,
        app_id: '',
        app_secret: '',
        verification_token: '',
        encrypt_key: ''
      }
    }
  },
  created() {
    this.fetchSettings()
  },
  methods: {
    async fetchSettings() {
      this.loading = true
      this.errorMessage = ''
      this.successMessage = ''
      
      try {
        const res = await getFeishuConfig()
        if (res.id) {
          this.feishuForm = {
            enabled: res.enabled || false,
            app_id: res.app_id || '',
            app_secret: res.app_secret || '',
            verification_token: res.verification_token || '',
            encrypt_key: res.encrypt_key || ''
          }
        }
      } catch (error) {
        console.error('获取飞书配置失败:', error)
        this.errorMessage = '无法获取飞书配置，请检查后端连接'
      } finally {
        this.loading = false
      }
    },

    async saveSettings() {
      if (this.feishuForm.enabled && (!this.feishuForm.app_id || !this.feishuForm.app_secret)) {
        this.errorMessage = '启用飞书连接需要填写 App ID 和 App Secret'
        return
      }

      this.isSaving = true
      this.successMessage = ''
      this.errorMessage = ''

      try {
        await updateFeishuConfig(this.feishuForm)
        this.successMessage = '配置已保存，正在尝试连接飞书...'
        setTimeout(() => {
          this.successMessage = ''
        }, 5000)
      } catch (error) {
        console.error('保存飞书配置失败:', error)
        this.errorMessage = '保存失败，请重试'
      } finally {
        this.isSaving = false
      }
    }
  }
}
</script>

<style scoped>
.channel-settings {
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

.settings-section {
  background: white;
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  border: 1px solid #f0f0f0;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
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
  margin: 0;
}

.status-badge {
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 0.85em;
  background: #eee;
  color: #666;
}

.status-badge.active {
  background: #e8f5e9;
  color: #2e7d32;
  border: 1px solid #c8e6c9;
}

.setting-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 0;
  border-bottom: 1px solid #f0f0f0;
}

.toggle-item {
  margin-bottom: 20px;
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
}

.setting-desc {
  color: #888;
  font-size: 0.85em;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  margin-top: 20px;
}

.form-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-item label {
  font-size: 0.9em;
  font-weight: 500;
  color: #555;
}

.form-input {
  padding: 10px 12px;
  border: 1px solid #ddd;
  border-radius: 8px;
  outline: none;
  transition: all 0.2s;
}

.form-input:focus {
  border-color: #cc2a1f;
  box-shadow: 0 0 0 2px rgba(204, 42, 31, 0.1);
}

.help-box {
  margin-top: 24px;
  padding: 16px;
  background: #fff8f1;
  border-left: 4px solid #ff9800;
  border-radius: 4px;
  font-size: 0.9em;
  color: #5d4037;
}

.help-box a {
  color: #007bff;
  text-decoration: none;
}

.success-message {
  background: #e8f5e9;
  color: #2e7d32;
  padding: 12px 20px;
  border-radius: 8px;
  margin-bottom: 20px;
  border: 1px solid #c8e6c9;
}

.error-message {
  background: #ffebee;
  color: #c62828;
  padding: 12px 20px;
  border-radius: 8px;
  margin-bottom: 20px;
  border: 1px solid #ffcdd2;
}

/* 开关样式 (复用 UserSettings.vue 里的) */
.toggle-switch {
  position: relative;
  display: inline-block;
  width: 50px;
  height: 26px;
  cursor: pointer;
}

.toggle-switch input { opacity: 0; width: 0; height: 0; }

.toggle-slider {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  background-color: #ccc;
  transition: 0.3s;
  border-radius: 26px;
}

.toggle-slider:before {
  position: absolute;
  content: "";
  height: 20px; width: 20px;
  left: 3px; bottom: 3px;
  background-color: white;
  transition: 0.3s;
  border-radius: 50%;
}

input:checked + .toggle-slider { background-color: #cc2a1f; }
input:checked + .toggle-slider:before { transform: translateX(24px); }

/* 操作按钮 */
.settings-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 10px;
}

.btn-save, .btn-reset {
  padding: 10px 24px;
  border: none;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s;
}

.btn-save { background: #cc2a1f; color: white; }
.btn-save:hover { background: #b52217; box-shadow: 0 4px 12px rgba(204, 42, 31, 0.3); }

.btn-reset { background: #f5f5f5; color: #666; border: 1px solid #ddd; }

@media (max-width: 768px) {
  .form-grid { grid-template-columns: 1fr; }
}
</style>
