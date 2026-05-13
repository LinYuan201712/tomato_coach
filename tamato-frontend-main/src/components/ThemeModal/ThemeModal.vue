<template>
  <!-- 主题设置弹窗 -->
  <div v-if="visible" class="theme-modal-overlay" @click.self="handleClose">
    <div class="theme-modal">
      <!-- 弹窗头部 -->
      <div class="modal-header">
        <h3>主题设置</h3>
        <button class="close-btn" @click="handleClose" aria-label="关闭">
          <span>×</span>
        </button>
      </div>
      
      <!-- 弹窗内容 -->
      <div class="modal-content">
        <p class="modal-description">选择你喜欢的界面主题</p>
        
        <div class="theme-options">
          <!-- 默认主题 -->
          <div 
            class="theme-option"
            :class="{ 'theme-option--active': selectedTheme === 'default' }"
            @click="handleSelect('default')"
          >
            <div class="theme-option__preview theme-option__preview--default">
              <div class="preview-header"></div>
              <div class="preview-body"></div>
              <div class="preview-footer"></div>
            </div>
            <div class="theme-option__info">
              <h4>默认主题</h4>
              <p>明亮的橙色系</p>
            </div>
            <div class="theme-option__check" v-if="selectedTheme === 'default'">
              ✓
            </div>
          </div>
          
          <!-- 暖色系 -->
          <div 
            class="theme-option"
            :class="{ 'theme-option--active': selectedTheme === 'warm' }"
            @click="handleSelect('warm')"
          >
            <div class="theme-option__preview theme-option__preview--warm">
              <div class="preview-header"></div>
              <div class="preview-body"></div>
              <div class="preview-footer"></div>
            </div>
            <div class="theme-option__info">
              <h4>暖色系</h4>
              <p>温暖的红色调</p>
            </div>
            <div class="theme-option__check" v-if="selectedTheme === 'warm'">
              ✓
            </div>
          </div>
          
          <!-- 深色模式 -->
          <div 
            class="theme-option"
            :class="{ 'theme-option--active': selectedTheme === 'dark' }"
            @click="handleSelect('dark')"
          >
            <div class="theme-option__preview theme-option__preview--dark">
              <div class="preview-header"></div>
              <div class="preview-body"></div>
              <div class="preview-footer"></div>
            </div>
            <div class="theme-option__info">
              <h4>深色模式</h4>
              <p>夜间友好，护眼</p>
            </div>
            <div class="theme-option__check" v-if="selectedTheme === 'dark'">
              ✓
            </div>
          </div>
        </div>
      </div>
      
      <!-- 弹窗底部 -->
      <div class="modal-footer">
        <button class="confirm-btn" @click="handleClose">
          完成
        </button>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'ThemeModal',
  
  props: {
    visible: {
      type: Boolean,
      default: false
    },
    currentTheme: {
      type: String,
      default: 'default'
    }
  },
  
  data() {
    return {
      localSelectedTheme: this.currentTheme
    };
  },
  
  computed: {
    selectedTheme() {
      return this.localSelectedTheme;
    }
  },
  
  watch: {
    currentTheme(newVal) {
      this.localSelectedTheme = newVal;
    }
  },
  
  methods: {
    handleSelect(theme) {
      this.localSelectedTheme = theme;
      this.$emit('theme-select', theme);
    },
    
    handleClose() {
      this.$emit('update:visible', false);
    }
  }
};
</script>

<style scoped>
/* 弹窗遮罩层 */
.theme-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10000;
  backdrop-filter: blur(2px);
  animation: modal-fade-in 0.2s ease;
}

@keyframes modal-fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

/* 弹窗主体 */
.theme-modal {
  background-color: white;
  border-radius: 16px;
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  animation: modal-slide-up 0.3s ease;
}

@keyframes modal-slide-up {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 弹窗头部 */
.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 24px 30px;
  border-bottom: 1px solid #eee;
}

.modal-header h3 {
  font-size: 1.4em;
  color: #333;
  margin: 0;
  font-weight: 600;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.8em;
  color: #999;
  cursor: pointer;
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: all 0.2s ease;
}

.close-btn:hover {
  background-color: #f5f5f5;
  color: #666;
}

.close-btn span {
  line-height: 1;
}

/* 弹窗内容 */
.modal-content {
  padding: 30px;
}

.modal-description {
  color: #666;
  text-align: center;
  margin-bottom: 25px;
  font-size: 0.95em;
}

/* 主题选项 */
.theme-options {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 20px;
  margin-bottom: 20px;
}

.theme-option {
  position: relative;
  border: 2px solid #e0e0e0;
  border-radius: 12px;
  padding: 15px;
  cursor: pointer;
  transition: all 0.3s ease;
  background-color: white;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.theme-option:hover {
  border-color: #eeaa67;
  transform: translateY(-3px);
  box-shadow: 0 8px 16px rgba(238, 170, 103, 0.1);
}

.theme-option--active {
  border-color: #eeaa67;
  background-color: #fffaf5;
}

/* 主题预览 */
.theme-option__preview {
  width: 100%;
  height: 80px;
  border-radius: 8px;
  overflow: hidden;
  margin-bottom: 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 6px;
}

.preview-header {
  height: 20%;
  border-radius: 4px;
}

.preview-body {
  flex: 1;
  border-radius: 4px;
}

.preview-footer {
  height: 15%;
  border-radius: 4px;
}

/* 不同主题的预览颜色 */
.theme-option__preview--default .preview-header {
  background-color: #eeaa67;
}

.theme-option__preview--default .preview-body {
  background-color: #fefaf5;
}

.theme-option__preview--default .preview-footer {
  background-color: #ffe4cc;
}

.theme-option__preview--warm .preview-header {
  background-color: #ff6b6b;
}

.theme-option__preview--warm .preview-body {
  background-color: #fff5f5;
}

.theme-option__preview--warm .preview-footer {
  background-color: #ffc9c9;
}

.theme-option__preview--dark .preview-header {
  background-color: #333;
}

.theme-option__preview--dark .preview-body {
  background-color: #1a1a1a;
}

.theme-option__preview--dark .preview-footer {
  background-color: #404040;
}

/* 主题信息 */
.theme-option__info {
  text-align: center;
}

.theme-option__info h4 {
  margin: 0 0 4px 0;
  font-size: 1em;
  color: #333;
  font-weight: 600;
}

.theme-option__info p {
  margin: 0;
  font-size: 0.85em;
  color: #666;
  line-height: 1.4;
}

/* 选中标记 */
.theme-option__check {
  position: absolute;
  top: 10px;
  right: 10px;
  width: 24px;
  height: 24px;
  background-color: #eeaa67;
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  font-size: 0.9em;
}

/* 弹窗底部 */
.modal-footer {
  padding: 20px 30px;
  background-color: #f9fafb;
  border-top: 1px solid #eee;
  border-radius: 0 0 16px 16px;
  text-align: center;
}

.confirm-btn {
  background-color: #eeaa67;
  color: white;
  border: none;
  padding: 12px 40px;
  border-radius: 8px;
  font-size: 1em;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
}

.confirm-btn:hover {
  background-color: #e69c55;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(238, 170, 103, 0.3);
}

.confirm-btn:active {
  transform: translateY(0);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .theme-options {
    grid-template-columns: 1fr;
    max-width: 280px;
    margin-left: auto;
    margin-right: auto;
  }
  
  .theme-modal {
    width: 95%;
    max-height: 85vh;
  }
  
  .modal-header {
    padding: 20px;
  }
  
  .modal-content {
    padding: 20px;
  }
  
  .modal-footer {
    padding: 15px 20px;
  }
}
</style>