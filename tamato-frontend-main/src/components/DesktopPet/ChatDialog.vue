<template>
  <div v-if="visible" class="chat-dialog-overlay" @click.self="closeDialog">
    <div class="chat-dialog" :style="dialogStyle">
      <!-- 侧边栏：历史对话列表 -->
      <div v-show="!sidebarCollapsed" class="chat-sidebar" :style="{ width: sidebarWidth + 'px' }">
        <div class="sidebar-header">
          <div class="sidebar-title-group">
            <span class="sidebar-title">历史对话</span>
          </div>
          <button class="new-chat-btn" @click="startNewSession" title="新建对话">
            <span class="plus">+</span>
          </button>
        </div>
        <div class="session-list">
          <div 
            v-for="session in sessions" 
            :key="session.session_id"
            :class="['session-item', { active: currentSessionID === session.session_id }]"
            @click="selectSession(session)"
          >
            <span class="session-icon">💬</span>
            <div class="session-info">
              <div class="session-title-container">
                <input 
                  v-if="editingSessionID === session.session_id"
                  v-model="editingTitle"
                  class="edit-title-input"
                  @blur="saveTitle(session)"
                  @keyup.enter="saveTitle(session)"
                  ref="editInput"
                />
                <span v-else class="session-title">{{ session.title }}</span>
                <button 
                  v-if="editingSessionID !== session.session_id"
                  class="edit-btn" 
                  @click.stop="startEditTitle(session)"
                >
                  ✏️
                </button>
              </div>
              <span class="session-time">{{ formatDate(session.updated_at) }}</span>
            </div>
          </div>
          <div v-if="sessions.length === 0" class="no-sessions">
            暂无历史对话
          </div>
        </div>
      </div>

      <!-- 侧边栏收缩后的占位/展开按钮 -->
      <div v-if="sidebarCollapsed" class="collapsed-sidebar-tip" @click="sidebarCollapsed = false" title="展开对话历史">
        <div class="expand-icon-wrapper">
          <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M13 17l5-5-5-5M6 17l5-5-5-5" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
      </div>

      <!-- 拖拽条 1 -->
      <div 
        v-if="!sidebarCollapsed" 
        class="resizer resizer-h" 
        @mousedown="startResizing('sidebar')"
      ></div>

      <!-- 主聊天区域 -->
      <div class="chat-main">
        <!-- 对话框头部 -->
        <div class="chat-header">
          <div class="chat-header-left">
            <button class="collapse-toggle" @click="sidebarCollapsed = !sidebarCollapsed" :title="sidebarCollapsed ? '展开侧边栏' : '收起侧边栏'">
              <svg v-if="!sidebarCollapsed" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5">
                <path d="M15 18l-6-6 6-6" stroke-linecap round stroke-linejoin="round"/>
              </svg>
              <svg v-else viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2.5">
                <path d="M9 18l6-6-6-6" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </button>
            <div class="pet-avatar">🍅</div>
            <div class="chat-title">
              <div class="chat-name">{{ currentSessionTitle || '番茄小助手' }}</div>
              <div class="chat-status">在线 · AI学习教练</div>
            </div>
          </div>
          <div class="chat-header-right">
            <button class="header-action-btn" @click="openProfileModal" title="AI 记忆与画像管理">
              <span class="brain-icon">🧠</span>
              <span v-if="hasSuggestions" class="suggestion-badge"></span>
            </button>
            <button class="close-btn" @click="closeDialog">×</button>
          </div>
        </div>
        
        <!-- 消息列表 -->
        <div class="chat-messages" ref="messagesContainer">
          <div 
            v-for="(message, index) in messages" 
            :key="index"
            :class="['message', message.role]"
          >
            <div class="message-avatar" v-if="message.role === 'assistant'">🍅</div>
            <div class="message-content">
              <div v-if="message.reasoning" class="message-reasoning">
                <div class="reasoning-title">🔍 思考与检索中...</div>
                <div class="reasoning-text">{{ message.reasoning }}</div>
              </div>
              <div v-if="isPendingAssistantMessage(message, index)" class="message-text typing">
                <span></span>
                <span></span>
                <span></span>
              </div>
              <div 
                v-else
                class="message-text" 
                v-html="renderMarkdown(message.content)"
                @click="handleMessageClick"
              ></div>
              <div class="message-time" v-if="message.timestamp">
                {{ formatTime(message.timestamp) }}
                <span v-if="message.usage && message.usage.total_tokens" class="token-usage">
                  · Token: {{ message.usage.total_tokens }}
                </span>
              </div>
            </div>
            <div class="message-avatar" v-if="message.role === 'user'">👤</div>
          </div>
        </div>
        
        <!-- 输入框 -->
        <div class="chat-input-area">
          <div class="chat-input-wrapper">
            <textarea
              v-model="inputMessage"
              class="chat-input"
              placeholder="想聊点什么？(Enter发送, Shift+Enter换行)"
              @keydown.enter.exact.prevent="sendMessage"
              @keydown.enter.shift.exact="inputMessage += '\n'"
              rows="1"
              ref="inputRef"
            ></textarea>
            
            <div class="input-actions">
              <!-- 模式选择 (精简版) -->
              <div class="mode-selector-container">
                <button class="action-btn mode-btn" @click="showModeMenu = !showModeMenu">
                  <span class="action-icon">{{ chatMode === 'fast' ? '⚡' : '💡' }}</span>
                  <span class="action-text">{{ chatMode === 'fast' ? '快速' : '思考' }}</span>
                </button>
                
                <transition name="fade-up">
                  <div v-if="showModeMenu" class="mode-dropdown-compact">
                    <div class="mode-opt" :class="{ active: chatMode === 'fast' }" @click="setChatMode('fast')">
                      <span class="opt-icon">⚡</span>
                      <span>快速</span>
                    </div>
                    <div class="mode-opt" :class="{ active: chatMode === 'thinking' }" @click="setChatMode('thinking')">
                      <span class="opt-icon">💡</span>
                      <span>思考</span>
                    </div>
                  </div>
                </transition>
              </div>

              <!-- 知识库切换 (精简版) -->
              <button 
                class="action-btn rag-btn" 
                :class="{ active: useKnowledge }"
                @click="useKnowledge = !useKnowledge"
              >
                <span class="action-icon">📚</span>
                <span class="action-text">知识库</span>
              </button>

              <div class="actions-divider"></div>

              <button 
                class="send-btn" 
                @click="sendMessage"
                :disabled="!inputMessage.trim() || isLoading"
              >
                <span v-if="!isLoading">发送</span>
                <span v-else>...</span>
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 拖拽条 2: 主区域与预览区域之间 -->
      <div 
        v-if="previewVisible" 
        class="resizer resizer-h" 
        @mousedown="startResizing('preview', $event)"
      ></div>

      <!-- 右侧文件预览面板 -->
      <transition name="slide">
        <div v-if="previewVisible" class="preview-panel" :style="{ width: previewWidth + 'px' }">
          <div class="preview-header">
            <div class="preview-title-container">
              <span class="preview-icon">📄</span>
              <span class="preview-title">{{ previewFileName }}</span>
            </div>
            <button class="preview-close" @click="previewVisible = false">×</button>
          </div>
          <div class="preview-content-area">
            <div v-if="isPreviewLoading" class="preview-loading">
              <div class="spinner"></div>
              <span>正在加载文件内容...</span>
            </div>
            <div 
              v-else 
              class="preview-body" 
              v-html="renderMarkdown(previewContent)"
            ></div>
          </div>
        </div>
      </transition>
      <!-- 用户画像管理弹窗 -->
      <transition name="fade">
        <div v-if="profileModalVisible" class="profile-modal-overlay" @click.self="profileModalVisible = false">
          <div class="profile-modal">
            <div class="profile-modal-header">
              <h3>🧠 AI 记忆与画像管理</h3>
              <button class="modal-close" @click="profileModalVisible = false">×</button>
            </div>
            
            <div class="profile-modal-body">
              <div class="profile-section">
                <label>锁定模式 (Lock Mode)</label>
                <div class="lock-selector">
                  <div 
                    :class="['lock-opt', { active: userProfile.profile_lock === 'unlocked' }]" 
                    @click="userProfile.profile_lock = 'unlocked'"
                  >
                    🔓 Unlocked
                  </div>
                  <div 
                    :class="['lock-opt', { active: userProfile.profile_lock === 'soft' }]" 
                    @click="userProfile.profile_lock = 'soft'"
                  >
                    🛡️ SoftLock
                  </div>
                  <div 
                    :class="['lock-opt', { active: userProfile.profile_lock === 'hard' }]" 
                    @click="userProfile.profile_lock = 'hard'"
                  >
                    🔒 HardLock
                  </div>
                </div>
                <p class="lock-tip">
                  {{ lockTipText }}
                </p>
              </div>

              <div class="profile-section">
                <label>核心学习目标</label>
                <textarea v-model="userProfile.goals" placeholder="例如：掌握 Go 语言并发编程..."></textarea>
              </div>

              <div class="profile-section">
                <label>偏好学习风格</label>
                <input v-model="userProfile.preferred_style" placeholder="例如：硬核、深入、实战导向..." />
              </div>

              <!-- AI 建议列表 (仅在 SoftLock 时显示) -->
              <div v-if="userProfile.profile_lock === 'soft' && suggestions.length > 0" class="suggestions-section">
                <div class="section-title">💡 AI 发现的新线索 (建议更新)</div>
                <div class="suggestion-list">
                  <div v-for="(sug, i) in suggestions" :key="i" class="suggestion-item">
                    <span class="sug-op">[{{ sug.op === 'update' ? '更新' : '新增' }}]</span>
                    <span class="sug-key">{{ sug.key === 'learning_goal' ? '学习目标' : '风格' }}:</span>
                    <span class="sug-val">{{ sug.value }}</span>
                    <button class="apply-sug" @click="applySuggestion(sug, i)">采纳</button>
                  </div>
                </div>
                <button class="clear-sug" @click="clearSuggestions">忽略所有建议</button>
              </div>
            </div>

            <div class="profile-modal-footer">
              <button class="cancel-btn" @click="profileModalVisible = false">取消</button>
              <button class="save-btn" @click="saveProfile" :disabled="isSavingProfile">
                {{ isSavingProfile ? '保存中...' : '保存修改' }}
              </button>
            </div>
          </div>
        </div>
      </transition>
      </div>
  </div>
