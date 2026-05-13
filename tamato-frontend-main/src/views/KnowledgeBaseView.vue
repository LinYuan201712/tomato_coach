<template>
  <div class="knowledge-base-container">
    <!-- 顶部导航栏 (完美同步 HomeView 风格) -->
    <nav class="kb-navbar">
      <div class="nav-brand" @click="$router.push('/home')">Tomato</div>
      <div class="nav-links">
        <a class="nav-link" @click="$router.push('/friends')">好友</a>
        <a class="nav-link active">知识库</a>
        <a class="nav-link" @click="$router.push('/task-management')">任务管理</a>
        <a class="nav-link" @click="$router.push('/personal-center')">个人中心</a>
      </div>
      <div class="user-avatar-container">
        <div class="user-name">{{ userInfo.username || 'User' }}</div>
        <div class="user-avatar-box">
          <img :src="displayAvatar" alt="用户头像" @error="handleAvatarError" />
        </div>
      </div>
    </nav>

    <div class="kb-content" :style="{ marginRight: contentMarginRight + 'px' }">
      <!-- 页面标题与工具栏 -->
      <div class="page-header">
        <div class="header-left">
          <button class="back-circle-btn" @click="goBack" v-if="currentFolderID !== 0">
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2.5">
              <path d="M15 18l-6-6 6-6" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </button>
          <h1 class="page-title">{{ currentFolderName || '我的文档' }}</h1>
        </div>
        <div class="toolbar">
          <div class="search-box-container">
            <div class="search-input-wrapper">
              <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" class="search-icon">
                <circle cx="11" cy="11" r="8"></circle>
                <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
              </svg>
              <input 
                v-model="searchQuery" 
                type="text" 
                placeholder="搜索文件或文件夹..." 
                class="search-input"
              />
            </div>
          </div>
          <div class="new-btn-group">
            <button class="new-btn tomato-btn" @click.stop="showNewMenu = !showNewMenu">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="3">
                <path d="M12 5v14M5 12h14" stroke-linecap="round"/>
              </svg>
              新建
            </button>
            <transition name="dropdown">
              <div v-if="showNewMenu" class="dropdown-menu tomato-dropdown" @click.stop>
                <div class="dropdown-item" @click="triggerFileUpload">
                  <div class="item-icon-circle upload-bg">
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="white" stroke-width="2.5">
                      <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M17 8l-5-5-5 5M12 3v12" stroke-linecap="round" stroke-linejoin="round"/>
                    </svg>
                  </div>
                  <div class="item-text">
                    <span class="item-title">上传文件</span>
                    <span class="item-desc">PDF, Word, Markdown</span>
                  </div>
                </div>
                <div class="dropdown-item" @click="openModal('folder')">
                  <div class="item-icon-circle folder-bg">
                    <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="white" stroke-width="2.5">
                      <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" stroke-linecap="round" stroke-linejoin="round"/>
                    </svg>
                  </div>
                  <div class="item-text">
                    <span class="item-title">新建文件夹</span>
                    <span class="item-desc">整理您的知识档案</span>
                  </div>
                </div>
              </div>
            </transition>
          </div>
        </div>
      </div>

      <!-- 空状态 -->
      <div v-if="!isPageLoading && filteredFolders.length === 0 && filteredFiles.length === 0" class="empty-state-modern">
        <div class="empty-visual">
          <div class="empty-bg-circle"></div>
          <div class="empty-icon-main">
            <svg viewBox="0 0 24 24" width="64" height="64" fill="none" stroke="#eeaa67" stroke-width="1.5">
              <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" stroke-linecap="round" stroke-linejoin="round"/>
              <line x1="12" y1="11" x2="12" y2="17" stroke-linecap="round"/>
              <line x1="9" y1="14" x2="15" y2="14" stroke-linecap="round"/>
            </svg>
          </div>
        </div>
        <div class="empty-text-box">
          <h3>{{ searchQuery ? '未找到相关内容' : '开启您的知识库' }}</h3>
          <p>{{ searchQuery ? '尝试搜索其他关键词' : '在这里存储您的 PDF、Word 和 Markdown 资料，方便随时查阅' }}</p>
          <button v-if="!searchQuery" class="empty-action-btn tomato-btn" @click="triggerFileUpload">上传第一个文件</button>
        </div>
      </div>

      <!-- 文件夹区域 -->
      <section class="kb-section" v-if="filteredFolders.length > 0">
        <h2 class="section-title">文件夹</h2>
        <div class="folders-grid">
          <div 
            v-for="folder in filteredFolders" 
            :key="folder.id" 
            class="folder-card"
            @click="openFolder(folder)"
            @contextmenu.prevent="showContextMenu($event, folder, 'folder')"
          >
            <div class="folder-visual">
              <div class="folder-icon-glow"></div>
              <svg viewBox="0 0 24 24" width="48" height="48" class="main-folder-icon">
                <defs>
                  <linearGradient :id="'folderGrad' + folder.id" x1="0%" y1="0%" x2="100%" y2="100%">
                    <stop offset="0%" style="stop-color:#eeaa67;stop-opacity:1" />
                    <stop offset="100%" style="stop-color:#f39c12;stop-opacity:1" />
                  </linearGradient>
                </defs>
                <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" :fill="'url(#folderGrad' + folder.id + ')'"/>
              </svg>
            </div>
            <div class="folder-info">
              <h3 class="folder-name-text">{{ folder.name }}</h3>
              <p class="folder-stats-text">最近更新</p>
            </div>
          </div>
        </div>
      </section>

      <!-- 最近文件区域 -->
      <section class="kb-section" v-if="recentFiles.length > 0 && !searchQuery">
        <h2 class="section-title">最近</h2>
        <div class="recent-grid">
          <div 
            v-for="file in recentFiles" 
            :key="file.id" 
            class="file-card-box"
            :class="{ active: isFileInPreview(file) }"
            @click="previewFile(file)"
            @contextmenu.prevent="showContextMenu($event, file, 'file')"
          >
            <div class="file-icon-square">
              <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="#eeaa67" stroke-width="2">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" stroke-linecap="round" stroke-linejoin="round"/>
                <path d="M14 2v6h6" stroke-linecap="round" stroke-linejoin="round"/>
                <path d="M16 13H8M16 17H8M10 9H8" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </div>
            <div class="file-content-info">
              <h3 class="file-display-name">{{ file.displayName || file.name }}</h3>
              <p class="file-display-meta">{{ file.createdAt }} • {{ formatSize(file.size) }}</p>
            </div>
          </div>
        </div>
      </section>

      <!-- 所有文件列表 -->
      <section class="kb-section" v-if="filteredFiles.length > 0">
        <h2 class="section-title">{{ searchQuery ? '搜索结果' : '所有文件' }}</h2>
        <div class="files-table-container">
          <table class="files-data-table">
            <thead>
              <tr>
                <th>名称</th>
                <th>创建时间</th>
                <th>大小</th>
              </tr>
            </thead>
            <tbody>
              <tr 
                v-for="file in filteredFiles" 
                :key="file.id" 
                :class="{ active: isFileInPreview(file) }"
                @click="previewFile(file)"
                @contextmenu.prevent="showContextMenu($event, file, 'file')"
              >
                <td class="name-td">
                  <div class="table-file-cell">
                    <div class="table-file-icon">
                      <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M13 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9z" stroke-linecap="round" stroke-linejoin="round"/>
                        <path d="M13 2v7h7" stroke-linecap="round" stroke-linejoin="round"/>
                      </svg>
                    </div>
                    <span>{{ file.displayName || file.name }}</span>
                  </div>
                </td>
                <td class="date-td">{{ file.createdAt }}</td>
                <td class="size-td">{{ formatSize(file.size) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </div>

    <!-- 自定义模态框 (取代 prompt) -->
    <transition name="fade">
      <div v-if="modal.show" class="modal-overlay" @click.self="modal.show = false">
        <div class="modal-card tomato-modal">
          <div class="modal-header">
            <h3>{{ modal.title }}</h3>
            <button @click="modal.show = false" class="close-modal-btn">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M18 6L6 18M6 6l12 12" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </button>
          </div>
          <div class="modal-body">
            <label>{{ modal.label }}</label>
            <input 
              v-model="modal.inputValue" 
              :placeholder="modal.placeholder" 
              @keyup.enter="confirmModal"
              ref="modalInput"
            />
          </div>
          <div class="modal-footer">
            <button class="btn-cancel-modal" @click="modal.show = false">取消</button>
            <button class="btn-confirm-modal tomato-btn" @click="confirmModal">确定</button>
          </div>
        </div>
      </div>
    </transition>

    <!-- 移动文件模态框 -->
    <transition name="fade">
      <div v-if="moveModal.show" class="modal-overlay" @click.self="moveModal.show = false">
        <div class="modal-card tomato-modal move-modal">
          <div class="modal-header">
            <h3>移动文件</h3>
            <button @click="moveModal.show = false" class="close-modal-btn">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M18 6L6 18M6 6l12 12" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </button>
          </div>
          <div class="modal-body">
            <p class="move-info">将 <b>{{ moveModal.target?.displayName || moveModal.target?.name }}</b> 移动到：</p>
            <div class="folder-list-select">
              <div 
                class="folder-select-item" 
                :class="{ selected: moveModal.targetFolderID === 0 }"
                @click="moveModal.targetFolderID = 0"
              >
                <span class="folder-icon">📂</span>
                <span>根目录</span>
              </div>
              <div 
                v-for="f in allFoldersList" 
                :key="f.id"
                class="folder-select-item"
                :class="{ selected: moveModal.targetFolderID === f.id }"
                @click="moveModal.targetFolderID = f.id"
              >
                <span class="folder-icon">📁</span>
                <span>{{ f.name }}</span>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button class="btn-cancel-modal" @click="moveModal.show = false">取消</button>
            <button class="btn-confirm-modal tomato-btn" @click="confirmMove">移动</button>
          </div>
        </div>
      </div>
    </transition>

    <!-- 右侧预览面板 -->
    <div 
      class="preview-side-panel" 
      :class="{ open: previewTabs.length > 0 }"
      :style="{ width: panelWidth + 'px', right: previewTabs.length > 0 ? '0' : '-' + panelWidth + 'px' }"
    >
      <!-- 拖拽手柄 -->
      <div class="panel-resizer" @mousedown="startResize"></div>
      
      <div class="panel-header">
        <div class="tabs-container">
          <div 
            v-for="(tab, index) in previewTabs" 
            :key="tab.file.id"
            class="preview-tab"
            :class="{ active: activeTabIndex === index }"
            @click="activeTabIndex = index"
          >
            <span class="tab-name">{{ tab.file.displayName || tab.file.name }}</span>
            <span class="tab-close" @click.stop="closeTab(index)">×</span>
          </div>
        </div>
        <button class="panel-close-btn" @click="closeAllTabs">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M18 6L6 18M6 6l12 12" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </button>
      </div>
      <div class="panel-content">
        <div v-if="previewTabs.length === 0" class="preview-empty">
          <div class="empty-icon">📄</div>
          <p>选择文件进行预览</p>
        </div>
        <div 
          v-for="(tab, index) in previewTabs" 
          :key="tab.file.id"
          v-show="activeTabIndex === index"
          class="tab-content-wrapper markdown-body"
          v-html="tab.content"
        ></div>
      </div>
    </div>

    <!-- 右键上下文菜单 -->
    <transition name="fade">
      <div 
        v-if="contextMenu.visible" 
        class="context-menu-box" 
        :style="{ top: contextMenu.y + 'px', left: contextMenu.x + 'px' }"
        @click.stop
      >
        <div class="menu-item-row" @click="openModal('rename')">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
          </svg>
          重命名
        </div>
        <div v-if="contextMenu.type === 'file'" class="menu-item-row" @click="openMoveModal">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z"/>
          </svg>
          移动到
        </div>
        <div class="menu-divider-line"></div>
        <div class="menu-item-row delete-red" @click="handleDelete">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="3 6 5 6 21 6"></polyline>
            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
          </svg>
          彻底删除
        </div>
      </div>
    </transition>

    <!-- 隐藏的上传控件 -->
    <input type="file" ref="fileInput" style="display: none" accept=".pdf,.doc,.docx,.md" @change="onFileSelected" />
    
    <!-- 全局加载遮罩 -->
    <div v-if="isPageLoading" class="kb-global-loader">
      <div class="kb-loader-spinner"></div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { 
  listKnowledge, listFolders, uploadKnowledge, deleteKnowledge, 
  renameKnowledge, createFolder, deleteFolder, getKnowledgePreview,
  moveKnowledge
} from '@/api/ai'
import MarkdownIt from 'markdown-it'
import defaultAvatar from '@/assets/images/avatar.png'

const md = new MarkdownIt({
  html: true,
  linkify: true,
  breaks: true
})

// 数据状态
const folders = ref([])
const allFiles = ref([])
const currentFolderID = ref(0)
const currentFolderName = ref('')
const userInfo = ref({
  username: localStorage.getItem('username') || 'User',
  avatar: localStorage.getItem('avatar') || ''
})

// 头像处理逻辑 (同步 HomeView)
const displayAvatar = computed(() => {
  const avatar = userInfo.value.avatar
  if (!avatar) return defaultAvatar
  if (avatar.startsWith('http') || avatar.startsWith('data:')) return avatar
  return `http://localhost:8090${avatar.startsWith('/') ? '' : '/'}${avatar}`
})

const handleAvatarError = (e) => {
  e.target.src = defaultAvatar
}

// 计算最近文件
const recentFiles = computed(() => {
  return [...allFiles.value]
    .sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt))
    .slice(0, 4)
})

