// 内容脚本：与网页通信，接收来自 background 的消息

// 监听来自 background 的消息
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.type === 'DISTRACTING_SITE_DETECTED') {
    // 检测到分心网站，显示警告
    showDistractionWarning(request.url);
    sendResponse({ received: true });
  }
  return true;
});

// 显示分心警告
function showDistractionWarning(url) {
  // 创建警告弹窗
  const warningDiv = document.createElement('div');
  warningDiv.id = 'tomato-distraction-warning';
  warningDiv.style.cssText = `
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0, 0, 0, 0.8);
    z-index: 999999;
    display: flex;
    justify-content: center;
    align-items: center;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  `;
  
  const warningContent = document.createElement('div');
  warningContent.style.cssText = `
    background: white;
    padding: 40px;
    border-radius: 16px;
    max-width: 500px;
    text-align: center;
    box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
  `;
  
  warningContent.innerHTML = `
    <div style="font-size: 48px; margin-bottom: 20px;">⚠️</div>
    <h2 style="color: #ff6b6b; margin: 0 0 16px 0; font-size: 24px;">检测到分心网站！</h2>
    <p style="color: #333; margin: 0 0 24px 0; line-height: 1.6;">
      您正在专注学习中，请关闭此页面并回到学习页面继续专注。
    </p>
    <div style="display: flex; gap: 12px; justify-content: center;">
      <button id="tomato-close-tab" style="
        background: #ff6b6b;
        color: white;
        border: none;
        padding: 12px 24px;
        border-radius: 8px;
        cursor: pointer;
        font-size: 16px;
        font-weight: 500;
      ">关闭此标签页</button>
      <button id="tomato-close-warning" style="
        background: #e0e0e0;
        color: #333;
        border: none;
        padding: 12px 24px;
        border-radius: 8px;
        cursor: pointer;
        font-size: 16px;
        font-weight: 500;
      ">我知道了</button>
    </div>
  `;
  
  warningDiv.appendChild(warningContent);
  document.body.appendChild(warningDiv);
  
  // 关闭标签页按钮
  document.getElementById('tomato-close-tab').addEventListener('click', () => {
    // 直接关闭当前标签页
    window.close();
    // 如果无法关闭（某些浏览器限制），尝试跳转到学习页面
    setTimeout(() => {
      if (!document.hidden) {
        window.location.href = 'http://localhost:8080/';
      }
    }, 100);
  });
  
  // 关闭警告按钮（5秒后自动关闭）
  document.getElementById('tomato-close-warning').addEventListener('click', () => {
    warningDiv.remove();
  });
  
  // 5秒后自动关闭警告（但标签页仍然打开）
  setTimeout(() => {
    if (document.getElementById('tomato-distraction-warning')) {
      document.getElementById('tomato-distraction-warning').remove();
    }
  }, 5000);
}

