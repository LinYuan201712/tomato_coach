<template>
  <div class="task-sidebar">
    <div class="sidebar-header">
      <h3 class="sidebar-title">待办任务</h3>
      <button @click="toggleCollapsed" class="collapse-btn">
        {{ isCollapsed ? '展开' : '收起' }}
      </button>
    </div>
    
    <div v-if="!isCollapsed" class="sidebar-content">
      <!-- 任务列表 -->
      <div class="task-list">
        <div 
          v-for="task in pendingTasks" 
          :key="task.task_id || task.taskId"
          class="task-item"
          :class="{ 'active': (task.status || task.taskStatus) === '进行中' }"
          @click="goToTaskManagement"
        >
          <div class="task-item-main">
            <div class="task-item-header">
              <span class="task-name">{{ task.task_name || task.taskName }}</span>
              <span class="task-duration">{{ task.duration || 25 }}分钟</span>
            </div>
            <div class="task-item-footer">
              <span class="task-status">{{ getStatusText(task.status || task.taskStatus) }}</span>
              <span class="task-time">{{ formatTime(task.created_at || task.createdAt) }}</span>
            </div>
          </div>
        </div>
      </div>
      
      <!-- 空状态 -->
      <div v-if="pendingTasks.length === 0" class="empty-tasks">
        <p>暂无待办任务</p>
        <button @click="goToTaskManagement" class="create-task-btn">
          去创建任务
        </button>
      </div>
      
      <!-- 底部统计 -->
      <div class="sidebar-footer">
        <div class="task-stats">
          <span>共 {{ totalTasks }} 个任务</span>
          <span>{{ completedTasks }} 个已完成</span>
        </div>
        <button @click="refreshTasks" class="refresh-btn" :disabled="loading">
          {{ loading ? '刷新中...' : '刷新' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import { getTasks, updateTask } from '@/api/tasks'
import { getCurrentUser } from '@/api/user'

export default {
  name: 'TaskSidebar',
  props: {
    // 可以从父组件传入初始状态
    initiallyCollapsed: {
      type: Boolean,
      default: false
    }
  },
  data() {
    return {
      tasks: [],
      userId: null, // 当前用户ID
      loading: false,
      isCollapsed: this.initiallyCollapsed
    }
  },
  computed: {
    // 未完成的任务（包括进行中和未完成），已完成的不显示在侧边栏
    pendingTasks() {
      return this.tasks.filter(task => {
        const statusText = this.getStatusText(task.status || task.taskStatus)
        return statusText !== '已完成'
      })
    },
    totalTasks() {
      return this.tasks.length
    },
    completedTasks() {
      return this.tasks.filter(task => {
        const status = task.status || task.taskStatus
        return status === '已完成'
      }).length
    }
  },
  async mounted() {
    await this.initUser()
    await this.loadTasks()
    // 可选：定时刷新任务列表
    // this.refreshInterval = setInterval(this.loadTasks, 60000) // 每分钟刷新一次
  },
  beforeUnmount() {
    // if (this.refreshInterval) {
    //   clearInterval(this.refreshInterval)
    // }
  },
  methods: {
    async initUser() {
      try {
        const userResult = await getCurrentUser()
        console.log('TaskSidebar用户信息响应:', userResult)
        
        if (userResult && userResult.data) {
          // 根据user.js的返回结构，user_id可能在不同的位置
          if (userResult.data.user_id) {
            this.userId = userResult.data.user_id
          } else if (userResult.data.id) {
            this.userId = userResult.data.id
          } else {
            console.warn('未找到user_id字段，用户数据:', userResult.data)
          }
        }
        console.log('TaskSidebar获取到的userId:', this.userId)
      } catch (error) {
        console.error('初始化用户失败:', error)
      }
    },

    async loadTasks() {
      if (!this.userId) {
        console.error('用户ID未获取，无法加载任务')
        return
      }

      try {
        this.loading = true
        const response = await getTasks(this.userId)
        console.log('TaskSidebar获取任务列表响应:', response)
        
        // 处理后端返回的数据格式
        if (response.success === true || response.success === "true" || response.code === 200) {
          // 可能是 {success: true, data: []} 或 {code: 200, data: []} 格式
          this.tasks = response.data || []
        } else if (Array.isArray(response)) {
          // 如果直接返回数组
          this.tasks = response
        } else {
          console.error('获取任务失败:', response.message || '未知错误')
          this.tasks = []
        }
      } catch (error) {
        console.error('加载任务失败:', error)
        this.tasks = []
      } finally {
        this.loading = false
      }
    },

    async toggleTaskStatus(task) {
      try {
        const currentStatus = task.status || task.taskStatus
        let newStatus
        if (currentStatus === '未完成' || currentStatus === '未开始') {
          newStatus = '进行中'
        } else if (currentStatus === '进行中') {
          newStatus = '已完成'
        } else {
          newStatus = '未完成'
        }

        const taskId = task.task_id || task.taskId
        const updateData = {
          task_name: task.task_name || task.taskName,
          task_note: task.task_note || task.taskNote || null,
          user_id: this.userId,
          status: newStatus
        }
        
        const response = await updateTask(taskId, updateData)
        console.log('TaskSidebar更新任务状态响应:', response)
        
        // 检查响应
        if (response.success === false || response.success === "false") {
          console.error('更新任务状态失败:', response.message)
          return
        }
        
        // 重新加载任务列表
        await this.loadTasks()
        
        // 如果任务刚刚完成，显示获得番茄的通知
        if (newStatus === '已完成' && (currentStatus === '进行中' || currentStatus === '未完成' || currentStatus === '未开始')) {
          const { showEarnTomatoNotification } = await import('@/utils/tomatoNotification')
          showEarnTomatoNotification(1, '任务完成奖励！继续加油哦！')
        }
        
        // 触发事件通知父组件
        this.$emit('task-status-changed', {
          taskId: taskId,
          newStatus: newStatus
        })
      } catch (error) {
        console.error('更新任务状态失败:', error)
      }
    },

    // 与任务管理页保持一致的状态文案
    getStatusText(status) {
      if (!status) return '未完成'
      
      const statusStr = String(status).trim().toLowerCase()
      
      if (
        statusStr === '已完成' || 
        statusStr === '完成' ||
        statusStr === 'finished' || 
        statusStr === 'done' ||
        statusStr === '已完'
      ) {
        return '已完成'
      }
      
      if (
        statusStr === '进行中' ||
        statusStr === '进行' ||
        statusStr === 'in_progress' ||
        statusStr === 'in progress' ||
        statusStr === 'doing'
      ) {
        return '进行中'
      }
      
      return '未完成'
    },

    formatTime(timeStr) {
      if (!timeStr) return '-'
      const date = new Date(timeStr)
      return `${date.getMonth() + 1}/${date.getDate()}`
    },

    toggleCollapsed() {
      this.isCollapsed = !this.isCollapsed
      this.$emit('toggle-collapsed', this.isCollapsed)
    },

    refreshTasks() {
      this.loadTasks()
    },

    goToTaskManagement() {
      this.$router.push('/task-management')
    }
  }
}
</script>

<style scoped>
.task-sidebar {
  width: 100%;
  background: white;
  border: 1px solid #e9ecef;
  border-radius: 12px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  max-height: 500px; /* 添加最大高度 */
}

.sidebar-header {
  padding: 20px;
  border-bottom: 1px solid #e9ecef;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.sidebar-title {
  margin: 0;
  font-size: 1.2em;
  color: #333;
  font-weight: 600;
}

.collapse-btn {
  padding: 6px 12px;
  border: 1px solid #e0e0e0;
  background: white;
  border-radius: 6px;
  font-size: 0.85em;
  cursor: pointer;
  color: #666;
  transition: all 0.2s ease;
}

.collapse-btn:hover {
  background: #f8f9fa;
  border-color: #eeaa67;
  color: #eeaa67;
}

.sidebar-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.task-list {
  flex: 1;
  overflow-y: auto;
  padding: 15px;
}

.task-item {
  background: white;
  border: 1px solid #e9ecef;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 10px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.task-item:hover {
  border-color: #eeaa67;
  transform: translateX(-2px);
}

.task-item.active {
  border-left: 4px solid #eeaa67;
  background: #fff9f2;
}

.task-item-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 8px;
}

.task-name {
  font-weight: 500;
  color: #333;
  flex: 1;
  font-size: 0.95em;
  line-height: 1.4;
}

.task-duration {
  font-size: 0.8em;
  color: #eeaa67;
  font-weight: 500;
  margin-left: 8px;
}

.task-item-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.8em;
}

