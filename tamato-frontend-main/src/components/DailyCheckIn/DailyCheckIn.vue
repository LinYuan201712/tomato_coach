<template>
  <div class="daily-checkin-card">
    <div class="checkin-header">
      <span class="checkin-title">📅 {{ currentMonth }}</span>
      <span class="checkin-days">{{ checkInDays }}天</span>
    </div>
    
    <!-- 本月日历 -->
    <div class="calendar-container">
      <!-- 星期标题 -->
      <div class="calendar-weekdays">
        <div class="weekday" v-for="day in weekdays" :key="day">{{ day }}</div>
      </div>
      
      <!-- 日期网格 -->
      <div class="calendar-grid">
        <div 
          v-for="(day, index) in calendarDays" 
          :key="index"
          :class="['calendar-day', { 
            'other-month': day.isOtherMonth,
            'today': day.isToday,
            'checked': day.isChecked,
            'makeup-available': day.isMakeupAvailable
          }]"
          :title="getDayTitle(day)"
          @click="handleDayClick(day)"
        >
          <span v-if="day.isChecked" class="tomato-icon">🍅</span>
          <span v-else class="day-number">{{ day.day }}</span>
        </div>
      </div>
    </div>
    
    <!-- 签到按钮 -->
    <div class="checkin-button-container">
      <button 
        v-if="!hasCheckedInToday"
        class="checkin-button" 
        @click="handleCheckIn"
        :disabled="checkingIn"
      >
        <span v-if="checkingIn">签到中...</span>
        <span v-else>立即签到</span>
      </button>
      <div v-else class="checked-message">
        <span class="checked-icon">✓</span>
        <span>今日已签到</span>
      </div>
    </div>

    <!-- 补签确认对话框 -->
    <div v-if="showMakeupDialog" class="makeup-dialog-overlay" @click="showMakeupDialog = false">
      <div class="makeup-dialog" @click.stop>
        <div class="makeup-dialog-header">
          <h3>补签确认</h3>
        </div>
        <div class="makeup-dialog-content">
          <p>确定要补签 <strong>{{ formatMakeupDate(selectedMakeupDate) }}</strong> 吗？</p>
          <p class="makeup-cost">将消耗 <span class="tomato-cost">10 个番茄</span></p>
          <p v-if="userTomatoes < 10" class="makeup-warning">
            ⚠️ 当前只有 {{ userTomatoes }} 个番茄，无法补签
          </p>
        </div>
        <div class="makeup-dialog-actions">
          <button class="makeup-btn-cancel" @click="showMakeupDialog = false">取消</button>
          <button 
            class="makeup-btn-confirm" 
            @click="confirmMakeupCheckIn"
            :disabled="userTomatoes < 10 || makingUpCheckIn"
          >
            <span v-if="makingUpCheckIn">补签中...</span>
            <span v-else>确认补签</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { getCurrency, checkIn, getCurrentMonthCheckInDates, makeupCheckIn, getCurrentUser } from '@/api/user'

