<template>
  <div class="friends-list-container">
    <div class="friends-header">
      <h3>👥 好友列表</h3>
      <span class="online-count">
        共 {{ friendsList.length }} 位好友
      </span>
    </div>

    <div v-if="loading" class="status-message loading-state">
      <div class="spinner"></div>
      正在加载好友数据...
    </div>
    <div v-else-if="error" class="status-message error-state">
      {{ error }}
      <button @click="fetchFriends" class="btn-retry">重试</button>
    </div>

    <div v-else class="friends-scroll-area">
      <div v-if="friendsList.length === 0" class="status-message empty-state">
        暂无好友，快去添加吧！
      </div>

      <template v-else>
        <div v-for="friend in friendsList" :key="friend.user_id">
          
          <div 
            :class="['friend-item', { 'is-expanded': expandedUserId === friend.user_id }]"
            @click="toggleDetails(friend.user_id)"
          >
            <div class="friend-avatar">
              <img 
                :src="getAvatarUrl(friend.avatar)" 
                alt="好友头像"
                @error="handleAvatarError"
              />
            </div>
            
            <div class="friend-name-wrapper">
              <div class="friend-name">{{ friend.username }}</div>
            </div>
            
            <div :class="['friend-status', friend.statusClass]">
              <span class="status-dot"></span>
              <span class="status-text">{{ friend.status || '未知状态' }}</span>
            </div>
          </div>

          <div v-if="expandedUserId === friend.user_id" class="friend-details-card">
            <div class="detail-row">
              <label>ID:</label>
              <span>{{ friend.user_id }}</span>
            </div>
            <div class="detail-row">
              <label>生日:</label>
              <span>{{ friend.birthday ? formatBirthday(friend.birthday) : '未知' }}</span>
            </div>
            <div class="detail-row">
              <label>地区:</label>
              <span>{{ friend.province || '未知' }}</span>
            </div>
            <div class="detail-row">
              <label>专注:</label>
              <span>{{ friend.tomato || 0 }} 个番茄</span>
            </div>
          </div>
        </div>
      </template>
    </div>
    
    </div>
</template>

<script>
import { getFriends } from '@/api/friends'
import { API_BASE_URL } from '@/api/config'
import avatarImage from '@/assets/images/avatar.png'

