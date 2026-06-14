<template>
  <div class="capture-page">
    <!-- 抓包主机列表 -->
    <div class="card">
      <div class="card-header collapsible-header" @click="hostListExpanded = !hostListExpanded">
        <h3 class="card-title">
          <span class="collapse-arrow" :class="{ collapsed: !hostListExpanded }">▶</span>
          抓包主机
        </h3>
        <div class="card-actions">
          <button class="btn btn-sm btn-secondary" @click.stop="goBackToHosts">+ 添加主机</button>
        </div>
      </div>
      <div class="card-body" v-show="hostListExpanded">
        <div v-if="store.hosts.length === 0" class="empty-hosts">
          <div class="empty-icon">📡</div>
          <p>暂无抓包主机，请先在"主机管理"页面选择主机</p>
          <button class="btn btn-primary" @click="goBackToHosts">前往选择</button>
        </div>
        <div v-else class="host-task-list">
          <div
            v-for="host in store.hosts"
            :key="host.id"
            class="host-task-card"
          >
            <!-- 主机信息头 -->
            <div class="host-header collapsible-header" @click="toggleHostExpand(host.id)">
              <div class="host-info">
                <span class="collapse-arrow" :class="{ collapsed: !hostExpanded[host.id] }">▶</span>
                <span class="host-icon">🖥️</span>
                <div>
                  <div class="host-name">{{ host.hostName }}</div>
                  <div class="host-meta">{{ host.ip }} / {{ host.userName }}</div>
                </div>
              </div>
              <div class="host-actions" @click.stop>
                <button class="btn btn-sm btn-secondary" @click="loadInterfaces(host)">加载网卡</button>
                <button class="btn btn-sm btn-primary" @click="addTask(host)">+ 添加抓包流</button>
              </div>
            </div>

            <!-- 网卡列表 + 抓包流配置列表（可折叠） -->
            <div v-show="hostExpanded[host.id]">
              <div v-if="hostInterfaces[host.id] && hostInterfaces[host.id]!.length > 0" class="interface-chip-list">
                <span class="chip-label">可用网卡：</span>
                <span
                  v-for="iface in hostInterfaces[host.id]"
                  :key="iface.name"
                  class="chip"
                >{{ iface.name }} ({{ iface.ip }})</span>
              </div>

              <div class="task-streams">
                <div
                  v-for="(task, idx) in getHostTasks(host.id)"
                  :key="task.streamId"
                  class="task-stream-item"
                >
                  <div class="stream-header">
                    <span class="stream-label">抓包流 #{{ task.streamId }}</span>
                    <button class="btn btn-sm btn-danger" @click="removeTask(host.id, idx)">移除</button>
                  </div>
                  <div class="stream-config">
                    <div class="form-row">
                      <div class="form-group">
                        <label class="form-label">网卡</label>
                        <div class="interface-checkboxes">
                          <label
                            v-for="iface in (hostInterfaces[host.id] || [])"
                            :key="iface.name"
                            class="checkbox-label"
                          >
                            <input
                              type="checkbox"
                              :checked="task.interfaces.includes(iface.name)"
                              @change="toggleTaskInterface(host.id, idx, iface.name)"
                            />
                            {{ iface.name }}
                          </label>
                          <span v-if="(!hostInterfaces[host.id] || hostInterfaces[host.id]!.length === 0)" class="text-muted">
                            请先加载网卡
                          </span>
                        </div>
                      </div>
                    </div>
                    <div class="form-row">
                      <div class="form-group">
                        <label class="form-label">BPF 过滤器</label>
                        <input
                          type="text"
                          class="form-input"
                          :value="task.bpfFilter"
                          @input="updateTaskField(host.id, idx, 'bpfFilter', ($event.target as HTMLInputElement).value)"
                          placeholder="例如: tcp port 80"
                        />
                      </div>
                      <div class="form-group">
                        <label class="form-label">Wireshark 过滤器</label>
                        <input
                          type="text"
                          class="form-input"
                          :value="task.wiresharkFilter"
                          @input="updateTaskField(host.id, idx, 'wiresharkFilter', ($event.target as HTMLInputElement).value)"
                          placeholder="例如: http"
                        />
                      </div>
                    </div>
                  </div>
                </div>
                <div v-if="getHostTasks(host.id).length === 0" class="no-stream">
                  暂无抓包流配置，请点击"+ 添加抓包流"
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 全局配置和操作 -->
    <div class="card" v-if="store.hosts.length > 0">
      <div class="card-header collapsible-header" @click="captureControlExpanded = !captureControlExpanded">
        <h3 class="card-title">
          <span class="collapse-arrow" :class="{ collapsed: !captureControlExpanded }">▶</span>
          抓包控制
        </h3>
      </div>
      <div class="card-body" v-show="captureControlExpanded">
        <div class="global-config">
          <div class="form-row">
            <div class="form-group">
              <label class="form-label">任务名称</label>
              <input type="text" class="form-input" v-model="taskName" placeholder="例如: 生产环境抓包任务" />
            </div>
            <div class="form-group">
              <label class="form-label">详情格式</label>
              <select class="form-select" v-model="detailFormat">
                <option value="normal">normal</option>
                <option value="json">json</option>
                <option value="xml">xml</option>
              </select>
            </div>
          </div>
          <div class="form-row inline-checkboxes">
            <label class="checkbox-label">
              <input type="checkbox" v-model="onlyCapture" />
              仅捕获（不解析）
            </label>
            <label class="checkbox-label">
              <input type="checkbox" v-model="parseDetail" />
              解析详情
            </label>
          </div>
        </div>
        <div class="capture-actions">
          <button class="btn btn-success" @click="startCapture" :disabled="!canStart || capturing">
            {{ capturing ? '启动中...' : '🚀 开始抓包' }}
          </button>
          <button class="btn btn-secondary" @click="resetAll">重置</button>
        </div>
      </div>
    </div>

    <!-- 停止确认弹窗 -->
    <div v-if="showStopDialog" class="modal-overlay" @click.self="cancelStop">
      <div class="modal-dialog modal-sm">
        <div class="modal-header">
          <h3>停止抓包</h3>
          <button class="modal-close" @click="cancelStop">✕</button>
        </div>
        <div class="modal-body">
          <p class="stop-dialog-text">请选择停止方式：</p>
          <div class="stop-options">
            <button
              v-if="taskGroupId"
              class="btn btn-warning stop-option-btn"
              :class="{ selected: stopDialogType === 'group' }"
              @click="stopDialogType = 'group'"
            >
              <span class="stop-option-icon">📦</span>
              <div class="stop-option-content">
                <div class="stop-option-title">停止整个任务组</div>
                <div class="stop-option-desc">停止任务组 #{{ taskGroupId }} 下的所有任务</div>
              </div>
            </button>
            <button
              class="btn stop-option-btn"
              :class="{ selected: stopDialogType === 'single', 'btn-warning': !taskGroupId }"
              @click="stopDialogType = 'single'"
            >
              <span class="stop-option-icon">📌</span>
              <div class="stop-option-content">
                <div class="stop-option-title">{{ taskGroupId ? '仅停止当前任务' : '停止任务' }}</div>
                <div class="stop-option-desc">{{ taskGroupId ? `停止当前任务 #${currentTaskId}` : `停止任务 #${currentTaskId}` }}</div>
              </div>
            </button>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="cancelStop">取消</button>
          <button
            class="btn btn-danger"
            :disabled="stopDialogType === null"
            @click="confirmStop"
          >确认停止</button>
        </div>
      </div>
    </div>

    <!-- 抓包结果卡片 -->
    <div class="card" v-if="capturing || showCaptureSection">
      <div class="card-header">
        <h3 class="card-title">抓包结果
          <span class="badge" v-if="taskGroupId">任务组: {{ taskGroupId }}</span>
          <span class="badge" v-else-if="currentTaskId">任务: {{ currentTaskId }}</span>
        </h3>
        <div class="card-actions">
          <span class="badge">{{ packetCount }} 个数据包</span>
          <button
            class="btn"
            :class="autoScroll ? 'btn-primary' : 'btn-secondary'"
            @click="toggleAutoScroll"
          >
            {{ autoScroll ? '✓ 跟随最新' : '跟随最新' }}
          </button>
          <button class="btn btn-danger" @click="handleStopCapture">
            停止抓包
          </button>
          <button class="btn btn-secondary" @click="handleClearPackets">
            清空
          </button>
        </div>
      </div>
      <div class="card-body">
        <div class="packet-table-container">
          <vxe-grid
            ref="gridRef"
            :data="packets"
            :columns="columns"
            :row-config="{ isCurrent: true, isHover: true }"
            :scroll-y="{ enabled: true, gt: 20 }"
            height="600"
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
        <div class="packet-detail-container">
          <h4 class="detail-title">数据包详情</h4>
          <pre class="packet-detail">{{ selectedPacketDetail }}</pre>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, nextTick, inject, onMounted, onBeforeUnmount, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import type { Packet, NetworkInterface, CaptureTask, HostCaptureConfig, CaptureRequest } from '@/types';