</template>

<script>
import { chatStreamWithAI, getSessions, createSession, getHistory, updateSessionTitle, getKnowledgePreview, getUserProfile, updateUserProfile } from '@/api/ai'
import MarkdownIt from 'markdown-it'

const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true,
  breaks: true // 支持回车换行
})

export default {
  name: 'ChatDialog',
  props: {
    visible: {
      type: Boolean,
      default: false
    }
  },
  data() {
    return {
      messages: [],
      inputMessage: '',
      isLoading: false,
      useKnowledge: true,
      sessions: [],
      currentSessionID: 'default',
      currentSessionTitle: '番茄小助手',
      editingSessionID: null,
      editingTitle: '',
      chatMode: 'thinking',
      showModeMenu: false,
      // 预览相关
      previewVisible: false,
      previewFileName: '',
      previewContent: '',
      isPreviewLoading: false,
      // 布局相关
      sidebarCollapsed: false,
      sidebarWidth: 260,
      previewWidth: 400,
      dialogWidth: 1000,
      dialogHeight: 700,
      resizingType: null,
      // 画像管理相关
      profileModalVisible: false,
      isSavingProfile: false,
      userProfile: {
        goals: '',
        preferred_style: '',
        profile_lock: 'soft',
        lock_suggestions: '[]'
      }
    }
  },
  watch: {
    async visible(newVal) {
      if (newVal) {
        await this.loadSessions()
        this.$nextTick(() => {
          if (this.$refs.inputRef) this.$refs.inputRef.focus()
        })
      }
    },
    messages: {
      handler() {
        this.$nextTick(() => {
          this.scrollToBottom()
        })
      },
      deep: true
    }
  },
  computed: {
    dialogStyle() {
      return {
        width: `${this.dialogWidth}px`,
        height: `${this.dialogHeight}px`,
        resize: 'both',
        overflow: 'hidden',
        minWidth: '600px',
        minHeight: '400px'
      }
    },
    hasSuggestions() {
      try {
        const sugs = JSON.parse(this.userProfile.lock_suggestions || '[]')
        return sugs.length > 0
      } catch (e) {
        return false
      }
    },
    suggestions() {
      try {
        return JSON.parse(this.userProfile.lock_suggestions || '[]')
      } catch (e) {
        return []
      }
    },
    lockTipText() {
      const mode = this.userProfile.profile_lock
      if (mode === 'unlocked') return 'AI 将根据对话自动实时更新您的画像。'
      if (mode === 'soft') return 'AI 不会自动修改，但会根据发现的新兴趣向您提出建议。'
      if (mode === 'hard') return '完全锁定，AI 将停止对该部分记忆的任何提取和更新。'
      return ''
    }
  },
  mounted() {
    this.loadSessions()
    window.addEventListener('mousemove', this.doResize)
    window.addEventListener('mouseup', this.stopResizing)
  },
  beforeUnmount() {
    window.removeEventListener('mousemove', this.doResize)
    window.removeEventListener('mouseup', this.stopResizing)
  },
  methods: {
    async loadSessions() {
      const list = await getSessions()
      this.sessions = list
      if (this.sessions.length > 0 && this.currentSessionID === 'default') {
        this.selectSession(this.sessions[0])
      }
    },

    async startNewSession() {
      const newID = 'sess_' + Math.random().toString(36).substring(2, 11)
      const title = '新对话'
      const success = await createSession(newID, title)
      if (success) {
        await this.loadSessions()
        const newSession = this.sessions.find(s => s.session_id === newID)
        if (newSession) this.selectSession(newSession)
      }
    },

    async selectSession(session) {
      this.currentSessionID = session.session_id
      this.currentSessionTitle = session.title
      this.isLoading = true
      this.isFirstMessageInSession = true
      try {
        const history = await getHistory(session.session_id)
        if (history && history.length > 0) {
          this.isFirstMessageInSession = false
          this.messages = history.map(m => ({
            role: m.role || m.Role,
            content: m.content || m.Content,
            reasoning: m.reasoning_content || m.ReasoningContent || m.reasoning,
            timestamp: m.created_at || m.CreatedAt || m.timestamp
          }))
        } else {
          this.messages = []
          this.addWelcomeMessage()
        }
      } catch (error) {
        console.error('加载历史记录失败:', error)
      } finally {
        this.isLoading = false
      }
    },

    startEditTitle(session) {
      this.editingSessionID = session.session_id
      this.editingTitle = session.title
      this.$nextTick(() => {
        if (this.$refs.editInput && this.$refs.editInput[0]) {
          this.$refs.editInput[0].focus()
        }
      })
    },

    async saveTitle(session) {
      if (!this.editingSessionID) return
      if (this.editingTitle.trim() && this.editingTitle !== session.title) {
        const success = await updateSessionTitle(session.session_id, this.editingTitle.trim())
        if (success) {
          session.title = this.editingTitle.trim()
          if (this.currentSessionID === session.session_id) {
            this.currentSessionTitle = session.title
          }
        }
      }
      this.editingSessionID = null
    },

    addWelcomeMessage() {
      if (this.messages.length === 0) {
        this.messages.push({
          role: 'assistant',
          content: '你好！我是你的 AI 学习教练。我们可以聊聊学习计划，或者帮你解答课业难题。',
          timestamp: new Date()
        })
      }
    },
    
    async sendMessage() {
      if (!this.inputMessage.trim() || this.isLoading) return
      
      const userMessage = this.inputMessage.trim()
      this.inputMessage = ''
      
      const isFirst = this.isFirstMessageInSession
      
      this.messages.push({
        role: 'user',
        content: userMessage,
        timestamp: new Date()
      })
      
      this.isLoading = true
      this.isFirstMessageInSession = false
      
      const assistantMsg = {
        role: 'assistant',
        content: '',
        reasoning: '',
        timestamp: new Date(),
        usage: null
      }
      this.messages.push(assistantMsg)
      const streamMessage = this.messages[this.messages.length - 1]

      try {
        await chatStreamWithAI(
          {
            message: userMessage,
            use_knowledge: this.useKnowledge,
            session_id: this.currentSessionID,
            chat_mode: this.chatMode
          },
          (data) => {
            // 处理实时消息
            if (data.type === 'content' && data.content) {
              streamMessage.content += data.content
            } else if (data.type === 'reasoning' && data.reasoning) {
              streamMessage.reasoning += data.reasoning
            }
            if (data.usage) {
              streamMessage.usage = { ...data.usage }
            }
            this.$nextTick(this.scrollToBottom)
        },
        (data) => {
          // 完成
          this.isLoading = false
          if (data && data.usage) {
            streamMessage.usage = { ...data.usage }
          }
          if (isFirst) {
            setTimeout(() => this.loadSessions(), 3000)
          }
        },
          () => {
            this.isLoading = false
            streamMessage.content += '\n\n[出错了: 连接中断]'
          }
        )
      } catch (error) {
        this.isLoading = false
        console.error('流式传输异常:', error)
      }

      this.$nextTick(() => {
        if (this.$refs.inputRef) this.$refs.inputRef.focus()
        this.scrollToBottom()
      })
    },
    
    closeDialog() {
      this.$emit('close')
    },
    
    scrollToBottom() {
      const container = this.$refs.messagesContainer
      if (container) container.scrollTop = container.scrollHeight
    },
    
    formatTime(timestamp) {
      const date = new Date(timestamp)
      return `${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`
    },

    formatDate(dateStr) {
      if (!dateStr) return ''
      const date = new Date(dateStr)
      return `${date.getMonth() + 1}-${date.getDate()}`
    },

    renderMarkdown(text) {
      if (!text) return ''
      
      let processedText = text
      
      // 1. 处理双括号格式 [[文件名.md]]
      processedText = processedText.replace(/\[\[(.*?)\]\]/g, (match, fileName) => {
        return `<a class="file-citation" href="preview:${fileName}" data-preview="${fileName}">📄 ${fileName}</a>`
      })
      
      // 2. 处理可能被写错的 Markdown 格式 [文件名](preview:文件名) 
      // 这里的正则专门处理文件名中带空格导致渲染器解析失败的情况
      processedText = processedText.replace(/\[(.*?)\]\(preview:(.*?)\)/g, (match, title, fileName) => {
        const actualFile = fileName.trim()
        return `<a class="file-citation" href="preview:${actualFile}" data-preview="${actualFile}">📄 ${actualFile}</a>`
      })
      
      return md.render(processedText)
    },

    isPendingAssistantMessage(message, index) {
      return this.isLoading &&
        index === this.messages.length - 1 &&
        message.role === 'assistant' &&
        !message.content &&
        !message.reasoning
    },

    async handleMessageClick(event) {
      // 检查点击的是否是带有 data-preview 属性的元素，或者是 preview: 协议的链接
      const target = event.target
      
      // 处理 data-preview 属性 (自定义标签)
      if (target.dataset.preview) {
        event.preventDefault()
        this.showFilePreview(target.dataset.preview)
        return
      }

      // 处理 Markdown 生成的 <a> 标签
      const link = target.closest('a')
      if (link && link.getAttribute('href')?.startsWith('preview:')) {
        event.preventDefault()
        const fileName = link.getAttribute('href').replace('preview:', '')
        this.showFilePreview(fileName)
      }
    },

    async showFilePreview(fileName) {
      this.previewFileName = fileName
      this.previewVisible = true
      this.isPreviewLoading = true
      try {
        const content = await getKnowledgePreview(fileName)
        this.previewContent = content || '该文件暂无内容。'
      } catch (error) {
        this.previewContent = '加载预览失败，请稍后重试。'
      } finally {
        this.isPreviewLoading = false
      }
    },

    setChatMode(mode) {
      this.chatMode = mode
      this.showModeMenu = false
    },

    // 拖拽逻辑
    startResizing(type) {
      this.isResizing = true
      this.resizingType = type
      document.body.style.cursor = 'col-resize'
    },
    stopResizing() {
      this.isResizing = false
      this.resizingType = null
      document.body.style.cursor = 'default'
    },
    doResize(event) {
      if (!this.isResizing) return
      
      const dialogRect = this.$el.querySelector('.chat-dialog').getBoundingClientRect()
      
      if (this.resizingType === 'sidebar') {
        const newWidth = event.clientX - dialogRect.left
        if (newWidth > 150 && newWidth < 400) {
          this.sidebarWidth = newWidth
        }
      } else if (this.resizingType === 'preview') {
        const newWidth = dialogRect.right - event.clientX
        if (newWidth > 200 && newWidth < 600) {
          this.previewWidth = newWidth
        }
      }
    },

    // 画像管理方法
    async openProfileModal() {
      this.profileModalVisible = true
      const profile = await getUserProfile()
      if (profile) {
        this.userProfile = profile
      }
    },

    async saveProfile() {
      this.isSavingProfile = true
      const success = await updateUserProfile({
        goals: this.userProfile.goals,
        preferred_style: this.userProfile.preferred_style,
        profile_lock: this.userProfile.profile_lock
      })
      if (success) {
        this.profileModalVisible = false
      }
      this.isSavingProfile = false
    },

    applySuggestion(sug, index) {
      const normalize = (str) => str.replace(/\s+/g, '').toLowerCase()
      const targetVal = sug.value
      const normalizedTarget = normalize(targetVal)

      if (sug.key === 'learning_goal') {
        const currentGoals = this.userProfile.goals.split(/[；;]/).filter(g => g.trim())
        const isDuplicate = currentGoals.some(g => normalize(g) === normalizedTarget)

        if (!isDuplicate) {
          if (sug.op === 'update') this.userProfile.goals = targetVal
          else {
            if (this.userProfile.goals) this.userProfile.goals += '；' + targetVal
            else this.userProfile.goals = targetVal
          }
        }
      } else if (sug.key === 'preferred_style') {
        if (normalize(this.userProfile.preferred_style) !== normalizedTarget) {
          this.userProfile.preferred_style = targetVal
        }
      }
      
      // 移除已处理的建议
      const currentSugs = [...this.suggestions]
      currentSugs.splice(index, 1)
      this.userProfile.lock_suggestions = JSON.stringify(currentSugs)
    },

    async clearSuggestions() {
      const success = await updateUserProfile({
        clear_suggestions: true
      })
      if (success) {
        this.userProfile.lock_suggestions = '[]'
      }
    }
  }
}
</script>