export default {
  name: 'FriendList',
  data() {
    return {
      loading: false,
      error: '',
      friendsList: [], 
      expandedUserId: null,
      defaultAvatar: avatarImage,
      refreshTimer: null, // 自动刷新定时器
      refreshInterval: 10000 // 每10秒刷新一次
    };
  },
  created() {
    this.fetchFriends();
    // 启动自动刷新
    this.startAutoRefresh();
  },
  beforeUnmount() {
    // 组件销毁时清理定时器
    this.stopAutoRefresh();
  },
  methods: {
    getStatusClass(status) {
      switch (status) {
        case '在线':
          return 'online';
        case '专注中':
          return 'busy';
        case '离线':
          return 'offline';
        default:
          return 'default';
      }
    },

    formatUserDetails(data) {
      return {
        user_id: data.user_id,
        username: data.username || data.friend_name || data.friend_username,
        status: data.status || data.friend_status || '未知',
        statusClass: this.getStatusClass(data.status || data.friend_status),
        avatar: data.avatar,
        phone: data.phone,
        sex: data.sex,
        birthday: data.birthday,
        tomato: data.tomato,
        province: data.province
      };
    },

    async fetchUserDetails(username) {
      try {
        const response = await fetch(`${API_BASE_URL}/users/${encodeURIComponent(username)}`, {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('token') || ''}`
          }
        });
        const result = await response.json();

        if (response.ok && result.success && result.data) {
          return this.formatUserDetails(result.data);
        } else {
          console.error(`获取用户 ${username} 详情失败:`, result.message);
          return { user_id: null, username, status: '未知', statusClass: 'default' };
        }
      } catch (err) {
        console.error(`请求用户 ${username} 详情失败:`, err);
        return { user_id: null, username, status: '网络错误', statusClass: 'default' };
      }
    },

    async fetchFriends() {
      this.loading = true;
      this.error = '';
      this.expandedUserId = null; 

      try {
        // 从API获取好友列表
        const response = await getFriends();
        
        if (response.success === true || response.success === "true" || response.code === 200) {
          const friendsList = response.data || [];
          
          // 获取每个好友的详细信息
          const detailPromises = friendsList.map(friend => {
            const username = friend.friend_name || friend.friend_username;
            if (username) {
              return this.fetchUserDetails(username);
            } else {
              // 如果已经有足够的信息，直接格式化
              return Promise.resolve(this.formatUserDetails(friend));
            }
          });
          
          const detailedFriends = await Promise.all(detailPromises);
          
          // 过滤掉没有user_id的好友（用户不存在）
          this.friendsList = detailedFriends.filter(f => f.user_id);
        } else {
          this.error = response.message || '获取好友列表失败';
          this.friendsList = [];
        }
      } catch (err) {
        this.error = '好友列表加载失败，请检查服务连接。';
        console.error('Fetch Friends Error:', err);
        this.friendsList = [];
      } finally {
        this.loading = false;
      }
    },

    toggleDetails(userId) {
      this.expandedUserId = this.expandedUserId === userId ? null : userId;
    },

    getAvatarUrl(avatar) {
      if (!avatar || avatar === this.defaultAvatar) {
        return this.defaultAvatar;
      }
      // 如果已经是完整URL，直接返回
      if (avatar.startsWith('http')) {
        return avatar;
      }
      // 如果是相对路径，添加完整URL（头像通常存储在/uploads目录）
      if (avatar.startsWith('/')) {
        return `http://localhost:8090${avatar}`;
      }
      // 否则拼接完整URL
      return `http://localhost:8090${avatar}`;
    },

    handleAvatarError(event) {
      event.target.src = this.defaultAvatar;
    },

    // 格式化生日
    formatBirthday(birthday) {
      if (!birthday) return '未设置';
      // 如果是时间戳（毫秒）
      if (typeof birthday === 'number') {
        const date = new Date(birthday);
        return `${date.getFullYear()}年${date.getMonth() + 1}月${date.getDate()}日`;
      }
      // 如果是日期字符串
      if (typeof birthday === 'string') {
        // 尝试解析日期字符串
        const date = new Date(birthday);
        if (!isNaN(date.getTime())) {
          return `${date.getFullYear()}年${date.getMonth() + 1}月${date.getDate()}日`;
        }
        // 如果已经是格式化好的字符串，直接返回
        return birthday;
      }
      return '未设置';
    },

    // 启动自动刷新好友列表
    startAutoRefresh() {
      // 如果已经有定时器，先清除
      if (this.refreshTimer) {
        clearInterval(this.refreshTimer);
      }
      
      // 设置定时器，每10秒刷新一次好友列表
      this.refreshTimer = setInterval(() => {
        // 只有在不在加载中时才刷新
        if (!this.loading) {
          this.fetchFriends();
        }
      }, this.refreshInterval);
      
      console.log('✅ FriendList 自动刷新已启动，间隔:', this.refreshInterval / 1000, '秒');
    },

    // 停止自动刷新
    stopAutoRefresh() {
      if (this.refreshTimer) {
        clearInterval(this.refreshTimer);
        this.refreshTimer = null;
        console.log('⏹️ FriendList 自动刷新已停止');
      }
    }
  },
};
</script>

<style scoped>
/* 容器样式 */
.friends-list-container {
  padding: 15px;
  background: white;
  border-radius: 10px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  display: flex;
  flex-direction: column;
  height: 100%; 
  max-height: 100%;
  min-height: 200px; 
}

.friends-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 10px;
  margin-bottom: 5px;
  border-bottom: 2px solid #f0f0f0;
  flex-shrink: 0; 
}