export default {
  name: 'DailyCheckIn',
  data() {
    return {
      checkInDays: 0,
      hasCheckedInToday: false,
      checkingIn: false,
      makingUpCheckIn: false,
      currentDate: new Date(),
      checkedDates: new Set(), // 存储已签到的日期
      showMakeupDialog: false,
      selectedMakeupDate: null,
      userTomatoes: 0 // 用户当前番茄数
    }
  },
  computed: {
    currentMonth() {
      const year = this.currentDate.getFullYear()
      const month = this.currentDate.getMonth() + 1
      return `${year}年${month}月`
    },
    weekdays() {
      return ['日', '一', '二', '三', '四', '五', '六']
    },
    calendarDays() {
      const year = this.currentDate.getFullYear()
      const month = this.currentDate.getMonth()
      const today = new Date()
      today.setHours(0, 0, 0, 0)
      
      // 获取本月第一天是星期几
      const firstDay = new Date(year, month, 1)
      const firstDayWeek = firstDay.getDay()
      
      // 获取本月天数
      const daysInMonth = new Date(year, month + 1, 0).getDate()
      
      // 获取上个月需要显示的天数
      const prevMonthDays = firstDayWeek
      
      // 获取上个月的最后几天
      const prevMonth = new Date(year, month, 0)
      const daysInPrevMonth = prevMonth.getDate()
      
      const days = []
      
      // 添加上个月的日期（灰色显示）
      for (let i = prevMonthDays - 1; i >= 0; i--) {
        const day = daysInPrevMonth - i
        const date = new Date(year, month - 1, day)
        date.setHours(0, 0, 0, 0)
        const dateStr = this.formatDate(date)
        const isMakeupAvailable = this.isMakeupAvailable(date, dateStr)
        
        days.push({
          day: day,
          date: dateStr,
          isOtherMonth: true,
          isToday: false,
          isChecked: this.checkedDates.has(dateStr),
          isMakeupAvailable: isMakeupAvailable
        })
      }
      
      // 添加本月的日期
      for (let day = 1; day <= daysInMonth; day++) {
        const date = new Date(year, month, day)
        date.setHours(0, 0, 0, 0)
        const dateStr = this.formatDate(date)
        const isToday = date.getTime() === today.getTime()
        const isMakeupAvailable = this.isMakeupAvailable(date, dateStr)
        
        days.push({
          day: day,
          date: dateStr,
          isOtherMonth: false,
          isToday: isToday,
          isChecked: this.checkedDates.has(dateStr),
          isMakeupAvailable: isMakeupAvailable
        })
      }
      
      // 添加下个月的日期（填充到6行，42个格子）
      const remainingDays = 42 - days.length
      for (let day = 1; day <= remainingDays; day++) {
        const date = new Date(year, month + 1, day)
        date.setHours(0, 0, 0, 0)
        const dateStr = this.formatDate(date)
        const isMakeupAvailable = this.isMakeupAvailable(date, dateStr)
        
        days.push({
          day: day,
          date: dateStr,
          isOtherMonth: true,
          isToday: false,
          isChecked: this.checkedDates.has(dateStr),
          isMakeupAvailable: isMakeupAvailable
        })
      }
      
      return days
    }
  },
  mounted() {
    this.loadCheckInStatus()
  },
  methods: {
    formatDate(date) {
      const year = date.getFullYear()
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      return `${year}-${month}-${day}`
    },
    
    async loadCheckInStatus() {
      try {
        // 并行获取资产信息、签到记录和用户信息（获取番茄数）
        const [currency, checkInDates, userResponse] = await Promise.all([
          getCurrency(),
          getCurrentMonthCheckInDates().catch(() => []), // 如果API不存在，返回空数组
          getCurrentUser().catch(() => null) // 获取用户信息以获取番茄数
        ])
        
        if (currency) {
          this.checkInDays = currency.check_day || 0
          this.hasCheckedInToday = currency.has_checked_in_today || false
        }
        
        // 处理用户信息：getCurrentUser返回的是ApiResponse格式 {success, data, message}
        let userInfo = null
        if (userResponse) {
          if (userResponse.success && userResponse.data) {
            userInfo = userResponse.data
          } else if (userResponse.tomato !== undefined) {
            // 如果直接返回UserResponse对象（没有包装在ApiResponse中）
            userInfo = userResponse
          }
        }
        
        if (userInfo && userInfo.tomato !== undefined) {
          this.userTomatoes = userInfo.tomato || 0
        }
        
        // 更新已签到的日期集合
        this.updateCheckedDates(checkInDates || [])
      } catch (error) {
        console.error('加载签到状态失败:', error)
        // 如果获取签到记录失败，至少显示今天的签到状态
        try {
          const currency = await getCurrency()
          if (currency) {
            this.checkInDays = currency.check_day || 0
            this.hasCheckedInToday = currency.has_checked_in_today || false
          }
          const userResponse = await getCurrentUser().catch(() => null)
          // 处理用户信息：getCurrentUser返回的是ApiResponse格式
          let userInfo = null
          if (userResponse) {
            if (userResponse.success && userResponse.data) {
              userInfo = userResponse.data
            } else if (userResponse.tomato !== undefined) {
              // 如果直接返回UserResponse对象
              userInfo = userResponse
            }
          }
          if (userInfo && userInfo.tomato !== undefined) {
            this.userTomatoes = userInfo.tomato || 0
          }
        } catch (e) {
          console.error('获取资产信息失败:', e)
        }
        this.updateCheckedDates([])
      }
    },
    
    updateCheckedDates(checkInDateStrings = []) {
      // 清空已签到的日期集合
      this.checkedDates.clear()
      
      // 添加所有已签到的日期
      checkInDateStrings.forEach(dateStr => {
        this.checkedDates.add(dateStr)
      })
      
      // 如果今天已签到但不在列表中，确保添加今天
      if (this.hasCheckedInToday) {
        const today = new Date()
        const todayStr = this.formatDate(today)
        this.checkedDates.add(todayStr)
      }
    },
    
    async handleCheckIn() {
      if (this.hasCheckedInToday || this.checkingIn) {
        return
      }
      
      this.checkingIn = true
      try {
        await checkIn()
        // 签到成功后，重新加载签到状态
        await this.loadCheckInStatus()
        
        // 显示成功提示
        this.$emit('checkin-success', {
          message: '签到成功！获得 1 个番茄 🍅',
          tomatoes: 1
        })
      } catch (error) {
        console.error('签到失败:', error)
        const errorMessage = error.message || '签到失败，请稍后重试'
        if (errorMessage.includes('今日已签到')) {
          alert('今日已签到，请明天再来')
          await this.loadCheckInStatus()
        } else {
          alert(errorMessage)
        }
      } finally {
        this.checkingIn = false
      }
    },
    
    // 判断日期是否可补签
    isMakeupAvailable(date, dateStr) {
      const today = new Date()
      today.setHours(0, 0, 0, 0)
      
      // 不能补签今天或未来
      if (date >= today) {
        return false
      }
      
      // 不能补签已签到的日期
      if (this.checkedDates.has(dateStr)) {
        return false
      }
      
      // 只能补签过去7天内
      const daysDiff = Math.floor((today - date) / (1000 * 60 * 60 * 24))
      if (daysDiff > 7) {
        return false
      }
      
      return true
    },
    
    // 获取日期标题（用于tooltip）
    getDayTitle(day) {
      if (day.isChecked) {
        return `${day.date} 已签到`
      } else if (day.isMakeupAvailable) {
        return `${day.date} 点击补签（消耗10个番茄）`
      } else if (day.isToday) {
        return `${day.date} 今天`
      } else {
        return day.date
      }
    },
    
    // 处理日期点击
    handleDayClick(day) {
      // 如果是可补签的日期，显示补签对话框
      if (day.isMakeupAvailable && !day.isOtherMonth) {
        this.selectedMakeupDate = day.date
        this.showMakeupDialog = true
      }
    },
    
    // 确认补签
    async confirmMakeupCheckIn() {
      if (!this.selectedMakeupDate || this.userTomatoes < 10 || this.makingUpCheckIn) {
        return
      }
      
      const makeupDate = this.selectedMakeupDate // 保存日期，因为后面会清空
      this.makingUpCheckIn = true
      try {
        await makeupCheckIn(makeupDate)
        // 补签成功后，重新加载签到状态
        await this.loadCheckInStatus()
        
        // 显示成功提示（在关闭对话框之前）
        this.$emit('checkin-success', {
          message: `补签成功！已补签 ${this.formatMakeupDate(makeupDate)}，消耗 10 个番茄`,
          tomatoes: -10
        })
        
        // 关闭对话框
        this.showMakeupDialog = false
        this.selectedMakeupDate = null
      } catch (error) {
        console.error('补签失败:', error)
        const errorMessage = error.message || '补签失败，请稍后重试'
        alert(errorMessage)
      } finally {
        this.makingUpCheckIn = false
      }
    },
    
    // 格式化补签日期显示
    formatMakeupDate(dateStr) {
      if (!dateStr) return ''
      const date = new Date(dateStr + 'T00:00:00')
      const month = date.getMonth() + 1
      const day = date.getDate()
      return `${month}月${day}日`
    }
  }
}
</script>

