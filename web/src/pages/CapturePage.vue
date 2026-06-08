<template>
  <div class="capture-page">
    <!-- 连接配置卡片 -->
    <div class="card">
      <div class="card-header">
        <h3 class="card-title">远程主机连接</h3>
        <div class="card-actions">
          <button class="btn-sm" @click="showHelp">❓</button>
        </div>
      </div>
      <div class="card-body">
        <div class="form-grid">
          <div class="form-item">
            <label class="form-label">主机地址</label>
            <input
              type="text"
              class="form-input"
              v-model="connectionForm.host"
              placeholder="例如: 192.168.1.100"
            />
          </div>
          <div class="form-item">
            <label class="form-label">用户名</label>
            <input
              type="text"
              class="form-input"
              v-model="connectionForm.username"
              placeholder="例如: root"
            />
          </div>
          <div class="form-item">
            <label class="form-label">密码</label>
            <input
              type="password"
              class="form-input"
              v-model="connectionForm.password"
              placeholder="输入密码"
            />
          </div>
          <div class="form-item form-actions">
            <button class="btn btn-primary" @click="handleConnect">
              获取网卡列表
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 网卡选择卡片 -->
    <div class="card" v-if="showInterfaceSection">
      <div class="card-header">
        <h3 class="card-title">选择网卡</h3>
        <div class="card-actions">
          <button class="btn-sm" @click="selectAllInterfaces">全选</button>
          <button class="btn-sm" @click="deselectAllInterfaces">取消全选</button>
        </div>
      </div>
      <div class="card-body">
        <div class="interface-grid">
          <div
            v-for="(iface, index) in interfaces"
            :key="index"
            class="interface-item"
            :class="{ selected: selectedInterfaces.includes(iface.name) }"
            @click="toggleInterface(iface.name)"
          >
            <span class="interface-icon">🌐</span>
            <div class="interface-info">
              <span class="interface-name">{{ iface.name }}</span>
              <span class="interface-ip">{{ iface.ip }}</span>
            </div>
            <span class="interface-check" v-if="selectedInterfaces.includes(iface.name)">✓</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 过滤器配置卡片 -->
    <div class="card" v-if="showFilterSection">
      <div class="card-header">
        <h3 class="card-title">过滤器配置</h3>
      </div>
      <div class="card-body">
        <div class="form-grid">
          <div class="form-item full-width">
            <label class="form-label">BPF 过滤器 (tcpdump)</label>
            <input
              type="text"
              class="form-input"
              v-model="filters.bpfFilter"
              placeholder="例如: tcp port 80"
            />
            <small class="form-hint">常用示例: tcp port 80, host 192.168.1.1, not port 22</small>
          </div>
          <div class="form-item full-width">
            <label class="form-label">Wireshark 过滤器</label>
            <input
              type="text"
              class="form-input"
              v-model="filters.wiresharkFilter"
              placeholder="例如: http.request.method == GET"
            />
            <small class="form-hint">常用示例: http, dns, ip.addr == 192.168.1.1</small>
          </div>
          <div class="form-item full-width form-actions">
            <button class="btn btn-success" @click="handleStartCapture">
              开始抓包
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 抓包结果卡片 -->
    <div class="card" v-if="showCaptureSection">
      <div class="card-header">
        <h3 class="card-title">抓包结果</h3>
        <div class="card-actions">
          <span class="badge">{{ packetCount }} 个数据包</span>
          <button class="btn btn-danger" @click="handleStopCapture">
            停止抓包
          </button>
          <button class="btn btn-secondary" @click="handleClearPackets">
            清空
          </button>
        </div>
      </div>
      <div class="card-body">
        <!-- 使用VXE-Table实现虚拟滚动 -->
        <div class="packet-table-container">
          <vxe-grid
            ref="gridRef"
            :data="packets"
            :columns="columns"
            :row-config="{ isCurrent: true, isHover: true }"
            :scroll-y="{ enabled: true, gt: 20 }"
            height="400"
            :show-header="true"
            :stripe="true"
            :border="true"
            :show-overflow="true"
            :row-class-name="getRowClassName"
            @cell-click="handleCellClick"
            @current-change="handleCurrentChange"
          >
            <template #noData>
              <div class="no-data">暂无数据包</div>
            </template>
          </vxe-grid>
        </div>

        <!-- 数据包详情 -->
        <div class="packet-detail-container">
          <h4 class="detail-title">数据包详情</h4>
          <pre class="packet-detail">{{ selectedPacketDetail }}</pre>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, nextTick } from 'vue';
import { inject } from 'vue';
import type { Packet, NetworkInterface } from '@/types';
import type { VxeGridInstance, VxeGridProps } from 'vxe-table';
import { ApiService } from '@/services/api.service';
import { WebSocketService } from '@/services/websocket.service';

// 注入父组件的上下文
const appContext = inject('appContext') as any;