<style scoped>
.chat-dialog-overlay {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10000;
}

.chat-dialog {
  width: 90%;
  max-width: 850px;
  height: 80vh;
  background: white;
  border-radius: 20px;
  display: flex;
  overflow: hidden;
  box-shadow: 0 20px 50px rgba(0,0,0,0.15);
}

/* 侧边栏样式 */
.chat-sidebar {
  width: 240px;
  background: #f8f9fa;
  border-right: 1px solid #eee;
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  padding: 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid #eee;
}

.sidebar-title {
  font-weight: 700;
  color: #444;
  font-size: 15px;
}

.new-chat-btn {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  border: 1px solid #ddd;
  background: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #666;
  transition: all 0.2s;
}

.new-chat-btn:hover {
  background: #eeaa67;
  color: white;
  border-color: #eeaa67;
}

.session-list {
  flex: 1;
  overflow-y: auto;
  padding: 10px;
}

.session-item {
  padding: 12px;
  border-radius: 12px;
  display: flex;
  align-items: flex-start;
  gap: 10px;
  cursor: pointer;
  margin-bottom: 4px;
  transition: background 0.2s;
}

.session-item:hover {
  background: #f0f0f0;
}

.session-item.active {
  background: #fff3e0;
}

.message-time {
  font-size: 11px;
  color: #94a3b8;
  margin-top: 4px;
  display: flex;
  align-items: center;
}