// 状态控制
const showNewMenu = ref(false)
const searchQuery = ref('')

// 过滤后的数据
const filteredFolders = computed(() => {
  if (!searchQuery.value) return folders.value
  const query = searchQuery.value.toLowerCase()
  return folders.value.filter(f => f.name.toLowerCase().includes(query))
})

const filteredFiles = computed(() => {
  if (!searchQuery.value) return allFiles.value
  const query = searchQuery.value.toLowerCase()
  return allFiles.value.filter(f => 
    (f.displayName || f.name).toLowerCase().includes(query)
  )
})

// 预览状态
const previewTabs = ref([]) // { file: object, content: string }
const activeTabIndex = ref(-1)

const isFileInPreview = (file) => {
  return previewTabs.value.some(tab => tab.file.id === file.id)
}

const contextMenu = ref({ visible: false, x: 0, y: 0, target: null, type: '' })
const fileInput = ref(null)
const isPageLoading = ref(false)

// 侧边栏宽度逻辑
const panelWidth = ref(450)
const isResizing = ref(false)

const startResize = () => {
  isResizing.value = true
  document.addEventListener('mousemove', handleResize)
  document.addEventListener('mouseup', stopResize)
  document.body.style.cursor = 'col-resize'
  document.body.style.userSelect = 'none'
}

