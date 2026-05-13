<template>
  <div class="forgot-password-page">
    <div class="forgot-password-container">
      <AuthForm
        :is-login="false"
        :loading="loading"
        :error="error"
        :success-message="successMessage"
        :logo="logoUrl"
        :hide-submit-button="true"
        :is-forgot-password="true"
        class="forgot-password-form"
      >
        <template #form-fields>
          <div class="step-indicator">
            <div class="step" :class="{ active: currentStep === 1, completed: currentStep > 1 }">
              <span class="step-number">1</span>
              <span class="step-label">验证邮箱</span>
            </div>
            <div class="step-line" :class="{ completed: currentStep > 1 }"></div>
            <div class="step" :class="{ active: currentStep === 2, completed: currentStep > 2 }">
              <span class="step-number">2</span>
              <span class="step-label">重置密码</span>
            </div>
          </div>

          <!-- 步骤1: 输入邮箱和验证码 -->
          <div v-if="currentStep === 1" class="step-content">
            <FormInput
              v-model="formData.email"
              type="email"
              label="邮箱地址"
              placeholder="请输入您的注册邮箱"
              required
              :error="errors.email"
            />
            
            <div class="verification-code-group">
              <FormInput
                v-model="formData.verificationCode"
                type="text"
                label="验证码"
                placeholder="请输入邮箱验证码"
                required
                :error="errors.verificationCode"
              />
              <button
                type="button"
                class="send-code-btn"
                :disabled="!canSendCode || sendingCode || countdown > 0"
                @click="sendVerificationCode"
              >
                <span v-if="countdown > 0">{{ countdown }}秒后重试</span>
                <span v-else-if="sendingCode">发送中...</span>
                <span v-else>发送验证码</span>
              </button>
            </div>

            <button 
              type="button" 
              class="submit-btn"
              :disabled="loading || !canProceedToStep2"
              :class="{ 'loading': loading }"
              @click="proceedToStep2"
            >
              <span v-if="loading">验证中...</span>
              <span v-else>下一步</span>
            </button>
          </div>

          <!-- 步骤2: 设置新密码 -->
          <div v-if="currentStep === 2" class="step-content">
            <FormInput
              v-model="formData.newPassword"
              type="password"
              label="新密码"
              placeholder="请输入6-15位新密码"
              required
              :error="errors.newPassword"
            />
            
            <FormInput
              v-model="formData.confirmPassword"
              type="password"
              label="确认新密码"
              placeholder="请再次输入新密码"
              required
              :error="errors.confirmPassword"
            />

            <button 
              type="button" 
              class="submit-btn"
              :disabled="loading || !canResetPassword"
              :class="{ 'loading': loading }"
              @click="resetPassword"
            >
              <span v-if="loading">重置中...</span>
              <span v-else>重置密码</span>
            </button>
          </div>
        </template>
      </AuthForm>
    </div>
  </div>
</template>

<script>
import AuthForm from '@/components/Login/AuthForm.vue'
import FormInput from '@/components/Login/FormInput.vue'
import { sendResetPasswordCode, resetPassword } from '@/api/auth'
import { getFriendlyErrorMessage } from '@/utils/errorMessages'

