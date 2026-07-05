<template>
  <div 
    class="desktop-pet"
    :style="petStyle"
    @mousedown="startDrag"
    @click="handleClick"
    @mouseenter="showTooltip = true"
    @mouseleave="showTooltip = false"
  >
    <!-- 桌宠主体 -->
    <div class="pet-body" :class="animationClass" :style="bodyStyle">
      <svg class="tomato-pet" viewBox="0 0 160 150" aria-hidden="true">
        <g class="tomato-body">
          <polygon class="tomato-outline" points="40,42 120,42 120,48 134,48 134,56 144,56 144,70 152,70 152,112 144,112 144,126 134,126 134,136 118,136 118,144 42,144 42,136 26,136 26,126 16,126 16,112 8,112 8,70 16,70 16,56 26,56 26,48 40,48" />
          <polygon class="tomato-fill" points="42,50 118,50 118,56 132,56 132,64 140,64 140,76 148,76 148,108 140,108 140,122 130,122 130,132 114,132 114,138 46,138 46,132 30,132 30,122 20,122 20,108 14,108 14,76 20,76 20,64 30,64 30,56 42,56" />
          <polygon class="tomato-bright" points="28,64 42,64 42,58 54,58 54,70 42,70 42,82 34,82 34,106 26,106 26,122 20,122 20,76 28,76" />
          <polygon class="tomato-edge-shadow" points="132,64 140,64 140,76 148,76 148,108 140,108 140,122 130,122 130,132 114,132 114,138 94,138 94,132 116,132 116,126 126,126 126,116 134,116 134,98 140,98 140,78 132,78" />
          <rect class="tomato-bottom-shadow" x="52" y="136" width="56" height="8" />
        </g>

        <g class="tomato-crown">
          <polygon class="leaf-outline" points="28,84 28,48 44,48 44,36 56,36 56,28 64,28 64,4 96,4 96,28 104,28 104,36 116,36 116,48 132,48 132,84 104,84 104,76 92,76 92,66 84,66 84,78 76,78 76,66 68,66 68,76 56,76 56,84" />
          <polygon class="leaf-main" points="36,76 36,56 50,56 50,44 62,44 62,36 72,36 72,12 88,12 88,36 98,36 98,44 110,44 110,56 124,56 124,76 100,76 100,68 88,68 88,58 82,58 82,70 78,70 78,58 72,58 72,68 60,68 60,76" />
          <polygon class="leaf-light" points="38,56 52,56 52,66 62,66 62,74 38,74" />
          <polygon class="leaf-mid" points="66,38 78,38 78,56 96,56 96,68 84,68 84,60 76,60 76,68 66,68" />
          <polygon class="leaf-dark" points="88,12 96,12 96,36 106,36 106,48 126,48 126,76 110,76 110,66 98,66 98,54 88,54" />
          <rect class="leaf-cut" x="74" y="18" width="12" height="18" />
        </g>

        <g class="tomato-face-pixels">
          <rect class="tomato-eye" x="52" y="84" width="8" height="18" />
          <rect class="tomato-eye" x="100" y="84" width="8" height="18" />
          <rect class="tomato-cheek" x="22" y="104" width="15" height="10" />
          <rect class="tomato-cheek" x="120" y="104" width="15" height="10" />
          <rect class="tomato-mouth" x="64" y="112" width="32" height="8" />
          <rect class="tomato-mouth" x="64" y="120" width="8" height="4" />
        </g>
      </svg>
    </div>
    
    <!-- 提示信息 -->
    <div v-if="showTooltip" class="pet-tooltip">
      {{ tooltipText }}
    </div>
    
    <!-- 对话气泡 - 放在外层，不受transform影响 -->
    <div v-if="showBubble" class="pet-bubble" :class="{ 'bubble-visible': showBubble }">
      {{ bubbleText }}
    </div>
    
    <!-- 功能菜单 -->
    <div v-if="showMenu || chatNotice.visible" class="pet-menu" :class="{ 'menu-visible': showMenu || chatNotice.visible, 'notice-menu': chatNotice.visible }">
      <div class="menu-item" @click="openChat">
        <span class="menu-icon">{{ chatNotice.icon || '💬' }}</span>
        <span class="menu-text">{{ chatNotice.text || '聊天' }}</span>
      </div>
    </div>
    
    <!-- 聊天对话框 - 增加 @mousedown.stop 阻止拖拽冲突 -->
    <ChatDialog 
      :visible="showChatDialog" 
      @close="closeChatDialog" 
      @message-started="handleMessageStarted"
      @message-generated="handleMessageGenerated"
      @mousedown.stop
    />
  </div>