import type { VxeGridInstance, VxeGridProps } from 'vxe-table';
import { CaptureService } from '@/services/capture.service';
import { WebSocketService } from '@/services/websocket.service';
import { TaskService } from '@/services/task.service';
import { HostService } from '@/services/host.service';
import { useCaptureHosts } from '@/composables/useCaptureHosts';

const router = useRouter();
const route = useRoute();
const { store, clearHosts, removeHost } = useCaptureHosts();
const appContext = inject('appContext') as any;

// 每个主机的网卡列表
const hostInterfaces = reactive<Record<number, NetworkInterface[]>>({});

// 每个主机的抓包流任务配置
const hostTasks = reactive<Record<number, CaptureTask[]>>({});

// 全局配置
const taskName = ref('');
const detailFormat = ref('normal');
const onlyCapture = ref(false);
const parseDetail = ref(true);

// 抓包状态
const capturing = ref(false);
const showCaptureSection = ref(false);

// 任务标识
const taskGroupId = ref<number | null>(null);
const currentTaskId = ref<number | null>(null);

// WebSocket
const wsService = new WebSocketService();

// 数据包
const packets = ref<Packet[]>([]);
const selectedPacketIndex = ref<number>(-1);
const gridRef = ref<VxeGridInstance>();
const autoScroll = ref(true);

