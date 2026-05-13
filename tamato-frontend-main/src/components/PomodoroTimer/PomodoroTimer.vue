<template>
  <div class="pomodoro-timer">
    <!-- 初始状态 -->
    <div v-if="!isActive" class="timer-start-screen">
      <img src="@/assets/logo.png" alt="番茄时钟" class="pomodoro-logo" />
      <h2 class="timer-title">番茄时钟</h2>
      <p class="timer-subtitle">专注 · 高效 · 平衡</p>
      <div class="encouragement-card">
        <div class="quote-icon">"</div>
        <p class="encouragement-text">{{ currentEncouragement }}</p>
      </div>
      <button @click="startTimer" class="start-button" :disabled="isLoading">
        <span v-if="isLoading" class="loading-spinner-small"></span>
        <span v-else>{{ isLoading ? "准备中..." : "开始专注" }}</span>
      </button>
    </div>

    <!-- 激活状态 -->
    <div v-else class="timer-active-screen">
      <!-- 顶部模式指示器 -->
      <div class="mode-indicator" :class="modeClass">
        <span class="mode-dot"></span>
        <span class="mode-text">{{ timerModeText }}</span>
      </div>

      <!-- 当前任务信息 -->
      <div v-if="selectedTask && !isBreak" class="current-task-info">
        <div class="task-info-header">
          <span class="task-icon">📚</span>
          <span class="task-name">{{ getTaskName(selectedTask) }}</span>
        </div>
        <div v-if="getTaskNote(selectedTask)" class="task-note-text">
          {{ getTaskNote(selectedTask) }}
        </div>
      </div>

      <!-- 圆形计时器 -->
      <div class="circular-timer">
        <!-- 外圈光环效果 -->
        <div class="timer-halo" :class="modeClass"></div>

        <!-- 计时数字 -->
        <div class="timer-display">
          <div class="timer-digits">{{ formattedTime }}</div>
          <div class="timer-label">{{ isBreak ? "休息时间" : "专注时间" }}</div>
        </div>

        <!-- 进度环 -->
        <svg class="progress-ring" width="280" height="280">
          <defs>
            <linearGradient
              id="gradient-focus"
              x1="0%"
              y1="0%"
              x2="100%"
              y2="100%"
            >
              <stop offset="0%" :stop-color="gradientStart" />
              <stop offset="100%" :stop-color="gradientEnd" />
            </linearGradient>
            <linearGradient
              id="gradient-break"
              x1="0%"
              y1="0%"
              x2="100%"
              y2="100%"
            >
              <stop offset="0%" stop-color="#63e6be" />
              <stop offset="100%" stop-color="#96f2d7" />
            </linearGradient>
          </defs>
          <circle class="progress-ring-background" cx="140" cy="140" r="130" />
          <circle
            class="progress-ring-fill"
            cx="140"
            cy="140"
            r="130"
            :stroke="isBreak ? 'url(#gradient-break)' : 'url(#gradient-focus)'"
            :stroke-dasharray="circumference"
            :stroke-dashoffset="progressOffset"
          />
        </svg>
      </div>

      <!-- 当前鼓励语 -->
      <div class="current-encouragement">
        <p>{{ currentEncouragement }}</p>
      </div>

      <!-- 控制按钮组 -->
      <div class="control-buttons">
        <button
          v-if="!isPaused && !isBreak"
          @click="pauseTimer"
          class="control-button pause-button"
        >
          <span class="button-icon"></span>
          <span class="button-text">暂停</span>
        </button>
        <button
          v-if="isPaused && !isBreak"
          @click="resumeTimer"
          class="control-button resume-button"
        >
          <span class="button-icon"></span>
          <span class="button-text">继续</span>
        </button>
        <button @click="stopTimer" class="control-button stop-button">
          <span class="button-icon"></span>
          <span class="button-text">终止</span>
        </button>

        <!-- 休息跳过按钮 -->
        <button
          v-if="isBreak && !isPaused"
          @click="skipBreak"
          class="control-button skip-button"
        >
          <span class="button-icon">⏭️</span>
          <span class="button-text">跳过休息</span>
        </button>
      </div>

      <!-- 统计数据 -->
      <div class="stats-panel">
        <div class="stat-item">
          <div class="stat-content">
            <div class="stat-value">{{ completedSessions }}</div>
            <div class="stat-label">已完成</div>
          </div>
        </div>
        <div class="stat-item">
          <div class="stat-content">
            <div class="stat-value">{{ formatStatsTime(totalFocusTime) }}</div>
            <div class="stat-label">总专注</div>
          </div>
        </div>
      </div>
      
    </div>

    <!-- 暂停模态框 -->
    <div v-if="showPauseModal" class="modal-container">
      <div class="modal-card">
        <div class="modal-header">
          <div class="modal-icon"></div>
          <h3 class="modal-title">暂停中</h3>
        </div>
        <div class="modal-body">
          <p class="modal-message">
            暂停限制时长为3分钟，避免过长的打断影响专注流程。
          </p>
          <div class="pause-timer">
            <div class="pause-progress" :style="pauseProgressStyle"></div>
            <div class="pause-time">{{ formatTime(pauseTimeLeft) }}</div>
          </div>
        </div>
        <button @click="resumeTimer" class="modal-action-button">
          立即继续专注
        </button>
      </div>
    </div>

    <!-- 跳过休息确认框 -->
    <div v-if="showSkipConfirm" class="modal-container">
      <div class="modal-card">
        <div class="modal-header">
          <div class="modal-icon"></div>
          <h3 class="modal-title">跳过休息？</h3>
        </div>
        <div class="modal-body">
          <p class="modal-message">
            确定要跳过休息时间，立即开始下一轮专注吗？
          </p>
        </div>
        <div class="modal-actions">
          <button @click="confirmSkip" class="modal-button confirm-button">
            是的，继续专注
          </button>
          <button @click="cancelSkip" class="modal-button cancel-button">
            继续休息
          </button>
        </div>
      </div>
    </div>

    <!-- 提前结束任务确认框 -->
    <div v-if="showTaskEndConfirm" class="modal-container">
      <div class="modal-card">
        <div class="modal-header">
          <div class="modal-icon">⚠️</div>
          <h3 class="modal-title">提前结束任务？</h3>
        </div>
        <div class="modal-body">
          <p class="modal-message">
            确定要提前结束当前任务"{{ selectedTask ? getTaskName(selectedTask) : '' }}"吗？
          </p>
          <p class="modal-tip">
            如果选择"是"，该任务将被标记为已完成。
          </p>
        </div>
        <div class="modal-actions">
          <button @click="confirmTaskEnd" class="modal-button confirm-button">
            是的，提前完成
          </button>
          <button @click="cancelTaskEnd" class="modal-button cancel-button">
            取消
          </button>
        </div>
      </div>
    </div>

    <!-- 切屏警告模态框 -->
    <div v-if="showDistractionModal" class="modal-container distraction-modal-overlay" @click.self="closeDistractionModal">
      <div class="modal-card distraction-modal">
        <div class="distraction-modal-header">
          <div class="distraction-icon">
            <svg viewBox="0 0 64 64" width="64" height="64">
              <circle cx="32" cy="32" r="30" fill="#fff3cd" stroke="#ffc107" stroke-width="2"/>
              <path d="M32 20 L32 36 M32 40 L32 44" stroke="#ffc107" stroke-width="3" stroke-linecap="round"/>
            </svg>
          </div>
          <h3 class="distraction-title">检测到分心行为</h3>
        </div>
        <div class="distraction-modal-body">
          <p class="distraction-message">
            您刚才切换了标签页或应用窗口
          </p>
          <p class="distraction-tip">
            为了保持专注，请关闭抖音、视频网站等分心应用，<br/>
            回到当前页面继续专注学习。
          </p>
          <div class="distraction-note">
            <span class="note-icon">⏱️</span>
            <span>专注时间仍在继续，请尽快回到学习状态！</span>
          </div>
        </div>
        <button @click="closeDistractionModal" class="distraction-button">
          我知道了
        </button>
      </div>
    </div>

    <!-- 任务选择弹窗 -->
    <div v-if="showTaskSelect" class="modal-container">
      <div class="modal-card task-select-modal">
        <div class="modal-header">
          <h3 class="modal-title">选择任务</h3>
        </div>
        <div class="modal-body">
          <div v-if="loadingTasks" class="loading-tasks">
            <div class="loading-spinner-small"></div>
            <p>加载任务中...</p>
          </div>
          <div v-else-if="availableTasks.length === 0" class="no-tasks">
            <p class="modal-message">暂无可用任务</p>
            <p class="modal-tip">请先到任务管理页面创建任务</p>
          </div>
          <div v-else class="task-list">
            <div
              v-for="task in availableTasks"
              :key="getTaskId(task)"
              @click="selectTask(task)"
              class="task-item"
              :class="{ 'task-selected': selectedTask && getTaskId(selectedTask) === getTaskId(task) }"
            >
              <div class="task-item-content">
                <div class="task-item-title">{{ getTaskName(task) }}</div>
                <div class="task-item-meta">
                  <span class="task-duration-badge">{{ task.duration || 25 }}分钟</span>
                  <span v-if="getTaskNote(task)" class="task-note-preview">{{ getTaskNote(task) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-actions">
          <button @click="confirmTaskSelection" class="modal-button confirm-button" :disabled="!selectedTask">
            开始专注
          </button>
          <button @click="cancelTaskSelection" class="modal-button cancel-button">
            取消
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { getTasks } from '@/api/tasks'
import { getCurrentUser } from '@/api/user'

// 鼓励话语语料库
const ENCOURAGEMENTS = [
  "行远自迩，笃行不怠",
  "焚膏继晷，兀兀穷年",
  "怀瑾握瑜，风禾尽起",
  "尽小者大，慎微者著",
  "浮舟沧海，立马昆仑",
  "流水不争先，争的是滔滔不绝",
  "想，全是问题；做，才有答案",
  "百舸争流，奋楫者先",
  "青矜之志，履践致远",
  "为者常成，行者常至",
  "马行千里，不洗尘沙",
  "日拱一卒，功不唐捐",
  "与其抱怨天黑，不如提灯前行",
  "站在山顶的人，不会嘲笑半山腰的攀登者",
  "把行动交给现在，把结果交给时间",
  "向下扎根，向上生花",
  "追风赶月莫停留，平芜尽处是春山",
];

export default {
  name: "PomodoroTimer",
  props: {
    active: {
      type: Boolean,
      default: false,
    },
    userId: {
      type: [Number, String],
      default: null,
    },
    roomId: {
      type: [Number, String],
      default: null,
    },
  },
  emits: [
    'timer-started',
    'timer-paused',
    'timer-resumed',
    'timer-stopped',
    'user-status-change',
    'task-selected',
    'focus-completed',
    'break-skipped',
    'task-completed'
  ],
  data() {
    return {
      // 计时器状态
      isActive: false,
      isPaused: false,
      isBreak: false,
      isLoading: false,

      // 时间设置（秒）
      focusTime: 25 * 60,
      breakTime: 5 * 60,
      maxPauseTime: 3 * 60,

      // 当前时间
      timeLeft: 25 * 60,
      pauseTimeLeft: 3 * 60,

      // 计时器引用
      timer: null,
      pauseTimerRef: null,

      // 统计数据
      completedSessions: 0,
      totalFocusTime: 0,

      // 弹窗控制
      showPauseModal: false,
      showSkipConfirm: false,
      showTaskSelect: false,
      showTaskEndConfirm: false,

      // 任务相关
      availableTasks: [],
      selectedTask: null,
      loadingTasks: false,
      currentUserId: null,

      // 随机鼓励语
      currentEncouragement: "",
      encouragements: [...ENCOURAGEMENTS],

      // 清爽的橘黄红色渐变
      gradientStart: "#ffa94d", // 更淡的橙色
      gradientEnd: "#ff8787", // 更淡的珊瑚红

      // 页面可见性监听
      isFocusing: false, // 是否正在专注中
      pageHiddenTime: null, // 页面隐藏的时间
      visibilityHandler: null, // 可见性变化处理器
      focusHandler: null, // 窗口焦点变化处理器
      blurHandler: null, // 窗口失去焦点处理器
      showDistractionModal: false, // 是否显示分心警告模态框
      
    };
  },
  computed: {
    formattedTime() {
      const minutes = Math.floor(this.timeLeft / 60);
      const seconds = this.timeLeft % 60;
      return `${minutes.toString().padStart(2, "0")}:${seconds
        .toString()
        .padStart(2, "0")}`;
    },

    timerModeText() {
      if (this.isBreak) return "休息时间";
      if (this.isPaused) return "已暂停";
      return "专注时间";
    },

    modeClass() {
      if (this.isBreak) return "mode-break";
      if (this.isPaused) return "mode-pause";
      return "mode-focus";
    },

    circumference() {
      return 2 * Math.PI * 130; // 更新为130
    },

    progressOffset() {
      const totalTime = this.isBreak ? this.breakTime : this.focusTime;
      const progress = this.timeLeft / totalTime;
      return this.circumference * (1 - progress);
    },

    pauseProgressStyle() {
      const progress = (this.pauseTimeLeft / this.maxPauseTime) * 100;
      return { width: `${progress}%` };
    },
  },
  watch: {
    active(newVal) {
      if (newVal && !this.isActive) {
        this.startTimer();
      }
    },
  },
  async mounted() {
    this.pickRandomEncouragement();
    // 获取当前用户ID
    await this.loadCurrentUserId();
    // 设置页面可见性监听
    this.setupVisibilityListener();

    // 检查是否有通过 URL 传过来的任务 ID
    const urlTaskId = this.$route.query.taskId;
    if (urlTaskId) {
      console.log("[Timer] 发现 URL 传参任务 ID:", urlTaskId);
      await this.loadTasks();
      const task = this.availableTasks.find(t => String(this.getTaskId(t)) === String(urlTaskId));
      if (task) {
        console.log("[Timer] 成功匹配到任务，自动选择并准备开始:", task);
        this.selectTask(task);
        // 这里不要直接 confirmTaskSelection，因为可能还在加载中
        setTimeout(() => {
          this.confirmTaskSelection();
        }, 500);
      } else {
        console.warn("[Timer] 未能匹配到任务 ID:", urlTaskId);
      }
    }
  },
  beforeUnmount() {
    this.clearTimers();
    // 移除页面可见性监听
    this.removeVisibilityListener();
  },
  methods: {
    pickRandomEncouragement() {
      const randomIndex = Math.floor(
        Math.random() * this.encouragements.length
      );
      this.currentEncouragement = this.encouragements[randomIndex];
    },

    formatTime(seconds) {
      const minutes = Math.floor(seconds / 60);
      const remainingSeconds = seconds % 60;
      return `${minutes}:${remainingSeconds.toString().padStart(2, "0")}`;
    },

    formatStatsTime(seconds) {
      const minutes = Math.floor(seconds / 60);
      if (minutes < 60) return `${minutes}m`;
      const hours = Math.floor(minutes / 60);
      return `${hours}h`;
    },

    async startTimer() {
      console.log("startTimer 被调用");
      // 先显示任务选择弹窗
      await this.showTaskSelection();
    },

    async showTaskSelection() {
      console.log("showTaskSelection 被调用");
      // 确保先设置 showTaskSelect 为 true，这样弹窗会立即显示
      this.showTaskSelect = true;
      this.selectedTask = null;
      this.loadingTasks = true;
      console.log("[Timer] 尝试显示任务选择弹窗, showTaskSelect:", this.showTaskSelect);
      
      // 强制更新视图
      await this.$nextTick();

      try {
        // 获取任务列表
        await this.loadTasks();
        console.log("任务列表加载完成，任务数量:", this.availableTasks.length);
      } catch (error) {
        console.error("加载任务失败:", error);
        // 即使加载失败，也显示弹窗（显示错误信息）
      } finally {
        this.loadingTasks = false;
      }
    },

    async loadCurrentUserId() {
      try {
        if (this.userId) {
          this.currentUserId = this.userId;
          return;
        }
        const response = await getCurrentUser();
        if (response.success && response.data) {
          this.currentUserId = response.data.id || response.data.userId || response.data.user_id;
        }
      } catch (error) {
        console.error("获取用户ID失败:", error);
      }
    },

    async loadTasks() {
      if (!this.currentUserId) {
        await this.loadCurrentUserId();
      }
      if (!this.currentUserId) {
        console.warn("无法加载任务：用户ID为空");
        this.availableTasks = [];
        return;
      }

      try {
        const response = await getTasks(this.currentUserId);
        console.log("[Timer] 获取任务成功:", response);
        // 过滤出未完成的任务
        const allTasks = response.data || response || [];
        this.availableTasks = allTasks.filter(task => {
          const status = task.status || '';
          const statusText = typeof status === 'string' ? status : '';
          return statusText !== '已完成' && statusText !== 'completed';
        });
        console.log("[Timer] 过滤后的可用任务列表:", JSON.stringify(this.availableTasks, null, 2));
        if (this.availableTasks.length === 0) {
          console.warn("[Timer] 警告：没有可用的未完成任务");
        }
      } catch (error) {
        console.error("加载任务列表失败:", error);
        this.availableTasks = [];
      }
    },

    getTaskId(task) {
      return task.task_id || task.taskId || task.id;
    },

    getTaskName(task) {
      return task.task_name || task.taskName || '未命名任务';
    },

    getTaskNote(task) {
      return task.task_note || task.taskNote || '';
    },

    selectTask(task) {
      this.selectedTask = task;
    },

    async confirmTaskSelection() {
      if (!this.selectedTask) {
        return;
      }

      // 根据任务时长设置 focusTime
      const duration = this.selectedTask.duration || 25; // 默认25分钟
      this.focusTime = duration * 60; // 转换为秒
      this.timeLeft = this.focusTime;

      // 关闭任务选择弹窗
      this.showTaskSelect = false;

      // 调用startFocus API创建专注会话记录（这会自动更新任务状态为"进行中"）
      try {
        const taskId = this.getTaskId(this.selectedTask);
        const taskName = this.getTaskName(this.selectedTask);
        
        // 先调用startFocus API，传全量数据
        const { startFocus } = await import('@/api/user');
        await startFocus({
          task_id: String(taskId),
          task_name: taskName,
          room_id: this.roomId ? String(this.roomId) : null,
          duration: duration,
          session_type: 'Task'
        });
        console.log("已调用 startFocus API，数据:", { taskId, taskName, roomId: this.roomId });
        
        // startFocus 已经更新了任务状态为"进行中"和开始时间
        // 如果需要更新其他字段（如task_note、duration），可以在这里更新
        try {
          const { updateTask } = await import('@/api/tasks');
          await updateTask(taskId, {
            task_name: taskName,
            task_note: this.getTaskNote(this.selectedTask) || null,
            duration: duration
            // 注意：不更新status，因为startFocus已经更新了
          });
          console.log("任务其他字段已更新");
        } catch (updateError) {
          console.warn("更新任务其他字段失败，但不影响专注:", updateError);
          // 即使更新失败，也不影响专注流程
        }
      } catch (startFocusError) {
        console.error("调用 startFocus API 失败:", startFocusError);
        // 如果startFocus失败，尝试手动更新任务状态，保证至少任务状态正确
        try {
          const taskId = this.getTaskId(this.selectedTask);
          const taskName = this.getTaskName(this.selectedTask);
          const { updateTask } = await import('@/api/tasks');
          await updateTask(taskId, {
            task_name: taskName,
            task_note: this.getTaskNote(this.selectedTask) || null,
            duration: duration,
            status: '进行中',
            taskStatus: '进行中'
          });
          console.warn("startFocus失败，已手动更新任务状态为'进行中'");
        } catch (fallbackError) {
          console.error("手动更新任务状态也失败:", fallbackError);
        }
        // 即使startFocus失败，也继续开始计时器，保证用户体验
      }

      // 立即发出状态变更事件，确保状态同步及时
      this.$emit("user-status-change", "focusing"); // 通知状态变为专注
      
      // 开始计时器
      this.isLoading = true;
      setTimeout(() => {
        this.isActive = true;
        this.isBreak = false;
        this.isPaused = false;
        this.isFocusing = true; // 标记正在专注中
        this.pickRandomEncouragement();
        this.startCountdown();
        this.isLoading = false;
        this.$emit("timer-started");
        this.$emit("task-selected", this.selectedTask); // 通知任务已选择
      }, 300);
    },

    cancelTaskSelection() {
      this.showTaskSelect = false;
      this.selectedTask = null;
    },

    startCountdown() {
      this.clearTimers();

      this.timer = setInterval(() => {
        if (this.timeLeft > 0) {
          this.timeLeft--;
          if (!this.isBreak && !this.isPaused) {
            this.totalFocusTime++;
          }
        } else {
          this.handleTimerEnd();
        }
      }, 1000);
    },

    // 处理计时器结束的完整逻辑
    handleTimerEnd() {
      if (this.isBreak) {
        this.startFocusSession(); // 调用重命名后的方法
      } else {
        this.startBreakSession(); // 调用重命名后的方法
      }
    },

    // 重命名为避免重复
    async startBreakSession() {
      // 专注完成，调用 stopFocus API 保存专注时长
      if (this.selectedTask && this.isFocusing) {
        try {
          const { stopFocus } = await import('@/api/user');
          await stopFocus();
          console.log("专注完成，已调用 stopFocus API 保存专注时长");
        } catch (stopError) {
          console.warn("调用 stopFocus API 失败:", stopError);
          // 即使 stopFocus 失败，也继续进入休息状态，保证用户体验
        }
      }
      
      this.completedSessions++;
      this.isBreak = true;
      this.isFocusing = false; // 休息时不在专注中
      this.timeLeft = this.breakTime;
      this.pickRandomEncouragement();
      this.$emit("focus-completed", this.completedSessions);
      this.$emit("user-status-change", "resting"); // 专注完成进入休息
    },

    // 重命名为避免重复
    startFocusSession() {
      this.isBreak = false;
      this.isFocusing = true; // 开始新的专注
      this.timeLeft = this.focusTime;
      this.pickRandomEncouragement();
      // 休息结束，自动开始新的专注
      this.startCountdown();
      this.$emit("timer-started");
      this.$emit("user-status-change", "focusing"); // 跳过休息时状态变为专注
    },

    pauseTimer() {
      this.isPaused = true;
      this.showPauseModal = true;
      this.pauseTimeLeft = this.maxPauseTime;

      if (this.timer) {
        clearInterval(this.timer);
        this.timer = null;
      }

      this.isFocusing = false; // 暂停时不监听页面切换
      this.startPauseCountdown();
      this.$emit("timer-paused");
      // 注意：暂停时状态仍然是专注，只是暂停了，所以不改变状态
    },

    startPauseCountdown() {
      this.pauseTimerRef = setInterval(() => {
        if (this.pauseTimeLeft > 0) {
          this.pauseTimeLeft--;
        } else {
          this.resumeTimer();
        }
      }, 1000);
    },

    resumeTimer() {
      this.isPaused = false;
      this.isFocusing = true; // 恢复专注，重新启用监听
      this.showPauseModal = false;

      if (this.pauseTimerRef) {
        clearInterval(this.pauseTimerRef);
        this.pauseTimerRef = null;
      }

      this.startCountdown();
      this.$emit("timer-resumed");
      this.$emit("user-status-change", "focusing"); // 恢复时状态变为专注
    },

    async stopTimer() {
      // 如果有选中的任务，显示任务提前结束确认框
      if (this.selectedTask && !this.isBreak) {
        this.showTaskEndConfirm = true;
      } else {
        // 没有任务或处于休息状态，直接终止
        if (confirm("确定要终止当前的专注吗？")) {
          // 如果正在专注，调用 stopFocus API 保存专注时长
          if (this.isFocusing && this.selectedTask) {
            try {
              const { stopFocus } = await import('@/api/user');
              await stopFocus();
              console.log("已调用 stopFocus API 保存专注时长");
            } catch (stopError) {
              console.warn("调用 stopFocus API 失败:", stopError);
              // 即使 stopFocus 失败，也继续终止，保证用户体验
            }
          }
          
          this.isFocusing = false; // 停止专注
          this.clearTimers();
          this.resetTimer();
          this.$emit("timer-stopped");
          this.$emit("user-status-change", "resting"); // 终止时状态变为休息
        }
      }
    },

    async confirmTaskEnd() {
      if (!this.selectedTask) {
        return;
      }

      try {
        const taskId = this.getTaskId(this.selectedTask);
        // 计算已专注的时间（分钟）
        const elapsedMinutes = Math.floor((this.focusTime - this.timeLeft) / 60);
        
        // 先调用 stopFocus API 更新用户状态和任务状态
        // 这会确保用户状态从"专注中"变为"在线"，避免状态不一致
        const { stopFocus } = await import('@/api/user');
        try {
          await stopFocus();
          console.log("已调用 stopFocus API 更新用户状态");
        } catch (stopError) {
          console.warn("调用 stopFocus API 失败，继续更新任务:", stopError);
          // 即使 stopFocus 失败，也继续更新任务状态
        }
        
        // 更新任务状态为"已完成"
        const { completeTaskApi } = await import('@/api/tasks');
        await completeTaskApi(taskId);
        console.log("任务已提前完成，已专注时间:", elapsedMinutes, "分钟");

        // 关闭确认框
        this.showTaskEndConfirm = false;

        // 清除计时器并重置
        this.isFocusing = false; 
        this.clearTimers();
        this.resetTimer();
        this.$emit("timer-stopped");
        this.$emit("user-status-change", "resting"); 

        // 提示并回到首页
        try {
          const { showEarnTomatoNotification } = await import('@/utils/tomatoNotification');
          showEarnTomatoNotification(1, '专注完成！任务已标记为已完成。');
        } catch (e) {
          console.warn("通知组件加载失败");
        }
        
        setTimeout(() => {
          this.$router.push('/home');
        }, 1500);

        this.$emit("task-completed", {
          task: this.selectedTask,
          elapsedMinutes: elapsedMinutes,
          isEarlyEnd: true
        });
      } catch (error) {
        console.error("更新任务状态失败:", error);
        try {
          const { showEarnTomatoNotification } = await import('@/utils/tomatoNotification');
          showEarnTomatoNotification(0, '更新任务状态失败，请稍后重试');
        } catch (e) {
          console.warn("通知组件加载失败");
        }
      }
    },

    cancelTaskEnd() {
      this.showTaskEndConfirm = false;
    },

    skipBreak() {
      this.showSkipConfirm = true;
    },

    confirmSkip() {
      this.showSkipConfirm = false;
      this.startFocusSession(); // 调用重命名后的方法
      this.$emit("break-skipped");
      this.$emit("user-status-change", "focusing"); // 跳过休息时状态变为专注
    },

    cancelSkip() {
      this.showSkipConfirm = false;
    },

    resetTimer() {
      this.isActive = false;
      this.isPaused = false;
      this.isBreak = false;
      this.timeLeft = this.focusTime;
      this.pauseTimeLeft = this.maxPauseTime;
      this.showPauseModal = false;
      this.showSkipConfirm = false;
      this.showTaskEndConfirm = false;
      // 注意：不重置 selectedTask，因为可能需要显示任务信息
      // this.selectedTask = null;
    },

    clearTimers() {
      if (this.timer) clearInterval(this.timer);
      if (this.pauseTimerRef) clearInterval(this.pauseTimerRef);
      this.timer = null;
      this.pauseTimerRef = null;
    },

    // 设置页面可见性监听
    setupVisibilityListener() {
      // 页面可见性变化监听（切换标签页、最小化窗口等）
      this.visibilityHandler = () => {
        if (this.isFocusing && !this.isBreak && !this.isPaused) {
          if (document.hidden) {
            // 页面隐藏时记录时间
            this.pageHiddenTime = Date.now();
            console.log("检测到页面隐藏，可能切换到了其他标签页");
          } else {
            // 页面重新可见时检查
            if (this.pageHiddenTime) {
              const hiddenDuration = Date.now() - this.pageHiddenTime;
              // 如果隐藏时间超过3秒，提醒用户
              if (hiddenDuration > 3000) {
                this.showDistractionWarning();
              }
              this.pageHiddenTime = null;
            }
          }
        }
      };
      
      // 监听来自扩展的消息
      window.addEventListener('message', (event) => {
        // 只接受来自扩展的消息
        if (event.data && event.data.type === 'DISTRACTING_SITE_DETECTED') {
          this.showDistractionWarning();
        }
      });

      // 窗口失去焦点监听（切换到其他应用）
      this.blurHandler = () => {
        if (this.isFocusing && !this.isBreak && !this.isPaused) {
          this.pageHiddenTime = Date.now();
          console.log("检测到窗口失去焦点，可能切换到了其他应用");
        }
      };

      // 窗口获得焦点监听
      this.focusHandler = () => {
        if (this.isFocusing && !this.isBreak && !this.isPaused) {
          if (this.pageHiddenTime) {
            const hiddenDuration = Date.now() - this.pageHiddenTime;
            // 如果隐藏时间超过3秒，提醒用户
            if (hiddenDuration > 3000) {
              this.showDistractionWarning();
            }
            this.pageHiddenTime = null;
          }
        }
      };

      // 添加事件监听器
      document.addEventListener("visibilitychange", this.visibilityHandler);
      window.addEventListener("blur", this.blurHandler);
      window.addEventListener("focus", this.focusHandler);
    },

    // 移除页面可见性监听
    removeVisibilityListener() {
      if (this.visibilityHandler) {
        document.removeEventListener("visibilitychange", this.visibilityHandler);
      }
      if (this.blurHandler) {
        window.removeEventListener("blur", this.blurHandler);
      }
      if (this.focusHandler) {
        window.removeEventListener("focus", this.focusHandler);
      }
    },

    // 显示分心警告
    showDistractionWarning() {
      // 显示漂亮的模态框而不是alert
      this.showDistractionModal = true;
      
      // 尝试让页面获得焦点（如果可能）
      if (window.focus) {
        window.focus();
      }
    },

    // 关闭分心警告模态框
    closeDistractionModal() {
      this.showDistractionModal = false;
    },

  },
};
</script>

<style scoped>
/* 基础容器 - 修改：移除最大宽度，填满容器 */
.pomodoro-timer {
  background: #ffffff;
  border-radius: 20px;
  padding: 32px; /* 增大内边距 */
  width: 100%; /* 填满父容器 */
  min-height: 550px; /* 增加高度 */
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.06), 0 1px 4px rgba(0, 0, 0, 0.04);
  border: 1px solid #f0f0f0;
  position: relative;
  overflow: hidden;
  display: flex; /* flex布局填满高度 */
  flex-direction: column;
}

/* 微妙的背景纹理 */
.pomodoro-timer::before {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-image: radial-gradient(
      circle at 20% 30%,
      rgba(255, 169, 77, 0.03) 0%,
      transparent 50%
    ),
    radial-gradient(
      circle at 80% 70%,
      rgba(255, 135, 135, 0.03) 0%,
      transparent 50%
    );
  pointer-events: none;
  z-index: 0;
}

/* 所有内容都在上面 */
.timer-start-screen,
.timer-active-screen {
  position: relative;
  z-index: 1;
  flex: 1; /* 填满容器高度 */
  display: flex;
  flex-direction: column;
}

/* 初始界面 */
.timer-start-screen {
  text-align: center;
  justify-content: center; /* 垂直居中 */
}

.pomodoro-logo {
  width: 120px;
  height: 120px;
  margin-bottom: 24px;
  object-fit: contain;
  animation: float 3s ease-in-out infinite;
  filter: drop-shadow(0 6px 12px rgba(255, 169, 77, 0.25));
}

@keyframes float {
  0%,
  100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-8px);
  }
}

