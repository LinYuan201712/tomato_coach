<template>
  <div class="auth-form">
    <div class="auth-header">
      <img :src="logo" alt="Logo" class="logo" />
      <h2>{{ isLogin ? '欢迎回来' : (isForgotPassword ? '找回密码' : '创建账户') }}</h2>
      <p>{{ isLogin ? '请登录您的账户' : (isForgotPassword ? '通过邮箱验证重置您的密码' : '注册新账户开始使用') }}</p>
    </div>

    <form @submit.prevent="handleSubmit" class="auth-form-container">
      <slot name="form-fields"></slot>
      
      <div v-if="isLogin" class="form-options">
        <!-- <label class="remember-me">
          <input type="checkbox" v-model="rememberMe" />
          记住我
        </label> -->
        <a href="#" class="forgot-password" @click.prevent="$emit('forgot-password')">忘记密码？</a>
      </div>

      <button 
        v-if="!hideSubmitButton"
        type="submit" 
        class="submit-btn"
        :disabled="loading"
        :class="{ 'loading': loading }"
      >
        <span v-if="loading">处理中...</span>
        <span v-else>{{ isLogin ? '登录' : '注册' }}</span>
      </button>
    </form>

    <div class="auth-footer">
      <p>
        {{ isLogin ? '还没有账户？' : '已有账户？' }}
        <a href="#" @click.prevent="$emit('toggle-mode')">
          {{ isLogin ? '立即注册' : '立即登录' }}
        </a>
      </p>
    </div>

    <div v-if="error" class="error-message">
      {{ error }}
    </div>
    <div v-if="successMessage" class="success-message">
      {{ successMessage }}
    </div>
  </div>
</template>

<script>
export default {
  name: 'AuthForm',
  props: {
    isLogin: {
      type: Boolean,
      default: true
    },
    loading: {
      type: Boolean,
      default: false
    },
    error: {
      type: String,
      default: ''
    },
    successMessage: {
      type: String,
      default: ''
    },
    logo: {
      type: String,
      default: ''
    },
    hideSubmitButton: {
      type: Boolean,
      default: false
    },
    isForgotPassword: {
      type: Boolean,
      default: false
    }
  },
  data() {
    return {
      rememberMe: false
    }
  },
  methods: {
    handleSubmit() {
      this.$emit('submit', {
        rememberMe: this.rememberMe
      })
    }
  }
}
</script>

<style scoped>
.auth-form {
  max-width: 400px;
  margin: 0 auto;
  padding: 2rem;
  background: white;
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
}

.auth-header {
  text-align: center;
  margin-bottom: 2rem;
}

.logo {
  width: 60px;
  height: 60px;
  margin-bottom: 1rem;
}

.auth-header h2 {
  color: #333;
  margin-bottom: 0.5rem;
  font-size: 1.5rem;
}

.auth-header p {
  color: #666;
  margin: 0;
}

.auth-form-container {
  margin-bottom: 1.5rem;
}

.form-options {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
  font-size: 0.9rem;
}

.remember-me {
  display: flex;
  align-items: center;
  color: #666;
}

.remember-me input {
  margin-right: 0.5rem;
}

.forgot-password {
  color: #cc2a1f;
  text-decoration: none;
}

.forgot-password:hover {
  text-decoration: underline;
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

.auth-footer {
  text-align: center;
  padding-top: 1.5rem;
  border-top: 1px solid #eee;
}

.auth-footer p {
  color: #666;
  margin: 0;
}

.auth-footer a {
  color: #cc2a1f;
  text-decoration: none;
  font-weight: 500;
}

.auth-footer a:hover {
  text-decoration: underline;
}

.error-message {
  margin-top: 1rem;
  padding: 14px 16px;
  background: linear-gradient(135deg, #fff5f5 0%, #ffe0e0 100%);
  color: #c53030;
  border: 1.5px solid #fc8181;
  border-radius: 8px;
  text-align: center;
  font-size: 0.9rem;
  font-weight: 500;
  box-shadow: 0 2px 8px rgba(197, 48, 48, 0.1);
  animation: slideDown 0.3s ease-out;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.error-message::before {
  content: '⚠️';
  font-size: 1.1rem;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.success-message {
  margin-top: 1rem;
  padding: 14px 16px;
  background: linear-gradient(135deg, #f0fff4 0%, #c6f6d5 100%);
  color: #22543d;
  border: 1.5px solid #68d391;
  border-radius: 8px;
  text-align: center;
  font-size: 0.9rem;
  font-weight: 500;
  box-shadow: 0 2px 8px rgba(34, 84, 61, 0.1);
  animation: slideDown 0.3s ease-out;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.success-message::before {
  content: '✓';
  font-size: 1.1rem;
  color: #22543d;
  font-weight: bold;
}
</style>