// 列配置
const columns = reactive<VxeGridProps.Columns>([
  { field: 'no', title: 'No.', width: 70, align: 'center' },
  { field: 'frameNumber', title: 'Frame', width: 70, align: 'center' },
  { field: 'timestamp', title: 'Time', width: 180, align: 'left' },
  { field: 'source', title: 'Source', width: 140, align: 'left',
    slots: { default: ({ row }: { row: Packet }) => row.ip4Src || row.ip6Src || row.ethSrc || '' }
  },
  { field: 'dest', title: 'Destination', width: 140, align: 'left',
    slots: { default: ({ row }: { row: Packet }) => row.ip4Dst || row.ip6Dst || row.ethDst || '' }
  },
  { field: 'protocol', title: 'Protocol', width: 90, align: 'center' },
  { field: 'length', title: 'Length', width: 80, align: 'center' },
  { field: 'info', title: 'Info', align: 'left', showOverflow: true },
]);

const packetCount = computed(() => packets.value.length);
const selectedPacketDetail = ref('点击数据包查看详细信息');
const detailLoading = ref(false);

// 是否可以开始抓包
const canStart = computed(() => {
  if (store.hosts.length === 0) return false;
  for (const host of store.hosts) {
    const tasks = hostTasks[host.id] || [];
    if (tasks.length === 0) return false;
    for (const t of tasks) {
      if (t.interfaces.length === 0) return false;
    }
  }
  return true;
});

// 获取某主机的任务列表
const getHostTasks = (hostId: number): CaptureTask[] => {
  return hostTasks[hostId] || [];
};

// 加载网卡
const loadInterfaces = async (host: { id: number }) => {
  try {
    const res = await CaptureService.getInterfaces(host.id);
    if (res.code === 0 && res.data) {
      hostInterfaces[host.id] = res.data;
    } else {
      alert('获取网卡失败: ' + res.message);
    }
  } catch (e: any) {
    alert('网络错误: ' + e.message);
  }
};

