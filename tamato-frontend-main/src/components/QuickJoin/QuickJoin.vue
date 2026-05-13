<template>
  <div class="quick-join-widget">
    <div class="widget-header">
      <h3 class="widget-title">快速加入自习室</h3>
      <div class="header-actions">
        <button @click="refreshRooms" class="refresh-btn" :disabled="loading">
          {{ loading ? '刷新中...' : '🔄' }}
        </button>
      </div>
    </div>
    
    <!-- 房间列表显示 -->
    <div v-if="!loading && displayedRooms.length > 0" class="rooms-container">
      <div 
        v-for="room in displayedRooms" 
        :key="room.room_id"
        class="room-card"
      >
        <div class="room-header">
          <div class="room-avatar" :style="{ backgroundColor: getAvatarColor(room.room_name) }">
            {{ getRoomInitials(room.room_name) }}
          </div>
          <div class="room-info">
            <h4 class="room-name">{{ room.room_name }}</h4>
            <div class="room-meta">
              <span class="room-members">
                👥 {{ room.current_members || 0 }}/{{ room.max_members }}
              </span>
              <span class="room-status" :class="getStatusClass(room)">
                {{ getStatusText(room) }}
              </span>
            </div>
          </div>
        </div>
        
        <div class="room-footer">
          <div class="room-music">
            <span v-if="room.music_name && room.music_name !== '无' && room.music_name !== '无背景音乐'">
              🎵 {{ room.music_name }}
            </span>
            <span v-else>🔇 无音乐</span>
          </div>
          <button 
            @click="joinRoom(room.room_id)"
            class="join-btn"
            :disabled="room.current_members >= room.max_members"
          >
            {{ room.current_members >= room.max_members ? '已满' : '加入' }}
          </button>
        </div>
      </div>
    </div>
    
    <!-- 翻页控件 - 只有当有足够数据时才显示 -->
    <div v-if="!loading && totalPages > 1" class="pagination-controls">
      <button 
        @click="prevPage" 
        :disabled="currentPage === 1" 
        class="page-btn prev-btn"
      >
        ◀ 上一页
      </button>
      
      <div class="page-indicator">
        第 {{ currentPage }} 页 / 共 {{ totalPages }} 页
      </div>
      
      <button 
        @click="nextPage" 
        :disabled="currentPage === totalPages" 
        class="page-btn next-btn"
      >
        下一页 ▶
      </button>
    </div>
    
    <!-- 空状态 -->
    <div v-if="!loading && displayedRooms.length === 0" class="empty-state">
      <div class="empty-icon">🏠</div>
      <p>暂无自习室</p>
      <button @click="refreshRooms" class="retry-btn">
        重新加载
      </button>
    </div>
    
    <!-- 加载状态 -->
    <div v-if="loading" class="loading-state">
      <div class="loading-spinner"></div>
      <p>加载自习室中...</p>
    </div>
  </div>
</template>

<script>
import { getRoomsList, getRoomMembers } from '@/api/studyRooms'

