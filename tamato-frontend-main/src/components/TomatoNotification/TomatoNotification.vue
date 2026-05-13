<template>
  <transition name="tomato-notification">
    <div v-if="visible" class="tomato-notification" :class="type">
      <div class="notification-content">
        <div class="tomato-icon-wrapper">
          <span class="tomato-emoji">🍅</span>
          <span v-if="type === 'earn'" class="plus-icon">+</span>
          <span v-else class="minus-icon">-</span>
        </div>
        <div class="notification-text">
          <div class="notification-title">{{ title }}</div>
          <div class="notification-message">{{ message }}</div>
        </div>
      </div>
      <div class="notification-progress" v-if="autoClose"></div>
    </div>
  </transition>
</template>

<script>
export default {
  name: 'TomatoNotification',
  props: {
    visible: {
      type: Boolean,
      default: false
    },
    type: {
      type: String,
      default: 'earn', // 'earn' 或 'spend'
      validator: (value) => ['earn', 'spend'].includes(value)
    },
    title: {
      type: String,
      default: ''
    },
    message: {
      type: String,
      default: ''
    },
    duration: {
      type: Number,
      default: 3000 // 3秒后自动关闭
    },
    autoClose: {
      type: Boolean,
      default: true
    }
  },
  data() {
    return {
      autoCloseTimer: null
    }
  },
  watch: {
    visible(newVal) {
      if (newVal && this.autoClose) {
        this.startAutoClose()
      } else if (!newVal) {
        // 如果visible变为false，清除定时器
        this.clearTimer()
      }
    }
  },
  mounted() {
    // 组件挂载时，如果visible已经是true，立即启动自动关闭
    if (this.visible && this.autoClose) {
      this.startAutoClose()
    }
  },
  methods: {
    startAutoClose() {
      this.clearTimer()
      this.autoCloseTimer = setTimeout(() => {
        this.$emit('close')
      }, this.duration)
    },
    clearTimer() {
      if (this.autoCloseTimer) {
        clearTimeout(this.autoCloseTimer)
        this.autoCloseTimer = null
      }
    }
  },
  beforeUnmount() {
    this.clearTimer()
  }
}
</script>

<style scoped>
.tomato-notification {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 99999;
  min-width: 300px;
  max-width: 400px;
  background: linear-gradient(135deg, #fff5eb 0%, #ffe4cc 100%);
  border-radius: 16px;
  box-shadow: 0 8px 24px rgba(238, 170, 103, 0.3),
              0 4px 8px rgba(0, 0, 0, 0.1);
  padding: 16px 20px;
  border: 2px solid #ffd4a3;
  animation: slideInRight 0.4s cubic-bezier(0.68, -0.55, 0.265, 1.55);
}

.tomato-notification.earn {
  background: linear-gradient(135deg, #fff5eb 0%, #ffe4cc 100%);
  border-color: #ffd4a3;
}

.tomato-notification.spend {
  background: linear-gradient(135deg, #fff0f0 0%, #ffe0e0 100%);
  border-color: #ffb3b3;
}

.notification-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.tomato-icon-wrapper {
  position: relative;
  width: 60px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: white;
  border-radius: 50%;
  box-shadow: 0 4px 12px rgba(238, 170, 103, 0.2);
  animation: bounce 0.6s ease-in-out;
}

.tomato-emoji {
  font-size: 36px;
  animation: rotate 0.5s ease-in-out;
}

.plus-icon,
.minus-icon {
  position: absolute;
  top: -8px;
  right: -8px;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  font-weight: bold;
  color: white;
  animation: popIn 0.4s cubic-bezier(0.68, -0.55, 0.265, 1.55);
}

.plus-icon {
  background: linear-gradient(135deg, #4ade80, #22c55e);
  box-shadow: 0 2px 8px rgba(34, 197, 94, 0.4);
}

.minus-icon {
  background: linear-gradient(135deg, #f87171, #ef4444);
  box-shadow: 0 2px 8px rgba(239, 68, 68, 0.4);
}

.notification-text {
  flex: 1;
}

.notification-title {
  font-size: 16px;
  font-weight: 600;
  color: #333;
  margin-bottom: 4px;
  line-height: 1.4;
}

.notification-message {
  font-size: 14px;
  color: #666;
  line-height: 1.5;
}

.notification-progress {
  position: absolute;
  bottom: 0;
  left: 0;
  height: 3px;
  background: linear-gradient(90deg, #eeaa67, #f5b877);
  border-radius: 0 0 14px 14px;
  animation: progressBar 3s linear forwards;
}

/* 动画 */
@keyframes slideInRight {
  from {
    transform: translateX(400px);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}

@keyframes bounce {
  0%, 100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.1);
  }
}

@keyframes rotate {
  from {
    transform: rotate(-10deg);
  }
  to {
    transform: rotate(0deg);
  }
}

@keyframes popIn {
  from {
    transform: scale(0);
    opacity: 0;
  }
  to {
    transform: scale(1);
    opacity: 1;
  }
}

@keyframes progressBar {
  from {
    width: 100%;
  }
  to {
    width: 0%;
  }
}

.tomato-notification-enter-active,
.tomato-notification-leave-active {
  transition: all 0.4s cubic-bezier(0.68, -0.55, 0.265, 1.55);
}

.tomato-notification-enter-from {
  transform: translateX(400px);
  opacity: 0;
}

.tomato-notification-leave-to {
  transform: translateX(400px);
  opacity: 0;
}

/* 响应式 */
@media (max-width: 768px) {
  .tomato-notification {
    top: 10px;
    right: 10px;
    left: 10px;
    min-width: auto;
    max-width: none;
  }
}
</style>