.token-usage {
  margin-left: 6px;
  color: #cbd5e1;
  font-weight: normal;
}

.session-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.session-title-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.session-title {
  font-size: 13px;
  font-weight: 500;
  color: #333;
  white-space: nowrap;
  text-overflow: ellipsis;
  overflow: hidden;
  max-width: 140px;
}

.edit-title-input {
  flex: 1;
  font-size: 13px;
  border: 1px solid #eeaa67;
  border-radius: 4px;
  padding: 2px 4px;
  width: 100%;
}

.edit-btn {
  background: none;
  border: none;
  font-size: 10px;
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.2s;
}

.session-item:hover .edit-btn {
  opacity: 0.5;
}

.edit-btn:hover {
  opacity: 1 !important;
}

.session-time {
  font-size: 11px;
  color: #999;
}

.no-sessions {
  text-align: center;
  color: #aaa;
  margin-top: 40px;
  font-size: 12px;
}

/* 主区域样式 */
.chat-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: white;
}

.chat-header {
  padding: 15px 25px;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.chat-header-left {
  display: flex;
  align-items: center;
  gap: 15px;
}

.pet-avatar {
  font-size: 28px;
  background: #fff3e0;
  width: 45px;
  height: 45px;
  border-radius: 15px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.chat-name {
  font-weight: 700;
  color: #333;
  font-size: 16px;
}

.chat-status {
  font-size: 12px;
  color: #27ae60;
}

.close-btn {
  font-size: 24px;
  background: none;
  border: none;
  color: #ccc;
  cursor: pointer;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 25px;
  background: #fafafa;
}

.message {
  margin-bottom: 20px;
  display: flex;
  gap: 12px;
}

.message.user {
  flex-direction: row-reverse;
}

.message-avatar {
  width: 38px;
  height: 38px;
  border-radius: 12px;
  background: white;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 10px rgba(0,0,0,0.05);
}

.message-content {
  max-width: 75%;
}

.message.user .message-content {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
}

.message-text {
  padding: 12px 18px;
  border-radius: 18px;
  font-size: 14.5px;
  line-height: 1.6;
  user-select: text !important; /* 确保文字可以被选中复制 */
  word-break: break-word;
}

.message.user .message-text {
  background: #eeaa67;
  color: white;
  border-top-right-radius: 4px;
}

.message.assistant .message-text {
  background: white;
  color: #333;
  border-top-left-radius: 4px;
  box-shadow: 0 4px 15px rgba(0,0,0,0.03);
}

/* Markdown 渲染样式补丁 */
.message-text :deep(p) {
  margin: 0 0 10px 0;
}

.message-text :deep(p):last-child {
  margin-bottom: 0;
}

.message-text :deep(ul), .message-text :deep(ol) {
  margin: 10px 0;
  padding-left: 25px;
}

.message-text :deep(li) {
  margin-bottom: 6px;
}

.message-text :deep(strong) {
  color: #eeaa67;
  font-weight: 700;
}

.message-text :deep(code) {
  background: #fdf6ec;
  color: #e6a23c;
  padding: 2px 4px;
  border-radius: 4px;
  font-family: monospace;
}

.message-text :deep(pre) {
  background: #2d2d2d;
  color: #ccc;
  padding: 12px;
  border-radius: 8px;
  overflow-x: auto;
  margin: 12px 0;
}

.message-text :deep(h1), .message-text :deep(h2), .message-text :deep(h3) {
  margin: 15px 0 10px 0;
  color: #333;
  font-weight: 700;
}

.message-reasoning {
  background: #f1f3f5;
  border-left: 4px solid #eeaa67;
  padding: 10px 15px;
  margin-bottom: 10px;
  border-radius: 8px;
  font-size: 13px;
}

.reasoning-title {
  color: #888;
  font-weight: 600;
  margin-bottom: 5px;
}

.reasoning-text {
  color: #666;
  font-style: italic;
}

.chat-input-area {
  padding: 20px 25px;
  background: white;
}

.chat-input-wrapper {
  display: flex;
  gap: 12px;
  align-items: flex-end;
  background: #f8f9fa;
  padding: 8px 15px;
  border-radius: 18px;
  border: 1px solid #eee;
}

.chat-input {
  flex: 1;
  border: none;
  background: none;
  padding: 10px 0;
  font-size: 14px;
  resize: none;
  max-height: 100px;
}

.chat-input:focus {
  outline: none;
}

.input-actions {
  display: flex;
  gap: 8px;
  padding-bottom: 5px;
}

.rag-toggle-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  border-radius: 20px;
  border: 1px solid #ddd;
  background: #f0f0f0;
  color: #777;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  font-weight: 500;
}

.rag-toggle-btn.active {
  background: #fff3e0;
  border-color: #eeaa67;
  color: #eeaa67;
  box-shadow: 0 2px 8px rgba(238, 170, 103, 0.2);
}

.rag-status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #bbb;
}