.timer-title {
  font-size: 28px;
  font-weight: 700;
  background: linear-gradient(135deg, #ffa94d 0%, #ff8787 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  margin: 0 0 8px 0;
  letter-spacing: -0.3px;
}

.timer-subtitle {
  color: #8a8a8a;
  font-size: 15px;
  margin: 0 0 28px 0;
  font-weight: 400;
}

.encouragement-card {
  background: #f8f9fa;
  border-radius: 16px;
  padding: 20px;
  margin: 20px 0 28px;
  position: relative;
  border: 1px solid #e9ecef;
}

.quote-icon {
  position: absolute;
  top: 12px;
  left: 16px;
  font-size: 20px;
  color: #ffa94d;
  opacity: 0.4;
}

.encouragement-text {
  color: #495057;
  font-size: 16px;
  line-height: 1.6;
  margin: 0;
  font-style: italic;
  font-weight: 500;
}

.start-button {
  background: linear-gradient(135deg, #ffa94d 0%, hsl(42, 85%, 61%) 100%);
  color: white;
  border: none;
  padding: 16px 40px;
  border-radius: 14px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  width: 100%;
  position: relative;
  overflow: hidden;
}

.start-button:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(255, 169, 77, 0.25);
}

.start-button:active:not(:disabled) {
  transform: translateY(0);
}

.start-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.loading-spinner-small {
  display: inline-block;
  width: 18px;
  height: 18px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top: 2px solid white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}

/* 激活界面 */
.timer-active-screen {
  position: relative;
}

.mode-indicator {
  display: inline-flex;
  align-items: center;
  padding: 6px 14px;
  border-radius: 18px;
  margin-bottom: 20px;
  font-size: 13px;
  font-weight: 500;
  background: #f8f9fa;
  border: 1px solid #e9ecef;
}

.mode-indicator.mode-focus {
  color: #ffa94d;
}

.mode-indicator.mode-break {
  color: #63e6be;
}

.mode-indicator.mode-pause {
  color: #ffc078;
}

.mode-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  margin-right: 6px;
}