export default {
  name: 'ForgotPasswordView',
  components: {
    AuthForm,
    FormInput
  },
  data() {
    return {
      loading: false,
      error: '',
      successMessage: '',
      currentStep: 1,
      countdown: 0,
      sendingCode: false,
      formData: {
        email: '',
        verificationCode: '',
        newPassword: '',
        confirmPassword: ''
      },
      errors: {
        email: '',
        verificationCode: '',
        newPassword: '',
        confirmPassword: ''
      },
      logoUrl: require('@/assets/logo.png')
    }
  },
  computed: {
    isValidEmail() {
      const emailPattern = /^[A-Za-z0-9]+([._-][A-Za-z0-9]+)*@([A-Za-z0-9]+[-]?[A-Za-z0-9]+\.)+[A-Za-z]{2,}$/
      return this.formData.email && emailPattern.test(this.formData.email)
    },
    canSendCode() {
      return this.isValidEmail
    },
    canProceedToStep2() {
      return this.isValidEmail && 
             this.formData.verificationCode && 
             this.formData.verificationCode.trim().length === 6
    },
    canResetPassword() {
      return this.formData.newPassword && 
             this.formData.confirmPassword &&
             this.formData.newPassword === this.formData.confirmPassword &&
             this.formData.newPassword.length >= 6 &&
             this.formData.newPassword.length <= 15
    }
  },
  watch: {
    'formData.email'() {
      this.errors.email = ''
      if (this.formData.email && !this.isValidEmail) {
        this.errors.email = '请输入有效的邮箱地址'
      }
    },
    'formData.verificationCode'() {
      this.errors.verificationCode = ''
    },
    'formData.newPassword'() {
      this.errors.newPassword = ''
      if (this.formData.newPassword && (this.formData.newPassword.length < 6 || this.formData.newPassword.length > 15)) {
        this.errors.newPassword = '密码长度应为6-15位'
      }
    },
    'formData.confirmPassword'() {
      this.errors.confirmPassword = ''
      if (this.formData.confirmPassword && this.formData.newPassword !== this.formData.confirmPassword) {
        this.errors.confirmPassword = '两次输入的密码不一致'
      }
    }
  },
  methods: {
    async sendVerificationCode() {
      if (!this.canSendCode) {
        this.errors.email = '请输入有效的邮箱地址'
        return
      }

      this.sendingCode = true
      this.error = ''
      this.errors.email = ''

      try {
        const result = await sendResetPasswordCode(this.formData.email)
        if (result.success) {
          this.successMessage = '验证码已发送到您的邮箱，请查收'
          this.startCountdown()
          setTimeout(() => {
            this.successMessage = ''
          }, 5000)
        } else {
          this.error = getFriendlyErrorMessage(result.message, 'forgot-password')
        }
      } catch (err) {
        if (err.isNetworkError) {
          this.error = '无法连接到服务器，请确保后端服务正在运行'
        } else if (err.message) {
          this.error = getFriendlyErrorMessage(err.message, 'forgot-password')
        } else {
          this.error = '发送验证码失败，请检查网络连接后重试'
        }
        console.error('发送验证码失败:', err)
      } finally {
        this.sendingCode = false
      }
    },
    startCountdown() {
      this.countdown = 60
      const timer = setInterval(() => {
        this.countdown--
        if (this.countdown <= 0) {
          clearInterval(timer)
        }
      }, 1000)
    },
    async proceedToStep2() {
      if (!this.canProceedToStep2) {
        if (!this.isValidEmail) {
          this.errors.email = '请输入有效的邮箱地址'
        }
        if (!this.formData.verificationCode) {
          this.errors.verificationCode = '请输入验证码'
        } else if (this.formData.verificationCode.trim().length !== 6) {
          this.errors.verificationCode = '验证码应为6位数字'
        }
        return
      }

      // 验证码会在重置密码时验证，这里直接进入下一步
      this.currentStep = 2
      this.error = ''
      this.successMessage = ''
    },
    async resetPassword() {
      if (!this.validateStep2()) {
        return
      }

      this.loading = true
      this.error = ''
      this.successMessage = ''

      try {
        const result = await resetPassword(
          this.formData.email,
          this.formData.verificationCode,
          this.formData.newPassword
        )

        if (result.success) {
          this.successMessage = '密码重置成功！正在跳转到登录页面...'
          setTimeout(() => {
            this.$router.push('/login')
          }, 2000)
        } else {
          this.error = getFriendlyErrorMessage(result.message, 'forgot-password')
        }
      } catch (err) {
        if (err.isNetworkError) {
          this.error = '无法连接到服务器，请确保后端服务正在运行'
        } else if (err.message) {
          this.error = getFriendlyErrorMessage(err.message, 'forgot-password')
        } else {
          this.error = '重置密码失败，请检查网络连接后重试'
        }
        console.error('重置密码失败:', err)
      } finally {
        this.loading = false
      }
    },
    validateStep2() {
      this.errors = {
        email: '',
        verificationCode: '',
        newPassword: '',
        confirmPassword: ''
      }
      let isValid = true

      if (!this.formData.newPassword) {
        this.errors.newPassword = '请输入新密码'
        isValid = false
      } else if (this.formData.newPassword.length < 6 || this.formData.newPassword.length > 15) {
        this.errors.newPassword = '密码长度应为6-15位'
        isValid = false
      }

      if (!this.formData.confirmPassword) {
        this.errors.confirmPassword = '请确认新密码'
        isValid = false
      } else if (this.formData.newPassword !== this.formData.confirmPassword) {
        this.errors.confirmPassword = '两次输入的密码不一致'
        isValid = false
      }

      return isValid
    }
  }
}
</script>

<style scoped>
.forgot-password-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #ffffff 0%, #ffd7b4 100%);
  padding: 1rem;
}

.forgot-password-container {
  width: 100%;
  max-width: 400px;
}

.step-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 2rem;
  padding: 0 1rem;
}

.step {
  display: flex;
  flex-direction: column;
  align-items: center;
  position: relative;
}

.step-number {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: #e0e0e0;
  color: #999;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  transition: all 0.3s ease;
}

.step.active .step-number {
  background: #cc2a1f;
  color: white;
}

.step.completed .step-number {
  background: #4caf50;
  color: white;
}

.step-label {
  margin-top: 0.5rem;
  font-size: 0.85rem;
  color: #666;
  transition: color 0.3s ease;
}

.step.active .step-label {
  color: #cc2a1f;
  font-weight: 500;
}

.step.completed .step-label {
  color: #4caf50;
}

.step-line {
  width: 60px;
  height: 2px;
  background: #e0e0e0;
  margin: 0 0.5rem;
  margin-top: -1rem;
  transition: background 0.3s ease;
}

.step-line.completed {
  background: #4caf50;
}

.step-content {
  margin-top: 1rem;
}

.verification-code-group {
  display: flex;
  gap: 0.5rem;
  align-items: flex-end;
}

.verification-code-group .form-input {
  flex: 1;
}

.send-code-btn {
  padding: 12px 16px;
  background: #cc2a1f;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 0.9rem;
  cursor: pointer;
  white-space: nowrap;
  transition: background 0.3s ease;
  min-width: 120px;
}

.send-code-btn:hover:not(:disabled) {
  background: #7f180f;
}

.send-code-btn:disabled {
  background: #ccc;
  cursor: not-allowed;
}

.submit-btn {
  width: 100%;
  padding: 12px;
  background: #cc2a1f;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 1rem;
  cursor: pointer;
  transition: background 0.3s ease;
  margin-top: 1rem;
}

.submit-btn:hover:not(:disabled) {
  background: #7f180f;
}

.submit-btn:disabled {
  background: #ccc;
  cursor: not-allowed;
}

.submit-btn.loading {
  background: #6c757d;
}

/* 隐藏注册/登录切换链接 */
.forgot-password-form :deep(.auth-footer) {
  display: none;
}
</style>