<style scoped>
.daily-checkin-card {
  background: white;
  border-radius: 10px;
  padding: 12px;
  box-shadow: 0 2px 6px rgba(238, 170, 103, 0.1);
  border: 1px solid #ffe4cc;
  margin-bottom: 15px;
}

.checkin-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
  padding-bottom: 8px;
  border-bottom: 1px solid #f5f5f5;
}

.checkin-title {
  font-size: 0.9em;
  color: #333;
  font-weight: 600;
}

.checkin-days {
  font-size: 0.75em;
  color: #eeaa67;
  font-weight: 600;
  background: #fff5eb;
  padding: 2px 8px;
  border-radius: 10px;
}

.calendar-container {
  margin-bottom: 10px;
}

.calendar-weekdays {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 2px;
  margin-bottom: 4px;
}

.weekday {
  text-align: center;
  font-size: 0.7em;
  color: #999;
  font-weight: 500;
  padding: 4px 0;
}

.calendar-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 2px;
}

.calendar-day {
  aspect-ratio: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  font-size: 0.75em;
  transition: all 0.2s;
  position: relative;
}

.calendar-day.other-month {
  color: #ddd;
}

.calendar-day.today {
  background: #fff5eb;
  border: 1px solid #ffe4cc;
  font-weight: 600;
  color: #eeaa67;
}