</template>

<script>
import ChatDialog from './ChatDialog.vue'

export default {
  name: 'DesktopPet',
  components: {
    ChatDialog
  },
  data() {
    return {
      // 位置
      x: window.innerWidth - 120,
      y: window.innerHeight - 120,
      
      // 拖拽状态
      isDragging: false,
      dragOffset: { x: 0, y: 0 },
      
      // 动画状态
      animationState: 'idle', // idle, walking, happy, sleep
      direction: 1, // 1: 向右, -1: 向左
      
      // 自动移动
      autoMoveTimer: null,
      targetX: null,
      targetY: null,
      moveSpeed: 2,
      
      // 交互
      showTooltip: false,
      showBubble: false,
      bubbleText: '',
      clickCount: 0,
      showMenu: false,
      menuTimer: null,
      showChatDialog: false,
      chatNotice: {
        visible: false,
        text: '',
        icon: ''
      },
      chatNoticeTimer: null,
      
      // 待机动画
      idleAnimationTimer: null
    }
  },
  computed: {
    petStyle() {
      return {
        left: `${this.x}px`,
        top: `${this.y}px`
      }
    },
    bodyStyle() {
      return {
        transform: `scaleX(${this.direction})`
      }
    },
    animationClass() {
      return {
        'pet-idle': this.animationState === 'idle',
        'pet-walking': this.animationState === 'walking',
        'pet-happy': this.animationState === 'happy',
        'pet-sleep': this.animationState === 'sleep'
      }
    },
    tooltipText() {
      return '点击我试试！拖拽可以移动'
    }
  },
  mounted() {
    this.initPet()
    this.startIdleAnimation()
    this.startAutoMove()
    window.addEventListener('mousemove', this.onMouseMove)
    window.addEventListener('mouseup', this.stopDrag)
    window.addEventListener('resize', this.handleResize)
    // 登录时显示时间段问候语
    this.showTimeGreeting()
  },
  beforeUnmount() {
    this.stopAllTimers()
    window.removeEventListener('mousemove', this.onMouseMove)
    window.removeEventListener('mouseup', this.stopDrag)
    window.removeEventListener('resize', this.handleResize)
  },
  methods: {
    initPet() {
      // 初始化位置（右下角）
      this.x = Math.max(20, window.innerWidth - 150)
      this.y = Math.max(20, window.innerHeight - 140)
    },
    
    // 拖拽功能
    startDrag(e) {
      this.isDragging = true
      this.dragOffset.x = e.clientX - this.x
      this.dragOffset.y = e.clientY - this.y
      this.animationState = 'walking'
      this.stopAutoMove()
    },
    
    onMouseMove(e) {
      if (this.isDragging) {
        this.x = Math.max(0, Math.min(e.clientX - this.dragOffset.x, window.innerWidth - 136))
        this.y = Math.max(0, Math.min(e.clientY - this.dragOffset.y, window.innerHeight - 128))
        
        // 根据移动方向改变朝向
        if (e.movementX > 0) {
          this.direction = 1
        } else if (e.movementX < 0) {
          this.direction = -1
        }
      }
    },
    
    stopDrag() {
      if (this.isDragging) {
        this.isDragging = false
        this.animationState = 'idle'
        this.startAutoMove()
      }
    },
    
    // 点击交互
    handleClick(e) {
      // 如果点击的是菜单项，不处理
      if (e.target.closest('.pet-menu')) {
        return
      }
      
      this.clickCount++
      
      // 显示功能菜单
      this.showMenu = true
      this.hideMenuAfterDelay()
      
      // 点击动画
      this.animationState = 'happy'
      setTimeout(() => {
        if (!this.isDragging) {
          this.animationState = 'idle'
        }
      }, 500)
    },
    
    // 延迟隐藏菜单
    hideMenuAfterDelay() {
      if (this.menuTimer) {
        clearTimeout(this.menuTimer)
      }
      this.menuTimer = setTimeout(() => {
        this.showMenu = false
      }, 3000)
    },
    
    // 打开聊天对话框
    openChat() {
      this.showChatDialog = true
      this.showMenu = false
      this.hideChatNotice()
      this.stopAutoMove() // 开启聊天时停止移动
      if (this.menuTimer) {
        clearTimeout(this.menuTimer)
      }
    },
    
    // 关闭聊天对话框
    closeChatDialog() {
      this.showChatDialog = false
      this.startAutoMove() // 关闭聊天时恢复移动
    },

    handleMessageStarted() {
      if (this.chatNoticeTimer) {
        clearTimeout(this.chatNoticeTimer)
        this.chatNoticeTimer = null
      }
      this.chatNotice = {
        visible: true,
        text: '对话生成中...',
        icon: '💬'
      }
    },

    handleMessageGenerated() {
      this.chatNotice = {
        visible: true,
        text: '对话已生成',
        icon: '✅'
      }
      this.animationState = 'happy'
      setTimeout(() => {
        if (!this.isDragging) {
          this.animationState = 'idle'
        }
      }, 700)
      this.chatNoticeTimer = setTimeout(() => {
        this.hideChatNotice()
      }, 8000)
    },

    hideChatNotice() {
      if (this.chatNoticeTimer) {
        clearTimeout(this.chatNoticeTimer)
        this.chatNoticeTimer = null
      }
      this.chatNotice = {
        visible: false,
        text: '',
        icon: ''
      }
    },
    
    showBubbleMessage(text) {
      this.bubbleText = text
      this.showBubble = true
      setTimeout(() => {
        this.showBubble = false
      }, 3000)
    },
    
    // 自动移动
    startAutoMove() {
      this.stopAutoMove()
      
      // 如果聊天框打开，不开启自动移动
      if (this.showChatDialog) return
      
      // 每5-10秒随机移动一次
      const delay = Math.random() * 5000 + 5000
      
      this.autoMoveTimer = setTimeout(() => {
        if (!this.isDragging) {
          this.moveToRandomPosition()
        }
        this.startAutoMove()
      }, delay)
    },
    
    stopAutoMove() {
      if (this.autoMoveTimer) {
        clearTimeout(this.autoMoveTimer)
        this.autoMoveTimer = null
      }
      this.targetX = null
      this.targetY = null
    },
    
    moveToRandomPosition() {
      // 随机目标位置
      this.targetX = Math.random() * (window.innerWidth - 150)
      this.targetY = Math.random() * (window.innerHeight - 140)
      this.animationState = 'walking'
      
      // 根据目标位置决定方向
      if (this.targetX > this.x) {
        this.direction = 1
      } else {
        this.direction = -1
      }
      
      // 移动动画
      this.animateMove()
    },
    
    animateMove() {
      if (this.isDragging || !this.targetX || !this.targetY) {
        this.animationState = 'idle'
        return
      }
      
      const dx = this.targetX - this.x
      const dy = this.targetY - this.y
      const distance = Math.sqrt(dx * dx + dy * dy)
      
      if (distance < 5) {
        // 到达目标
        this.x = this.targetX
        this.y = this.targetY
        this.targetX = null
        this.targetY = null
        this.animationState = 'idle'
        return
      }
      
      // 继续移动
      const moveX = (dx / distance) * this.moveSpeed
      const moveY = (dy / distance) * this.moveSpeed
      
      this.x = Math.max(0, Math.min(this.x + moveX, window.innerWidth - 136))
      this.y = Math.max(0, Math.min(this.y + moveY, window.innerHeight - 128))
      
      requestAnimationFrame(() => this.animateMove())
    },
    
    // 待机动画
    startIdleAnimation() {
      this.idleAnimationTimer = setInterval(() => {
        if (!this.isDragging && this.animationState === 'idle') {
          // 偶尔做一些小动作
          if (Math.random() < 0.3) {
            this.animationState = 'happy'
            setTimeout(() => {
              if (this.animationState === 'happy') {
                this.animationState = 'idle'
              }
            }, 800)
          }
        }
      }, 3000)
    },
    
    // 窗口大小改变
    handleResize() {
      // 确保桌宠不会超出窗口
      this.x = Math.max(0, Math.min(this.x, window.innerWidth - 136))
      this.y = Math.max(0, Math.min(this.y, window.innerHeight - 128))
    },
    
    // 清理所有定时器
    stopAllTimers() {
      this.stopAutoMove()
      if (this.idleAnimationTimer) {
        clearInterval(this.idleAnimationTimer)
        this.idleAnimationTimer = null
      }
      if (this.menuTimer) {
        clearTimeout(this.menuTimer)
        this.menuTimer = null
      }
      if (this.chatNoticeTimer) {
        clearTimeout(this.chatNoticeTimer)
        this.chatNoticeTimer = null
      }
    },
    
    // 获取时间段问候语
    getTimeGreeting() {
      const hour = new Date().getHours()
      
      if (hour >= 6 && hour < 9) {
        // 早上 6:00 - 9:00
        return '早上好呀，一日之计在于晨，快开始今天的学习吧！'
      } else if (hour >= 9 && hour < 12) {
        // 上午 9:00 - 12:00
        return '上午好！新的一天开始了，让我们一起加油学习吧！'
      } else if (hour >= 12 && hour < 14) {
        // 中午 12:00 - 14:00
        return '中午好！记得适当休息，下午继续努力哦！'
      } else if (hour >= 14 && hour < 18) {
        // 下午 14:00 - 18:00
        return '下午好！下午时光很宝贵，保持专注继续学习吧！'
      } else if (hour >= 18 && hour < 22) {
        // 晚上 18:00 - 22:00
        return '晚上好！今天的学习任务完成得怎么样？继续加油！'
      } else {
        // 深夜 22:00 - 6:00
        return '夜深了，还在学习吗？记得早点休息，保持身体健康哦！'
      }
    },
    
    // 显示时间段问候语
    showTimeGreeting() {
      const greeting = this.getTimeGreeting()
      // 延迟一点显示，让组件完全加载
      setTimeout(() => {
        this.showBubbleMessage(greeting)
      }, 500)
    }
  }
}
</script>