const handleResize = (e) => {
  if (!isResizing.value) return
  const newWidth = window.innerWidth - e.clientX
  if (newWidth > 300 && newWidth < 1000) {
    panelWidth.value = newWidth
  }
}

const stopResize = () => {
  isResizing.value = false
  document.removeEventListener('mousemove', handleResize)
  document.removeEventListener('mouseup', stopResize)
  document.body.style.cursor = 'default'
  document.body.style.userSelect = 'auto'
}

// 动态计算主内容边距
const contentMarginRight = computed(() => {
  return previewTabs.value.length > 0 ? panelWidth.value : 0
})

// 模态框状态
const modal = ref({
  show: false,
  title: '',
  label: '',
  inputValue: '',
  placeholder: '',
  type: '', 
  target: null
})

const moveModal = ref({
  show: false,
  target: null,
  targetFolderID: 0
})

const allFoldersList = ref([]) // 用于移动时的全局文件夹列表

const openModal = (type) => {
  showNewMenu.value = false
  contextMenu.value.visible = false
  
  if (type === 'folder') {
    modal.value = {
      show: true,
      title: 'New Folder',
      label: 'Folder Name',
      inputValue: '',
      placeholder: 'Enter folder name...',
      type: 'folder'
    }
  } else if (type === 'rename') {
    const target = contextMenu.value.target
    modal.value = {
      show: true,
      title: 'Rename',
      label: 'New Name',
      inputValue: target.displayName || target.name,
      placeholder: 'Enter new name...',
      type: 'rename',
      target: target
    }
  }
}

