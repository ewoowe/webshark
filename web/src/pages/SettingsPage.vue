<template>
  <div class="settings-page">
    <!-- 系统设置 -->
    <div class="card">
      <div class="card-header">
        <h3 class="card-title">系统设置</h3>
      </div>
      <div class="card-body">
        <div class="settings-list">
          <div class="setting-item">
            <div class="setting-info">
              <span class="setting-label">深色模式</span>
              <span class="setting-desc">启用深色主题界面</span>
            </div>
            <div class="setting-control">
              <input type="checkbox" v-model="settings.darkMode" @change="updateSettings" />
            </div>
          </div>

          <div class="setting-item">
            <div class="setting-info">
              <span class="setting-label">自动滚动</span>
              <span class="setting-desc">新数据包到达时自动滚动到底部</span>
            </div>
            <div class="setting-control">
              <input type="checkbox" v-model="settings.autoScroll" @change="updateSettings" />
            </div>
          </div>

          <div class="setting-item">
            <div class="setting-info">
              <span class="setting-label">声音提示</span>
              <span class="setting-desc">操作成功/失败时播放提示音</span>
            </div>
            <div class="setting-control">
              <input type="checkbox" v-model="settings.soundEnabled" @change="updateSettings" />
            </div>
          </div>

          <div class="setting-item">
            <div class="setting-info">
              <span class="setting-label">显示时间戳</span>
              <span class="setting-desc">在数据包列表中显示详细时间戳</span>
            </div>
            <div class="setting-control">
              <input type="checkbox" v-model="settings.showTimestamp" @change="updateSettings" />
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 性能设置 -->
    <div class="card">
      <div class="card-header">
        <h3 class="card-title">性能设置</h3>
      </div>
      <div class="card-body">
        <div class="settings-list">
          <div class="setting-item">
            <div class="setting-info">
              <span class="setting-label">虚拟滚动阈值</span>
              <span class="setting-desc">启用虚拟滚动的数据包数量阈值</span>
            </div>
            <div class="setting-control">
              <input
                type="number"
                class="form-input"
                v-model.number="settings.virtualScrollThreshold"
                min="10"
                max="1000"
                @change="updateSettings"
              />
            </div>
          </div>

          <div class="setting-item">
            <div class="setting-info">
              <span class="setting-label">最大数据包数</span>
              <span class="setting-desc">单次会话最大保存的数据包数量</span>
            </div>
            <div class="setting-control">
              <input
                type="number"
                class="form-input"
                v-model.number="settings.maxPackets"
                min="100"
                max="100000"
                @change="updateSettings"
              />
            </div>
          </div>

          <div class="setting-item">
            <div class="setting-info">
              <span class="setting-label">刷新频率</span>
              <span class="setting-desc">数据刷新间隔（毫秒）</span>
            </div>
            <div class="setting-control">
              <select class="form-select" v-model="settings.refreshRate" @change="updateSettings">
                <option value="100">100ms</option>
                <option value="250">250ms</option>
                <option value="500">500ms</option>
                <option value="1000">1000ms</option>
              </select>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 网络设置 -->
    <div class="card">
      <div class="card-header">
        <h3 class="card-title">网络设置</h3>
      </div>
      <div class="card-body">
        <div class="settings-list">
          <div class="setting-item">
            <div class="setting-info">
              <span class="setting-label">连接超时</span>
              <span class="setting-desc">网络连接超时时间（秒）</span>
            </div>
            <div class="setting-control">
              <input
                type="number"
                class="form-input"
                v-model.number="settings.connectionTimeout"
                min="5"
                max="120"
                @change="updateSettings"
              />
            </div>
          </div>

          <div class="setting-item">
            <div class="setting-info">
              <span class="setting-label">重试次数</span>
              <span class="setting-desc">连接失败后的自动重试次数</span>
            </div>
            <div class="setting-control">
              <input
                type="number"
                class="form-input"
                v-model.number="settings.retryCount"
                min="0"
                max="10"
                @change="updateSettings"
              />
            </div>
          </div>

          <div class="setting-item">
            <div class="setting-info">
              <span class="setting-label">WebSocket 心跳</span>
              <span class="setting-desc">WebSocket 心跳间隔（秒）</span>
            </div>
            <div class="setting-control">
              <input
                type="number"
                class="form-input"
                v-model.number="settings.heartbeatInterval"
                min="10"
                max="300"
                @change="updateSettings"
              />
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 关于信息 -->
    <div class="card">
      <div class="card-header">
        <h3 class="card-title">关于</h3>
      </div>
      <div class="card-body">
        <div class="about-info">
          <div class="about-item">
            <span class="about-label">版本</span>
            <span class="about-value">1.0.0</span>
          </div>
          <div class="about-item">
            <span class="about-label">技术栈</span>
            <span class="about-value">Vue 3 + TypeScript + Vite</span>
          </div>
          <div class="about-item">
            <span class="about-label">数据表格</span>
            <span class="about-value">VXE-Table</span>
          </div>
          <div class="about-item">
            <span class="about-label">开发团队</span>
            <span class="about-value">WebShark Team</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 操作按钮 -->
    <div class="action-buttons">
      <button class="btn btn-primary" @click="resetSettings">
        重置设置
      </button>
      <button class="btn btn-secondary" @click="exportSettings">
        导出设置
      </button>
      <button class="btn btn-secondary" @click="importSettings">
        导入设置
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue';