// UI 状态
const showInterfaceSection = ref(false);
const showFilterSection = ref(false);
const showCaptureSection = ref(false);

// 数据
const interfaces = ref<NetworkInterface[]>([]);
const selectedInterfaces = ref<string[]>([]);
const packets = ref<Packet[]>([]);
const selectedPacketIndex = ref<number>(-1);

// 表单数据
const connectionForm = reactive({
  host: '',
  username: '',
  password: '',
});

const filters = reactive({
  bpfFilter: '',
  wiresharkFilter: '',
});

// WebSocket 服务
const wsService = new WebSocketService();
let sessionId: string | null = null;

// 表格引用
const gridRef = ref<VxeGridInstance>();

// 表格列配置
const columns = reactive<VxeGridProps.Columns>([
  {
    field: 'index',
    title: 'No.',
    width: 80,
    align: 'center',
    slots: {
      default: ({ row, $rowIndex }) => $rowIndex + 1,
    },
  },
  {
    field: 'timestamp',
    title: 'Time',
    width: 160,
    align: 'left',
  },
  {
    field: 'source',
    title: 'Source',
    width: 150,
    align: 'left',
  },
  {
    field: 'dest',
    title: 'Destination',
    width: 150,
    align: 'left',
  },
  {
    field: 'protocol',
    title: 'Protocol',
    width: 100,
    align: 'center',
  },
  {
    field: 'length',
    title: 'Length',
    width: 100,
    align: 'center',
  },
  {
    field: 'info',
    title: 'Info',
    align: 'left',
    showOverflow: true,
  },
]);

const packetCount = computed(() => packets.value.length);

const selectedPacketDetail = computed(() => {
  if (selectedPacketIndex.value === -1) {
    return '点击数据包查看详细信息';
  }
  const packet = packets.value[selectedPacketIndex.value];
  return JSON.stringify(packet, null, 2);
});

// 方法
const showHelp = () => {
  alert('连接帮助：\n\n请填写远程主机的SSH连接信息：\n- 主机地址：IP地址或域名\n- 用户名：SSH用户名\n- 密码：SSH密码');
};

const handleConnect = async () => {
  if (!connectionForm.host || !connectionForm.username || !connectionForm.password) {
    alert('请填写所有连接信息');
    return;
  }

  try {
    appContext?.setConnectionStatus('connecting');

    const response = await ApiService.getInterfaces(
      connectionForm.host,
      connectionForm.username,
      connectionForm.password
    );

    if (response.success && response.data) {
      interfaces.value = response.data as NetworkInterface[];
      showInterfaceSection.value = true;
      showFilterSection.value = true;
      appContext?.setConnectionStatus('connected');
      alert(`成功获取 ${response.data.length} 个网卡`);
    } else {
      appContext?.setConnectionStatus('disconnected');
      alert(`获取网卡失败: ${response.error}`);
    }
  } catch (error) {
    appContext?.setConnectionStatus('disconnected');
    alert(`网络错误: ${error instanceof Error ? error.message : 'Unknown error'}`);
  }
};

const toggleInterface = (name: string) => {
  const index = selectedInterfaces.value.indexOf(name);
  if (index > -1) {
    selectedInterfaces.value.splice(index, 1);
  } else {
    selectedInterfaces.value.push(name);
  }
};

const selectAllInterfaces = () => {
  selectedInterfaces.value = interfaces.value.map(iface => iface.name);
};

const deselectAllInterfaces = () => {
  selectedInterfaces.value = [];
};

const handleStartCapture = async () => {
  if (selectedInterfaces.value.length === 0) {
    alert('请选择至少一个网卡');
    return;
  }

  try {
    const response = await ApiService.startCapture({
      host: connectionForm.host,
      username: connectionForm.username,
      password: connectionForm.password,
      interfaces: selectedInterfaces.value,
      bpf_filter: filters.bpfFilter,
      wireshark_filter: filters.wiresharkFilter,
    });

    if (response.success && response.message) {
      sessionId = response.message;
      showCaptureSection.value = true;
      alert('抓包已启动');
      connectWebSocket(sessionId);
    } else {
      alert(`启动抓包失败: ${response.error}`);
    }
  } catch (error) {
    alert(`网络错误: ${error instanceof Error ? error.message : 'Unknown error'}`);
  }
};

const connectWebSocket = (sid: string) => {
  wsService.onOpen(() => {
    console.log('WebSocket 连接已建立');
  });

  wsService.onMessage((packet: Packet) => {
    addPacket(packet);
  });

  wsService.onError(() => {
    console.error('WebSocket 连接错误');
  });

  wsService.onClose(() => {
    console.log('WebSocket 连接已关闭');
  });

  wsService.connect(sid);
};

const addPacket = (packet: Packet) => {
  packets.value.push(packet);
  nextTick(() => {
    scrollToBottom();
  });
};