const confirmModal = async () => {
  if (!modal.value.inputValue.trim()) return
  const val = modal.value.inputValue.trim()
  modal.value.show = false
  isPageLoading.value = true
  try {
    if (modal.value.type === 'folder') {
      await createFolder(val, currentFolderID.value)
    } else if (modal.value.type === 'rename') {
      await renameKnowledge(modal.value.target.id, val)
      // 更新预览标签中的名称
      const tabIndex = previewTabs.value.findIndex(t => t.file.id === modal.value.target.id)
      if (tabIndex !== -1) {
        previewTabs.value[tabIndex].file.displayName = val
      }
    }
    loadData()
  } catch (err) {
    console.error('Operation failed:', err)
  } finally {
    isPageLoading.value = false
  }
}

const openMoveModal = async () => {
  contextMenu.value.visible = false
  moveModal.value.target = contextMenu.value.target
  moveModal.value.targetFolderID = currentFolderID.value
  moveModal.value.show = true
  
  // 获取所有文件夹以便移动
  try {
    const res = await listFolders(0) // 获取根目录下的文件夹作为一级
    allFoldersList.value = res.data || res
  } catch (err) {
    console.error('获取文件夹列表失败', err)
  }
}

const confirmMove = async () => {
  if (!moveModal.value.target) return
  const fileID = moveModal.value.target.id
  const targetID = moveModal.value.targetFolderID
  
  moveModal.value.show = false
  isPageLoading.value = true
  try {
    await moveKnowledge(fileID, targetID)
    loadData()
  } catch (err) {
    console.error('移动失败', err)
  } finally {
    isPageLoading.value = false
  }
}

