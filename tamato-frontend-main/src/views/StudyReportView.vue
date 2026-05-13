<template>
  <div class="report-container">
    <nav class="navbar">
      <div class="nav-brand" @click="$router.push('/home')">Tomato</div>
      <div class="nav-title">学习报告成就看板</div>
      <div class="nav-actions">
        <button class="btn-back" @click="$router.push('/home')">返回主页</button>
      </div>
    </nav>

    <main class="report-content">
      <div v-if="loading" class="loading-state">
        <div class="loader"></div>
        <p>AI 教练正在深度分析您的学习成果...</p>
      </div>

      <div v-else-if="report" class="report-card animate-fade-in">
        <header class="report-header">
          <div class="report-date-badge">
            <span class="day">{{ formatDate(report.report_date, 'DD') }}</span>
            <span class="month">{{ formatDate(report.report_date, 'MMM') }}</span>
          </div>
          <div class="report-main-info">
            <h1>昨日学习日报</h1>
            <p class="subtitle">记录每一份努力，见证每一刻成长</p>
          </div>
          <div class="report-actions">
            <button class="btn-regenerate" @click="handleRegenerate" :disabled="regenerating">
              <span v-if="!regenerating">重新生成</span>
              <span v-else class="mini-loader"></span>
            </button>
          </div>
        </header>

        <section class="stats-grid">
          <div class="stat-item glass">
            <div class="stat-icon">⏱️</div>
            <div class="stat-value">{{ report.total_duration }} min</div>
            <div class="stat-label">专注时长</div>
          </div>
          <div class="stat-item glass">
            <div class="stat-icon">✅</div>
            <div class="stat-value">{{ report.completed_tasks }}</div>
            <div class="stat-label">完成任务</div>
          </div>
          <div class="stat-item glass">
            <div class="stat-icon">🔥</div>
            <div class="stat-value">{{ report.session_count }}</div>
            <div class="stat-label">专注次数</div>
          </div>
          <div class="stat-item glass">
            <div class="stat-icon">📊</div>
            <div class="stat-value">{{ report.average_duration }} min</div>
            <div class="stat-label">平均时长</div>
          </div>
        </section>

        <section class="ai-summary glass">
          <div class="summary-header">
            <span class="ai-icon">🤖</span>
            <h2>AI 教练深度点评</h2>
          </div>
          <div class="markdown-content" v-html="renderedContent"></div>
        </section>
      </div>

      <div v-else class="empty-state glass">
        <div class="empty-icon">📂</div>
        <h3>暂无报告数据</h3>
        <p>您昨天好像没有开启过专注会话哦，快去开启一段专注之旅吧！</p>
        <div class="empty-actions">
          <button class="btn-primary" @click="$router.push('/home')">前往专注</button>
          <button class="btn-secondary" @click="handleRegenerate" :disabled="regenerating">
            <span v-if="!regenerating">尝试生成报告</span>
            <span v-else class="mini-loader"></span>
          </button>
        </div>
      </div>
    </main>
  </div>
</template>

<script>
import { getDailyReport, regenerateDailyReport } from '@/api/reports'
import MarkdownIt from 'markdown-it'

const md = new MarkdownIt()

export default {
  name: 'StudyReportView',
  data() {
    return {
      report: null,
      loading: true,
      regenerating: false
    }
  },
  computed: {
    renderedContent() {
      if (!this.report || !this.report.content) return ''
      return md.render(this.report.content)
    }
  },
  async created() {
    await this.fetchReport()
  },
  methods: {
    async fetchReport() {
      this.loading = true
      try {
        const res = await getDailyReport()
        if (res.success) {
          this.report = res.data
        }
      } catch (err) {
        console.error('获取报告失败:', err)
      } finally {
        this.loading = false
      }
    },
    async handleRegenerate() {
      if (this.regenerating) return
      this.regenerating = true
      try {
        const res = await regenerateDailyReport()
        if (res.success) {
          this.report = res.data
        }
      } catch (err) {
        alert('生成失败，请稍后重试')
      } finally {
        this.regenerating = false
      }
    },
    formatDate(dateStr, part) {
      if (!dateStr) return ''
      const date = new Date(dateStr)
      if (part === 'DD') {
        return date.getDate().toString().padStart(2, '0')
      }
      if (part === 'MMM') {
        const months = ['JAN', 'FEB', 'MAR', 'APR', 'MAY', 'JUN', 'JUL', 'AUG', 'SEP', 'OCT', 'NOV', 'DEC']
        return months[date.getMonth()]
      }
      return date.toLocaleDateString()
    }
  }
}
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Outfit:wght@300;400;600;700&display=swap');

.report-container {
  min-height: 100vh;
  background: linear-gradient(135deg, #fdfbfb 0%, #ebedee 100%);
  font-family: 'Outfit', sans-serif;
  color: #2d3436;
}

.navbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem 5%;
  background: rgba(255, 255, 255, 0.7);
  backdrop-filter: blur(10px);
  position: sticky;
  top: 0;
  z-index: 100;
  border-bottom: 1px solid rgba(255, 255, 255, 0.3);
}

