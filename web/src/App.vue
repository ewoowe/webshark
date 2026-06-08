<template>
  <div class="app-container">
    <!-- 侧边栏 -->
    <aside class="sidebar" :class="{ collapsed: sidebarCollapsed }">
      <div class="sidebar-header">
        <h1 class="app-title" v-if="!sidebarCollapsed">🦈 WebShark</h1>
        <span class="app-title-icon" v-if="sidebarCollapsed">🦈</span>
        <p class="app-subtitle" v-if="!sidebarCollapsed">网络抓包分析工具</p>
      </div>

      <nav class="sidebar-nav">
        <router-link
          v-for="route in navRoutes"
          :key="route.path"
          :to="route.path"
          class="nav-item"
          :class="{ active: isActive(route.path) }"
        >
          <span class="nav-icon">{{ route.meta?.icon }}</span>
          <span class="nav-text" v-if="!sidebarCollapsed">{{ route.meta?.title }}</span>
        </router-link>
      </nav>

      <div class="sidebar-footer">
        <button class="collapse-btn" @click="toggleSidebar">
          <span>{{ sidebarCollapsed ? '→' : '←' }}</span>
        </button>
      </div>
    </aside>

    <!-- 主内容区 -->
    <div class="main-content">
      <!-- 顶部导航栏 -->
      <header class="top-header">
        <div class="header-left">
          <button class="menu-toggle" @click="toggleSidebar">
            <span>☰</span>
          </button>
          <h2 class="page-title">{{ pageTitle }}</h2>
        </div>
        <div class="header-right">
          <div class="status-indicator" :class="connectionStatus">
            <span class="status-dot"></span>
            <span class="status-text">{{ statusText }}</span>
          </div>
          <div class="user-info">
            <span class="user-avatar">👤</span>
            <span class="user-name">Admin</span>
          </div>
        </div>
      </header>

      <!-- 内容区域 -->
      <main class="content-area">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, provide } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import router from '@/router';

const route = useRoute();
const vueRouter = useRouter();
const sidebarCollapsed = ref(false);
const connectionStatus = ref<'connected' | 'disconnected' | 'connecting'>('disconnected');

// 导航路由列表
const navRoutes = router.getRoutes().filter(r => r.meta?.title);

const isActive = (path: string) => {
  return route.path === path || (path === '/dashboard' && route.path === '/');
};

const pageTitle = computed(() => {
  return (route.meta?.title as string) || '仪表盘';
});

const statusText = computed(() => {
  const statusMap = {
    connected: '已连接',
    disconnected: '未连接',
    connecting: '连接中'
  };
  return statusMap[connectionStatus.value];
});

const navigateTo = (path: string) => {
  vueRouter.push(path);
};

const toggleSidebar = () => {
  sidebarCollapsed.value = !sidebarCollapsed.value;
};

const setConnectionStatus = (status: 'connected' | 'disconnected' | 'connecting') => {
  connectionStatus.value = status;
};

provide('appContext', {
  setConnectionStatus,
  navigateTo,
});

defineExpose({
  setConnectionStatus,
  navigateTo,
});
</script>

<style scoped>
.app-container {
  display: flex;
  height: 100vh;
  overflow: hidden;
  background: #f0f2f5;
}