// 折叠状态
const hostListExpanded = ref(true); // "抓包主机"整张卡片
const captureControlExpanded = ref(true); // "抓包控制"整张卡片
const hostExpanded = reactive<Record<number, boolean>>({}); // 每台主机的折叠状态

// 切换主机折叠
const toggleHostExpand = (hostId: number) => {
  if (hostExpanded[hostId] === undefined) {
    hostExpanded[hostId] = true; // 默认展开
  }
  hostExpanded[hostId] = !hostExpanded[hostId];
};

// 全局抓包流 ID 计数器
let globalStreamIdCounter = 0;

// 生成下一个全局唯一的 streamId
const nextStreamId = (): number => {
  return ++globalStreamIdCounter;
};

// 添加抓包流
const addTask = (host: { id: number }) => {
  if (!hostTasks[host.id]) {
    hostTasks[host.id] = [];
  }
  hostTasks[host.id].push({
    streamId: nextStreamId(),
    interfaces: [],
    bpfFilter: '',
    wiresharkFilter: '',
  });
};

// 移除抓包流并重新编号（全局）
const removeTask = (hostId: number, idx: number) => {
  hostTasks[hostId].splice(idx, 1);
  renumberAllTasks();
};

// 全局重新编号所有主机的所有抓包流，确保从 1 开始顺序递增
const renumberAllTasks = () => {
  globalStreamIdCounter = 0;
  for (const host of store.hosts) {
    const tasks = hostTasks[host.id];
    if (!tasks) continue;
    for (const task of tasks) {
      task.streamId = nextStreamId();
    }
  }
};

// 更新抓包流字段
const updateTaskField = (hostId: number, idx: number, field: keyof CaptureTask, value: any) => {
  const tasks = hostTasks[hostId];
  if (tasks && tasks[idx]) {
    (tasks[idx] as any)[field] = value;
  }
};

// 切换网卡选择
const toggleTaskInterface = (hostId: number, idx: number, ifaceName: string) => {
  const tasks = hostTasks[hostId];
  if (!tasks || !tasks[idx]) return;
  const task = tasks[idx];
  const i = task.interfaces.indexOf(ifaceName);
  if (i > -1) {
    task.interfaces.splice(i, 1);
  } else {
    task.interfaces.push(ifaceName);
  }
};

// 开始抓包
const startCapture = async () => {
  if (!canStart.value) {
    alert('请确保每个主机至少有一个抓包流且选择了网卡');
    return;
  }

  capturing.value = true;
  try {
    const hostCaptures: HostCaptureConfig[] = store.hosts.map(host => ({
      hostId: host.id,
      captures: (hostTasks[host.id] || []).map(t => ({
        streamId: t.streamId,
        interfaces: t.interfaces,
        bpfFilter: t.bpfFilter,
        wiresharkFilter: t.wiresharkFilter,
      })),
    }));

    const req: CaptureRequest = {
      taskName: taskName.value || `抓包任务_${Date.now()}`,
      onlyCapture: onlyCapture.value,
      parseDetail: parseDetail.value,
      detailFormat: detailFormat.value,
      hostCaptures,
    };

    const res = await CaptureService.startCapture(req);
    console.log('[CapturePage] startCapture response:', res);
    if (res.code === 0 && res.data) {
      const info = res.data;
      console.log('[CapturePage] taskInfo:', info);
      taskGroupId.value = info.taskGroupId > 0 ? info.taskGroupId : null;
      currentTaskId.value = info.taskIds ? Object.values(info.taskIds)[0] || null : null;

      console.log('[CapturePage] taskGroupId:', taskGroupId.value, 'currentTaskId:', currentTaskId.value);

      showCaptureSection.value = true;
      alert('抓包已启动');

      // 连接 WebSocket
      if (taskGroupId.value) {
        console.log('[CapturePage] connecting WebSocket via taskGroup:', taskGroupId.value);
        connectWebSocket('taskGroup', taskGroupId.value);
      } else if (currentTaskId.value) {
        console.log('[CapturePage] connecting WebSocket via task:', currentTaskId.value);
        connectWebSocket('task', currentTaskId.value);
      } else {
        console.warn('[CapturePage] No valid taskGroupId or taskId to connect WebSocket!');
      }
    } else {
      alert('启动抓包失败: ' + res.message);
    }
  } catch (e: any) {
    alert('启动抓包失败: ' + e.message);
  } finally {
    capturing.value = false;
  }
};