export default {
  name: 'QuickJoin',
  props: {
    roomsPerPage: {
      type: Number,
      default: 4
    },
    autoRefresh: {
      type: Boolean,
      default: false
    },
    refreshInterval: {
      type: Number,
      default: 30000
    }
  },
  data() {
    return {
      allRooms: [],      // 所有获取到的房间
      displayedRooms: [], // 当前页显示的房间
      loading: false,
      refreshTimer: null,
      currentPage: 1,
      totalPages: 1      // 基于实际数据计算
    }
  },
  mounted() {
    this.loadRooms()
    if (this.autoRefresh) {
      this.startAutoRefresh()
    }
  },
  beforeUnmount() {
    this.stopAutoRefresh()
  },
  watch: {
    // 监听当前页码变化，自动更新显示的房间
    currentPage: {
      handler() {
        this.updateDisplayedRooms()
      },
      immediate: true
    }
  },
  methods: {
    async loadRooms() {
      try {
        this.loading = true
        const response = await getRoomsList()
        console.log('📊 自习室列表API响应:', response)
        
        // 支持两种数据格式：
        // 1. response.data 直接是数组
        // 2. response.data.list 是数组
        let roomList = null
        if (response && response.data) {
          if (Array.isArray(response.data)) {
            // 格式1: data 直接是数组
            roomList = response.data
          } else if (response.data.list && Array.isArray(response.data.list)) {
            // 格式2: data.list 是数组
            roomList = response.data.list
          }
        }
        
        if (roomList && roomList.length > 0) {
          console.log(`✅ 获取到 ${roomList.length} 个自习室`)
          console.log('📊 原始房间数据示例（第一个）:', roomList[0])
          
          // 处理所有房间数据
          this.allRooms = roomList
            .filter(room => room)
            .map(room => this.formatRoomData(room))
          
          console.log('📋 格式化后的房间数据:', this.allRooms)
          
          // 计算实际的总页数
          this.calculateTotalPages()
          
          // 更新显示的房间
          this.updateDisplayedRooms()
          
          console.log(`🎯 总房间数: ${this.allRooms.length}, 每页: ${this.roomsPerPage}, 总页数: ${this.totalPages}`)
          
          // 异步获取每个房间的真实成员数（如果API没有返回）
          this.updateRoomMemberCounts()
        } else {
          console.warn('⚠️ API返回数据格式异常或数据为空', response)
          this.allRooms = []
          this.calculateTotalPages()
          this.updateDisplayedRooms()
        }
      } catch (error) {
        console.error('❌ 加载自习室列表失败:', error)
        this.allRooms = []
        this.calculateTotalPages()
        this.updateDisplayedRooms()
      } finally {
        this.loading = false
      }
    },

    // 计算总页数（基于实际数据）
    calculateTotalPages() {
      if (this.allRooms.length === 0) {
        this.totalPages = 1
      } else {
        this.totalPages = Math.ceil(this.allRooms.length / this.roomsPerPage)
      }
      
      // 如果当前页超过了总页数，回到第一页
      if (this.currentPage > this.totalPages) {
        this.currentPage = 1
      }
      
      console.log(`📄 计算总页数: ${this.totalPages} (房间数: ${this.allRooms.length}, 每页: ${this.roomsPerPage})`)
    },

    // 更新当前显示的房间
    updateDisplayedRooms() {
      console.log(`🔄 更新第 ${this.currentPage} 页的房间`)
      
      if (this.allRooms.length === 0) {
        this.displayedRooms = []
        return
      }
      
      // 计算当前页的房间索引
      const startIndex = (this.currentPage - 1) * this.roomsPerPage
      const endIndex = Math.min(startIndex + this.roomsPerPage, this.allRooms.length)
      
      console.log(`📄 索引范围: ${startIndex} - ${endIndex}`)
      
      // 获取当前页的房间
      this.displayedRooms = this.allRooms.slice(startIndex, endIndex)
      
      console.log(`✅ 显示 ${this.displayedRooms.length} 个房间`)
    },

    // 上一页
    prevPage() {
      if (this.currentPage > 1) {
        this.currentPage--
        console.log(`⬅️ 翻到上一页: ${this.currentPage}`)
      }
    },

    // 下一页
    nextPage() {
      if (this.currentPage < this.totalPages) {
        this.currentPage++
        console.log(`➡️ 翻到下一页: ${this.currentPage}`)
      }
    },

    formatRoomData(room) {
      // 支持驼峰命名和下划线命名两种格式
      const roomId = room.roomId || room.room_id || room.id || 'unknown_' + Date.now()
      const roomName = room.roomName || room.room_name || room.name || '自习室'
      const maxMembers = Math.max(room.maxMembers || room.max_members || room.max_member || 4, 1)
      
      // 从API返回的数据中读取现有人数，支持多种字段名
      // 后端可能返回活跃成员数（状态为'专注中'或'休息中'的成员）
      // 优先使用：activeMemberCount, active_member_count, countActiveMembers, count_active_members
      // 其次使用：currentMembers, current_members, current_member, memberCount, member_count
      let currentMembers = room.activeMemberCount || 
                          room.active_member_count ||
                          room.countActiveMembers ||
                          room.count_active_members ||
                          room.currentMembers || 
                          room.current_members || 
                          room.current_member || 
                          room.memberCount || 
                          room.member_count ||
                          room.activeMembers ||
                          room.active_members ||
                          0
      
      // 确保人数在合理范围内（0 到 maxMembers）
      currentMembers = Math.min(Math.max(Number(currentMembers) || 0, 0), maxMembers)
      
      // 支持 endTime (时间戳毫秒) 和 end_time (时间戳秒) 两种格式
      const endTime = room.endTime || room.end_time
      const isActive = endTime ? (Date.now() < (endTime > 1000000000000 ? endTime : endTime * 1000)) : true
      
      return {
        room_id: roomId,
        room_name: roomName,
        create_person: room.createPerson || room.create_person || room.creator || room.owner || '未知用户',
        max_members: maxMembers,
        current_members: currentMembers, // 使用真实的人数，不再随机生成
        music_name: room.musicName || room.music_name || room.music || '无背景音乐',
        is_active: isActive
      }
    },

    getAvatarColor(name) {
      const colors = [
        '#eeaa67', // 橘黄色
        '#4dabf7', // 蓝色
        '#69db7c', // 绿色
        '#ff922b', // 橙色
        '#748ffc', // 紫色
        '#20c997', // 青色
        '#fa5252', // 红色
        '#7950f2'  // 深紫色
      ]
      if (!name) return colors[0]
      const index = name.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0) % colors.length
      return colors[index]
    },

    getRoomInitials(name) {
      if (!name) return '🏠'
      const initials = name
        .split(' ')
        .map(word => word.charAt(0))
        .join('')
        .toUpperCase()
        .substring(0, 2)
      return initials || name.charAt(0)
    },

    getStatusClass(room) {
      if (room.current_members >= room.max_members) {
        return 'status-full'
      }
      // 房间存在就一定有成员（房主离开会解散房间），所以总是显示为活跃状态
      return 'status-active'
    },

    getStatusText(room) {
      if (room.current_members >= room.max_members) {
        return '已满'
      }
      // 房间存在就一定有成员（房主离开会解散房间）
      // 不再显示"专注中"，因为无法从房间列表API获取成员的实际状态
      // 统一显示"进行中"
      return '进行中'
    },

    joinRoom(roomId) {
      console.log('🎯 请求加入房间:', roomId)
      this.$emit('join-room', roomId)
    },

    refreshRooms() {
      console.log('🔄 刷新房间列表')
      this.currentPage = 1
      this.loadRooms()
    },

    startAutoRefresh() {
      if (this.refreshTimer) {
        clearInterval(this.refreshTimer)
      }
      this.refreshTimer = setInterval(() => {
        this.refreshRooms()
      }, this.refreshInterval)
    },

    stopAutoRefresh() {
      if (this.refreshTimer) {
        clearInterval(this.refreshTimer)
        this.refreshTimer = null
      }
    },

    // 异步更新房间的成员数（如果API没有返回或需要验证）
    async updateRoomMemberCounts() {
      // 检查哪些房间需要更新成员数
      // 注意：后端统计的是状态为'专注中'或'休息中'的活跃成员
      // 如果现有人数为0，可能是：
      // 1. API没有返回真实数据
      // 2. 房间确实没有活跃成员（但创建者刚加入可能还没开始专注）
      const roomsToUpdate = this.allRooms.filter(room => {
        // 如果现有人数为0，可能需要更新（但创建者可能在房间里但还没开始专注）
        // 为了确保准确性，我们仍然尝试获取一次真实数据
        return room.current_members === 0 || room.current_members === undefined || room.current_members === null
      })

      if (roomsToUpdate.length === 0) {
        console.log('✅ 所有房间都有现有人数，无需更新')
        return
      }

      console.log(`🔄 需要更新 ${roomsToUpdate.length} 个房间的成员数（现有人数为0或未返回）`)
      console.log('📝 注意：后端统计的是状态为"专注中"或"休息中"的活跃成员数')

      // 并行获取所有房间的成员数（限制并发数，避免过多请求）
      const BATCH_SIZE = 5 // 每批处理5个房间
      for (let i = 0; i < roomsToUpdate.length; i += BATCH_SIZE) {
        const batch = roomsToUpdate.slice(i, i + BATCH_SIZE)
        const promises = batch.map(async (room) => {
          try {
            const response = await getRoomMembers(room.room_id, null)
            console.log(`📊 房间 ${room.room_id} 的成员列表响应:`, response)

            // 解析成员列表
            // 注意：后端返回的成员列表可能只包含状态为'专注中'或'休息中'的活跃成员
            const data = response?.data
            const memberList = Array.isArray(data?.list) ? data.list
              : Array.isArray(data?.members) ? data.members
              : Array.isArray(data?.content) ? data.content
              : Array.isArray(data) ? data
              : []

            // 统计活跃成员数（状态为'专注中'或'休息中'的成员）
            // 如果后端只返回活跃成员，memberList.length 就是活跃成员数
            const activeMemberCount = memberList.length
            
            // 如果API返回了总成员数，优先使用
            const totalMemberCount = data?.totalMembers || data?.total_members || data?.memberCount || data?.member_count
            
            // 使用总成员数（如果存在），否则使用活跃成员数
            const memberCount = totalMemberCount !== undefined ? totalMemberCount : activeMemberCount
            
            console.log(`✅ 房间 ${room.room_id} 的成员数: ${memberCount} (活跃成员: ${activeMemberCount}${totalMemberCount ? `, 总成员: ${totalMemberCount}` : ''})`)

            // 更新房间的成员数
            const roomIndex = this.allRooms.findIndex(r => r.room_id === room.room_id)
            if (roomIndex !== -1) {
              // Vue 3 中不需要 $set，直接赋值即可
              this.allRooms[roomIndex].current_members = memberCount
              // 如果当前页显示了这个房间，也需要更新显示
              this.updateDisplayedRooms()
            }
          } catch (error) {
            console.error(`❌ 获取房间 ${room.room_id} 的成员数失败:`, error)
            // 失败时保持原值，不做处理
          }
        })

        // 等待当前批次完成
        await Promise.all(promises)
        
        // 批次之间稍作延迟，避免请求过于密集
        if (i + BATCH_SIZE < roomsToUpdate.length) {
          await new Promise(resolve => setTimeout(resolve, 100))
        }
      }

      console.log('✅ 房间成员数更新完成')
    }
  }
}
</script>