<style scoped>
.desktop-pet {
  position: fixed;
  width: 136px;
  height: 128px;
  z-index: 9999;
  cursor: grab;
  user-select: none;
  pointer-events: auto;
  transition: transform 0.1s ease;
}

.desktop-pet:active {
  cursor: grabbing;
}

.pet-body {
  width: 100%;
  height: 100%;
  position: relative;
  transform-origin: center bottom;
}

.tomato-pet {
  --pet-scale: 1;
  position: absolute;
  left: 50%;
  bottom: 2px;
  width: 132px;
  height: 124px;
  transform: translateX(-50%) scale(var(--pet-scale));
  transform-origin: center bottom;
  shape-rendering: crispEdges;
  filter: drop-shadow(0 8px 0 rgba(72, 23, 11, 0.14));
  transition: filter 0.2s ease;
}

.tomato-outline {
  fill: #910506;
}

.tomato-fill {
  fill: #ef070d;
}

.tomato-bright {
  fill: #ff2930;
}

.tomato-edge-shadow,
.tomato-bottom-shadow {
  fill: #b30508;
}

.leaf-outline {
  fill: #057522;
}

.leaf-main {
  fill: #00ad2f;
}

.leaf-light {
  fill: #25df51;
}

.leaf-mid {
  fill: #0f932c;
}

