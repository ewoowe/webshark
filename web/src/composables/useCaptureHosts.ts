import { reactive } from 'vue';
import type { CaptureHost } from '@/types';

// 共享的抓包主机状态，用于在主机管理页和抓包页之间传递数据
export const captureHostStore = reactive<{
  hosts: CaptureHost[];
}>({
  hosts: [],
});

export function useCaptureHosts() {
  const addHosts = (hosts: CaptureHost[]) => {
    // 去重添加
    const existingIds = new Set(captureHostStore.hosts.map(h => h.id));
    for (const host of hosts) {
      if (!existingIds.has(host.id)) {
        captureHostStore.hosts.push(host);
        existingIds.add(host.id);
      }
    }
  };

  const removeHost = (hostId: number) => {
    const index = captureHostStore.hosts.findIndex(h => h.id === hostId);
    if (index > -1) {
      captureHostStore.hosts.splice(index, 1);
    }
  };

  const clearHosts = () => {
    captureHostStore.hosts.splice(0);
  };

  const hasHosts = () => captureHostStore.hosts.length > 0;

  const hostCount = () => captureHostStore.hosts.length;

  return {
    store: captureHostStore,
    addHosts,
    removeHost,
    clearHosts,
    hasHosts,
    hostCount,
  };
}