.nav-brand {
  font-size: 1.8rem;
  font-weight: 700;
  color: #eeaa67;
  cursor: pointer;
}

.nav-title {
  font-size: 1.2rem;
  font-weight: 600;
  color: #636e72;
}

.report-content {
  max-width: 1000px;
  margin: 3rem auto;
  padding: 0 1.5rem;
}

.loading-state {
  text-align: center;
  margin-top: 10rem;
}

.loader {
  width: 50px;
  height: 50px;
  border: 5px solid #f3f3f3;
  border-top: 5px solid #eeaa67;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 1.5rem;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.report-card {
  display: flex;
  flex-direction: column;
  gap: 2.5rem;
}

.report-header {
  display: flex;
  align-items: center;
  gap: 2rem;
}

.report-date-badge {
  background: #eeaa67;
  color: white;
  padding: 1rem;
  border-radius: 15px;
  display: flex;
  flex-direction: column;
  align-items: center;
  min-width: 80px;
  box-shadow: 0 10px 20px rgba(238, 170, 103, 0.3);
}

.day { font-size: 1.8rem; font-weight: 700; }
.month { font-size: 0.9rem; text-transform: uppercase; font-weight: 600; opacity: 0.9; }

.report-main-info h1 {
  font-size: 2.5rem;
  font-weight: 700;
  margin: 0;
  background: linear-gradient(to right, #2d3436, #eeaa67);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.subtitle {
  color: #636e72;
  margin-top: 0.5rem;
  font-size: 1.1rem;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
}

.glass {
  background: rgba(255, 255, 255, 0.6);
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 24px;
  box-shadow: 0 8px 32px 0 rgba(31, 38, 135, 0.05);
}

.stat-item {
  padding: 2rem;
  text-align: center;
  transition: transform 0.3s ease;
}

.stat-item:hover {
  transform: translateY(-5px);
}

.stat-icon { font-size: 2rem; margin-bottom: 1rem; }
.stat-value { font-size: 1.8rem; font-weight: 700; color: #2d3436; }
.stat-label { font-size: 0.9rem; color: #636e72; font-weight: 600; margin-top: 0.5rem; }

.ai-summary {
  padding: 3rem;
}

.summary-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 2rem;
}

.ai-icon { font-size: 2.5rem; }
.summary-header h2 { font-size: 1.8rem; font-weight: 700; color: #2d3436; margin: 0; }

.markdown-content {
  line-height: 1.8;
  color: #2d3436;
  font-size: 1.1rem;
}

.markdown-content :deep(h1), .markdown-content :deep(h2), .markdown-content :deep(h3) {
  margin-top: 2rem;
  color: #eeaa67;
}

.markdown-content :deep(p) {
  margin-bottom: 1.2rem;
}

.markdown-content :deep(ul), .markdown-content :deep(ol) {
  padding-left: 1.5rem;
  margin-bottom: 1.5rem;
}

.markdown-content :deep(li) {
  margin-bottom: 0.8rem;
}

.btn-regenerate {
  background: transparent;
  border: 2px solid #eeaa67;
  color: #eeaa67;
  padding: 0.8rem 1.5rem;
  border-radius: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
}

.btn-regenerate:hover {
  background: #eeaa67;
  color: white;
  box-shadow: 0 5px 15px rgba(238, 170, 103, 0.3);
}

.btn-regenerate:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.empty-state {
  text-align: center;
  padding: 5rem 2rem;
}

.empty-icon { font-size: 4rem; margin-bottom: 2rem; }
.empty-state h3 { font-size: 1.5rem; margin-bottom: 1rem; }
.empty-state p { color: #636e72; margin-bottom: 2rem; }

.empty-actions {
  display: flex;
  justify-content: center;
  gap: 1.5rem;
  margin-top: 2rem;
}

.btn-primary {
  background: #eeaa67;
  border: none;
  color: white;
  padding: 1rem 2.5rem;
  border-radius: 15px;
  font-size: 1.1rem;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.3s;
}

.btn-primary:hover {
  transform: scale(1.05);
  box-shadow: 0 10px 25px rgba(238, 170, 103, 0.4);
}

.btn-secondary {
  background: transparent;
  border: 2px solid #eeaa67;
  color: #eeaa67;
  padding: 1rem 2.5rem;
  border-radius: 15px;
  font-size: 1.1rem;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.3s;
}

.btn-secondary:hover {
  background: rgba(238, 170, 103, 0.1);
  transform: scale(1.05);
}

.mini-loader {
  width: 15px;
  height: 15px;
  border: 2px solid #eeaa67;
  border-top: 2px solid white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  display: inline-block;
}

.animate-fade-in {
  animation: fadeIn 0.8s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
