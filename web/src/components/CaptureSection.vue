<template>
  <section class="panel" v-if="visible">
    <div class="capture-header">
      <h2>抓包结果</h2>
      <div class="capture-controls">
        <span class="badge">{{ packetCount }} 个数据包</span>
        <button class="btn btn-danger" @click="handleStopCapture">
          停止抓包
        </button>
        <button class="btn btn-secondary" @click="handleClearPackets">
          清空
        </button>
      </div>
    </div>

    <!-- 数据包列表 - 使用VXE-Table实现虚拟滚动 -->
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
      <h3>数据包详情</h3>
      <pre class="packet-detail">{{ selectedPacketDetail }}</pre>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, reactive, onUnmounted } from 'vue';
import type { Packet } from '@/types';
import type { VxeGridInstance, VxeGridPropTypes, VxeColumnSlotTypes } from 'vxe-table';

interface Props {
  visible: boolean;
}

defineProps<Props>();
const emit = defineEmits<{
  stopCapture: [];
  clearPackets: [];
}>();

// 数据包列表 - 限制最大数量防止内存泄漏
const MAX_PACKETS = 10000;
const packets = ref<Packet[]>([]);
const selectedPacketIndex = ref<number>(-1);
const gridRef = ref<VxeGridInstance>();

// 滚动节流：限制 scrollToBottom 的调用频率
let scrollTimer: ReturnType<typeof setTimeout> | null = null;
const SCROLL_THROTTLE_MS = 100;

// 表格列配置
const columns = reactive<VxeGridPropTypes.Columns<Packet>>([
  {
    field: 'frame',
    title: 'No.',
    width: 80,
    align: 'center',
  },
  {
    field: 'timestamp',
    title: 'Time',
    width: 200,
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

// 获取行样式
const getRowClassName = ({ rowIndex }: { rowIndex: number }) => {
  return rowIndex === selectedPacketIndex.value ? 'selected-row' : '';
};

// 处理单元格点击
const handleCellClick = ({ rowIndex }: { rowIndex: number }) => {
  selectedPacketIndex.value = rowIndex;
};

// 处理当前行变化
const handleCurrentChange = ({ rowIndex }: { rowIndex: number }) => {
  selectedPacketIndex.value = rowIndex;
};

// 添加数据包（带数量限制，防止内存溢出）
const addPacket = (packet: Packet) => {
  // 超出上限时移除最旧的数据
  if (packets.value.length >= MAX_PACKETS) {
    packets.value.shift();
    // 如果当前选中的是被移除的包，则取消选中
    if (selectedPacketIndex.value === 0) {
      selectedPacketIndex.value = -1;
      if (gridRef.value) {
        gridRef.value.clearCurrentRow();
      }
    } else if (selectedPacketIndex.value > 0) {
      // 调整索引：所有索引前移一位
      selectedPacketIndex.value--;
    }
  }

  packets.value.push(packet);

  // 节流滚动：高频数据包到达时避免频繁 DOM 操作
  if (scrollTimer === null) {
    scrollTimer = setTimeout(() => {
      scrollTimer = null;
      nextTick(() => {
        scrollToBottom();
      });
    }, SCROLL_THROTTLE_MS);
  }
};

// 滚动到底部
const scrollToBottom = () => {
  if (gridRef.value) {
    const tableEl = gridRef.value.$el;
    const bodyEl = tableEl.querySelector('.vxe-table--body-wrapper');
    if (bodyEl) {
      bodyEl.scrollTop = bodyEl.scrollHeight;
    }
  }
};

const handleStopCapture = () => {
  emit('stopCapture');
};

const handleClearPackets = () => {
  packets.value = [];
  selectedPacketIndex.value = -1;
  if (gridRef.value) {
    gridRef.value.clearCurrentRow();
  }
  emit('clearPackets');
};

// 组件卸载时清理定时器，防止内存泄漏
onUnmounted(() => {
  if (scrollTimer !== null) {
    clearTimeout(scrollTimer);
    scrollTimer = null;
  }
});

defineExpose({
  addPacket,
  clear: () => {
    packets.value = [];
    selectedPacketIndex.value = -1;
    if (gridRef.value) {
      gridRef.value.clearCurrentRow();
    }
  },
});
</script>

<style scoped>
.panel {
  background: white;
  border-radius: 10px;
  padding: 25px;
  margin-bottom: 20px;
  box-shadow: 0 10px 30px rgba(0,0,0,0.2);
}

.panel h2 {
  color: #333;
  margin-bottom: 20px;
  padding-bottom: 10px;
  border-bottom: 2px solid #667eea;
}

.capture-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.capture-controls {
  display: flex;
  align-items: center;
  gap: 15px;
}

.badge {
  background: #667eea;
  color: white;
  padding: 6px 12px;
  border-radius: 20px;
  font-size: 14px;
  font-weight: 600;
}

.packet-table-container {
  margin-bottom: 20px;
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  overflow: hidden;
}

/* VXE-Table 自定义样式 */
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
  transition: background-color 0.2s;
}

.packet-table-container :deep(.vxe-body--row:hover) {
  background-color: #f8f9fa !important;
}

.packet-table-container :deep(.vxe-body--row.selected-row) {
  background-color: #e3f2fd !important;
}

.packet-table-container :deep(.vxe-body--row.selected-row:hover) {
  background-color: #bbdefb !important;
}

.packet-table-container :deep(.vxe-body--column) {
  padding: 10px 12px;
}

.packet-table-container :deep(.vxe-body--column) {
  border-bottom: 1px solid #e0e0e0;
}

.packet-table-container :deep(.no-data) {
  padding: 40px;
  text-align: center;
  color: #888;
  font-size: 14px;
}

/* 自定义滚动条样式 */
.packet-table-container :deep(.vxe-table--body-wrapper)::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

.packet-table-container :deep(.vxe-table--body-wrapper)::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 4px;
}

.packet-table-container :deep(.vxe-table--body-wrapper)::-webkit-scrollbar-thumb {
  background: #888;
  border-radius: 4px;
}

.packet-table-container :deep(.vxe-table--body-wrapper)::-webkit-scrollbar-thumb:hover {
  background: #555;
}

.packet-detail-container {
  background: #f8f9fa;
  border: 1px solid #e0e0e0;
  border-radius: 6px;
  padding: 20px;
}

.packet-detail-container h3 {
  margin-bottom: 15px;
  color: #333;
}

.packet-detail {
  background: #2d2d2d;
  color: #f8f8f2;
  padding: 15px;
  border-radius: 6px;
  overflow-x: auto;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  max-height: 400px;
  overflow-y: auto;
}

.btn {
  padding: 12px 24px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
  margin-right: 10px;
}

.btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 5px 15px rgba(0,0,0,0.2);
}

.btn-secondary {
  background: #6c757d;
  color: white;
}

.btn-danger {
  background: linear-gradient(135deg, #eb3349 0%, #f45c43 100%);
  color: white;
}
</style>