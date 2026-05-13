// 分心网站列表（可以根据需要扩展）
const DISTRACTING_SITES = [
  'douyin.com',
  'tiktok.com',
  'bilibili.com',
  'youtube.com',
  'iqiyi.com',
  'youku.com',
  'tencent.com/video',
  'acfun.cn',
  'huya.com',
  'douyu.com',
  'kuaishou.com',
  'xiaohongshu.com',
  'weibo.com',
  'zhihu.com/hot',
  'taobao.com',
  'tmall.com',
  'jd.com'
];

// 检查URL是否是分心网站
function isDistractingSite(url) {
  if (!url) return false;
  try {
    const urlObj = new URL(url);
    const hostname = urlObj.hostname.toLowerCase();
    return DISTRACTING_SITES.some(site => hostname.includes(site));
  } catch (e) {
    return false;
  }
}

// 监听标签页更新
chrome.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
  if (changeInfo.status === 'complete' && tab.url) {
    if (isDistractingSite(tab.url)) {
      // 检查是否在专注模式
      chrome.storage.local.get(['isFocusing'], (result) => {
        if (result.isFocusing) {
          // 通知内容脚本显示警告
          chrome.tabs.sendMessage(tabId, {
            type: 'DISTRACTING_SITE_DETECTED',
            url: tab.url
          }).catch(() => {
            // 如果内容脚本未加载，尝试关闭标签页或显示通知
            console.log('检测到分心网站，但无法发送消息');
          });
        }
      });
    }
  }
});

// 监听标签页激活（切换标签页）
chrome.tabs.onActivated.addListener((activeInfo) => {
  chrome.tabs.get(activeInfo.tabId, (tab) => {
    if (tab && tab.url && isDistractingSite(tab.url)) {
      chrome.storage.local.get(['isFocusing'], (result) => {
        if (result.isFocusing) {
          // 通知内容脚本
          chrome.tabs.sendMessage(activeInfo.tabId, {
            type: 'DISTRACTING_SITE_DETECTED',
            url: tab.url
          }).catch(() => {
            console.log('检测到分心网站');
          });
        }
      });
    }
  });
});

// 监听来自内容脚本的消息
chrome.runtime.onMessage.addListener((request, sender, sendResponse) => {
  if (request.type === 'SET_FOCUSING') {
    // 设置专注状态
    chrome.storage.local.set({ isFocusing: request.isFocusing }, () => {
      sendResponse({ success: true });
    });
    return true; // 保持消息通道开放
  }
  
  if (request.type === 'CHECK_DISTRACTING_TABS') {
    // 检查所有打开的标签页
    chrome.tabs.query({}, (tabs) => {
      const distractingTabs = tabs.filter(tab => 
        tab.url && isDistractingSite(tab.url)
      );
      sendResponse({ distractingTabs: distractingTabs.map(tab => ({
        id: tab.id,
        url: tab.url,
        title: tab.title
      })) });
    });
    return true;
  }
  
  if (request.type === 'CLOSE_DISTRACTING_TAB') {
    // 关闭分心标签页
    if (request.tabId) {
      chrome.tabs.remove(request.tabId, () => {
        sendResponse({ success: true });
      });
    }
    return true;
  }
});

// 监听存储变化，当专注状态改变时检查所有标签页
chrome.storage.onChanged.addListener((changes, areaName) => {
  if (areaName === 'local' && changes.isFocusing) {
    if (changes.isFocusing.newValue === true) {
      // 开始专注，检查所有打开的标签页
      chrome.tabs.query({}, (tabs) => {
        tabs.forEach(tab => {
          if (tab.url && isDistractingSite(tab.url)) {
            chrome.tabs.sendMessage(tab.id, {
              type: 'DISTRACTING_SITE_DETECTED',
              url: tab.url
            }).catch(() => {
              // 如果标签页没有内容脚本，尝试关闭它
              if (tab.url.startsWith('http://') || tab.url.startsWith('https://')) {
                chrome.tabs.remove(tab.id);
              }
            });
          }
        });
      });
    }
  }
});