.active .rag-status-dot {
  background: #eeaa67;
  box-shadow: 0 0 5px #eeaa67;
}

.file-citation {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: #fff5eb;
  color: #eeaa67 !important;
  border: 1px solid #ffe4cc;
  padding: 2px 8px;
  border-radius: 6px;
  text-decoration: none !important;
  font-size: 0.9em;
  margin: 0 4px;
  font-weight: 600;
  transition: all 0.2s;
}

.file-citation:hover {
  background: #eeaa67;
  color: white !important;
  transform: translateY(-1px);
}

.send-btn {
  background: #eeaa67;
  color: white;
  border: none;
  width: 36px;
  height: 36px;
  border-radius: 12px;
  cursor: pointer;
  font-weight: 700;
}

.send-btn:disabled {
  background: #ddd;
}

.typing span {
  width: 6px;
  height: 6px;
  background: #eeaa67;
  border-radius: 50%;
  display: inline-block;
  margin: 0 2px;
  animation: bounce 1.4s infinite;
}

@keyframes bounce {
  0%, 80%, 100% { transform: scale(0); }
  40% { transform: scale(1.0); }
}

@media (max-width: 768px) {
  .chat-sidebar {
    display: none;
  }
}

/* 预览面板样式 */
.preview-panel {
  background: white;
  border-left: 1px solid #eee;
  display: flex;
  flex-direction: column;
  z-index: 10;
  box-shadow: -5px 0 15px rgba(0,0,0,0.05);
}