/* 侧边栏样式 */
.sidebar {
  width: 200px;
  background: linear-gradient(180deg, #0f1a2e 0%, #141e30 30%, #1a2840 100%);
  color: white;
  display: flex;
  flex-direction: column;
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  flex-shrink: 0;
  box-shadow: 2px 0 20px rgba(0, 0, 0, 0.15);
  position: relative;
}

.sidebar::after {
  content: '';
  position: absolute;
  right: 0;
  top: 0;
  bottom: 0;
  width: 1px;
  background: linear-gradient(180deg, rgba(102, 126, 234, 0.3), transparent 50%, rgba(118, 75, 162, 0.3));
}

.sidebar.collapsed {
  width: 100px;
}

.sidebar.collapsed .app-subtitle {
  display: none;
}

.app-title-icon {
  display: none;
  font-size: 28px;
  line-height: 1;
  text-align: center;
}

.sidebar.collapsed .app-title-icon {
  display: block;
}

.sidebar-header {
  padding: 24px 20px 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  text-align: center;
}

.app-title {
  font-size: 22px;
  margin: 0;
  font-weight: 800;
  letter-spacing: 1px;
  background: linear-gradient(135deg, #a78bfa 0%, #7c8cf8 40%, #667eea 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.app-subtitle {
  font-size: 16px;
  color: rgba(255, 255, 255, 0.4);
  margin: 8px 0 0 0;
  font-weight: 400;
  letter-spacing: 2px;
  text-transform: uppercase;
}

/* 侧边栏导航 */
.sidebar-nav {
  flex: 1;
  padding: 16px 12px;
  overflow-y: auto;
}

.nav-item {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  margin-bottom: 4px;
  cursor: pointer;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  color: rgba(255, 255, 255, 0.55);
  border-left: 3px solid transparent;
  border-radius: 8px;
  text-decoration: none;
  font-weight: 500;
  position: relative;
  overflow: hidden;
}

.nav-item::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  background: linear-gradient(180deg, #667eea, #764ba2);
  border-radius: 0 3px 3px 0;
  opacity: 0;
  transform: scaleY(0.6);
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}

.nav-item:hover {
  color: rgba(255, 255, 255, 0.9);
  background: rgba(255, 255, 255, 0.06);
}

.nav-item:hover::before {
  opacity: 0.5;
  transform: scaleY(1);
}

.nav-item.active {
  color: #fff;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.25), rgba(118, 75, 162, 0.15));
  border-left-color: transparent;
  box-shadow: inset 0 0 20px rgba(102, 126, 234, 0.1);
}

.nav-item.active::before {
  opacity: 1;
  transform: scaleY(1);
}

.nav-icon {
  font-size: 24px;
  min-width: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.25s ease;
}

.nav-item:hover .nav-icon {
  transform: scale(1.1);
}

.nav-item.active .nav-icon {
  filter: drop-shadow(0 0 6px rgba(102, 126, 234, 0.5));
}

.nav-text {
  margin-left: 10px;
  white-space: nowrap;
  font-size: 15px;
  letter-spacing: 0.5px;
}

/* 侧边栏底部 */
.sidebar-footer {
  padding: 12px 16px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
}

.collapse-btn {
  width: 100%;
  padding: 10px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  color: rgba(255, 255, 255, 0.5);
  cursor: pointer;
  border-radius: 8px;
  transition: all 0.3s;
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 1px;
}

.collapse-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  color: rgba(255, 255, 255, 0.85);
  border-color: rgba(255, 255, 255, 0.15);
}

/* 主内容区 */
.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* 顶部导航栏 */
.top-header {
  background: white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  padding: 0 24px;
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.menu-toggle {
  background: none;
  border: none;
  font-size: 20px;
  cursor: pointer;
  padding: 8px;
  border-radius: 4px;
  transition: background 0.3s;
}

.menu-toggle:hover {
  background: rgba(0, 0, 0, 0.05);
}

.page-title {
  margin: 0;
  font-size: 18px;
  color: #333;
  font-weight: 500;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 20px;
}

/* 状态指示器 */
.status-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: #f0f2f5;
  border-radius: 20px;
  font-size: 14px;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #999;
}

.status-indicator.connected .status-dot {
  background: #52c41a;
  box-shadow: 0 0 0 2px rgba(82, 196, 26, 0.2);
}

.status-indicator.disconnected .status-dot {
  background: #f5222d;
}

.status-indicator.connecting .status-dot {
  background: #faad14;
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.status-text {
  color: #666;
  font-weight: 500;
}

/* 用户信息 */
.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  background: #f0f2f5;
  border-radius: 20px;
  cursor: pointer;
  transition: all 0.3s;
}

.user-info:hover {
  background: #e6f7ff;
}

.user-avatar {
  font-size: 18px;
}

.user-name {
  font-size: 14px;
  color: #333;
  font-weight: 500;
}

/* 内容区域 */
.content-area {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .sidebar {
    position: absolute;
    height: 100%;
    z-index: 1000;
    transform: translateX(-100%);
  }

  .sidebar.mobile-open {
    transform: translateX(0);
  }

  .main-content {
    width: 100%;
  }

  .content-area {
    padding: 16px;
  }
}
</style>