.leaf-dark {
  fill: #04701e;
}

.leaf-cut {
  fill: #128e2e;
  opacity: 0.7;
}

.tomato-eye,
.tomato-mouth {
  fill: #020202;
}

.tomato-cheek {
  fill: #ffd7d9;
}

.pet-body:hover .tomato-pet {
  filter: drop-shadow(0 12px 0 rgba(64, 18, 10, 0.18)) saturate(1.08);
}

/* 待机动画 */
.pet-idle .tomato-pet {
  animation: idleBounce 2s ease-in-out infinite;
}

@keyframes idleBounce {
  0%, 100% {
    transform: translateX(-50%) translateY(0) scale(var(--pet-scale));
  }
  50% {
    transform: translateX(-50%) translateY(-5px) scale(calc(var(--pet-scale) * 1.05));
  }
}

/* 走路动画 */
.pet-walking .tomato-pet {
  animation: walking 0.6s ease-in-out infinite;
}

@keyframes walking {
  0%, 100% {
    transform: translateX(-50%) translateY(0) rotate(-5deg) scale(var(--pet-scale));
  }
  50% {
    transform: translateX(-50%) translateY(-8px) rotate(5deg) scale(var(--pet-scale));
  }
}

/* 开心动画 */
.pet-happy .tomato-pet {
  animation: happy 0.5s ease-in-out;
}