<style scoped>
/* 保持原有的样式，只调整翻页控件的样式 */

.pagination-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 20px;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 8px;
  border: 1px solid #e9ecef;
}

.page-btn {
  background: white;
  border: 1px solid #dee2e6;
  border-radius: 6px;
  padding: 8px 16px;
  cursor: pointer;
  font-size: 0.9em;
  font-weight: 500;
  color: #495057;
  transition: all 0.2s ease;
  min-width: 80px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.page-btn:hover:not(:disabled) {
  background: #fff5eb;
  border-color: #eeaa67;
  color: #eeaa67;
  transform: translateY(-1px);
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  background: #f8f9fa;
}

.page-indicator {
  font-size: 0.9em;
  color: #666;
  font-weight: 500;
  padding: 0 15px;
}

/* 原有样式保持不变 */
.quick-join-widget {
  background: white;
  border-radius: 12px;
  padding: 20px;
}

.widget-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.widget-title {
  margin: 0;
  font-size: 1.2em;
  color: #333;
  font-weight: 600;
}

.header-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.refresh-btn {
  background: #f8f9fa;
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  padding: 6px 12px;
  cursor: pointer;
  font-size: 1em;
  transition: all 0.2s ease;
}

.refresh-btn:hover:not(:disabled) {
  background: #fff5eb;
  border-color: #eeaa67;
  color: #eeaa67;
}

.refresh-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.rooms-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 300px;
}

