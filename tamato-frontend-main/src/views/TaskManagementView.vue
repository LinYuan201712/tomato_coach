<template>
  <div class="task-management-view">
    <!-- 顶部导航栏 -->
    <nav class="navbar">
      <div class="nav-brand">Tomato</div>
      <div class="nav-links">
        <a class="nav-link" @click="goToHome">返回首页</a>
      </div>
    </nav>

    <main class="main-content">
      <!-- 页面标题 -->
      <div class="page-header">
        <h1 class="page-title">任务管理</h1>
        <p class="page-subtitle">简单高效的任务清单</p>
      </div>

      <div class="task-layout">
        <!-- 左侧：任务列表 -->
        <div class="task-list-section">
          <div class="section-header">
            <h2 class="section-title">我的任务</h2>
            <div class="header-actions">
              <div class="range-picker">
                <input type="date" v-model="startDate" class="date-picker" />
                <span class="range-sep">至</span>
                <input type="date" v-model="endDate" class="date-picker" />
              </div>
              <button @click="openCreateModal" class="create-task-btn">
                <span class="btn-icon">+</span>
                新建任务
              </button>
            </div>
          </div>

          <!-- 任务状态筛选 -->
          <div class="filter-tabs">
            <button 
              v-for="tab in statusTabs" 
              :key="tab.value"
              @click="activeTab = tab.value"
              :class="['tab-btn', { active: activeTab === tab.value }]"
            >
              {{ tab.label }}
              <span class="tab-count">{{ getTaskCount(tab.value) }}</span>
            </button>
          </div>

          <!-- 任务列表 -->
          <div class="tasks-container">
            <div 
              v-for="task in filteredTasks" 
              :key="getTaskId(task)"
              class="task-card"
              :class="getTaskStatusClass(task.status)"
            >
              <div class="task-main">
                <div class="task-header">
                  <div class="task-title-section">
                    <h3 class="task-title">{{ getTaskName(task) }}</h3>
                    <span class="task-duration">{{ task.duration || 25 }}分钟</span>
                  </div>
                  <div class="task-actions">
                     <!-- 执行按钮 -->
                     <button 
                       v-if="!isTaskCompleted(task)"
                       @click="executeTask(task)"
                       class="action-btn execute-btn"
                       title="进入个人自习室执行任务"
                     >
                       执行
                     </button>
                     <!-- 完成按钮 -->
                    <button 
                      v-if="!isTaskCompleted(task)"
                      @click="completeTask(task)"
                      class="action-btn complete-btn"
                    >
                      完成
                    </button>
                    <!-- 编辑按钮 - 已完成的任务不可编辑 -->
                    <button 
                      v-if="!isTaskCompleted(task)"
                      @click="editTask(task)" 
                      class="action-btn edit-btn"
                    >
                      编辑
                    </button>
                    <!-- 删除按钮 -->
                    <button @click="deleteTask(getTaskId(task))" class="action-btn delete-btn">
                      删除
                    </button>
                  </div>
                </div>
                
                <p v-if="getTaskNote(task)" class="task-note">{{ getTaskNote(task) }}</p>
                
                <div class="task-footer">
                  <span class="create-time">
                    创建: {{ formatTime(task.created_at || task.createdAt) }}
                  </span>
                  <span class="status-tag" :class="getStatusTagClass(task.status)">
                    {{ getStatusText(task.status) }}
                  </span>
                </div>
              </div>
            </div>

            <!-- 空状态 -->
            <div v-if="filteredTasks.length === 0" class="empty-state-premium">
              <div class="empty-illustration">
                <div class="icon-circle">
                  <span class="empty-icon-large">📋</span>
                </div>
              </div>
              <h3 class="empty-title">暂无任务</h3>
              <p class="empty-desc">在这个时间段内没有找到任务，开启新的学习计划吧</p>
              <button @click="openCreateModal" class="create-first-btn-premium">
                <span class="btn-plus">+</span> 创建第一个任务
              </button>
            </div>
          </div>
        </div>

        <!-- 右侧：统计信息 -->
        <div class="stats-section">
          <div class="stats-card">
            <h3 class="stats-title">今日统计</h3>
            <div class="stats-grid">
              <div class="stat-item">
                <div class="stat-value">{{ todayStats.total }}</div>
                <div class="stat-label">今日任务</div>
              </div>
              <div class="stat-item">
                <div class="stat-value">{{ todayStats.completed }}</div>
                <div class="stat-label">已完成</div>
              </div>
              <div class="stat-item">
                <div class="stat-value">{{ todayStats.pending }}</div>
                <div class="stat-label">未完成</div>
              </div>
            </div>
          </div>

          <div class="stats-card">
            <h3 class="stats-title">总任务统计</h3>
            <div class="stats-grid">
              <div class="stat-item">
                <div class="stat-value">{{ totalTasks }}</div>
                <div class="stat-label">总任务</div>
              </div>
              <div class="stat-item">
                <div class="stat-value">{{ completedTasks }}</div>
                <div class="stat-label">总完成</div>
              </div>
              <div class="stat-item">
                <div class="stat-value">{{ pendingTasks }}</div>
                <div class="stat-label">总未完</div>
              </div>
            </div>
          </div>

          <!-- 快速导航 -->
          <div class="quick-nav-card">
            <h3 class="nav-title">快速导航</h3>
            <div class="nav-buttons">
              <button @click="goToHome" class="nav-btn create-room-btn">
                返回个人中心
              </button>
              <button @click="goToJoinRoom" class="nav-btn join-room-btn">
                加入自习室
              </button>
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- 创建/编辑任务弹窗 -->
    <div v-if="showModal" class="modal-overlay-premium" @click.self="closeModal">
      <div class="task-modal-premium">
        <div class="modal-header-premium">
          <div class="modal-icon-circle">
            <span class="modal-icon">✍️</span>
          </div>
          <h3 class="modal-title-premium">{{ editingTask ? '编辑任务' : '创建新任务' }}</h3>
          <button @click="closeModal" class="close-x-btn">×</button>
        </div>
        
        <form @submit.prevent="submitTask" class="task-form-premium">
          <div class="form-group-premium">
            <label class="form-label-premium">任务名称 <span class="required-star">*</span></label>
            <div class="input-wrapper-premium">
              <input 
                type="text" 
                v-model="taskForm.task_name"
                placeholder="你想完成什么？"
                class="form-input-premium"
                required
              >
            </div>
          </div>
 
          <div class="form-group-premium">
            <label class="form-label-premium">任务备注</label>
            <div class="input-wrapper-premium">
              <textarea 
                v-model="taskForm.task_note"
                placeholder="添加一些细节备注..."
                class="form-textarea-premium"
                rows="3"
              ></textarea>
            </div>
          </div>
 
          <div class="form-group-premium">
            <label class="form-label-premium">计划时长</label>
            <div class="duration-chips-grid">
              <div 
                v-for="opt in durationOptions" 
                :key="opt.value"
                class="duration-chip"
                :class="{ active: taskForm.duration == opt.value }"
                @click="taskForm.duration = opt.value"
              >
                <span class="chip-icon">🍅</span>
                <div class="chip-content">
                  <span class="chip-minutes">{{ opt.label }}</span>
                  <span class="chip-tomatoes">{{ opt.sub }}</span>
                </div>
              </div>
            </div>
          </div>
 
          <div class="modal-footer-premium">
            <button type="button" @click="closeModal" class="cancel-btn-premium">
              取消
            </button>
            <button type="submit" class="confirm-btn-premium">
              {{ editingTask ? '保存修改' : '立即创建' }}
            </button>
          </div>
        </form>
      </div>
    </div>
    
    <!-- 自定义确认弹窗 (简约风格) -->
    <div v-if="confirmModal.show" class="modal-overlay" @click.self="closeConfirmModal">
      <div class="confirm-modal-simple" :class="confirmModal.type">
        <h3 class="confirm-title">{{ confirmModal.title }}</h3>
        <p class="confirm-message">{{ confirmModal.message }}</p>
        <div class="confirm-actions">
          <button @click="confirmModal.onConfirm" class="confirm-btn-main">确认</button>
          <button @click="closeConfirmModal" class="confirm-btn-cancel">取消</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { getTasks, createTask, updateTask, deleteTask, completeTaskApi } from '@/api/tasks'