.task-status {
  padding: 2px 8px;
  border-radius: 10px;
  background: #f8f9fa;
  color: #666;
}

.task-item.active .task-status {
  background: #fff9f2;
  color: #eeaa67;
}

.task-time {
  color: #999;
}

.empty-tasks {
  text-align: center;
  padding: 30px 15px;
  color: #666;
}

.empty-tasks p {
  margin: 0 0 15px 0;
}

.create-task-btn {
  background: #eeaa67;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 0.9em;
  cursor: pointer;
  transition: all 0.2s ease;
}

.create-task-btn:hover {
  background: #e69c55;
}

.sidebar-footer {
  padding: 15px;
  border-top: 1px solid #e9ecef;
  background: #f8f9fa;
}

.task-stats {
  display: flex;
  justify-content: space-between;
  margin-bottom: 10px;
  font-size: 0.85em;
  color: #666;
}

.refresh-btn {
  width: 100%;
  padding: 8px;
  border: 1px solid #e0e0e0;
  background: white;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9em;
  color: #666;
  transition: all 0.2s ease;
}

.refresh-btn:hover:not(:disabled) {
  background: #f0f0f0;
  border-color: #eeaa67;
  color: #eeaa67;
}

.refresh-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .task-sidebar {
    width: 100%;
    height: auto;
    position: relative;
    border-left: none;
    border-top: 1px solid #e9ecef;
  }
}
</style>