// 加载数据
const loadData = async () => {
  isPageLoading.value = true
  try {
    const [foldersRes, filesRes] = await Promise.all([
      listFolders(currentFolderID.value),
      listKnowledge(currentFolderID.value)
    ])
    folders.value = foldersRes.data || foldersRes 
    allFiles.value = filesRes.data || filesRes
  } catch (err) {
    console.error('加载失败:', err)
  } finally {
    isPageLoading.value = false
  }
}

const formatSize = (bytes) => {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const openFolder = (folder) => {
  currentFolderID.value = folder.id
  currentFolderName.value = folder.name
  loadData()
}

const goBack = () => {
  currentFolderID.value = 0
  currentFolderName.value = ''
  loadData()
}

const previewFile = async (file) => {
  // 检查是否已经打开
  const existingIndex = previewTabs.value.findIndex(tab => tab.file.id === file.id)
  if (existingIndex !== -1) {
    activeTabIndex.value = existingIndex
    return
  }

  // 新增标签
  const newTab = {
    file: file,
    content: '<div class="loading-preview">正在加载预览...</div>'
  }
  previewTabs.value.push(newTab)
  activeTabIndex.value = previewTabs.value.length - 1

  try {
    const content = await getKnowledgePreview(file.name)
    newTab.content = md.render(content || '# 文件内容为空')
  } catch (err) {
    newTab.content = '<div class="error-preview">内容加载失败</div>'
  }
}

const closeTab = (index) => {
  previewTabs.value.splice(index, 1)
  if (activeTabIndex.value >= previewTabs.value.length) {
    activeTabIndex.value = previewTabs.value.length - 1
  }
}

const closeAllTabs = () => {
  previewTabs.value = []
  activeTabIndex.value = -1
}

const showContextMenu = (e, target, type) => {
  contextMenu.value = {
    visible: true,
    x: e.clientX,
    y: e.clientY,
    target,
    type
  }
}

const handleDelete = async () => {
  const target = contextMenu.value.target
  if (confirm(`确定要彻底删除 ${target.displayName || target.name} 吗？`)) {
    let success = false
    if (contextMenu.value.type === 'file') {
      success = await deleteKnowledge(target.id)
    } else {
      success = await deleteFolder(target.id)
    }
    if (success) loadData()
  }
  contextMenu.value.visible = false
}

const triggerFileUpload = () => {
  fileInput.value.click()
  showNewMenu.value = false
}

const onFileSelected = async (e) => {
  const file = e.target.files[0]
  if (!file) return
  
  let targetFolderID = currentFolderID.value
  
  // 默认文件夹逻辑：如果在根目录，自动归类到“未分类”
  if (targetFolderID === 0) {
    try {
      isPageLoading.value = true
      const foldersRes = await listFolders(0)
      const folderList = foldersRes.data || foldersRes
      let uncategorizedFolder = folderList.find(f => f.name === '未分类')
      
      if (!uncategorizedFolder) {
        uncategorizedFolder = await createFolder('未分类', 0)
      }
      
      if (uncategorizedFolder) {
        targetFolderID = uncategorizedFolder.id
      }
    } catch (err) {
      console.error('获取或创建默认文件夹失败', err)
    } finally {
      isPageLoading.value = false
    }
  }

  const formData = new FormData()
  formData.append('file', file)
  formData.append('folderID', targetFolderID)
  try {
    isPageLoading.value = true
    await uploadKnowledge(formData)
    loadData()
    // 如果之前是在根目录上传，跳转到未分类文件夹
    if (currentFolderID.value === 0 && targetFolderID !== 0) {
      currentFolderID.value = targetFolderID
      currentFolderName.value = '未分类'
      loadData()
    }
  } catch (err) {
    alert('上传失败')
  } finally {
    isPageLoading.value = false
    e.target.value = ''
  }
}

const closeMenus = () => {
  showNewMenu.value = false
  contextMenu.value.visible = false
}

onMounted(() => {
  loadData()
  window.addEventListener('click', closeMenus)
})

onUnmounted(() => {
  window.removeEventListener('click', closeMenus)
})
</script>

<style scoped>
.knowledge-base-container {
  min-height: 100vh;
  background-color: #fefaf5; /* 浅橘黄色背景 */
  color: #1a1a1a;
  display: flex;
  flex-direction: column;
}

/* 顶部导航栏 (同步 HomeView) */
.kb-navbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 40px;
  height: 70px;
  background: white;
  border-bottom: 1px solid #ffe4cc;
  position: sticky;
  top: 0;
  z-index: 1000;
}