@keyframes happy {
  0% {
    transform: translateX(-50%) scale(var(--pet-scale)) rotate(0deg);
  }
  25% {
    transform: translateX(-50%) scale(calc(var(--pet-scale) * 1.2)) rotate(-10deg);
  }
  50% {
    transform: translateX(-50%) scale(calc(var(--pet-scale) * 1.3)) rotate(10deg);
  }
  75% {
    transform: translateX(-50%) scale(calc(var(--pet-scale) * 1.2)) rotate(-5deg);
  }
  100% {
    transform: translateX(-50%) scale(var(--pet-scale)) rotate(0deg);
  }
}

/* 睡觉动画 */
.pet-sleep .tomato-pet {
  animation: sleep 3s ease-in-out infinite;
  opacity: 0.7;
}

.pet-sleep .tomato-eye {
  transform: scaleY(0.25);
  transform-origin: center;
}

.pet-happy .tomato-cheek {
  fill: #ffc0c4;
  opacity: 1;
}

@keyframes sleep {
  0%, 100% {
    transform: translateX(-50%) translateY(0) rotate(0deg) scale(var(--pet-scale));
  }
  50% {
    transform: translateX(-50%) translateY(-3px) rotate(5deg) scale(var(--pet-scale));
  }
}

/* 提示信息 */
.pet-tooltip {
  position: absolute;
  bottom: 100%;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(0, 0, 0, 0.8);
  color: white;
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 12px;
  white-space: nowrap;
  margin-bottom: 8px;
  pointer-events: none;
  animation: tooltipFade 0.3s ease;
}

.pet-tooltip::after {
  content: '';
  position: absolute;
  top: 100%;
  left: 50%;
  transform: translateX(-50%);
  border: 6px solid transparent;
  border-top-color: rgba(0, 0, 0, 0.8);
}

@keyframes tooltipFade {
  from {
    opacity: 0;
    transform: translateX(-50%) translateY(5px);
  }
  to {
    opacity: 1;
    transform: translateX(-50%) translateY(0);
  }
}

/* 对话气泡 */
.pet-bubble {
  position: absolute;
  left: 100%;
  top: 50%;
  transform: translateY(-50%);
  background: white;
  color: #333;
  padding: 8px 12px;
  border-radius: 12px;
  font-size: 13px;
  max-width: 250px;
  margin-left: 15px;
  white-space: nowrap;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  pointer-events: none;
  opacity: 0;
  transition: opacity 0.3s ease, transform 0.3s ease;
  z-index: 10000;
  /* 确保文字不镜像 */
  direction: ltr;
  text-align: left;
}

.pet-bubble.bubble-visible {
  opacity: 1;
  transform: translateY(-50%) translateX(5px);
}

.pet-bubble::after {
  content: '';
  position: absolute;
  right: 100%;
  top: 50%;
  transform: translateY(-50%);
  border: 8px solid transparent;
  border-right-color: white;
}

/* 功能菜单 */
.pet-menu {
  position: absolute;
  left: 100%;
  top: 50%;
  transform: translateY(-50%);
  margin-left: 15px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
  padding: 8px;
  min-width: 120px;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.3s ease, transform 0.3s ease;
  z-index: 10001;
  direction: ltr;
  text-align: left;
}

.pet-menu.menu-visible {
  opacity: 1;
  pointer-events: auto;
  transform: translateY(-50%) translateX(5px);
}

.pet-menu.notice-menu {
  border: 1px solid #ffe4cc;
  box-shadow: 0 8px 24px rgba(238, 170, 103, 0.24);
}

.pet-menu::before {
  content: '';
  position: absolute;
  right: 100%;
  top: 50%;
  transform: translateY(-50%);
  border: 8px solid transparent;
  border-right-color: white;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.menu-item:hover {
  background-color: #f5f5f5;
}

.notice-menu .menu-item {
  background: #fffaf5;
}

.notice-menu .menu-item:hover {
  background: #fff3e6;
}

.menu-icon {
  font-size: 18px;
}

.menu-text {
  font-size: 14px;
  color: #333;
  font-weight: 500;
}

/* 响应式 */
@media (max-width: 768px) {
  .desktop-pet {
    width: 104px;
    height: 98px;
  }

  .tomato-pet {
    --pet-scale: 0.76;
    transform-origin: center bottom;
  }
  
  .pet-bubble {
    max-width: 150px;
    font-size: 12px;
  }
  
  .pet-menu {
    min-width: 100px;
  }
}
</style>