import { getCurrentUser } from '@/api/user'

export default {
  name: 'TaskManagementView',
  data() {
    return {
      loading: false, tasks: [], userId: null, activeTab: 'all', showModal: false, editingTask: null,
      startDate: new Date().toISOString().substr(0, 10), // 默认显示今天
      endDate: new Date().toISOString().substr(0, 10),
      taskForm: { task_name: '', task_note: '', duration: 25 },
      durationOptions: [
        { label: '25分钟', sub: '1个番茄', value: 25 },
        { label: '50分钟', sub: '2个番茄', value: 50 },
        { label: '75分钟', sub: '3个番茄', value: 75 },
        { label: '100分钟', sub: '4个番茄', value: 100 }
      ],
      statusTabs: [ { label: '全部', value: 'all' }, { label: '未完成', value: '未完成' }, { label: '已完成', value: '已完成' } ],
      confirmModal: { show: false, title: '', message: '', type: 'info', onConfirm: () => {} }
    }
  },
  computed: {
    sortedTasks() {
      return [...this.tasks].sort((a, b) => {
        const aC = this.getStatusText(a.status) === '已完成', bC = this.getStatusText(b.status) === '已完成'
        if (aC !== bC) return aC - bC
        return new Date(b.created_at || b.createdAt || 0) - new Date(a.created_at || a.createdAt || 0)
      })
    },
    filteredTasks() {
      if (this.activeTab === '未完成') {
        // 未完成任务永远显示所有时间段
        return this.sortedTasks.filter(t => this.getStatusText(t.status) === '未完成')
      }

      // 全部和已完成需要按时间段筛选
      let filtered = this.sortedTasks.filter(t => {
        const taskDate = new Date(t.created_at || t.createdAt).toISOString().substr(0, 10)
        return taskDate >= this.startDate && taskDate <= this.endDate
      })

      if (this.activeTab === 'all') {
        // "全部" 模式：时间段内的已完成 + 所有的未完成
        const allPending = this.sortedTasks.filter(t => this.getStatusText(t.status) === '未完成')
        const periodCompleted = filtered.filter(t => this.getStatusText(t.status) === '已完成')
        
        // 合并去重并排序
        const combined = [...allPending, ...periodCompleted]
        return combined.sort((a, b) => {
          const aC = this.getStatusText(a.status) === '已完成', bC = this.getStatusText(b.status) === '已完成'
          if (aC !== bC) return aC - bC
          return new Date(b.created_at || b.createdAt || 0) - new Date(a.created_at || a.createdAt || 0)
        })
      }

      return filtered.filter(t => this.getStatusText(t.status) === this.activeTab)
    },
    todayStats() {
      const today = new Date().toISOString().substr(0, 10)
      const todayTasks = this.tasks.filter(t => {
        const taskDate = new Date(t.created_at || t.createdAt).toISOString().substr(0, 10)
        return taskDate === today
      })
      return {
        total: todayTasks.length,
        completed: todayTasks.filter(t => this.isTaskCompleted(t)).length,
        pending: todayTasks.filter(t => !this.isTaskCompleted(t)).length
      }
    },
    totalTasks() { return this.tasks.length },
    completedTasks() { return this.tasks.filter(t => this.isTaskCompleted(t)).length },
    pendingTasks() { return this.tasks.filter(t => !this.isTaskCompleted(t)).length }
  },
  async mounted() { await this.initUser(); await this.loadTasks() },
  methods: {
    getTaskId(t) { return t.task_id || t.taskId },
    getTaskName(t) { return t.task_name || t.taskName },
    getTaskNote(t) { return t.task_note || t.taskNote },
    isTaskCompleted(t) { return this.getStatusText(t.status) === '已完成' },
    getTaskCount(s) {
      if (s === '未完成') return this.pendingTasks // 永远显示总数
      if (s === 'all') return this.filteredTasks.length // 显示当前筛选下的全部
      // 已完成显示筛选时间段内的
      return this.tasks.filter(t => {
        const statusMatch = this.getStatusText(t.status) === s
        const taskDate = new Date(t.created_at || t.createdAt).toISOString().substr(0, 10)
        const dateMatch = taskDate >= this.startDate && taskDate <= this.endDate
        return statusMatch && dateMatch
      }).length
    },
    getTaskStatusClass(s) { return this.getStatusText(s) === '已完成' ? 'task-completed' : 'task-pending' },
    getStatusTagClass(s) { return this.getStatusText(s) === '已完成' ? 'tag-completed' : 'tag-pending' },
    getStatusText(status) {
      if (!status) return '未完成'
      const s = String(status).trim().toUpperCase()
      // 适配多种后端可能的返回状态
      if (['已完成', '完成', 'FINISHED', 'DONE', 'COMPLETED', 'TRUE', '1'].includes(s)) {
        return '已完成'
      }
      if (['进行中', 'IN_PROGRESS', 'DOING', 'ACTIVE'].includes(s)) {
        return '进行中'
      }
      return '未完成'
    },
    formatTime(t) { return t ? new Date(t).toLocaleDateString() : '-' },
    async initUser() {
      try { const res = await getCurrentUser(); if (res?.data) this.userId = res.data.user_id || res.data.id } catch (e) { console.error('获取用户失败:', e) }
    },
    async loadTasks() {
      if (!this.userId) return
      try {
        this.loading = true
        const res = await getTasks(this.userId)
        console.log('✅ 加载任务原始数据:', res)
        this.tasks = res.data || (Array.isArray(res) ? res : [])
      } catch (e) { console.error('加载任务失败:', e) } finally { this.loading = false }
    },
    executeTask(t) {
      const taskId = this.getTaskId(t)
      this.$router.push({
        path: '/study-room/personal',
        query: { taskId }
      })
    },
    async completeTask(t) {
      this.confirmModal = {
        show: true, title: '确认完成', message: `确定要完成任务 "${this.getTaskName(t)}" 吗？`, type: 'info',
        onConfirm: async () => {
          this.closeConfirmModal()
          try {
            const taskId = this.getTaskId(t)
            console.log(`🚀 准备完成任务 ID: ${taskId}`)
            // 使用后端专门的完成接口
            const res = await completeTaskApi(taskId)
            console.log('✅ 完成任务响应:', res)
            await this.loadTasks()
          } catch (e) {
            console.error('完成任务失败:', e)
            alert('操作失败，请检查网络或权限')
          }
        }
      }
    },
    async submitTask() {
      try {
        const d = { ...this.taskForm, user_id: this.userId }
        if (this.editingTask) await updateTask(this.getTaskId(this.editingTask), d)
        else await createTask(d)
        this.closeModal(); await this.loadTasks()
      } catch (e) { console.error('提交任务失败:', e) }
    },
    async deleteTask(id) {
      this.confirmModal = {
        show: true, title: '确认删除', message: '确定要删除这个任务吗？', type: 'danger',
        onConfirm: async () => {
          this.closeConfirmModal()
          try {
            await deleteTask(id)
            await this.loadTasks()
          } catch (e) { console.error('删除任务失败:', e) }
        }
      }
    },
    openCreateModal() { this.editingTask = null; this.taskForm = { task_name: '', task_note: '', duration: 25 }; this.showModal = true },
    editTask(t) {
      this.editingTask = t
      this.taskForm = { task_name: this.getTaskName(t), task_note: this.getTaskNote(t) || '', duration: t.duration || 25 }
      this.showModal = true
    },
    closeModal() { this.showModal = false },
    closeConfirmModal() { this.confirmModal.show = false },
    goToHome() { this.$router.push('/home') },
    goToJoinRoom() { this.$router.push('/join-room') }
  }
}
</script>