.nav-brand {
  font-size: 24px;
  font-weight: 800;
  color: #eeaa67;
  cursor: pointer;
}

.nav-links {
  display: flex;
  gap: 32px;
}

.nav-link {
  text-decoration: none;
  color: #4a4a4a;
  font-weight: 500;
  cursor: pointer;
  position: relative;
}

.nav-link.active {
  color: #eeaa67;
}

.nav-link.active::after {
  content: '';
  position: absolute;
  bottom: -8px;
  left: 0;
  width: 100%;
  height: 2px;
  background: #eeaa67;
}

.user-avatar-container {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-name {
  font-weight: 500;
  color: #666;
  font-size: 14px;
}

.user-avatar-box {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  overflow: hidden;
  border: 2px solid #fff;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}

.user-avatar-box img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

/* 内容布局 */
.kb-content {
  flex: 1;
  max-width: 1200px;
  width: 100%;
  margin: 0 auto;
  padding: 40px 20px;
  transition: margin-right 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.knowledge-base-container {
  display: flex;
  flex-direction: column;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

/* 文件夹网格 */
.kb-section {
  margin-bottom: 40px;
}

.section-title {
  font-size: 18px;
  font-weight: 700;
  margin-bottom: 20px;
  color: #111;
}

.folders-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 20px;
}

.folder-card {
  background: white;
  padding: 20px;
  border-radius: 16px;
  border: 1px solid #ffe4cc;
  display: flex;
  align-items: center;
  gap: 16px;
  cursor: pointer;
  transition: all 0.2s;
}

.folder-card:hover {
  border-color: #eeaa67;
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(238, 170, 103, 0.1);
}

.folder-name-text {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 4px;
}

.folder-stats-text {
  font-size: 12px;
  color: #9ca3af;
}

/* 最近文件卡片 */
.recent-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 20px;
}

.file-card-box {
  background: white;
  padding: 16px;
  border-radius: 16px;
  border: 1px solid #ffe4cc;
  display: flex;
  align-items: center;
  gap: 16px;
  cursor: pointer;
  transition: all 0.2s;
}

.file-card-box:hover {
  border-color: #eeaa67;
  transform: translateY(-2px);
}

.file-card-box.active {
  border-color: #eeaa67;
  background: #fff5eb;
}

.file-icon-square {
  width: 44px;
  height: 44px;
  background: #fff5eb;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.file-content-info {
  flex: 1;
  overflow: hidden;
}

.file-display-name {
  font-size: 14px;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-bottom: 4px;
}

.file-display-meta {
  font-size: 12px;
  color: #9ca3af;
}

/* 表格样式 */
.files-table-container {
  background: white;
  border-radius: 16px;
  border: 1px solid #ffe4cc;
  overflow: hidden;
}

.files-data-table {
  width: 100%;
  border-collapse: collapse;
}

.files-data-table th {
  text-align: left;
  padding: 14px 20px;
  background: #fffaf5;
  font-size: 12px;
  color: #eeaa67;
  font-weight: 600;
}

.files-data-table tr {
  cursor: pointer;
  transition: background 0.2s;
}

.files-data-table tr:hover {
  background: #fffaf5;
}

.files-data-table tr.active {
  background: #fff5eb;
}

.files-data-table td {
  padding: 14px 20px;
  border-top: 1px solid #fffaf5;
  font-size: 14px;
}

.table-file-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.table-file-icon {
  color: #eeaa67;
}

/* 模态框样式 */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.4);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 2000;
}

.tomato-modal {
  background: white;
  width: 380px;
  border-radius: 24px;
  padding: 24px;
  box-shadow: 0 20px 60px rgba(0,0,0,0.15);
  border: 1px solid #ffe4cc;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.modal-header h3 {
  font-size: 20px;
  font-weight: 700;
  color: #333;
}