const settings = reactive({
  darkMode: false,
  autoScroll: true,
  soundEnabled: false,
  showTimestamp: true,
  virtualScrollThreshold: 20,
  maxPackets: 10000,
  refreshRate: 250,
  connectionTimeout: 30,
  retryCount: 3,
  heartbeatInterval: 30,
});

const updateSettings = () => {
  // 保存设置到 localStorage
  localStorage.setItem('webshark_settings', JSON.stringify(settings));
  console.log('设置已更新:', settings);
};

const resetSettings = () => {
  if (confirm('确定要重置所有设置为默认值吗？')) {
    Object.assign(settings, {
      darkMode: false,
      autoScroll: true,
      soundEnabled: false,
      showTimestamp: true,
      virtualScrollThreshold: 20,
      maxPackets: 10000,
      refreshRate: 250,
      connectionTimeout: 30,
      retryCount: 3,
      heartbeatInterval: 30,
    });
    updateSettings();
    alert('设置已重置');
  }
};

const exportSettings = () => {
  const settingsJson = JSON.stringify(settings, null, 2);
  const blob = new Blob([settingsJson], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = 'webshark_settings.json';
  a.click();
  URL.revokeObjectURL(url);
  alert('设置已导出');
};

const importSettings = () => {
  const input = document.createElement('input');
  input.type = 'file';
  input.accept = '.json';
  input.onchange = (e: any) => {
    const file = e.target.files[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (event: any) => {
        try {
          const importedSettings = JSON.parse(event.target.result);
          Object.assign(settings, importedSettings);
          updateSettings();
          alert('设置已导入');
        } catch (error) {
          alert('导入失败：无效的设置文件');
        }
      };
      reader.readAsText(file);
    }
  };
  input.click();
};

// 从 localStorage 加载设置
const loadSettings = () => {
  const savedSettings = localStorage.getItem('webshark_settings');
  if (savedSettings) {
    try {
      const parsedSettings = JSON.parse(savedSettings);
      Object.assign(settings, parsedSettings);
    } catch (error) {
      console.error('加载设置失败:', error);
    }
  }
};

// 初始化时加载设置
loadSettings();
</script>

<style scoped>
.settings-page {
  animation: fadeIn 0.3s ease-in-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 卡片样式 */
.card {
  background: white;
  border-radius: 8px;
  margin-bottom: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  overflow: hidden;
}

.card-header {
  padding: 20px 24px;
  border-bottom: 1px solid #f0f2f5;
}

.card-title {
  margin: 0;
  font-size: 16px;
  color: #333;
  font-weight: 600;
}

.card-body {
  padding: 24px;
}

/* 设置列表 */
.settings-list {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.setting-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 0;
  border-bottom: 1px solid #f0f2f5;
}

.setting-item:last-child {
  border-bottom: none;
}

.setting-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.setting-label {
  font-size: 14px;
  color: #333;
  font-weight: 500;
}

.setting-desc {
  font-size: 12px;
  color: #999;
}

.setting-control {
  margin-left: 20px;
}

/* 表单控件 */
.form-input {
  padding: 8px 12px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  font-size: 14px;
  width: 120px;
  transition: border-color 0.3s;
}

.form-input:focus {
  outline: none;
  border-color: #667eea;
}

.form-select {
  padding: 8px 12px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  font-size: 14px;
  width: 120px;
  background: white;
  cursor: pointer;
  transition: border-color 0.3s;
}

.form-select:focus {
  outline: none;
  border-color: #667eea;
}

/* 复选框样式 */
input[type="checkbox"] {
  width: 20px;
  height: 20px;
  cursor: pointer;
  accent-color: #667eea;
}

/* 关于信息 */
.about-info {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
}

.about-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.about-label {
  font-size: 14px;
  color: #999;
}

.about-value {
  font-size: 16px;
  color: #333;
  font-weight: 600;
}

/* 操作按钮 */
.action-buttons {
  display: flex;
  gap: 12px;
  padding: 20px 0;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s;
}

.btn:hover {
  opacity: 0.85;
}

.btn-primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.btn-secondary {
  background: #f0f2f5;
  color: #333;
}
</style>