// WebSocket 连接
const connectWebSocket = (taskType: string, id: number) => {
  console.log('[CapturePage] connectWebSocket called, taskType:', taskType, 'id:', id);
  wsService.onOpen(() => {
    console.log('[CapturePage] WebSocket 连接成功');
    appContext?.setConnectionStatus('connected');
  });
  wsService.onClose(() => {
    console.log('[CapturePage] WebSocket 连接关闭');
    appContext?.setConnectionStatus('disconnected');
  });
  wsService.onError((error) => {
    console.error('WebSocket 连接错误:', error);
    appContext?.setConnectionStatus('disconnected');
  });
  wsService.onMessage((packet: Packet) => {
    packets.value.push(packet);
    scrollToBottom();
  });
  wsService.connect(taskType, String(id));
};

// 停止确认弹窗
const showStopDialog = ref(false);
const stopDialogType = ref<'single' | 'group' | null>(null);

// 停止抓包 - 弹出确认对话框
const handleStopCapture = () => {
  if (taskGroupId.value) {
    // 有任务组：让用户选择停止任务组还是仅停止当前任务
    stopDialogType.value = null;
    showStopDialog.value = true;
  } else {
    // 没有任务组：直接确认停止
    stopDialogType.value = 'single';
    showStopDialog.value = true;
  }
};

// 确认停止
const confirmStop = async () => {
  if (stopDialogType.value === null) return;
  try {
    const isGroupStop = stopDialogType.value === 'group';
    const res = await CaptureService.stopCapture(
      isGroupStop ? taskGroupId.value || undefined : undefined,
      !isGroupStop ? (currentTaskId.value || undefined) : undefined,
    );
    if (res.code === 0) {
      alert('抓包已停止');
      wsService.disconnect();
      showStopDialog.value = false;
    } else {
      alert('停止失败: ' + res.message);
    }
  } catch (e: any) {
    alert('停止失败: ' + e.message);
  }
};

// 取消停止
const cancelStop = () => {
  showStopDialog.value = false;
  stopDialogType.value = null;
};

// 清空数据包
const handleClearPackets = () => {
  packets.value = [];
  selectedPacketIndex.value = -1;
  gridRef.value?.clearCurrentRow();
};

// 重置所有
const resetAll = () => {
  clearHosts();
  for (const key of Object.keys(hostInterfaces)) {
    delete hostInterfaces[Number(key)];
  }
  for (const key of Object.keys(hostTasks)) {
    delete hostTasks[Number(key)];
  }
  taskName.value = '';
  detailFormat.value = 'normal';
  onlyCapture.value = false;
  parseDetail.value = true;
};

// 跳回主机管理
const goBackToHosts = () => {
  router.push('/hosts');
};

// 表格事件
const handleCellClick = async ({ row, rowIndex }: { row: Packet; rowIndex: number }) => {
  selectedPacketIndex.value = rowIndex;
  await fetchPacketDetail(row);
};

const handleCurrentChange = async ({ rowIndex }: { rowIndex: number }) => {
  selectedPacketIndex.value = rowIndex;
  const packet = packets.value[rowIndex];
  if (packet) {
    await fetchPacketDetail(packet);
  }
};

// 获取数据包详情
const fetchPacketDetail = async (packet: Packet) => {
  detailLoading.value = true;
  selectedPacketDetail.value = '加载中...';
  try {
    // 用 packet.taskId 和 packet.frameNumber 请求详情
    const res = await CaptureService.getPacketDetail(packet.taskId, packet.frameNumber);
    if (res.code === 0 && res.data) {
      selectedPacketDetail.value = res.data;
    } else {
      selectedPacketDetail.value = `获取详情失败: ${res.message || '未知错误'}`;
    }
  } catch (e: any) {
    selectedPacketDetail.value = `获取详情失败: ${e.message}`;
  } finally {
    detailLoading.value = false;
  }
};

