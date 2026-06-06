<template>
  <div class="container">
    <Header />

    <ConnectionSection
      @connect="handleConnect"
    />

    <InterfaceSection
      v-model="selectedInterfaces"
      :visible="showInterfaceSection"
      :interfaces="interfaces"
    />

    <FilterSection
      ref="filterSectionRef"
      :visible="showFilterSection"
      @start-capture="handleStartCapture"
    />

    <CaptureSection
      ref="captureSectionRef"
      :visible="showCaptureSection"
      @stop-capture="handleStopCapture"
      @clear-packets="handleClearPackets"
    />

    <StatusMessage ref="statusMessageRef" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import Header from './components/Header.vue';
import ConnectionSection from './components/ConnectionSection.vue';
import InterfaceSection from './components/InterfaceSection.vue';
import FilterSection from './components/FilterSection.vue';
import CaptureSection from './components/CaptureSection.vue';
import StatusMessage from './components/StatusMessage.vue';
import { ApiService } from './services/api.service';
import { WebSocketService } from './services/websocket.service';
import type { NetworkInterface, Packet } from './types';

// UI state
const showInterfaceSection = ref(false);
const showFilterSection = ref(false);
const showCaptureSection = ref(false);

// Data
const interfaces = ref<NetworkInterface[]>([]);
const selectedInterfaces = ref<string[]>([]);

// Component refs
const statusMessageRef = ref<InstanceType<typeof StatusMessage> | null>(null);
const filterSectionRef = ref<InstanceType<typeof FilterSection> | null>(null);
const captureSectionRef = ref<InstanceType<typeof CaptureSection> | null>(null);

// WebSocket service
const wsService = new WebSocketService();

// Session state
let sessionId: string | null = null;

// Show status message
const showStatus = (message: string, type: 'success' | 'error' | 'info' = 'info') => {
  statusMessageRef.value?.show(message, type);
};

// Handle connection
const handleConnect = async (host: string, username: string, password: string) => {
  try {
    showStatus('正在获取网卡列表...', 'info');

    const response = await ApiService.getInterfaces(host, username, password);

    if (response.success && response.data) {
      interfaces.value = response.data as NetworkInterface[];
      showInterfaceSection.value = true;
      showFilterSection.value = true;
      showStatus(`成功获取 ${response.data.length} 个网卡`, 'success');
    } else {
      showStatus(`获取网卡失败: ${response.error}`, 'error');
    }
  } catch (error) {
    showStatus(`网络错误: ${error instanceof Error ? error.message : 'Unknown error'}`, 'error');
  }
};

// Start capture
const handleStartCapture = async () => {
  const connectionForm = getConnectionForm();
  const filters = filterSectionRef.value?.getFilters();

  if (!connectionForm) {
    showStatus('请填写连接信息', 'error');
    return;
  }

  if (selectedInterfaces.value.length === 0) {
    showStatus('请选择至少一个网卡', 'error');
    return;
  }

  try {
    showStatus('正在启动抓包...', 'info');

    const response = await ApiService.startCapture({
      host: connectionForm.host,
      username: connectionForm.username,
      password: connectionForm.password,
      interfaces: selectedInterfaces.value,
      bpf_filter: filters?.bpfFilter || '',
      wireshark_filter: filters?.wiresharkFilter || '',
    });

    if (response.success && response.message) {
      sessionId = response.message;
      showCaptureSection.value = true;
      showStatus('抓包已启动', 'success');

      // Connect WebSocket
      connectWebSocket(sessionId);
    } else {
      showStatus(`启动抓包失败: ${response.error}`, 'error');
    }
  } catch (error) {
    showStatus(`网络错误: ${error instanceof Error ? error.message : 'Unknown error'}`, 'error');
  }
};

// Connect WebSocket
const connectWebSocket = (sid: string) => {
  wsService.onOpen(() => {
    showStatus('WebSocket 连接成功', 'success');
  });

  wsService.onMessage((packet: Packet) => {
    captureSectionRef.value?.addPacket(packet);
  });

  wsService.onError((error) => {
    showStatus('WebSocket 连接错误', 'error');
    console.error('WebSocket error:', error);
  });

  wsService.onClose(() => {
    showStatus('WebSocket 连接已关闭', 'info');
  });

  wsService.connect(sid);
};

// Stop capture
const handleStopCapture = async () => {
  if (!sessionId) {
    showStatus('没有活跃的抓包会话', 'error');
    return;
  }

  try {
    const response = await ApiService.stopCapture(sessionId);

    if (response.success) {
      showStatus('抓包已停止', 'success');

      // Close WebSocket
      wsService.disconnect();
      sessionId = null;
    } else {
      showStatus(`停止抓包失败: ${response.error}`, 'error');
    }
  } catch (error) {
    showStatus(`网络错误: ${error instanceof Error ? error.message : 'Unknown error'}`, 'error');
  }
};

// Clear packets
const handleClearPackets = () => {
  captureSectionRef.value?.clear();
  showStatus('已清空数据包列表', 'success');
};

// Get connection form data
const getConnectionForm = () => {
  const host = document.getElementById('host') as HTMLInputElement;
  const username = document.getElementById('username') as HTMLInputElement;
  const password = document.getElementById('password') as HTMLInputElement;

  if (!host || !username || !password) {
    return null;
  }

  return {
    host: host.value,
    username: username.value,
    password: password.value,
  };
};

// Cleanup on unmount
import { onUnmounted } from 'vue';

onUnmounted(() => {
  wsService.disconnect();
});
</script>

<style scoped>
.container {
  max-width: 1400px;
  margin: 0 auto;
}
</style>