.room-card {
  background: #f8f9fa;
  border-radius: 8px;
  padding: 15px;
  border: 1px solid #e9ecef;
  transition: all 0.3s ease;
}

.room-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  border-color: #eeaa67;
}

/* ... 其余样式保持不变 ... */
</style>
<style scoped>
.quick-join-widget {
  background: white;
  border-radius: 12px;
  padding: 20px;
}

.widget-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.widget-title {
  margin: 0;
  font-size: 1.2em;
  color: #333;
  font-weight: 600;
}

.header-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.refresh-btn {
  background: #f8f9fa;
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  padding: 6px 12px;
  cursor: pointer;
  font-size: 1em;
  transition: all 0.2s ease;
}

.refresh-btn:hover:not(:disabled) {
  background: #fff5eb;
  border-color: #eeaa67;
  color: #eeaa67;
}

.refresh-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.rooms-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 300px; /* 保持高度稳定 */
}

.room-card {
  background: #f8f9fa;
  border-radius: 8px;
  padding: 15px;
  border: 1px solid #e9ecef;
  transition: all 0.3s ease;
}

.room-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  border-color: #eeaa67;
}

.room-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.room-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: bold;
  font-size: 1em;
  flex-shrink: 0;
}

.room-info {
  flex: 1;
  min-width: 0;
}