.close-modal-btn {
  background: #f8f9fa;
  border: none;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #999;
  transition: all 0.2s;
}

.close-modal-btn:hover {
  background: #fff5eb;
  color: #eeaa67;
}

.modal-body label {
  display: block;
  font-size: 14px;
  font-weight: 600;
  color: #666;
  margin-bottom: 8px;
}

.modal-body input {
  width: 100%;
  padding: 12px 16px;
  border: 2px solid #f0f0f0;
  border-radius: 12px;
  outline: none;
  font-size: 15px;
  transition: all 0.2s;
}

.modal-body input:focus {
  border-color: #eeaa67;
  background: #fffaf5;
}

.modal-footer {
  display: flex;
  gap: 12px;
  margin-top: 30px;
}

.btn-cancel-modal {
  flex: 1;
  background: #f8f9fa;
  color: #666;
  padding: 12px;
  border-radius: 12px;
  border: none;
  font-weight: 600;
  cursor: pointer;
}

.btn-confirm-modal {
  flex: 1;
  padding: 12px;
  border-radius: 12px;
  font-weight: 600;
  cursor: pointer;
}

/* 移动文件列表 */
.folder-list-select {
  max-height: 240px;
  overflow-y: auto;
  border: 1px solid #eee;
  border-radius: 12px;
  margin-top: 12px;
}

.folder-select-item {
  padding: 12px 16px;
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  transition: background 0.2s;
}

.folder-select-item:hover {
  background: #fffaf5;
}

.folder-select-item.selected {
  background: #fff5eb;
  color: #eeaa67;
  font-weight: 600;
}

/* 右侧预览面板 */
.preview-side-panel {
  position: fixed;
  top: 70px;
  right: -400px;
  width: 400px;
  height: calc(100vh - 70px);
  background: white;
  box-shadow: -4px 0 20px rgba(0,0,0,0.05);
  border-left: 1px solid #ffe4cc;
  transition: right 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  z-index: 900;
  display: flex;
  flex-direction: column;
}

.preview-side-panel.open {
  right: 0;
}

.panel-header {
  padding: 10px 16px;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #fffaf5;
}

.tabs-container {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  padding-bottom: 4px;
}

.preview-tab {
  padding: 6px 12px;
  background: white;
  border: 1px solid #eee;
  border-radius: 8px;
  font-size: 13px;
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  white-space: nowrap;
  max-width: 150px;
}

.preview-tab.active {
  border-color: #eeaa67;
  background: #fff5eb;
  color: #eeaa67;
  font-weight: 600;
}

.tab-name {
  overflow: hidden;
  text-overflow: ellipsis;
}

.tab-close {
  color: #ccc;
  font-size: 16px;
}

.tab-close:hover {
  color: #ff6b6b;
}

.panel-close-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: #999;
}