.preview-header {
  padding: 18px 20px;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fafafa;
}

.preview-title-container {
  display: flex;
  align-items: center;
  gap: 10px;
  overflow: hidden;
}

.preview-title {
  font-weight: 700;
  color: #333;
  font-size: 15px;
  white-space: nowrap;
  text-overflow: ellipsis;
  overflow: hidden;
}

.preview-close {
  background: none;
  border: none;
  font-size: 20px;
  color: #999;
  cursor: pointer;
  padding: 0 5px;
}

.preview-content-area {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  background: white;
}

.preview-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
  color: #999;
  gap: 15px;
}

.spinner {
  width: 30px;
  height: 30px;
  border: 3px solid #f3f3f3;
  border-top: 3px solid #eeaa67;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

/* 动画 */
.slide-enter-active, .slide-leave-active {
  transition: all 0.3s ease;
}
.slide-enter-from, .slide-leave-to {
  transform: translateX(100%);
  opacity: 0;
}

/* Markdown 预览样式补丁 */
.preview-body :deep(p) { line-height: 1.8; margin-bottom: 15px; }
.preview-body :deep(h1), .preview-body :deep(h2) { border-bottom: 1px solid #eee; padding-bottom: 10px; margin-top: 20px; }

/* 拖拽与折叠辅助样式 */
.chat-header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.collapse-toggle {
  background: white;
  border: 1px solid #eee;
  border-radius: 8px;
  cursor: pointer;
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #999;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 2px 5px rgba(0,0,0,0.05);
}

.collapse-toggle:hover {
  border-color: #eeaa67;
  color: #eeaa67;
  transform: scale(1.05);
  box-shadow: 0 4px 8px rgba(238, 170, 103, 0.15);
}

.collapsed-sidebar-tip {
  width: 12px;
  background: #fdfdfd;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  border-right: 1px solid #f0f0f0;
  transition: all 0.3s;
  position: relative;
  overflow: visible;
}

.collapsed-sidebar-tip:hover {
  width: 24px;
  background: #fff9f3;
  border-right-color: #ffe4cc;
}

.expand-icon-wrapper {
  opacity: 0;
  transform: translateX(-5px);
  transition: all 0.3s;
  color: #eeaa67;
}

.collapsed-sidebar-tip:hover .expand-icon-wrapper {
  opacity: 1;
  transform: translateX(0);
}

.collapsed-sidebar-tip::after {
  content: '';
  position: absolute;
  right: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  background: #eeaa67;
  opacity: 0.3;
}

.collapsed-sidebar-tip:hover::after {
  opacity: 1;
}

.resizer {
  background: transparent;
  transition: background 0.2s;
  z-index: 30;
}

.resizer:hover {
  background: rgba(238, 170, 103, 0.3);
}

.resizer-h {
  width: 6px;
  cursor: col-resize;
  margin: 0 -3px;
  flex-shrink: 0;
}

.chat-dialog {
  display: flex;
  flex-direction: row;
  background: white;
  border-radius: 24px;
  box-shadow: 0 20px 60px rgba(0,0,0,0.15);
  overflow: hidden;
  transition: none; /* 拖拽时不使用过渡 */
}

.chat-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 300px;
  background: #fff;
  z-index: 5;
}

.stream-toggle {
  background: #f0f9ff;
  border-color: #bae6fd;
  color: #0369a1;
  margin-left: 8px;
}

.stream-toggle.active {
  background: #0ea5e9;
  color: white;
  border-color: #0284c7;
}

.stream-toggle.active .rag-status-dot {
  background: #fff;
}
.chat-input-wrapper {
  background: white;
  border: 1px solid #eee;
  border-radius: 16px;
  padding: 8px 12px;
  display: flex;
  flex-direction: column;
  transition: all 0.3s;
  box-shadow: 0 2px 10px rgba(0,0,0,0.03);
}

.chat-input-wrapper:focus-within {
  border-color: #eeaa67;
  box-shadow: 0 4px 15px rgba(238, 170, 103, 0.15);
}

.chat-input {
  border: none;
  resize: none;
  width: 100%;
  padding: 8px;
  font-size: 14px;
  color: #333;
  outline: none;
  background: transparent;
  min-height: 40px;
}

.input-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 4px;
  padding-top: 4px;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  background: #f8f9fa;
  border: 1px solid #f0f0f0;
  padding: 4px 10px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  color: #64748b;
  font-size: 12px;
}

