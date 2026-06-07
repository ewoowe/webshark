import { createRouter, createWebHashHistory } from 'vue-router';
import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/dashboard',
  },
  {
    path: '/dashboard',
    name: 'dashboard',
    component: () => import('../pages/DashboardPage.vue'),
    meta: { title: '仪表盘', icon: '📊' },
  },
  {
    path: '/hosts',
    name: 'hosts',
    component: () => import('../pages/HostManagementPage.vue'),
    meta: { title: '主机管理', icon: '🖥️' },
  },
  {
    path: '/capture',
    name: 'capture',
    component: () => import('../pages/CapturePage.vue'),
    meta: { title: '抓包分析', icon: '🕸️' },
  },
  {
    path: '/settings',
    name: 'settings',
    component: () => import('../pages/SettingsPage.vue'),
    meta: { title: '系统设置', icon: '⚙️' },
  },
];

const router = createRouter({
  history: createWebHashHistory(),
  routes,
});

export default router;