.mode-focus .mode-dot {
  background: #ffa94d;
}
.mode-break .mode-dot {
  background: #63e6be;
}
.mode-pause .mode-dot {
  background: #ffc078;
}

/* 圆形计时器 - 增大尺寸 */
.circular-timer {
  position: relative;
  width: 280px; /* 从220px增大到280px */
  height: 280px;
  margin: 0 auto 28px;
}

.timer-halo {
  position: absolute;
  top: -15px;
  left: -15px;
  right: -15px;
  bottom: -15px;
  border-radius: 50%;
  opacity: 0.15;
  z-index: 0;
}

.timer-halo.mode-focus {
  background: #ffa94d;
}
.timer-halo.mode-break {
  background: #63e6be;
}

.timer-display {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
  z-index: 2;
}

.timer-digits {
  font-size: 52px; /* 从44px增大到52px */
  font-weight: 700;
  font-family: "Inter", system-ui, -apple-system, sans-serif;
  color: hwb(32 47% 4%);
  letter-spacing: -1px;
  margin-bottom: 4px;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.timer-label {
  font-size: 14px;
  color: #868e96;
  font-weight: 500;
  letter-spacing: 0.5px;
}

.progress-ring {
  width: 280px; /* 与容器一致 */
  height: 280px;
  transform: rotate(-90deg);
}

.progress-ring-background {
  fill: none;
  stroke: #f1f3f5;
  stroke-width: 10;
}

.progress-ring-fill {
  fill: none;
  stroke-width: 10;
  stroke-linecap: round;
  transition: stroke-dashoffset 1s linear;
  opacity: 0.6;
}

/* 当前任务信息 */
.current-task-info {
  margin: 0 0 20px 0;
  padding: 16px;
  background: linear-gradient(135deg, #fff9f2 0%, #ffe8cc 100%);
  border-radius: 14px;
  border: 2px solid #ffa94d;
  flex-shrink: 0;
  box-shadow: 0 2px 8px rgba(255, 169, 77, 0.15);
}

.task-info-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.task-icon {
  font-size: 20px;
}

.task-name {
  font-size: 16px;
  font-weight: 600;
  color: #212529;
  flex: 1;
}

.task-note-text {
  font-size: 13px;
  color: #868e96;
  line-height: 1.4;
  margin-top: 6px;
  padding-top: 8px;
  border-top: 1px solid rgba(255, 169, 77, 0.2);
}

/* 鼓励语 */
.current-encouragement {
  margin: 20px 0 28px;
  padding: 14px;
  background: #f8f9fa;
  border-radius: 14px;
  border: 1px solid #e9ecef;
  flex-shrink: 0; /* 防止被压缩 */
}

.current-encouragement p {
  margin: 0;
  color: #495057;
  font-size: 15px;
  font-weight: 500;
  text-align: center;
  line-height: 1.5;
}

/* 控制按钮 */
.control-buttons {
  display: flex;
  justify-content: center;
  gap: 10px;
  margin-bottom: 28px;
  flex-wrap: wrap;
  flex-shrink: 0;
}

.control-button {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 12px 20px;
  border: 1px solid #e9ecef;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  min-width: 120px;
  background: white;
}

.control-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.pause-button {
  background: white;
  color: #ffa94d;
  border-color: #ffe8cc;
}

.pause-button:hover {
  background: #fff9f2;
  border-color: #ffa94d;
}

.resume-button {
  background: white;
  color: #63e6be;
  border-color: #d3f9d8;
}

.resume-button:hover {
  background: #f3fef9;
  border-color: #63e6be;
}

.stop-button {
  background: white;
  color: #ff8787;
  border-color: #ffd8d8;
}

.stop-button:hover {
  background: #fff5f5;
  border-color: #ff8787;
}

.skip-button {
  background: white;
  color: #9775fa;
  border-color: #e5dbff;
}

.skip-button:hover {
  background: #f8f9ff;
  border-color: #9775fa;
}

/* 统计数据 */
.stats-panel {
  display: flex;
  justify-content: center;
  gap: 40px;
  padding-top: 20px;
  border-top: 1px solid #f1f3f5;
  margin-top: auto; /* 推到容器底部 */
  flex-shrink: 0;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.stat-content {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat-value {
  font-size: 28px; /* 从24px增大到28px */
  font-weight: 700;
  color: #212529;
  line-height: 1;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 15px; /* 从14px增大到15px */
  color: #868e96;
}

/* 模态框 */
.modal-container {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  backdrop-filter: blur(3px);
  animation: fadeIn 0.2s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.modal-card {
  background: white;
  border-radius: 20px;
  padding: 28px;
  max-width: 380px;
  width: 90%;
  box-shadow: 0 16px 40px rgba(0, 0, 0, 0.12);
  animation: slideUp 0.3s ease;
  border: 1px solid #f0f0f0;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(15px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.modal-header {
  text-align: center;
  margin-bottom: 20px;
}

.modal-icon {
  font-size: 40px;
  margin-bottom: 12px;
}

.modal-title {
  font-size: 20px;
  font-weight: 700;
  color: #212529;
  margin: 0;
}

.modal-body {
  margin-bottom: 20px;
}

.modal-message {
  color: #495057;
  font-size: 15px;
  line-height: 1.6;
  margin: 0 0 12px 0;
  text-align: center;
}

.modal-tip {
  color: #868e96;
  font-size: 13px;
  text-align: center;
  margin: 0;
  font-style: italic;
}

.pause-timer {
  background: #f8f9fa;
  border-radius: 10px;
  padding: 14px;
  margin-top: 12px;
  position: relative;
  overflow: hidden;
  border: 1px solid #e9ecef;
}

.pause-progress {
  position: absolute;
  top: 0;
  left: 0;
  bottom: 0;
  background: linear-gradient(90deg, #ffc078 0%, #ffd8a8 100%);
  transition: width 1s linear;
  opacity: 0.7;
}

.pause-time {
  position: relative;
  z-index: 1;
  text-align: center;
  font-size: 18px;
  font-weight: 600;
  color: #212529;
}

.modal-action-button,
.modal-button {
  width: 100%;
  padding: 14px;
  border: none;
  border-radius: 10px;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
}

.modal-action-button {
  background: linear-gradient(135deg, #ffa94d 0%, #ff8787 100%);
  color: white;
}

.modal-action-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(255, 169, 77, 0.25);
}

.modal-actions {
  display: flex;
  gap: 10px;
  margin-top: 12px;
}

.modal-button {
  flex: 1;
}

.confirm-button {
  background: linear-gradient(135deg, #63e6be 0%, #96f2d7 100%);
  color: white;
  border: none;
}

.cancel-button {
  background: white;
  color: #495057;
  border: 1px solid #dee2e6;
}

.confirm-button:hover,
.cancel-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

/* 任务选择弹窗样式 */
.task-select-modal {
  max-width: 500px;
}

.loading-tasks {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  gap: 16px;
}

.loading-tasks p {
  color: #868e96;
  font-size: 14px;
  margin: 0;
}

.no-tasks {
  text-align: center;
  padding: 40px 20px;
}

.task-list {
  max-height: 400px;
  overflow-y: auto;
  padding: 8px 0;
}

.task-item {
  padding: 16px;
  margin-bottom: 8px;
  background: #f8f9fa;
  border: 2px solid #e9ecef;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.task-item:hover {
  background: #f1f3f5;
  border-color: #ffa94d;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(255, 169, 77, 0.15);
}

.task-item.task-selected {
  background: #fff9f2;
  border-color: #ffa94d;
  box-shadow: 0 4px 12px rgba(255, 169, 77, 0.25);
}

.task-item-content {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.task-item-title {
  font-size: 16px;
  font-weight: 600;
  color: #212529;
}

.task-item-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.task-duration-badge {
  background: linear-gradient(135deg, #ffa94d 0%, #ff8787 100%);
  color: white;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 13px;
  font-weight: 600;
}

.task-note-preview {
  color: #868e96;
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 200px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .pomodoro-timer {
    padding: 20px;
    min-height: 500px;
  }

  .circular-timer {
    width: 240px;
    height: 240px;
  }

  .progress-ring {
    width: 240px;
    height: 240px;
  }

  .progress-ring-background,
  .progress-ring-fill {
    cx: 120;
    cy: 120;
    r: 110;
  }

  .timer-digits {
    font-size: 44px;
  }

  .control-button {
    min-width: 110px;
    padding: 10px 16px;
  }

  .modal-card {
    padding: 24px;
  }
}

/* 分心警告模态框样式 */
.distraction-modal-overlay {
  z-index: 10000;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  animation: fadeIn 0.3s ease;
}

.distraction-modal {
  max-width: 420px;
  width: 90%;
  background: linear-gradient(135deg, #fff9e6 0%, #fff3cd 100%);
  border: 2px solid #ffc107;
  box-shadow: 0 8px 32px rgba(255, 193, 7, 0.3);
  animation: slideUp 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.distraction-modal-header {
  text-align: center;
  padding: 24px 24px 16px;
}

.distraction-icon {
  display: flex;
  justify-content: center;
  align-items: center;
  margin-bottom: 16px;
  animation: pulse 2s ease-in-out infinite;
}

.distraction-icon svg {
  filter: drop-shadow(0 4px 8px rgba(255, 193, 7, 0.3));
}

.distraction-title {
  font-size: 22px;
  font-weight: 700;
  color: #856404;
  margin: 0;
  text-shadow: 0 1px 2px rgba(255, 255, 255, 0.5);
}

.distraction-modal-body {
  padding: 0 24px 24px;
  text-align: center;
}

.distraction-message {
  font-size: 16px;
  font-weight: 600;
  color: #856404;
  margin: 0 0 12px;
  line-height: 1.5;
}

.distraction-tip {
  font-size: 14px;
  color: #856404;
  margin: 0 0 16px;
  line-height: 1.6;
  opacity: 0.9;
}

.distraction-note {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 12px 16px;
  background: rgba(255, 255, 255, 0.6);
  border-radius: 8px;
  font-size: 13px;
  color: #856404;
  margin-top: 16px;
}

.note-icon {
  font-size: 18px;
}

.distraction-button {
  width: 100%;
  padding: 14px 24px;
  background: linear-gradient(135deg, #ffc107 0%, #ffb300 100%);
  color: #856404;
  border: none;
  border-radius: 10px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 4px 12px rgba(255, 193, 7, 0.25);
  margin-top: 8px;
}

.distraction-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(255, 193, 7, 0.35);
  background: linear-gradient(135deg, #ffb300 0%, #ffa000 100%);
}

.distraction-button:active {
  transform: translateY(0);
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(30px) scale(0.95);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

@keyframes pulse {
  0%, 100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.05);
  }
}
</style>

<style scoped>
/* 基础容器 - 修改：移除最大宽度，填满容器 */
.pomodoro-timer {
  background: #ffffff;
  border-radius: 20px;
  padding: 32px; /* 增大内边距 */
  width: 100%; /* 填满父容器 */
  min-height: 550px; /* 增加高度 */
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.06), 0 1px 4px rgba(0, 0, 0, 0.04);
  border: 1px solid #f0f0f0;
  position: relative;
  overflow: hidden;
  display: flex; /* flex布局填满高度 */
  flex-direction: column;
}

/* 微妙的背景纹理 */
.pomodoro-timer::before {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-image: radial-gradient(
      circle at 20% 30%,
      rgba(255, 169, 77, 0.03) 0%,
      transparent 50%
    ),
    radial-gradient(
      circle at 80% 70%,
      rgba(255, 135, 135, 0.03) 0%,
      transparent 50%
    );
  pointer-events: none;
  z-index: 0;
}

/* 所有内容都在上面 */
.timer-start-screen,
.timer-active-screen {
  position: relative;
  z-index: 1;
  flex: 1; /* 填满容器高度 */
  display: flex;
  flex-direction: column;
}

/* 初始界面 */
.timer-start-screen {
  text-align: center;
  justify-content: center; /* 垂直居中 */
}

.pomodoro-logo {
  width: 120px;
  height: 120px;
  margin-bottom: 24px;
  object-fit: contain;
  animation: float 3s ease-in-out infinite;
  filter: drop-shadow(0 6px 12px rgba(255, 169, 77, 0.25));
  align-self: center; /* 确保图片自身居中 */
}

@keyframes float {
  0%,
  100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-8px);
  }
}

.timer-title {
  font-size: 28px;
  font-weight: 700;
  background: linear-gradient(135deg, #ffa94d 0%, #ff8787 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  margin: 0 0 8px 0;
  letter-spacing: -0.3px;
}

.timer-subtitle {
  color: #8a8a8a;
  font-size: 15px;
  margin: 0 0 28px 0;
  font-weight: 400;
}

.encouragement-card {
  background: #f8f9fa;
  border-radius: 16px;
  padding: 20px;
  margin: 20px 0 28px;
  position: relative;
  border: 1px solid #e9ecef;
}

.quote-icon {
  position: absolute;
  top: 12px;
  left: 16px;
  font-size: 20px;
  color: #ffa94d;
  opacity: 0.4;
}

.encouragement-text {
  color: #495057;
  font-size: 16px;
  line-height: 1.6;
  margin: 0;
  font-style: italic;
  font-weight: 500;
}

.start-button {
  background: linear-gradient(135deg, #ffa94d 0%, hsl(42, 85%, 61%) 100%);
  color: white;
  border: none;
  padding: 16px 40px;
  border-radius: 14px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  width: 100%;
  position: relative;
  overflow: hidden;
}

.start-button:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(255, 169, 77, 0.25);
}

.start-button:active:not(:disabled) {
  transform: translateY(0);
}

.start-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.loading-spinner-small {
  display: inline-block;
  width: 18px;
  height: 18px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top: 2px solid white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}

/* 激活界面 */
.timer-active-screen {
  position: relative;
}

.mode-indicator {
  display: inline-flex;
  align-items: center;
  padding: 6px 14px;
  border-radius: 18px;
  margin-bottom: 20px;
  font-size: 13px;
  font-weight: 500;
  background: #f8f9fa;
  border: 1px solid #e9ecef;
}

.mode-indicator.mode-focus {
  color: #ffa94d;
}

.mode-indicator.mode-break {
  color: #63e6be;
}

.mode-indicator.mode-pause {
  color: #ffc078;
}

.mode-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  margin-right: 6px;
}

.mode-focus .mode-dot {
  background: #ffa94d;
}
.mode-break .mode-dot {
  background: #63e6be;
}
.mode-pause .mode-dot {
  background: #ffc078;
}

/* 圆形计时器 - 增大尺寸 */
.circular-timer {
  position: relative;
  width: 280px; /* 从220px增大到280px */
  height: 280px;
  margin: 0 auto 28px;
}

.timer-halo {
  position: absolute;
  top: -15px;
  left: -15px;
  right: -15px;
  bottom: -15px;
  border-radius: 50%;
  opacity: 0.15;
  z-index: 0;
}

.timer-halo.mode-focus {
  background: #ffa94d;
}
.timer-halo.mode-break {
  background: #63e6be;
}

.timer-display {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
  z-index: 2;
}

.timer-digits {
  font-size: 52px; /* 从44px增大到52px */
  font-weight: 700;
  font-family: "Inter", system-ui, -apple-system, sans-serif;
  color: hwb(32 47% 4%);
  letter-spacing: -1px;
  margin-bottom: 4px;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.timer-label {
  font-size: 14px;
  color: #868e96;
  font-weight: 500;
  letter-spacing: 0.5px;
}

.progress-ring {
  width: 280px; /* 与容器一致 */
  height: 280px;
  transform: rotate(-90deg);
}

.progress-ring-background {
  fill: none;
  stroke: #f1f3f5;
  stroke-width: 10;
}

.progress-ring-fill {
  fill: none;
  stroke-width: 10;
  stroke-linecap: round;
  transition: stroke-dashoffset 1s linear;
  opacity: 0.6;
}

/* 当前任务信息 */
.current-task-info {
  margin: 0 0 20px 0;
  padding: 16px;
  background: linear-gradient(135deg, #fff9f2 0%, #ffe8cc 100%);
  border-radius: 14px;
  border: 2px solid #ffa94d;
  flex-shrink: 0;
  box-shadow: 0 2px 8px rgba(255, 169, 77, 0.15);
}

.task-info-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.task-icon {
  font-size: 20px;
}

.task-name {
  font-size: 16px;
  font-weight: 600;
  color: #212529;
  flex: 1;
}

.task-note-text {
  font-size: 13px;
  color: #868e96;
  line-height: 1.4;
  margin-top: 6px;
  padding-top: 8px;
  border-top: 1px solid rgba(255, 169, 77, 0.2);
}

/* 鼓励语 */
.current-encouragement {
  margin: 20px 0 28px;
  padding: 14px;
  background: #f8f9fa;
  border-radius: 14px;
  border: 1px solid #e9ecef;
  flex-shrink: 0; /* 防止被压缩 */
}

.current-encouragement p {
  margin: 0;
  color: #495057;
  font-size: 15px;
  font-weight: 500;
  text-align: center;
  line-height: 1.5;
}

/* 控制按钮 */
.control-buttons {
  display: flex;
  justify-content: center;
  gap: 10px;
  margin-bottom: 28px;
  flex-wrap: wrap;
  flex-shrink: 0;
}

.control-button {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 12px 20px;
  border: 1px solid #e9ecef;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  min-width: 120px;
  background: white;
}

.control-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.pause-button {
  background: white;
  color: #ffa94d;
  border-color: #ffe8cc;
}

.pause-button:hover {
  background: #fff9f2;
  border-color: #ffa94d;
}

.resume-button {
  background: white;
  color: #63e6be;
  border-color: #d3f9d8;
}

.resume-button:hover {
  background: #f3fef9;
  border-color: #63e6be;
}

.stop-button {
  background: white;
  color: #ff8787;
  border-color: #ffd8d8;
}

.stop-button:hover {
  background: #fff5f5;
  border-color: #ff8787;
}

.skip-button {
  background: white;
  color: #9775fa;
  border-color: #e5dbff;
}

.skip-button:hover {
  background: #f8f9ff;
  border-color: #9775fa;
}

/* 统计数据 */
.stats-panel {
  display: flex;
  justify-content: center;
  gap: 40px;
  padding-top: 20px;
  border-top: 1px solid #f1f3f5;
  margin-top: auto; /* 推到容器底部 */
  flex-shrink: 0;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.stat-content {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat-value {
  font-size: 28px; /* 从24px增大到28px */
  font-weight: 700;
  color: #212529;
  line-height: 1;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 15px; /* 从14px增大到15px */
  color: #868e96;
}

/* 模态框 */
.modal-container {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(3px);
  animation: fadeIn 0.2s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.modal-card {
  background: white;
  border-radius: 20px;
  padding: 28px;
  max-width: 380px;
  width: 90%;
  box-shadow: 0 16px 40px rgba(0, 0, 0, 0.12);
  animation: slideUp 0.3s ease;
  border: 1px solid #f0f0f0;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(15px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.modal-header {
  text-align: center;
  margin-bottom: 20px;
}

.modal-icon {
  font-size: 40px;
  margin-bottom: 12px;
}

.modal-title {
  font-size: 20px;
  font-weight: 700;
  color: #212529;
  margin: 0;
}

.modal-body {
  margin-bottom: 20px;
}

.modal-message {
  color: #495057;
  font-size: 15px;
  line-height: 1.6;
  margin: 0 0 12px 0;
  text-align: center;
}

.modal-tip {
  color: #868e96;
  font-size: 13px;
  text-align: center;
  margin: 0;
  font-style: italic;
}

.pause-timer {
  background: #f8f9fa;
  border-radius: 10px;
  padding: 14px;
  margin-top: 12px;
  position: relative;
  overflow: hidden;
  border: 1px solid #e9ecef;
}

.pause-progress {
  position: absolute;
  top: 0;
  left: 0;
  bottom: 0;
  background: linear-gradient(90deg, #ffc078 0%, #ffd8a8 100%);
  transition: width 1s linear;
  opacity: 0.7;
}

.pause-time {
  position: relative;
  z-index: 1;
  text-align: center;
  font-size: 18px;
  font-weight: 600;
  color: #212529;
}

.modal-action-button,
.modal-button {
  width: 100%;
  padding: 14px;
  border: none;
  border-radius: 10px;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
}

.modal-action-button {
  background: linear-gradient(135deg, #ffa94d 0%, #ff8787 100%);
  color: white;
}

.modal-action-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(255, 169, 77, 0.25);
}

.modal-actions {
  display: flex;
  gap: 10px;
  margin-top: 12px;
}

.modal-button {
  flex: 1;
}

.confirm-button {
  background: linear-gradient(135deg, #63e6be 0%, #96f2d7 100%);
  color: white;
  border: none;
}

.cancel-button {
  background: white;
  color: #495057;
  border: 1px solid #dee2e6;
}

.confirm-button:hover,
.cancel-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .pomodoro-timer {
    padding: 20px;
    min-height: 500px;
  }

  .circular-timer {
    width: 240px;
    height: 240px;
  }

  .progress-ring {
    width: 240px;
    height: 240px;
  }

  .progress-ring-background,
  .progress-ring-fill {
    cx: 120;
    cy: 120;
    r: 110;
  }

  .timer-digits {
    font-size: 44px;
  }

  .control-button {
    min-width: 110px;
    padding: 10px 16px;
  }

  .modal-card {
    padding: 24px;
  }
}

/* 分心警告模态框样式 */
.distraction-modal-overlay {
  z-index: 10000;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  animation: fadeIn 0.3s ease;
}

.distraction-modal {
  max-width: 420px;
  width: 90%;
  background: linear-gradient(135deg, #fff9e6 0%, #fff3cd 100%);
  border: 2px solid #ffc107;
  box-shadow: 0 8px 32px rgba(255, 193, 7, 0.3);
  animation: slideUp 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.distraction-modal-header {
  text-align: center;
  padding: 24px 24px 16px;
}

.distraction-icon {
  display: flex;
  justify-content: center;
  align-items: center;
  margin-bottom: 16px;
  animation: pulse 2s ease-in-out infinite;
}

.distraction-icon svg {
  filter: drop-shadow(0 4px 8px rgba(255, 193, 7, 0.3));
}

.distraction-title {
  font-size: 22px;
  font-weight: 700;
  color: #856404;
  margin: 0;
  text-shadow: 0 1px 2px rgba(255, 255, 255, 0.5);
}

.distraction-modal-body {
  padding: 0 24px 24px;
  text-align: center;
}

.distraction-message {
  font-size: 16px;
  font-weight: 600;
  color: #856404;
  margin: 0 0 12px;
  line-height: 1.5;
}

.distraction-tip {
  font-size: 14px;
  color: #856404;
  margin: 0 0 16px;
  line-height: 1.6;
  opacity: 0.9;
}

.distraction-note {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 12px 16px;
  background: rgba(255, 255, 255, 0.6);
  border-radius: 8px;
  font-size: 13px;
  color: #856404;
  margin-top: 16px;
}

.note-icon {
  font-size: 18px;
}

.distraction-button {
  width: 100%;
  padding: 14px 24px;
  background: linear-gradient(135deg, #ffc107 0%, #ffb300 100%);
  color: #856404;
  border: none;
  border-radius: 10px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 4px 12px rgba(255, 193, 7, 0.25);
  margin-top: 8px;
}

.distraction-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(255, 193, 7, 0.35);
  background: linear-gradient(135deg, #ffb300 0%, #ffa000 100%);
}

.distraction-button:active {
  transform: translateY(0);
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(30px) scale(0.95);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

@keyframes pulse {
  0%, 100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.05);
  }
}
</style>
