<template>
  <div class="login-page">
    <div class="login-container">
      <AuthForm
        :is-login="true"
        :loading="loading"
        :error="error"
        :success-message="successMessage"
        :logo="logoUrl"
        @submit="handleLogin"
        @toggle-mode="$router.push('/register')"
        @forgot-password="$router.push('/forgot-password')"
      >
        <template #form-fields>
          <FormInput
            v-model="formData.username"
            type="text"
            label="用户名"
            placeholder="请输入您的用户名"
            required
            :error="errors.username"
          />
          
          <FormInput
            v-model="formData.password"
            type="password"
            label="密码"
            placeholder="请输入您的密码"
            required
            :error="errors.password"
          />
        </template>
      </AuthForm>
    </div>
  </div>
</template>

<script>
import AuthForm from '@/components/Login/AuthForm.vue'
import FormInput from '@/components/Login/FormInput.vue'
import { login } from '@/api/auth'
import { getFriendlyErrorMessage } from '@/utils/errorMessages'

export default {
  name: 'LoginView',
  components: {
    AuthForm,
    FormInput
  },
  data() {
    return {
      loading: false,
      error: '',
      successMessage: '',
      formData: {
        username: '',
        password: ''
      },
      errors: {
        username: '',
        password: ''
      },
      logoUrl: require('@/assets/logo.png')
    }
  },
  mounted() {
    // 检查是否从注册页面跳转过来
    if (this.$route.query.registered === 'true') {
      this.successMessage = '注册成功！请使用您的账号登录'
      // 自动填充用户名
      if (this.$route.query.username) {
        this.formData.username = this.$route.query.username
      }
      // 3秒后自动清除成功提示
      setTimeout(() => {
        this.successMessage = ''
      }, 5000)
    }
  },
  methods: {
    validateForm() {
      this.errors = {
        username: '',
        password: ''
      }
      let isValid = true

      if (!this.formData.username) {
        this.errors.username = '请输入用户名'
        isValid = false
      }

      if (!this.formData.password) {
        this.errors.password = '请输入密码'
        isValid = false
      }

      return isValid
    },

    async handleLogin() {
      if (!this.validateForm()) {
        return
      }

      this.loading = true
      this.error = ''

      try {
        const result = await login(this.formData.username, this.formData.password)

        if (result.success) {
          // 登录成功，跳转到首页
          setTimeout(() => {
            this.$router.push('/home')
          }, 800)
        } else {
          // 使用友好的错误消息
          this.error = getFriendlyErrorMessage(result.message, 'login')
        }
      } catch (err) {
        // 显示友好的错误信息
        if (err.isNetworkError) {
          this.error = '无法连接到服务器，请确保后端服务正在运行'
        } else if (err.message) {
          this.error = getFriendlyErrorMessage(err.message, 'login')
        } else {
          this.error = '登录失败，请检查网络连接后重试'
        }
        console.error('登录请求失败:', err)
      } finally {
        this.loading = false
      }
    }
  },
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #ffffff 0%, #ffd7b4 100%);
  padding: 1rem;
}

.login-container {
  width: 100%;
  max-width: 400px;
}
</style>