.action-btn:hover {
  background: #fff5eb;
  border-color: #ffe4cc;
  color: #eeaa67;
}

.action-btn.active {
  background: #fff3e0;
  border-color: #eeaa67;
  color: #eeaa67;
}

.action-icon {
  font-size: 14px;
}

.actions-divider {
  width: 1px;
  height: 16px;
  background: #eee;
  margin: 0 4px;
}

.mode-selector-container {
  position: relative;
}

.mode-dropdown-compact {
  position: absolute;
  bottom: calc(100% + 10px);
  right: 0;
  background: white;
  border-radius: 12px;
  box-shadow: 0 8px 25px rgba(0,0,0,0.12);
  border: 1px solid #eee;
  padding: 6px;
  z-index: 100;
  min-width: 100px;
}

.mode-opt {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 8px;
  cursor: pointer;
  font-size: 13px;
  color: #444;
  transition: all 0.2s;
  white-space: nowrap;
}

.mode-opt:hover {
  background: #f8f9fa;
}

.mode-opt.active {
  background: #fff3e0;
  color: #eeaa67;
  font-weight: 700;
}

.opt-icon {
  font-size: 16px;
}

.send-btn {
  background: #eeaa67;
  color: white;
  border: none;
  padding: 6px 16px;
  border-radius: 10px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 13px;
}