.room-name {
  margin: 0 0 4px 0;
  font-size: 1em;
  font-weight: 600;
  color: #333;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.room-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 0.85em;
}

.room-members {
  color: #666;
}

.room-status {
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 0.8em;
  font-weight: 500;
}

.status-active {
  background: #e7f5e9;
  color: #2b8a3e;
}

.status-inactive {
  background: #f8f9fa;
  color: #6c757d;
}

.status-full {
  background: #fff5f5;
  color: #c92a2a;
}

.room-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 12px;
  border-top: 1px solid #e9ecef;
}

.room-music {
  font-size: 0.85em;
  color: #666;
  display: flex;
  align-items: center;
  gap: 4px;
}

.join-btn {
  background: #eeaa67;
  color: white;
  border: none;
  padding: 6px 16px;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s ease;
}

.join-btn:hover:not(:disabled) {
  background: #e69c55;
  transform: translateY(-1px);
}

.join-btn:disabled {
  background: #ccc;
  cursor: not-allowed;
  transform: none;
}

/* 翻页控件样式 */
.pagination-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 20px;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 8px;
  border: 1px solid #e9ecef;
}

.page-btn {
  background: white;
  border: 1px solid #dee2e6;
  border-radius: 6px;
  padding: 8px 16px;
  cursor: pointer;
  font-size: 0.9em;
  font-weight: 500;
  color: #495057;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  gap: 4px;
}

.page-btn:hover:not(:disabled) {
  background: #fff5eb;
  border-color: #eeaa67;
  color: #eeaa67;
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  background: #f8f9fa;
}

.prev-btn {
  order: 1;
}

.next-btn {
  order: 3;
}

.page-numbers {
  order: 2;
  flex: 1;
  text-align: center;
}

.page-info {
  font-size: 0.9em;
  color: #666;
  font-weight: 500;
}

.empty-state {
  text-align: center;
  padding: 30px 20px;
  color: #666;
}

.empty-icon {
  font-size: 3em;
  margin-bottom: 10px;
}

.empty-state p {
  margin: 0 0 15px 0;
}

.retry-btn {
  background: #eeaa67;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9em;
  transition: all 0.2s ease;
}

.retry-btn:hover {
  background: #e69c55;
}

.loading-state {
  text-align: center;
  padding: 30px 20px;
  color: #666;
}

.loading-spinner {
  width: 30px;
  height: 30px;
  border: 3px solid #f3f3f3;
  border-top: 3px solid #eeaa67;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin: 0 auto 15px;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}
</style>