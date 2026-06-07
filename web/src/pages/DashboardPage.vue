<template>
  <div class="dashboard-page">
    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon total-packets">
          <span>📦</span>
        </div>
        <div class="stat-content">
          <h3 class="stat-title">总数据包数</h3>
          <p class="stat-value">{{ totalPackets }}</p>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon active-connections">
          <span>🔗</span>
        </div>
        <div class="stat-content">
          <h3 class="stat-title">活跃连接</h3>
          <p class="stat-value">{{ activeConnections }}</p>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon protocols">
          <span>🌐</span>
        </div>
        <div class="stat-content">
          <h3 class="stat-title">协议数量</h3>
          <p class="stat-value">{{ protocolCount }}</p>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon bandwidth">
          <span>📊</span>
        </div>
        <div class="stat-content">
          <h3 class="stat-title">总流量</h3>
          <p class="stat-value">{{ totalBandwidth }}</p>
        </div>
      </div>
    </div>

    <!-- 快速操作 -->
    <div class="quick-actions">
      <h2 class="section-title">快速操作</h2>
      <div class="action-buttons">
        <button class="action-btn primary" @click="startNewCapture">
          <span class="btn-icon">🚀</span>
          <span class="btn-text">新建抓包</span>
        </button>
        <button class="action-btn secondary" @click="viewRecentPackets">
          <span class="btn-icon">📋</span>
          <span class="btn-text">查看最近数据包</span>
        </button>
        <button class="action-btn secondary" @click="exportData">
          <span class="btn-icon">📤</span>
          <span class="btn-text">导出数据</span>
        </button>
        <button class="action-btn secondary" @click="clearAllData">
          <span class="btn-icon">🗑️</span>
          <span class="btn-text">清空所有数据</span>
        </button>
      </div>
    </div>

    <!-- 最近活动 -->
    <div class="recent-activity">
      <h2 class="section-title">最近活动</h2>
      <div class="activity-list">
        <div v-for="(activity, index) in recentActivities" :key="index" class="activity-item">
          <span class="activity-icon">{{ activity.icon }}</span>
          <div class="activity-content">
            <p class="activity-text">{{ activity.text }}</p>
            <span class="activity-time">{{ activity.time }}</span>
          </div>
        </div>
        <div v-if="recentActivities.length === 0" class="no-activity">
          暂无活动记录
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { inject } from 'vue';

// 注入父组件的方法
const appContext = inject('appContext') as any;

// 模拟数据
const totalPackets = ref(0);
const activeConnections = ref(0);
const protocolCount = ref(0);
const totalBandwidth = ref('0 MB');

const recentActivities = ref<Array<{
  icon: string;
  text: string;
  time: string;
}>>([]);

// 添加活动记录
const addActivity = (icon: string, text: string) => {
  const now = new Date();
  const timeString = `${now.getHours().toString().padStart(2, '0')}:${now.getMinutes().toString().padStart(2, '0')}:${now.getSeconds().toString().padStart(2, '0')}`;

  recentActivities.value.unshift({
    icon,
    text,
    time: timeString,
  });

  // 只保留最近 10 条记录
  if (recentActivities.value.length > 10) {
    recentActivities.value = recentActivities.value.slice(0, 10);
  }
};

// 更新统计数据
const updateStats = (packetCount: number, protocols: string[]) => {
  totalPackets.value = packetCount;
  protocolCount.value = protocols.length;
  activeConnections.value = Math.ceil(Math.random() * 10);

  // 计算总流量（模拟）
  const mbValue = (packetCount * 0.0015).toFixed(2);
  totalBandwidth.value = `${mbValue} MB`;
};

// 快速操作方法
const startNewCapture = () => {
  appContext?.navigateTo('capture');
};

const viewRecentPackets = () => {
  appContext?.navigateTo('capture');
};

const exportData = () => {
  // 实现数据导出功能
  addActivity('📤', '数据导出操作');
};

const clearAllData = () => {
  if (confirm('确定要清空所有数据吗？')) {
    totalPackets.value = 0;
    activeConnections.value = 0;
    protocolCount.value = 0;
    totalBandwidth.value = '0 MB';
    recentActivities.value = [];
    addActivity('🗑️', '已清空所有数据');
  }
};

// 暴露方法给父组件
defineExpose({
  addActivity,
  updateStats,
});
</script>

<style scoped>
.dashboard-page {
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

/* 统计卡片 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin-bottom: 30px;
}

.stat-card {
  background: white;
  border-radius: 8px;
  padding: 24px;
  display: flex;
  align-items: center;
  gap: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  transition: all 0.3s;
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
}

.stat-icon.total-packets {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.stat-icon.active-connections {
  background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
}

.stat-icon.protocols {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.stat-icon.bandwidth {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
}

.stat-content {
  flex: 1;
}

.stat-title {
  font-size: 14px;
  color: #666;
  margin: 0 0 8px 0;
  font-weight: 500;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #333;
  margin: 0;
}

/* 快速操作 */
.quick-actions {
  background: white;
  border-radius: 8px;
  padding: 24px;
  margin-bottom: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.section-title {
  font-size: 16px;
  color: #333;
  margin: 0 0 20px 0;
  font-weight: 600;
}

.action-buttons {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 16px 24px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s;
}

.action-btn.primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.action-btn.primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.action-btn.secondary {
  background: #f0f2f5;
  color: #333;
}

.action-btn.secondary:hover {
  background: #e6f7ff;
  transform: translateY(-2px);
}

.btn-icon {
  font-size: 18px;
}

/* 最近活动 */
.recent-activity {
  background: white;
  border-radius: 8px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.activity-list {
  max-height: 400px;
  overflow-y: auto;
}

.activity-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 0;
  border-bottom: 1px solid #f0f2f5;
  transition: background 0.2s;
}

.activity-item:hover {
  background: #fafafa;
}

.activity-item:last-child {
  border-bottom: none;
}

.activity-icon {
  font-size: 20px;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f0f2f5;
  border-radius: 8px;
}

.activity-content {
  flex: 1;
}

.activity-text {
  font-size: 14px;
  color: #333;
  margin: 0 0 4px 0;
  font-weight: 500;
}

.activity-time {
  font-size: 12px;
  color: #999;
}

.no-activity {
  text-align: center;
  padding: 40px;
  color: #999;
  font-size: 14px;
}

/* 滚动条样式 */
.activity-list::-webkit-scrollbar {
  width: 6px;
}

.activity-list::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

.activity-list::-webkit-scrollbar-thumb {
  background: #d9d9d9;
  border-radius: 3px;
}

.activity-list::-webkit-scrollbar-thumb:hover {
  background: #bfbfbf;
}
</style>