.send-btn:hover:not(:disabled) {
  background: #e69a55;
  transform: translateY(-1px);
  box-shadow: 0 4px 10px rgba(238, 170, 103, 0.3);
}

.send-btn:disabled {
  background: #f1f5f9;
  color: #cbd5e1;
  cursor: not-allowed;
}

.fade-up-enter-active, .fade-up-leave-active {
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}
.fade-up-enter-from, .fade-up-leave-to {
  opacity: 0;
  transform: translateY(8px);
}

/* 画像管理相关样式 */
.header-action-btn {
  background: none;
  border: none;
  font-size: 18px;
  cursor: pointer;
  padding: 5px;
  position: relative;
  transition: transform 0.2s;
  margin-right: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.header-action-btn:hover {
  transform: scale(1.1);
}

.suggestion-badge {
  position: absolute;
  top: 0;
  right: 0;
  width: 8px;
  height: 8px;
  background: #ff4d4f;
  border-radius: 50%;
  border: 1px solid white;
}

.profile-modal-overlay {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
  backdrop-filter: blur(2px);
}

.profile-modal {
  width: 450px;
  background: white;
  border-radius: 16px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  animation: modal-up 0.3s ease-out;
}

@keyframes modal-up {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

.profile-modal-header {
  padding: 15px 20px;
  border-bottom: 1px solid #eee;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fcfcfc;
}

.profile-modal-header h3 {
  margin: 0;
  font-size: 16px;
  color: #333;
}

.modal-close {
  background: none;
  border: none;
  font-size: 20px;
  color: #999;
  cursor: pointer;
}

.profile-modal-body {
  padding: 20px;
  max-height: 500px;
  overflow-y: auto;
}

.profile-section {
  margin-bottom: 20px;
}

.profile-section label {
  display: block;
  font-size: 13px;
  font-weight: 600;
  color: #64748b;
  margin-bottom: 8px;
}

.lock-selector {
  display: flex;
  background: #f1f5f9;
  padding: 4px;
  border-radius: 8px;
  gap: 4px;
}

.lock-opt {
  flex: 1;
  text-align: center;
  font-size: 12px;
  padding: 6px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  color: #64748b;
}

.lock-opt.active {
  background: white;
  color: #eeaa67;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  font-weight: 600;
}

.lock-tip {
  font-size: 11px;
  color: #94a3b8;
  margin-top: 8px;
  line-height: 1.4;
}

.profile-section textarea, .profile-section input {
  width: 100%;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 10px;
  font-size: 14px;
  outline: none;
  transition: all 0.2s;
}

.profile-section textarea:focus, .profile-section input:focus {
  border-color: #eeaa67;
  box-shadow: 0 0 0 3px rgba(238, 170, 103, 0.1);
}

.profile-section textarea {
  height: 80px;
  resize: none;
}

.suggestions-section {
  background: #fff7ed;
  border: 1px solid #ffedd5;
  border-radius: 12px;
  padding: 12px;
  margin-top: 10px;
}

.section-title {
  font-size: 12px;
  font-weight: 700;
  color: #c2410c;
  margin-bottom: 10px;
}

.suggestion-list {
  max-height: 150px;
  overflow-y: auto;
}

.suggestion-item {
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 8px;
  color: #475569;
  padding: 4px;
  border-radius: 4px;
  background: white;
}

.sug-op { color: #f97316; font-weight: 700; }
.sug-key { color: #64748b; font-weight: 500; }
.sug-val { flex: 1; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

.apply-sug {
  background: #f97316;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 2px 8px;
  font-size: 11px;
  cursor: pointer;
}

.clear-sug {
  width: 100%;
  background: none;
  border: 1px dashed #fdba74;
  color: #f97316;
  padding: 6px;
  border-radius: 6px;
  font-size: 11px;
  cursor: pointer;
  margin-top: 5px;
}

.profile-modal-footer {
  padding: 15px 20px;
  border-top: 1px solid #eee;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  background: #fcfcfc;
}

.cancel-btn {
  background: none;
  border: 1px solid #ddd;
  padding: 8px 16px;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  color: #666;
}

.save-btn {
  background: #eeaa67;
  color: white;
  border: none;
  padding: 8px 24px;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
}

.save-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.fade-enter-active, .fade-leave-active {
  transition: opacity 0.3s;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}
</style>