const handleCellClick = ({ row, rowIndex }: { row: Packet; rowIndex: number }) => {
  selectedPacketIndex.value = rowIndex;
};

const handleCurrentChange = ({ row, rowIndex }: { row: Packet; rowIndex: number }) => {
  selectedPacketIndex.value = rowIndex;
};

const getRowClassName = ({ row, rowIndex }: { row: Packet; rowIndex: number }) => {
  return rowIndex === selectedPacketIndex.value ? 'selected-row' : '';
};

const scrollToBottom = () => {
  if (gridRef.value) {
    const tableEl = gridRef.value.$el;
    const bodyEl = tableEl.querySelector('.vxe-table--body-wrapper');
    if (bodyEl) {
      bodyEl.scrollTop = bodyEl.scrollHeight;
    }
  }
};

const handleStopCapture = async () => {
  if (!sessionId) {
    alert('没有活跃的抓包会话');
    return;
  }

  try {
    const response = await ApiService.stopCapture(sessionId);
    if (response.success) {
      alert('抓包已停止');
      wsService.disconnect();
      sessionId = null;
    } else {
      alert(`停止抓包失败: ${response.error}`);
    }
  } catch (error) {
    alert(`网络错误: ${error instanceof Error ? error.message : 'Unknown error'}`);
  }
};

const handleClearPackets = () => {
  packets.value = [];
  selectedPacketIndex.value = -1;
  if (gridRef.value) {
    gridRef.value.clearCurrentRow();
  }
  alert('已清空数据包列表');
};
</script>

<style scoped>
.capture-page {
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
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-title {
  margin: 0;
  font-size: 16px;
  color: #333;
  font-weight: 600;
}

.card-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.card-body {
  padding: 24px;
}

/* 表单样式 */
.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
}

.form-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-item.full-width {
  grid-column: 1 / -1;
}

.form-item.form-actions {
  justify-content: flex-end;
  align-items: flex-end;
}

.form-label {
  font-size: 14px;
  color: #333;
  font-weight: 500;
}

.form-input {
  padding: 10px 12px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  font-size: 14px;
  transition: border-color 0.3s;
}

.form-input:focus {
  outline: none;
  border-color: #667eea;
}

.form-hint {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}

/* 按钮样式 */
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

.btn-success {
  background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
  color: white;
}

.btn-danger {
  background: linear-gradient(135deg, #eb3349 0%, #f45c43 100%);
  color: white;
}

.btn-secondary {
  background: #f0f2f5;
  color: #333;
}

.btn-sm {
  padding: 6px 12px;
  background: #f0f2f5;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.3s;
}

.btn-sm:hover {
  background: #e6f7ff;
  border-color: #667eea;
}

/* 网卡选择样式 */
.interface-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 16px;
}

.interface-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background: #f9f9f9;
  border: 2px solid transparent;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s;
}

.interface-item:hover {
  background: #f0f2f5;
}

.interface-item.selected {
  border-color: #667eea;
  background: #f0f5ff;
}

.interface-icon {
  font-size: 24px;
}

.interface-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.interface-name {
  font-size: 14px;
  font-weight: 600;
  color: #333;
}

.interface-ip {
  font-size: 12px;
  color: #999;
}

.interface-check {
  width: 24px;
  height: 24px;
  background: #667eea;
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
}

/* 徽章样式 */
.badge {
  background: #667eea;
  color: white;
  padding: 6px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

/* 数据包表格 */
.packet-table-container {
  border: 1px solid #f0f2f5;
  border-radius: 8px;
  overflow: hidden;
  margin-bottom: 20px;
}

.packet-table-container :deep(.vxe-table) {
  font-size: 13px;
}

.packet-table-container :deep(.vxe-table--header) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.packet-table-container :deep(.vxe-table--header .vxe-header--column) {
  color: white;
  font-weight: 600;
  background: transparent;
}

.packet-table-container :deep(.vxe-body--row) {
  cursor: pointer;
}

.packet-table-container :deep(.vxe-body--row:hover) {
  background-color: #f9f9f9 !important;
}

.packet-table-container :deep(.vxe-body--row.selected-row) {
  background-color: #f0f5ff !important;
}

.packet-table-container :deep(.no-data) {
  padding: 40px;
  text-align: center;
  color: #999;
  font-size: 14px;
}

/* 数据包详情 */
.packet-detail-container {
  background: #f9f9f9;
  border: 1px solid #f0f2f5;
  border-radius: 8px;
  padding: 20px;
}

.detail-title {
  font-size: 14px;
  color: #333;
  margin: 0 0 16px 0;
  font-weight: 600;
}

.packet-detail {
  background: #2d2d2d;
  color: #f8f8f2;
  padding: 16px;
  border-radius: 6px;
  overflow-x: auto;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  max-height: 400px;
  overflow-y: auto;
}
</style>