<!-- src/components/MusicPlayerSimple.vue -->
<template>
  <div class="music-player-simple">
    <div class="player-title">
      <span>背景音乐: {{ musicName }}</span>
      <span v-if="isPlaying" class="playing-indicator">循环播放中...</span>
    </div>
    
    <div class="player-controls">
      <button @click="togglePlay" class="play-btn">
        {{ isPlaying ? '⏸ 暂停' : '▶ 播放' }}
      </button>
      <button @click="stop" class="stop-btn">⏹ 停止</button>
    </div>
  </div>
</template>

<script>
export default {
  name: 'MusicPlayerSimple',
  props: {
    musicName: {
      type: String,
      required: true
    }
  },
  data() {
    return {
      audio: null,
      isPlaying: false
    };
  },
  watch: {
    musicName(newName) {
      if (newName) {
        this.loadMusic();
      }
    }
  },
  mounted() {
    this.loadMusic();
  },
  beforeUnmount() {
    if (this.audio) {
      this.audio.pause();
      this.audio = null;
    }
  },
  methods: {
    loadMusic() {
      // 停止当前播放
      if (this.audio) {
        this.audio.pause();
        this.audio = null;
        this.isPlaying = false;
      }
      
      // 创建新的音频对象
      this.audio = new Audio();
      
      // 设置自动循环
      this.audio.loop = true; // 关键：设置循环播放
      
      // 根据音乐名称设置音频文件路径
      const fileName = `${this.musicName}.mp3`;
      
      try {
        // 使用 require 动态导入音乐文件
        const musicPath = require(`@/assets/music/${fileName}`);
        this.audio.src = musicPath;
        this.audio.volume = 0.5; // 设置默认音量
        
        console.log('音乐加载成功:', fileName);
      } catch (error) {
        console.error('加载音乐失败:', error);
      }
    },
    
    togglePlay() {
      if (!this.audio || !this.audio.src) return;
      
      if (this.isPlaying) {
        this.audio.pause();
      } else {
        this.audio.play().catch(error => {
          console.error('播放失败:', error);
        });
      }
      this.isPlaying = !this.isPlaying;
    },
    
    stop() {
      if (this.audio) {
        this.audio.pause();
        this.audio.currentTime = 0;
        this.isPlaying = false;
      }
    }
  }
};
</script>



<style scoped>
.music-player-simple {
  background: white;
  border-radius: 10px;
  padding: 15px;
  margin-top: 20px;
  border: 1px solid #e9ecef;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}

.player-title {
  margin-bottom: 12px;
  font-weight: 500;
  color: #333;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.playing-indicator {
  font-size: 0.85em;
  color: #eeaa67;
}

.player-controls {
  display: flex;
  gap: 8px; /* 减小间距 */
}

.play-btn, .stop-btn {
  padding: 8px 12px; /* 减小水平内边距 */
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s ease;
  width: 70px; /* 固定宽度 */
}

.play-btn {
  background: #eeaa67;
  color: white;
}

.play-btn:hover {
  background: #e69c55;
}

.stop-btn {
  background: #f8f9fa;
  color: #666;
  border: 1px solid #e0e0e0;
}

.stop-btn:hover {
  background: #e9ecef;
}
</style>