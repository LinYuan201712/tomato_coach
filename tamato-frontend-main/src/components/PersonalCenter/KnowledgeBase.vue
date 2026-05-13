<template>
  <div class="knowledge-base">
    <div class="upload-section">
      <div class="upload-card" @click="triggerUpload">
        <div class="upload-icon">📤</div>
        <div class="upload-text">点击上传课件或学习资料</div>
        <div class="upload-tip">支持 PDF, Word, Markdown (最大 10MB)</div>
        <input 
          type="file" 
          ref="fileInput" 
          style="display: none" 
          @change="handleFileChange"
          accept=".pdf,.doc,.docx,.md,.txt"
        >
      </div>
    </div>

    <div class="document-list-section">
      <div class="section-header">
        <h3>已上传文档 ({{ documents.length }})</h3>
        <button class="refresh-btn" @click="fetchDocuments" :disabled="loading">刷新</button>
      </div>

      <div v-if="loading" class="loading-state">正在加载文档列表...</div>
      
      <div v-else-if="documents.length === 0" class="empty-state">
        <div class="empty-icon">📂</div>
        <p>暂无文档，上传一些资料来增强你的 AI 助手吧！</p>
      </div>

      <div v-else class="document-grid">
        <div v-for="doc in documents" :key="doc.id" class="document-card">
          <div class="doc-icon">{{ getFileIcon(doc.name) }}</div>
          <div class="doc-info">
            <div class="doc-name" :title="doc.name">{{ doc.name }}</div>
            <div class="doc-meta">{{ formatSize(doc.size) }} · {{ formatDate(doc.uploadTime) }}</div>
          </div>
          <button class="delete-btn" @click="handleDelete(doc.id)" title="删除">🗑️</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { uploadDocument, getDocumentList, deleteDocument } from '@/api/knowledge'

export default {
  name: 'KnowledgeBase',
  data() {
    return {
      documents: [],
      loading: false
    }
  },
  mounted() {
    this.fetchDocuments()
  },
  methods: {
    async fetchDocuments() {
      this.loading = true
      try {
        const res = await getDocumentList()
        console.log('📄 文档列表响应:', res)
        if (res.success) {
          this.documents = res.data || []
        } else {
          console.error('获取文档列表失败:', res.message)
        }
      } catch (err) {
        console.error('获取文档列表出错:', err)
      } finally {
        this.loading = false
      }
    },
    triggerUpload() {
      this.$refs.fileInput.click()
    },
    async handleFileChange(e) {
      const file = e.target.files[0]
      if (!file) return

      const formData = new FormData()
      formData.append('file', file)

      const loadingMsg = this.$message ? this.$message.loading('正在上传并解析文档...', 0) : null
      
      try {
        const res = await uploadDocument(formData)
        if (res.success) {
          if (this.$message) {
            this.$message.success('上传成功！AI 已学习该文档。')
          } else {
            alert('上传成功！AI 已学习该文档。')
          }
          this.fetchDocuments()
        } else {
          const msg = res.message || '上传失败'
          if (this.$message) {
            this.$message.error(msg)
          } else {
            alert(msg)
          }
        }
      } catch (err) {
        console.error('上传过程出错:', err)
        if (this.$message) {
          this.$message.error('上传出错，请稍后再试: ' + err.message)
        } else {
          alert('上传出错，请稍后再试: ' + err.message)
        }
      } finally {
        if (loadingMsg) loadingMsg()
        e.target.value = ''
      }
    },
    async handleDelete(id) {
      if (!confirm('确定要删除该文档吗？这将从 AI 的知识库中移除相关内容。')) return

      try {
        const res = await deleteDocument(id)
        if (res.success) {
          this.fetchDocuments()
        }
      } catch (err) {
        console.error('删除失败:', err)
      }
    },
    getFileIcon(name) {
      const ext = name.split('.').pop().toLowerCase()
      if (ext === 'pdf') return '📕'
      if (['doc', 'docx'].includes(ext)) return '📘'
      if (ext === 'md') return '📝'
      return '📄'
    },
    formatSize(bytes) {
      if (!bytes) return '0 B'
      const k = 1024
      const sizes = ['B', 'KB', 'MB', 'GB']
      const i = Math.floor(Math.log(bytes) / Math.log(k))
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
    },
    formatDate(timestamp) {
      if (!timestamp) return ''
      return new Date(timestamp).toLocaleDateString()
    }
  }
}
</script>

<style scoped>
.knowledge-base {
  padding: 10px;
}

.upload-section {
  margin-bottom: 30px;
}

.upload-card {
  border: 2px dashed #e0e0e0;
  border-radius: 12px;
  padding: 40px;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s;
  background: #fafafa;
}

.upload-card:hover {
  border-color: #eeaa67;
  background: #fffefb;
}

.upload-icon {
  font-size: 40px;
  margin-bottom: 15px;
}

.upload-text {
  font-size: 18px;
  font-weight: 600;
  color: #333;
  margin-bottom: 8px;
}

.upload-tip {
  font-size: 14px;
  color: #999;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.refresh-btn {
  background: none;
  border: 1px solid #ddd;
  padding: 4px 12px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
}

.document-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.document-card {
  display: flex;
  align-items: center;
  padding: 16px;
  background: white;
  border: 1px solid #eee;
  border-radius: 8px;
  transition: box-shadow 0.2s;
}

.document-card:hover {
  box-shadow: 0 4px 12px rgba(0,0,0,0.05);
}

.doc-icon {
  font-size: 28px;
  margin-right: 12px;
}

.doc-info {
  flex: 1;
  min-width: 0;
}

.doc-name {
  font-weight: 600;
  font-size: 14px;
  margin-bottom: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.doc-meta {
  font-size: 12px;
  color: #999;
}

.delete-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 8px;
  font-size: 16px;
  opacity: 0.3;
  transition: opacity 0.2s;
}

.document-card:hover .delete-btn {
  opacity: 1;
}

.delete-btn:hover {
  color: #f44336;
}

.empty-state, .loading-state {
  text-align: center;
  padding: 60px 20px;
  color: #999;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 15px;
  opacity: 0.5;
}
</style>