.panel-content {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

.preview-empty {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #999;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
  opacity: 0.3;
}

/* 上下文菜单 */
.context-menu-box {
  position: fixed;
  background: white;
  border-radius: 12px;
  box-shadow: 0 8px 24px rgba(0,0,0,0.12);
  border: 1px solid #eee;
  padding: 6px;
  min-width: 160px;
  z-index: 3000;
}

.menu-item-row {
  padding: 10px 14px;
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  cursor: pointer;
  border-radius: 8px;
  transition: all 0.2s;
}

.menu-item-row:hover {
  background: #fff5eb;
  color: #eeaa67;
}

.delete-red:hover {
  background: #fff5f5;
  color: #ff4d4f;
}

.menu-divider-line {
  height: 1px;
  background: #f0f0f0;
  margin: 4px 0;
}

/* 下拉菜单 */
.tomato-dropdown {
  background: white;
  border-radius: 16px;
  border: 1px solid #ffe4cc;
  box-shadow: 0 10px 30px rgba(238, 170, 103, 0.15);
  overflow: hidden;
}

/* 加载动画 */
.kb-global-loader {
  position: fixed;
  inset: 0;
  background: rgba(255,255,255,0.7);
  backdrop-filter: blur(2px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
}

.kb-loader-spinner {
  width: 40px;
  height: 40px;
  border: 4px solid #fff5eb;
  border-top-color: #eeaa67;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Markdown 预览样式补丁 */
.markdown-body {
  font-size: 15px;
  line-height: 1.6;
}

.loading-preview, .error-preview {
  text-align: center;
  padding: 40px 0;
  color: #999;
}

/* 移动动画 */
.fade-enter-active, .fade-leave-active { transition: opacity 0.2s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }

.dropdown-enter-active, .dropdown-leave-active { transition: all 0.2s ease; }
.dropdown-enter-from, .dropdown-leave-to { opacity: 0; transform: translateY(-10px); }

/* 工具栏补丁 */
.search-box-container {
  flex: 1;
  max-width: 400px;
  margin-right: 20px;
}

.search-input-wrapper {
  position: relative;
  width: 100%;
}

.search-icon {
  position: absolute;
  left: 14px;
  top: 50%;
  transform: translateY(-50%);
  color: #9ca3af;
  pointer-events: none;
}

.search-input {
  width: 100%;
  padding: 10px 16px 10px 42px;
  background: white;
  border: 1.5px solid #ffe4cc;
  border-radius: 12px;
  outline: none;
  font-size: 14px;
  transition: all 0.2s;
}

.search-input:focus {
  border-color: #eeaa67;
  box-shadow: 0 0 0 4px rgba(238, 170, 103, 0.1);
  background: #fffaf5;
}

.new-btn-group {
  position: relative;
}

/* 下拉菜单重构 */
.tomato-dropdown {
  position: absolute;
  top: 100%;
  right: 0;
  min-width: 240px;
  padding: 8px;
  margin-top: 8px;
  background: white;
  border-radius: 16px;
  border: 1px solid #ffe4cc;
  box-shadow: 0 10px 30px rgba(238, 170, 103, 0.15);
  z-index: 1100;
  overflow: hidden;
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.dropdown-item:hover {
  background: #fff5eb;
}

.item-icon-circle {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.upload-bg { background: #4ecdc4; }
.folder-bg { background: #eeaa67; }

.item-text {
  display: flex;
  flex-direction: column;
}

.item-title {
  font-size: 14px;
  font-weight: 600;
  color: #1a1a1a;
}

.item-desc {
  font-size: 11px;
  color: #9ca3af;
  margin-top: 2px;
}

/* 现代空状态 */
.empty-state-modern {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 20px;
  text-align: center;
}

.empty-visual {
  position: relative;
  margin-bottom: 24px;
}

.empty-bg-circle {
  width: 120px;
  height: 120px;
  background: #fff5eb;
  border-radius: 50%;
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  z-index: 0;
}

.empty-icon-main {
  position: relative;
  z-index: 1;
}

.empty-text-box h3 {
  font-size: 20px;
  font-weight: 700;
  margin-bottom: 8px;
  color: #111;
}

.empty-text-box p {
  font-size: 14px;
  color: #6b7280;
  max-width: 300px;
  line-height: 1.5;
  margin-bottom: 24px;
}

.empty-action-btn {
  padding: 10px 24px;
  border-radius: 12px;
  font-weight: 600;
  cursor: pointer;
}

/* 针对文件夹内的返回按钮美化 */
.back-circle-btn {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  border: 1px solid #ffe4cc;
  background: white;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #eeaa67;
  transition: all 0.2s;
}

.back-circle-btn:hover {
  background: #fff5eb;
  transform: translateX(-4px);
  border-color: #eeaa67;
}

.page-title {
  font-size: 28px;
  font-weight: 700;
  color: #111;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
}

.tomato-btn {
  background: #eeaa67 !important;
  color: white !important;
  border: none !important;
  transition: all 0.2s ease;
  box-shadow: 0 4px 12px rgba(238, 170, 103, 0.2);
}

.tomato-btn:hover {
  background: #f39c12 !important;
  transform: translateY(-1px);
  box-shadow: 0 6px 16px rgba(238, 170, 103, 0.3);
}

.new-btn {
  padding: 10px 24px;
  border-radius: 12px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  border: none;
  font-family: inherit;
}

/* 侧边面板补丁 */
.preview-side-panel {
  position: fixed;
  top: 70px;
  height: calc(100vh - 70px);
  background: white;
  box-shadow: -4px 0 20px rgba(0,0,0,0.05);
  border-left: 1px solid #ffe4cc;
  transition: right 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  z-index: 900;
  display: flex;
  flex-direction: column;
}

.panel-resizer {
  position: absolute;
  left: -4px;
  top: 0;
  width: 8px;
  height: 100%;
  cursor: col-resize;
  z-index: 10;
  transition: background 0.2s;
}

.panel-resizer:hover {
  background: rgba(238, 170, 103, 0.2);
}
</style>
