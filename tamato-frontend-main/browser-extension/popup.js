// 弹出窗口脚本

document.addEventListener('DOMContentLoaded', () => {
  const statusDiv = document.getElementById('status');
  const toggleBtn = document.getElementById('toggleFocus');
  const checkTabsBtn = document.getElementById('checkTabs');
  
  // 加载当前专注状态
  chrome.storage.local.get(['isFocusing'], (result) => {
    updateUI(result.isFocusing || false);
  });
  
  // 切换专注状态
  toggleBtn.addEventListener('click', () => {
    chrome.storage.local.get(['isFocusing'], (result) => {
      const newState = !(result.isFocusing || false);
      chrome.storage.local.set({ isFocusing: newState }, () => {
        updateUI(newState);
        // 通知所有标签页
        chrome.tabs.query({}, (tabs) => {
          tabs.forEach(tab => {
            chrome.tabs.sendMessage(tab.id, {
              type: 'FOCUS_STATE_CHANGED',
              isFocusing: newState
            }).catch(() => {
              // 忽略错误（某些标签页可能没有内容脚本）
            });
          });
        });
      });
    });
  });
  
  // 检查分心标签页
  checkTabsBtn.addEventListener('click', () => {
    chrome.runtime.sendMessage({ type: 'CHECK_DISTRACTING_TABS' }, (response) => {
      if (response && response.distractingTabs && response.distractingTabs.length > 0) {
        alert(`发现 ${response.distractingTabs.length} 个分心标签页`);
        // 可以在这里显示列表或直接关闭
      } else {
        alert('未发现分心标签页');
      }
    });
  });
  
  // 更新UI
  function updateUI(isFocusing) {
    if (isFocusing) {
      statusDiv.textContent = '正在专注中';
      statusDiv.className = 'status focusing';
      toggleBtn.textContent = '结束专注模式';
      toggleBtn.className = 'btn-secondary';
    } else {
      statusDiv.textContent = '未在专注';
      statusDiv.className = 'status resting';
      toggleBtn.textContent = '开始专注模式';
      toggleBtn.className = 'btn-primary';
    }
  }
});