const getRowClassName = ({ rowIndex }: { rowIndex: number }) => {
  return rowIndex === selectedPacketIndex.value ? 'selected-row' : '';
};

const toggleAutoScroll = () => {
  autoScroll.value = !autoScroll.value;
  if (autoScroll.value) {
    scrollToBottom(true);
  }
};

const scrollToBottom = (force = false) => {
  if (!force && !autoScroll.value) return;
  const lastRow = packets.value[packets.value.length - 1];
  if (!lastRow) return;
  nextTick(() => {
    nextTick(() => {
      if (gridRef.value) {
        gridRef.value.scrollToRow(lastRow);
      }
    });
  });
};

// 从任务管理页跳转过来，恢复抓包配置和 WebSocket 连接
const restoreFromTask = async () => {
  const queryTaskId = route.query.taskId;
  const queryTaskGroupId = route.query.taskGroupId;

  if (!queryTaskId && !queryTaskGroupId) return;

  try {
    let tasks: import('@/types').Task[] = [];

    if (queryTaskGroupId) {
      // 按任务组恢复
      const groupId = Number(queryTaskGroupId);
      const res = await TaskService.listTasksByGroupId(groupId, 1, 100);
      if (res.code === 0 && res.data) {
        tasks = res.data.items || [];
        taskGroupId.value = groupId;
      }
    } else if (queryTaskId) {
      // 按单个任务恢复
      const taskId = Number(queryTaskId);
      const res = await TaskService.getTask(taskId);
      if (res.code === 0 && res.data) {
        tasks = [res.data];
        currentTaskId.value = taskId;
      }
    }

    if (tasks.length === 0) {
      console.warn('[CapturePage] restoreFromTask: 未找到相关任务');
      return;
    }

    // 收集唯一的 hostId，加载主机信息和网卡
    const hostIdSet = new Set(tasks.map(t => t.hostId));
    for (const hostId of hostIdSet) {
      // 获取主机信息并加入 store
      const hostRes = await HostService.getHost(hostId);
      if (hostRes.code === 0 && hostRes.data) {
        const host = hostRes.data;
        const { addHosts } = useCaptureHosts();
        addHosts([{
          id: host.id,
          hostName: host.hostName,
          ip: host.ip,
          userName: host.userName,
        }]);

        // 加载网卡
        const ifRes = await CaptureService.getInterfaces(host.id);
        if (ifRes.code === 0 && ifRes.data) {
          hostInterfaces[host.id] = ifRes.data;
        }
      }
    }

    // 恢复每个主机的抓包流配置
    for (const task of tasks) {
      if (!hostTasks[task.hostId]) {
        hostTasks[task.hostId] = [];
      }
      hostTasks[task.hostId].push({
        streamId: task.streamId,
        interfaces: task.interfaces || [],
        bpfFilter: task.bpfFilter || '',
        wiresharkFilter: task.wiresharkFilter || '',
      });
    }

    // 恢复全局配置（取第一个任务的配置）
    const firstTask = tasks[0];
    taskName.value = firstTask.taskName || '';
    detailFormat.value = firstTask.detailFormat || 'normal';
    onlyCapture.value = firstTask.onlyCapture || false;
    parseDetail.value = firstTask.parseDetail || false;

    // 更新全局 streamId 计数器
    let maxStreamId = 0;
    for (const t of tasks) {
      if (t.streamId > maxStreamId) maxStreamId = t.streamId;
    }
    globalStreamIdCounter = maxStreamId;

    // 显示抓包结果区域
    showCaptureSection.value = true;

    // 检查并断开旧的 WebSocket 连接，然后建立新连接
    if (wsService.isConnected()) {
      console.log('[CapturePage] restoreFromTask: 检测到已有 WebSocket 连接，先断开旧连接');
      wsService.disconnect();
    }

    if (taskGroupId.value) {
      console.log('[CapturePage] restoreFromTask: connecting WebSocket via taskGroup:', taskGroupId.value);
      connectWebSocket('taskGroup', taskGroupId.value);
    } else if (currentTaskId.value) {
      console.log('[CapturePage] restoreFromTask: connecting WebSocket via task:', currentTaskId.value);
      connectWebSocket('task', currentTaskId.value);
    }
  } catch (e: any) {
    console.error('[CapturePage] restoreFromTask 失败:', e);
    alert('恢复抓包配置失败: ' + e.message);
  }
};