.friends-header h3 {
  margin: 0;
  color: #333;
  font-size: 1.2em;
}

.online-count {
  font-size: 0.9em;
  color: #666;
  font-weight: 500;
}

/* 滚动区域 */
.friends-scroll-area {
  flex-grow: 1; 
  overflow-y: auto; 
  padding-right: 5px; 
}

/* 列表项样式 */
.friend-item {
  display: flex;
  align-items: center;
  padding: 10px 5px;
  gap: 15px;
  border-radius: 6px;
  cursor: pointer;
  transition: background-color 0.2s;
  border-bottom: 1px solid #f5f5f5;
}

.friend-item:last-of-type {
  border-bottom: none;
}

.friend-item:hover {
  background-color: #f7f7f7;
}

/* 展开时的背景 */
.friend-item.is-expanded {
  background-color: #f0f0f0;
  border-radius: 6px 6px 0 0;
  border-bottom: none; 
}

.friend-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background-color: #ddd;
  flex-shrink: 0;
  overflow: hidden;
  background-image: linear-gradient(45deg, #cc2a1f, #eeaa67);
  display: flex;
  align-items: center;
  justify-content: center;
}

.friend-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 50%;
}

/* 名字包装容器：占据所有中间空间，将状态推到右边 */
.friend-name-wrapper {
  flex-grow: 1; 
  min-width: 0; 
}

.friend-name {
  font-weight: bold;
  color: #333;
  font-size: 1.05em;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 状态指示器：放在最右边 */
.friend-status {
  font-size: 0.9em;
  display: flex;
  align-items: center;
  flex-shrink: 0; 
  margin-left: 10px; 
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 6px;
}

/* 状态颜色 */
.friend-status.online {
  color: #4CAF50;
}
.friend-status.online .status-dot {
  background-color: #4CAF50; 
}

.friend-status.busy {
  color: #f1c40f;
}
.friend-status.busy .status-dot {
  background-color: #f1c40f; 
}

.friend-status.offline,
.friend-status.default {
  color: #95a5a6;
}
.friend-status.offline .status-dot,
.friend-status.default .status-dot {
  background-color: #95a5a6; 
}

/* -------------------------------------
   好友详情卡片：L-B-R 橙色硬边框样式
   ------------------------------------- */
.friend-details-card {
  /* 详情内容对齐 */
  padding: 15px 15px 15px 15px; 
  margin-bottom: 8px; 
  background-color: white;
  
  /* 关键修改：移除阴影 */
  box-shadow: none;
  
  /* 橙色边框线 (使用 #eeaa67 作为橙色) */
  border: 2px solid #ffe4cc; 
  border-top: none; /* 顶部无边框 */
  
  /* 增加圆角 */
  border-radius: 0 0 8px 8px; 

  animation: slideDown 0.3s ease-out;
}

.detail-row {
  display: flex;
  gap: 15px;
  line-height: 1.8;
  font-size: 0.9em;
}

.detail-row label {
  font-weight: 500;
  color: #888;
  width: 50px; 
  flex-shrink: 0;
  text-align: right;
}

.detail-row span {
  color: #333;
  font-weight: 400;
  word-break: break-all;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 状态信息/错误信息样式 (保持不变) */
.status-message {
  text-align: center;
  padding: 20px 10px;
  font-size: 0.95em;
  color: #666;
}

.error-state {
  color: #cc2a1f;
  background: #fff0f0;
  border-radius: 4px;
}

.btn-retry {
  margin-top: 10px;
  padding: 5px 10px;
  background: #cc2a1f;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-retry:hover {
  background: #b52217;
}

.spinner {
  border: 4px solid rgba(0, 0, 0, 0.1);
  border-top: 4px solid #cc2a1f;
  border-radius: 50%;
  width: 20px;
  height: 20px;
  animation: spin 1s linear infinite;
  margin: 0 auto 10px auto;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}
</style>