.calendar-day.checked {
  background: #fff5eb;
}

.calendar-day.makeup-available {
  background: #fff8f0;
  border: 1px dashed #ffd4a3;
  cursor: pointer;
  position: relative;
}

.calendar-day.makeup-available:hover {
  background: #ffe4cc;
  border-color: #eeaa67;
  transform: scale(1.05);
}

.calendar-day.makeup-available .day-number {
  color: #eeaa67;
  font-weight: 500;
}

.tomato-icon {
  font-size: 1.2em;
  display: block;
}

.day-number {
  display: block;
  color: #666;
}

.calendar-day.today .day-number {
  color: #eeaa67;
}

.calendar-day.other-month .day-number {
  color: #ddd;
}

.checkin-button-container {
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid #f5f5f5;
  text-align: center;
}

.checkin-button {
  width: 100%;
  padding: 8px;
  background: linear-gradient(135deg, #eeaa67, #f5b877);
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 0.85em;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow: 0 2px 4px rgba(238, 170, 103, 0.25);
}

.checkin-button:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 3px 8px rgba(238, 170, 103, 0.35);
}

.checkin-button:active:not(:disabled) {
  transform: translateY(0);
}

.checkin-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.checked-message {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  color: #eeaa67;
  font-size: 0.85em;
  font-weight: 500;
  padding: 8px;
}

.checked-icon {
  font-size: 1em;
  font-weight: bold;
}

@media (max-width: 768px) {
  .daily-checkin-card {
    padding: 10px;
  }
  
  .checkin-title {
    font-size: 0.85em;
  }
  
  .weekday {
    font-size: 0.65em;
  }
  
  .calendar-day {
    font-size: 0.7em;
  }
  
  .tomato-icon {
    font-size: 1em;
  }
}

/* 补签对话框样式 */
.makeup-dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.makeup-dialog {
  background: white;
  border-radius: 12px;
  padding: 20px;
  min-width: 300px;
  max-width: 90%;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  animation: dialogFadeIn 0.2s ease;
}

@keyframes dialogFadeIn {
  from {
    opacity: 0;
    transform: scale(0.9);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}

.makeup-dialog-header {
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 1px solid #f5f5f5;
}

.makeup-dialog-header h3 {
  margin: 0;
  font-size: 1.1em;
  color: #333;
  font-weight: 600;
}

.makeup-dialog-content {
  margin-bottom: 20px;
  line-height: 1.6;
  color: #666;
}

.makeup-dialog-content p {
  margin: 8px 0;
}

.makeup-dialog-content strong {
  color: #eeaa67;
  font-weight: 600;
}

.makeup-cost {
  font-size: 0.95em;
  margin-top: 10px;
}

.tomato-cost {
  color: #eeaa67;
  font-weight: 600;
  font-size: 1.1em;
}

.makeup-warning {
  color: #ff6b6b;
  font-size: 0.9em;
  margin-top: 8px;
  padding: 8px;
  background: #fff5f5;
  border-radius: 6px;
  border-left: 3px solid #ff6b6b;
}

.makeup-dialog-actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
}

.makeup-btn-cancel,
.makeup-btn-confirm {
  padding: 8px 20px;
  border: none;
  border-radius: 6px;
  font-size: 0.9em;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.makeup-btn-cancel {
  background: #f5f5f5;
  color: #666;
}

.makeup-btn-cancel:hover {
  background: #e8e8e8;
}

.makeup-btn-confirm {
  background: linear-gradient(135deg, #eeaa67, #f5b877);
  color: white;
  box-shadow: 0 2px 4px rgba(238, 170, 103, 0.25);
}

.makeup-btn-confirm:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 3px 8px rgba(238, 170, 103, 0.35);
}

.makeup-btn-confirm:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