onMounted(() => {
  restoreFromTask();
});

// 监听路由 query 变化：当 taskId/taskGroupId 变化时，先断开旧连接再恢复
watch(
  () => ({ taskId: route.query.taskId, taskGroupId: route.query.taskGroupId }),
  (newVal, oldVal) => {
    // 仅在已有旧连接且 query 确实变化时处理
    if (oldVal && (newVal.taskId !== oldVal.taskId || newVal.taskGroupId !== oldVal.taskGroupId)) {
      if (wsService.isConnected()) {
        console.log('[CapturePage] 路由 query 变化，断开旧 WebSocket 连接');
        wsService.disconnect();
      }
      // 重置数据包
      packets.value = [];
      selectedPacketIndex.value = -1;
      restoreFromTask();
    }
  },
);

onBeforeUnmount(() => {
  if (wsService.isConnected()) {
    console.log('[CapturePage] 组件卸载，断开 WebSocket 连接');
    wsService.disconnect();
  }
});
</script>

<style scoped>
.capture-page { animation: fadeIn 0.3s ease-in-out; }
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.card {
  background: white;
  border-radius: 8px;
  margin-bottom: 20px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.06);
  overflow: hidden;
}
.card-header {
  padding: 20px 24px;
  border-bottom: 1px solid #f0f2f5;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.collapsible-header {
  cursor: pointer;
  user-select: none;
  transition: background 0.2s;
}
.collapsible-header:hover {
  background: #f5f7fa;
}
.card-title { margin: 0; font-size: 16px; color: #333; font-weight: 600; display: flex; align-items: center; gap: 12px; }
.card-actions { display: flex; align-items: center; gap: 12px; }
.card-body { padding: 24px; }

/* 折叠箭头 */
.collapse-arrow {
  display: inline-block;
  font-size: 10px;
  transition: transform 0.25s ease;
  color: #999;
}
.collapse-arrow.collapsed {
  transform: rotate(-90deg);
}

/* 空状态 */
.empty-hosts { text-align: center; padding: 60px 20px; }
.empty-icon { font-size: 48px; margin-bottom: 16px; }
.empty-hosts p { color: #999; margin-bottom: 20px; }

/* 主机任务卡片 */
.host-task-card {
  border: 1px solid #f0f2f5;
  border-radius: 8px;
  margin-bottom: 16px;
  overflow: hidden;
}
.host-task-card:last-child { margin-bottom: 0; }

.host-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: #fafbfc;
  border-bottom: 1px solid #f0f2f5;
}
.host-info { display: flex; align-items: center; gap: 12px; }
.host-icon { font-size: 24px; }
.host-name { font-size: 15px; font-weight: 600; color: #333; }
.host-meta { font-size: 12px; color: #999; margin-top: 2px; }
.host-actions { display: flex; gap: 8px; }

/* 网卡 chip */
.interface-chip-list {
  padding: 10px 20px;
  background: #f9f9f9;
  border-bottom: 1px solid #f0f2f5;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
}
.chip-label { font-size: 12px; color: #666; }
.chip {
  display: inline-block;
  padding: 3px 10px;
  background: #e6f7ff;
  border: 1px solid #91d5ff;
  border-radius: 12px;
  font-size: 12px;
  color: #0066cc;
}

/* 抓包流 */
.task-streams { padding: 0; }
.task-stream-item {
  padding: 16px 20px;
  border-bottom: 1px solid #f0f2f5;
}
.task-stream-item:last-child { border-bottom: none; }

.stream-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
.stream-label { font-size: 14px; font-weight: 600; color: #667eea; }

.no-stream {
  padding: 24px;
  text-align: center;
  color: #999;
  font-size: 14px;
}

/* 表单 */
.form-row {
  display: flex;
  gap: 16px;
  margin-bottom: 12px;
}
.form-row:last-child { margin-bottom: 0; }
.form-row.inline-checkboxes {
  align-items: center;
  gap: 24px;
}
.form-group { flex: 1; display: flex; flex-direction: column; gap: 6px; }
.form-label { font-size: 13px; color: #666; font-weight: 500; }
.form-input, .form-select {
  padding: 8px 12px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  font-size: 13px;
  transition: border-color 0.3s;
}
.form-input:focus, .form-select:focus { outline: none; border-color: #667eea; }

.interface-checkboxes {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}
.checkbox-label {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  cursor: pointer;
}
.text-muted { color: #999; font-size: 12px; }

/* 全局配置 */
.global-config { margin-bottom: 20px; }

/* 抓包操作按钮 */
.capture-actions { display: flex; gap: 12px; }

/* 按钮 */
.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s;
  display: flex;
  align-items: center;
  gap: 6px;
}
.btn:hover { opacity: 0.85; }
.btn:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-primary { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; }
.btn-success { background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%); color: white; }
.btn-danger { background: linear-gradient(135deg, #eb3349 0%, #f45c43 100%); color: white; }
.btn-secondary { background: #f0f2f5; color: #333; }
.btn-sm { padding: 6px 12px; font-size: 12px; }

/* 徽章 */
.badge {
  background: #667eea;
  color: white;
  padding: 4px 10px;
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
.packet-table-container :deep(.vxe-table) { font-size: 12px; }
.packet-table-container :deep(.vxe-body--row) { height: 32px; }
.packet-table-container :deep(.vxe-body--column) { padding: 0 8px; height: 32px; line-height: 32px; }
.packet-table-container :deep(.vxe-table--header) { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; }
.packet-table-container :deep(.vxe-table--header .vxe-header--column) { color: white; font-weight: 600; background: transparent; padding: 0 8px; height: 36px; }
.packet-table-container :deep(.vxe-body--row) { cursor: pointer; }
.packet-table-container :deep(.vxe-body--row:hover) { background-color: #f9f9f9 !important; }
.packet-table-container :deep(.vxe-body--row.selected-row) { background-color: #f0f5ff !important; }
.packet-table-container :deep(.no-data) { padding: 40px; text-align: center; color: #999; font-size: 14px; }

.packet-detail-container {
  background: #f9f9f9;
  border: 1px solid #f0f2f5;
  border-radius: 8px;
  padding: 20px;
}
.detail-title { font-size: 14px; color: #333; margin: 0 0 16px 0; font-weight: 600; }
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

/* 模态框 */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}
.modal-dialog {
  background: white;
  border-radius: 12px;
  width: 560px;
  max-height: 80vh;
  overflow-y: auto;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
}
.modal-sm {
  width: 440px;
}
.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #f0f2f5;
}
.modal-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #333;
}
.modal-close {
  border: none;
  background: none;
  font-size: 18px;
  cursor: pointer;
  color: #999;
  padding: 4px;
}
.modal-close:hover { color: #333; }
.modal-body {
  padding: 24px;
}
.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding: 16px 24px;
  border-top: 1px solid #f0f2f5;
}

/* 停止弹窗 */
.stop-dialog-text {
  margin: 0 0 16px 0;
  font-size: 14px;
  color: #555;
}
.stop-options {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.stop-option-btn {
  display: flex;
  align-items: center;
  gap: 14px;
  width: 100%;
  height: auto;
  padding: 16px 20px;
  border: 2px solid #e8e8e8;
  border-radius: 8px;
  background: white;
  cursor: pointer;
  text-align: left;
  transition: all 0.2s;
}
.stop-option-btn:hover {
  border-color: #d9d9d9;
  background: #fafafa;
}
.stop-option-btn.selected {
  border-color: #4a90d9;
  background: #f0f5ff;
}
.stop-option-btn.btn-warning {
  border-color: #fa8c16;
}
.stop-option-btn.btn-warning.selected {
  border-color: #fa8c16;
  background: #fff7e6;
}
.stop-option-icon {
  font-size: 28px;
  flex-shrink: 0;
}
.stop-option-content {
  flex: 1;
}
.stop-option-title {
  font-size: 14px;
  font-weight: 600;
  color: #333;
  margin-bottom: 4px;
}
.stop-option-desc {
  font-size: 12px;
  color: #999;
}
</style>