<style scoped>
.task-management-view { min-height: 100vh; background-color: #ffffff; }
.navbar { display: flex; justify-content: space-between; align-items: center; padding: 16px 20px; background: white; border-bottom: 1px solid #ffe4cc; position: sticky; top: 0; z-index: 100; }
.nav-brand { font-size: 1.5rem; font-weight: bold; color: #eeaa67; }
.nav-links { display: flex; gap: 20px; }
.nav-link { cursor: pointer; padding: 8px 12px; border-radius: 6px; transition: 0.2s; color: #333; }
.nav-link:hover { background-color: #fff5eb; color: #eeaa67; }

.main-content { max-width: 1200px; margin: 0 auto; padding: 40px 20px; }
.page-header { text-align: center; margin-bottom: 40px; }
.page-title { font-size: 2.5rem; color: #333; font-weight: bold; margin-bottom: 10px; }
.page-subtitle { color: #666; }

.task-layout { display: grid; grid-template-columns: 1fr 350px; gap: 30px; }

.task-list-section { background: white; border-radius: 12px; padding: 30px; border: 1px solid #e9ecef; box-shadow: 0 2px 10px rgba(0,0,0,0.05); }
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 25px; }
.header-actions { display: flex; gap: 12px; align-items: center; }
.range-picker { display: flex; align-items: center; gap: 8px; background: #fff; border: 1px solid #e9ecef; border-radius: 8px; padding: 2px 8px; }
.range-sep { color: #999; font-size: 0.8rem; }
.date-picker { border: none; padding: 6px; font-size: 0.85rem; color: #333; outline: none; background: transparent; cursor: pointer; }
.date-picker:focus { color: #eeaa67; }
.create-task-btn { background: #eeaa67; color: white; border: none; padding: 10px 20px; border-radius: 8px; font-weight: bold; cursor: pointer; transition: 0.2s; white-space: nowrap; }
.create-task-btn:hover { background: #e69c55; }

/* 空状态美化 */
.empty-state-premium { text-align: center; padding: 60px 20px; background: #fdfdfd; border-radius: 12px; border: 1px dashed #e9ecef; margin-top: 20px; }
.empty-illustration { margin-bottom: 20px; display: flex; justify-content: center; }
.icon-circle { width: 80px; height: 80px; background: #fff5eb; border-radius: 50%; display: flex; align-items: center; justify-content: center; }
.empty-icon-large { font-size: 40px; }
.empty-title { font-size: 1.5rem; color: #333; margin: 0 0 10px 0; font-weight: 600; }
.empty-desc { color: #999; margin-bottom: 30px; font-size: 0.95rem; }
.create-first-btn-premium { background: #eeaa67; color: white; border: none; padding: 12px 24px; border-radius: 30px; font-weight: bold; cursor: pointer; transition: 0.3s; box-shadow: 0 4px 12px rgba(238, 170, 103, 0.3); display: inline-flex; align-items: center; gap: 8px; }
.create-first-btn-premium:hover { background: #e69c55; transform: translateY(-2px); box-shadow: 0 6px 15px rgba(238, 170, 103, 0.4); }
.btn-plus { font-size: 1.2rem; }

.filter-tabs { display: flex; gap: 8px; margin-bottom: 25px; background: #f8f9fa; padding: 4px; border-radius: 10px; }
.tab-btn { flex: 1; padding: 10px; border: none; background: none; cursor: pointer; border-radius: 6px; font-weight: 500; color: #666; }
.tab-btn.active { background: white; color: #eeaa67; box-shadow: 0 2px 5px rgba(0,0,0,0.1); }

.task-card { background: white; border-radius: 10px; padding: 20px; margin-bottom: 16px; border: 1px solid #e9ecef; transition: 0.2s; }
.task-card:hover { border-color: #eeaa67; box-shadow: 0 4px 12px rgba(0,0,0,0.05); }
.task-completed { border-left: 4px solid #28a745; opacity: 0.7; }

.task-header { display: flex; justify-content: space-between; align-items: flex-start; }
.task-title { font-size: 1.2rem; margin: 0; color: #333; }
.task-duration { color: #eeaa67; font-size: 0.9rem; font-weight: bold; }
.task-actions { display: flex; gap: 10px; }

.action-btn { padding: 6px 12px; border-radius: 6px; border: none; font-size: 0.85rem; font-weight: bold; cursor: pointer; transition: 0.2s; }
.complete-btn { background: #e7f5e9; color: #2b8a3e; }
.edit-btn { background: #fff9db; color: #f08c00; }
.delete-btn { background: #fff5f5; color: #fa5252; }

.task-note { color: #666; font-size: 0.9rem; margin: 10px 0; }
.task-footer { display: flex; justify-content: space-between; border-top: 1px solid #f8f9fa; padding-top: 12px; margin-top: 10px; }

.stats-card, .quick-nav-card { background: white; border-radius: 12px; padding: 25px; border: 1px solid #e9ecef; margin-bottom: 20px; }
.stats-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 15px; text-align: center; }
.stat-value { font-size: 1.5rem; font-weight: bold; color: #eeaa67; }
.stat-label { font-size: 0.8rem; color: #999; }

.nav-buttons { display: flex; flex-direction: column; gap: 12px; }
.nav-btn { padding: 12px; border-radius: 8px; border: none; font-weight: bold; cursor: pointer; transition: 0.2s; }
.create-room-btn { background: #eeaa67; color: white; }
.join-room-btn { background: #f8f9fa; color: #333; border: 1px solid #e9ecef; }

.modal-overlay-premium { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.4); backdrop-filter: blur(4px); display: flex; align-items: center; justify-content: center; z-index: 1000; }
.task-modal-premium { background: white; border-radius: 20px; padding: 0; width: 90%; max-width: 450px; overflow: hidden; box-shadow: 0 20px 40px rgba(0,0,0,0.15); animation: modalIn 0.3s ease; }
@keyframes modalIn { from { opacity: 0; transform: translateY(20px); } to { opacity: 1; transform: translateY(0); } }

.modal-header-premium { padding: 24px; background: #fff5eb; display: flex; align-items: center; gap: 15px; position: relative; }
.modal-icon-circle { width: 44px; height: 44px; background: white; border-radius: 12px; display: flex; align-items: center; justify-content: center; box-shadow: 0 4px 10px rgba(238, 170, 103, 0.15); }
.modal-icon { font-size: 20px; }
.modal-title-premium { margin: 0; font-size: 1.25rem; color: #333; font-weight: 700; flex: 1; }
.close-x-btn { border: none; background: none; font-size: 1.5rem; color: #999; cursor: pointer; padding: 0 5px; transition: 0.2s; }
.close-x-btn:hover { color: #eeaa67; transform: rotate(90deg); }

.task-form-premium { padding: 24px; }
.form-group-premium { margin-bottom: 20px; }
.form-label-premium { display: block; margin-bottom: 8px; font-size: 0.9rem; font-weight: 600; color: #555; }
.required-star { color: #ff6b6b; margin-left: 2px; }

.input-wrapper-premium { position: relative; width: 100%; }
.form-input-premium, .form-textarea-premium, .form-select-premium { 
  width: 100%; padding: 12px 16px; border: 2px solid #f1f3f5; border-radius: 12px; 
  font-size: 0.95rem; color: #333; transition: all 0.2s ease; outline: none; background: #f8f9fa;
}
.form-input-premium:focus, .form-textarea-premium:focus, .form-select-premium:focus { 
  border-color: #eeaa67; background: #fff; box-shadow: 0 0 0 4px rgba(238, 170, 103, 0.1); 
}

.form-textarea-premium { resize: none; min-height: 100px; }

/* 时长选项卡样式 */
.duration-chips-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 12px; }
.duration-chip { 
  display: flex; align-items: center; gap: 12px; padding: 12px 16px; 
  background: #f8f9fa; border: 2px solid #f1f3f5; border-radius: 16px; 
  cursor: pointer; transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}
.duration-chip:hover { border-color: #eeaa67; background: #fffcf9; transform: translateY(-2px); }
.duration-chip.active { 
  background: #eeaa67; border-color: #eeaa67; 
  box-shadow: 0 4px 12px rgba(238, 170, 103, 0.3);
}
.chip-icon { font-size: 1.2rem; filter: grayscale(1); transition: 0.2s; }
.duration-chip.active .chip-icon { filter: grayscale(0); }
.chip-content { display: flex; flex-direction: column; }
.chip-minutes { font-size: 0.95rem; font-weight: 700; color: #333; }
.chip-tomatoes { font-size: 0.75rem; color: #999; }
.duration-chip.active .chip-minutes, .duration-chip.active .chip-tomatoes { color: white; }

.modal-footer-premium { display: flex; gap: 12px; margin-top: 32px; }
.cancel-btn-premium { flex: 1; padding: 12px; border: 1px solid #e9ecef; border-radius: 12px; background: white; color: #666; font-weight: 600; cursor: pointer; transition: 0.2s; }
.cancel-btn-premium:hover { background: #f8f9fa; border-color: #ddd; }
.confirm-btn-premium { flex: 2; padding: 12px; border: none; border-radius: 12px; background: #eeaa67; color: white; font-weight: 700; cursor: pointer; transition: 0.3s; box-shadow: 0 4px 12px rgba(238, 170, 103, 0.3); }
.confirm-btn-premium:hover { background: #e69c55; transform: translateY(-2px); box-shadow: 0 6px 15px rgba(238, 170, 103, 0.4); }

/* 弹窗简约风格 */
.confirm-modal-simple { background: white; border-radius: 12px; padding: 25px; width: 90%; max-width: 400px; text-align: center; box-shadow: 0 10px 30px rgba(0,0,0,0.2); }
.confirm-title { margin: 0 0 15px 0; color: #333; }
.confirm-message { color: #666; margin-bottom: 25px; }
.confirm-actions { display: flex; gap: 12px; }
.confirm-btn-main, .confirm-btn-cancel { flex: 1; padding: 12px; border: none; border-radius: 8px; font-weight: bold; cursor: pointer; }
.confirm-btn-main { background: #eeaa67; color: white; }
.danger .confirm-btn-main { background: #ff6b6b; color: white; }
.confirm-btn-cancel { background: #f1f3f5; color: #666; }
</style>
.execute-btn { background: linear-gradient(135deg, #6366f1, #8b5cf6) !important; color: white !important; }
.execute-btn:hover { background: linear-gradient(135deg, #4f46e5, #7c3aed) !important; transform: translateY(-2px); box-shadow: 0 4px 12px rgba(99, 102, 241, 0.3); }
