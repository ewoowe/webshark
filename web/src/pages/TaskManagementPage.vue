<template>
  <div class="task-management-page">
    <!-- 搜索栏 -->
    <div class="card">
      <div class="card-body">
        <div class="toolbar">
          <div class="toolbar-left">
            <div class="search-box">
              <select v-model="searchType" class="search-type-select">
                <option value="all">全部任务</option>
                <option value="hostId">按主机ID</option>
                <option value="taskGroupId">按任务组ID</option>
              </select>
              <input
                v-if="searchType !== 'all'"
                type="number"
                class="search-input"
                v-model="searchId"
                placeholder="输入 ID 搜索"
                @keyup.enter="handleSearch"
              />
              <button class="search-btn" @click="handleSearch">🔍</button>
            </div>
          </div>
          <div class="toolbar-right">
            <button class="btn btn-primary" @click="showCreateDialog">
              <span class="btn-icon">➕</span>
              <span class="btn-text">创建任务</span>
            </button>
            <button class="btn btn-secondary" @click="refreshList">
              <span class="btn-icon">🔄</span>
              <span class="btn-text">刷新</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 任务列表 -->
    <div class="card">
      <div class="card-body">
        <div class="table-container">
          <vxe-grid
            ref="gridRef"
            :data="tasks"
            :columns="columns"
            :row-config="{ isCurrent: true, isHover: true }"
            :show-header="true"
            :stripe="true"
            :border="true"
            :show-overflow="true"
            :loading="loading"
          >
            <!-- 状态列模板 -->
            <template #status="{ row }">
              <span class="status-tag" :class="'status-' + row.status">
                {{ statusLabel(row.status) }}
              </span>
            </template>

            <!-- 网卡列模板 -->
            <template #interfaces="{ row }">
              <span v-if="row.interfaces && row.interfaces.length > 0">
                <span v-for="(iface, i) in row.interfaces" :key="i" class="interface-tag">
                  {{ iface }}
                </span>
              </span>
              <span v-else class="text-muted">any</span>
            </template>

            <!-- 创建时间列模板 -->
            <template #createdAt="{ row }">
              {{ formatTime(row.createdAt) }}
            </template>

            <!-- 停止时间列模板 -->
            <template #stopAt="{ row }">
              <span v-if="row.stopAt">{{ formatTime(row.stopAt) }}</span>
              <span v-else class="text-muted">-</span>
            </template>

            <!-- 操作列模板 -->
            <template #actions="{ row }">
              <div class="action-buttons">
                <button class="action-btn" @click="viewTask(row)">详情</button>
                <button
                  class="action-btn action-btn-warning"
                  @click="stopTask(row)"
                >停止</button>
                <button class="action-btn action-btn-danger" @click="deleteTask(row)">删除</button>
              </div>
            </template>
          </vxe-grid>
        </div>

        <!-- 分页 -->
        <div class="pagination">
          <div class="pagination-info">
            共 {{ total }} 条记录，第 {{ page }} / {{ totalPages || 1 }} 页
          </div>
          <div class="pagination-controls">
            <button
              class="page-btn"
              :disabled="page === 1"
              @click="goToPage(1)"
            >首页</button>
            <button
              class="page-btn"
              :disabled="page === 1"
              @click="goToPage(page - 1)"
            >上一页</button>
            <button
              v-for="p in visiblePages" :key="p"
              class="page-btn"
              :class="{ active: p === page }"
              @click="goToPage(p)"
            >{{ p }}</button>
            <button
              class="page-btn"
              :disabled="page >= totalPages"
              @click="goToPage(page + 1)"
            >下一页</button>
            <button
              class="page-btn"
              :disabled="page >= totalPages"
              @click="goToPage(totalPages)"
            >末页</button>
            <select v-model.number="pageSize" @change="handlePageSizeChange" class="page-size-select">
              <option :value="10">10条/页</option>
              <option :value="20">20条/页</option>
              <option :value="50">50条/页</option>
            </select>
          </div>
        </div>
      </div>
    </div>

    <!-- 创建任务对话框 -->
    <div v-if="showCreate" class="modal-overlay" @click.self="showCreate = false">
      <div class="modal-dialog">
        <div class="modal-header">
          <h3>创建任务</h3>
          <button class="modal-close" @click="showCreate = false">✕</button>
        </div>
        <div class="modal-body">
          <div class="form-group">
            <label>任务名称 <span class="required">*</span></label>
            <input v-model="createForm.taskName" class="form-input" placeholder="输入任务名称" />
          </div>
          <div class="form-group">
            <label>主机 ID <span class="required">*</span></label>
            <input v-model.number="createForm.hostId" type="number" class="form-input" placeholder="输入主机 ID" />
          </div>
          <div class="form-group">
            <label>网卡接口</label>
            <input v-model="createForm.interfacesStr" class="form-input" placeholder="多个网卡用逗号分隔，留空为 any" />
          </div>
          <div class="form-group">
            <label>BPF 过滤条件</label>
            <input v-model="createForm.bpfFilter" class="form-input" placeholder="例如: tcp port 80" />
          </div>
          <div class="form-group">
            <label>Wireshark 过滤条件</label>
            <input v-model="createForm.wiresharkFilter" class="form-input" placeholder="例如: http" />
          </div>
          <div class="form-group">
            <label>详细格式</label>
            <select v-model="createForm.detailFormat" class="form-input">
              <option value="normal">normal</option>
              <option value="json">json</option>
              <option value="pdml">pdml</option>
              <option value="ek">ek</option>
            </select>
          </div>
          <div class="form-row">
            <label class="checkbox-label">
              <input v-model="createForm.onlyCapture" type="checkbox" />
              仅抓包不解析
            </label>
            <label class="checkbox-label">
              <input v-model="createForm.parseDetail" type="checkbox" />
              解析详细内容
            </label>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showCreate = false">取消</button>
          <button class="btn btn-primary" @click="handleCreate">创建</button>
        </div>
      </div>
    </div>

    <!-- 任务详情对话框 -->
    <div v-if="showDetail" class="modal-overlay" @click.self="showDetail = false">
      <div class="modal-dialog">
        <div class="modal-header">
          <h3>任务详情 #{{ detailTask?.id }}</h3>
          <button class="modal-close" @click="showDetail = false">✕</button>
        </div>
        <div class="modal-body" v-if="detailTask">
          <div class="detail-grid">
            <div class="detail-item">
              <span class="detail-label">任务名称</span>
              <span class="detail-value">{{ detailTask.taskName }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">主机 ID</span>
              <span class="detail-value">{{ detailTask.hostId }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">任务组 ID</span>
              <span class="detail-value">{{ detailTask.taskGroupId }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">流 ID</span>
              <span class="detail-value">{{ detailTask.streamId }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">状态</span>
              <span class="detail-value">
                <span class="status-tag" :class="'status-' + detailTask.status">
                  {{ statusLabel(detailTask.status) }}
                </span>
              </span>
            </div>
            <div class="detail-item">
              <span class="detail-label">网卡</span>
              <span class="detail-value">{{ detailTask.interfaces?.join(', ') || 'any' }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">BPF 过滤</span>
              <span class="detail-value">{{ detailTask.bpfFilter || '-' }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">Wireshark 过滤</span>
              <span class="detail-value">{{ detailTask.wiresharkFilter || '-' }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">详细格式</span>
              <span class="detail-value">{{ detailTask.detailFormat }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">仅抓包</span>
              <span class="detail-value">{{ detailTask.onlyCapture ? '是' : '否' }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">解析详情</span>
              <span class="detail-value">{{ detailTask.parseDetail ? '是' : '否' }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">文件路径</span>
              <span class="detail-value" style="word-break: break-all;">{{ detailTask.filePath || '-' }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">完整命令</span>
              <span class="detail-value" style="word-break: break-all; font-family: monospace; font-size: 12px;">{{ detailTask.fullCommand || '-' }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">创建时间</span>
              <span class="detail-value">{{ formatTime(detailTask.createdAt) }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">停止时间</span>
              <span class="detail-value">{{ detailTask.stopAt ? formatTime(detailTask.stopAt) : '-' }}</span>
            </div>
          </div>
          <div v-if="detailTask.message" class="detail-message">
            <span class="detail-label">状态消息</span>
            <pre class="detail-value">{{ detailTask.message }}</pre>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="showDetail = false">关闭</button>
          <button
            v-if="detailTask?.status === 'running'"
            class="btn btn-primary"
            @click="viewCapture(detailTask!)"
          >查看抓包</button>
        </div>
      </div>
    </div>

    <!-- 停止确认弹窗 -->
    <div v-if="showStopDialog" class="modal-overlay" @click.self="cancelStop">
      <div class="modal-dialog modal-sm">
        <div class="modal-header">
          <h3>停止任务</h3>
          <button class="modal-close" @click="cancelStop">✕</button>
        </div>
        <div class="modal-body">
          <p class="stop-dialog-text">请选择停止方式：</p>
          <div class="stop-options">
            <button
              v-if="stopTargetTask?.taskGroupId && stopTargetTask.taskGroupId > 0"
              class="btn btn-warning stop-option-btn"
              :class="{ selected: stopDialogType === 'group' }"
              @click="stopDialogType = 'group'"
            >
              <span class="stop-option-icon">📦</span>
              <div class="stop-option-content">
                <div class="stop-option-title">停止整个任务组</div>
                <div class="stop-option-desc">停止任务组 #{{ stopTargetTask?.taskGroupId }} 下的所有任务</div>
              </div>
            </button>
            <button
              class="btn stop-option-btn"
              :class="{ selected: stopDialogType === 'single', 'btn-warning': !stopTargetTask?.taskGroupId }"
              @click="stopDialogType = 'single'"
            >
              <span class="stop-option-icon">📌</span>
              <div class="stop-option-content">
                <div class="stop-option-title">{{ stopTargetTask?.taskGroupId ? '仅停止此任务' : '停止此任务' }}</div>
                <div class="stop-option-desc">仅停止任务 #{{ stopTargetTask?.id }}</div>
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
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import type { VxeGridProps, VxeGridInstance } from 'vxe-table';
import type { Task } from '@/types';
import { TaskService } from '@/services/task.service';

const router = useRouter();

// 表格列定义
const columns: VxeGridProps.Columns = [
  { field: 'id', title: 'ID', width: 70, align: 'center' },
  { field: 'taskName', title: '任务名称', minWidth: 160, align: 'center' },
  { field: 'hostId', title: '主机ID', width: 90, align: 'center' },
  { field: 'taskGroupId', title: '任务组ID', width: 100, align: 'center' },
  { field: 'streamId', title: '流ID', width: 70, align: 'center' },
  { field: 'status', title: '状态', width: 100, align: 'center', slots: { default: 'status' } },
  { field: 'interfaces', title: '网卡', width: 160, align: 'center', slots: { default: 'interfaces' } },
  { field: 'detailFormat', title: '格式', width: 80, align: 'center' },
  { field: 'createdAt', title: '创建时间', width: 170, align: 'center', slots: { default: 'createdAt' } },
  { field: 'stopAt', title: '停止时间', width: 170, align: 'center', slots: { default: 'stopAt' } },
  { field: 'actions', title: '操作', width: 180, align: 'center', slots: { default: 'actions' }, fixed: 'right' },
];

// 数据
const tasks = ref<Task[]>([]);
const loading = ref(false);
const gridRef = ref<VxeGridInstance>();

// 分页
const page = ref(1);
const pageSize = ref(10);
const total = ref(0);

// 搜索
const searchType = ref<'all' | 'hostId' | 'taskGroupId'>('all');
const searchId = ref<number | null>(null);

// 创建任务
const showCreate = ref(false);
const createForm = reactive({
  taskName: '',
  hostId: null as number | null,
  interfacesStr: '',
  bpfFilter: '',
  wiresharkFilter: '',
  detailFormat: 'normal',
  onlyCapture: false,
  parseDetail: false,
});

// 详情
const showDetail = ref(false);
const detailTask = ref<Task | null>(null);

// 状态映射
const statusLabel = (status: string) => {
  const map: Record<string, string> = {
    created: '已创建',
    running: '运行中',
    failed: '失败',
    stopping: '停止中',
    stopped: '已停止',
  };
  return map[status] || status;
};

// 时间格式化
const formatTime = (time: string) => {
  if (!time) return '-';
  return new Date(time).toLocaleString('zh-CN');
};

// 分页计算
const totalPages = computed(() => Math.ceil(total.value / pageSize.value) || 1);

const visiblePages = computed(() => {
  const totalP = totalPages.value;
  const cur = page.value;
  if (totalP <= 7) {
    return Array.from({ length: totalP }, (_, i) => i + 1);
  }
  let start = Math.max(1, cur - 3);
  let end = Math.min(totalP, start + 6);
  if (end - start < 6) {
    start = Math.max(1, end - 6);
  }
  return Array.from({ length: end - start + 1 }, (_, i) => start + i);
});

// 跳转页面
const goToPage = (p: number) => {
  page.value = p;
  loadTasks();
};

// 每页数量变更
const handlePageSizeChange = () => {
  page.value = 1;
  loadTasks();
};

// 加载任务列表
const loadTasks = async () => {
  loading.value = true;
  try {
    let res;
    if (searchType.value === 'hostId' && searchId.value) {
      res = await TaskService.listTasksByHostId(searchId.value, page.value, pageSize.value);
    } else if (searchType.value === 'taskGroupId' && searchId.value) {
      res = await TaskService.listTasksByGroupId(searchId.value, page.value, pageSize.value);
    } else {
      res = await TaskService.listTasks(page.value, pageSize.value);
    }
    if (res.code === 0 && res.data) {
      tasks.value = res.data.items || [];
      total.value = res.data.total || 0;
      page.value = res.data.page || 1;
      pageSize.value = res.data.pageSize || 10;
    } else {
      tasks.value = [];
      total.value = 0;
    }
  } catch (e: any) {
    console.error('加载任务列表失败:', e);
  } finally {
    loading.value = false;
  }
};

// 搜索
const handleSearch = () => {
  page.value = 1;
  loadTasks();
};


// 刷新
const refreshList = () => {
  loadTasks();
};

// 创建
const showCreateDialog = () => {
  createForm.taskName = '';
  createForm.hostId = null;
  createForm.interfacesStr = '';
  createForm.bpfFilter = '';
  createForm.wiresharkFilter = '';
  createForm.detailFormat = 'normal';
  createForm.onlyCapture = false;
  createForm.parseDetail = false;
  showCreate.value = true;
};

const handleCreate = async () => {
  if (!createForm.taskName.trim()) {
    alert('请输入任务名称');
    return;
  }
  if (!createForm.hostId) {
    alert('请输入主机 ID');
    return;
  }
  try {
    const interfaces = createForm.interfacesStr
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean);
    const res = await TaskService.createTask({
      taskName: createForm.taskName.trim(),
      hostId: createForm.hostId,
      interfaces: interfaces.length > 0 ? interfaces : undefined,
      bpfFilter: createForm.bpfFilter || undefined,
      wiresharkFilter: createForm.wiresharkFilter || undefined,
      detailFormat: createForm.detailFormat,
      onlyCapture: createForm.onlyCapture,
      parseDetail: createForm.parseDetail,
    });
    if (res.code === 0) {
      alert('任务创建成功');
      showCreate.value = false;
      loadTasks();
    } else {
      alert('创建失败: ' + res.message);
    }
  } catch (e: any) {
    alert('创建失败: ' + e.message);
  }
};

// 查看详情
const viewTask = (task: Task) => {
  detailTask.value = task;
  showDetail.value = true;
};

// 查看抓包 - 跳转到抓包分析页
const viewCapture = (task: Task) => {
  showDetail.value = false;
  if (task.taskGroupId && task.taskGroupId > 0) {
    router.push({ path: '/capture', query: { taskGroupId: String(task.taskGroupId) } });
  } else {
    router.push({ path: '/capture', query: { taskId: String(task.id) } });
  }
};

// 停止任务 - 弹出确认对话框
const stopTargetTask = ref<Task | null>(null);
const showStopDialog = ref(false);
const stopDialogType = ref<'single' | 'group' | null>(null);

const stopTask = (task: Task) => {
  stopTargetTask.value = task;
  if (task.taskGroupId && task.taskGroupId > 0) {
    stopDialogType.value = null;
  } else {
    stopDialogType.value = 'single';
  }
  showStopDialog.value = true;
};

const confirmStop = async () => {
  const task = stopTargetTask.value;
  if (!task || stopDialogType.value === null) return;
  try {
    const isGroupStop = stopDialogType.value === 'group';
    // 通过 capture API 停止，与 CapturePage 一致
    const { CaptureService } = await import('@/services/capture.service');
    const res = await CaptureService.stopCapture(
      isGroupStop ? task.taskGroupId : undefined,
      !isGroupStop ? task.id : undefined,
    );
    if (res.code === 0) {
      alert('任务已停止');
      showStopDialog.value = false;
      loadTasks();
    } else {
      alert('停止失败: ' + res.message);
    }
  } catch (e: any) {
    alert('停止失败: ' + e.message);
  }
};

const cancelStop = () => {
  showStopDialog.value = false;
  stopTargetTask.value = null;
  stopDialogType.value = null;
};

// 删除任务
const deleteTask = async (task: Task) => {
  if (!confirm(`确定要删除任务 "${task.taskName}" (#${task.id}) 吗？此操作不可撤销。`)) return;
  try {
    const res = await TaskService.deleteTask(task.id);
    if (res.code === 0) {
      alert('任务已删除');
      loadTasks();
    } else {
      alert('删除失败: ' + res.message);
    }
  } catch (e: any) {
    alert('删除失败: ' + e.message);
  }
};

onMounted(() => {
  loadTasks();
});
</script>

<style scoped>
.task-management-page {
  animation: fadeIn 0.3s ease-in-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.card {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  margin-bottom: 16px;
  overflow: hidden;
}

.card-body {
  padding: 16px 20px;
}

/* 工具栏 */
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.toolbar-left {
  flex: 1;
  min-width: 200px;
}

.toolbar-right {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 0;
  max-width: 500px;
}

.search-type-select {
  height: 36px;
  padding: 0 12px;
  border: 1px solid #d9d9d9;
  border-right: none;
  border-radius: 6px 0 0 6px;
  background: #fafafa;
  font-size: 13px;
  color: #333;
  outline: none;
  cursor: pointer;
}

.search-input {
  flex: 1;
  height: 36px;
  padding: 0 12px;
  border: 1px solid #d9d9d9;
  border-right: none;
  font-size: 13px;
  outline: none;
  color: #333;
}

.search-input:focus {
  border-color: #4a90d9;
}

.search-btn {
  height: 36px;
  padding: 0 16px;
  border: 1px solid #d9d9d9;
  border-radius: 0 6px 6px 0;
  background: #f5f5f5;
  cursor: pointer;
  font-size: 14px;
}

.search-btn:hover {
  background: #e8e8e8;
}

/* 按钮 */
.btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  height: 36px;
  padding: 0 16px;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  white-space: nowrap;
}

.btn-primary {
  background: #4a90d9;
  color: white;
}

.btn-primary:hover {
  background: #357abd;
}

.btn-secondary {
  background: #f0f2f5;
  color: #555;
}

.btn-secondary:hover {
  background: #e0e3e8;
}
.btn-danger {
  background: linear-gradient(135deg, #eb3349 0%, #f45c43 100%);
  color: white;
}

.btn-icon {
  font-size: 14px;
}

/* 表格容器 */
.table-container {
  border: 1px solid #f0f2f5;
  border-radius: 8px;
  overflow: hidden;
  margin: -16px -20px;
}

.table-container :deep(.vxe-table--header) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.table-container :deep(.vxe-table--header .vxe-header--column) {
  color: white;
  font-weight: 600;
  background: transparent;
}

/* 分页 */
.pagination {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 0 8px;
}

.pagination-info {
  font-size: 13px;
  color: #999;
}

.pagination-controls {
  display: flex;
  align-items: center;
  gap: 6px;
}

.page-btn {
  height: 32px;
  min-width: 36px;
  padding: 0 10px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  background: white;
  font-size: 13px;
  cursor: pointer;
  color: #333;
  transition: all 0.2s;
}

.page-btn:hover:not(:disabled) {
  border-color: #4a90d9;
  color: #4a90d9;
}

.page-btn.active {
  background: #4a90d9;
  border-color: #4a90d9;
  color: white;
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-size-select {
  height: 32px;
  padding: 0 8px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  background: white;
  font-size: 13px;
  cursor: pointer;
  outline: none;
}

/* 状态标签 */
.status-tag {
  display: inline-block;
  padding: 2px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.status-running {
  background: #e6f7e6;
  color: #52c41a;
}

.status-stopped {
  background: #f0f2f5;
  color: #999;
}

.status-stopping {
  background: #fff7e6;
  color: #fa8c16;
}

.status-failed {
  background: #fff1f0;
  color: #f5222d;
}

.status-created {
  background: #e6f0ff;
  color: #4a90d9;
}

/* 网卡标签 */
.interface-tag {
  display: inline-block;
  padding: 1px 6px;
  margin: 1px 3px;
  background: #f0f5ff;
  color: #4a90d9;
  border-radius: 4px;
  font-size: 12px;
}

/* 操作按钮 */
.action-buttons {
  display: flex;
  gap: 6px;
  justify-content: center;
}

.action-btn {
  padding: 4px 12px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  background: white;
  font-size: 12px;
  cursor: pointer;
  color: #333;
}

.action-btn:hover {
  border-color: #4a90d9;
  color: #4a90d9;
}

.action-btn-warning {
  border-color: #fa8c16;
  color: #fa8c16;
}

.action-btn-warning:hover {
  background: #fff7e6;
}

.action-btn-danger {
  border-color: #f5222d;
  color: #f5222d;
}

.action-btn-danger:hover {
  background: #fff1f0;
}

/* 辅助文本 */
.text-muted {
  color: #bfbfbf;
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

.modal-close:hover {
  color: #333;
}

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

/* 表单 */
.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 6px;
  font-size: 13px;
  font-weight: 500;
  color: #555;
}

.required {
  color: #f5222d;
}

.form-input {
  width: 100%;
  height: 38px;
  padding: 0 12px;
  border: 1px solid #d9d9d9;
  border-radius: 6px;
  font-size: 13px;
  outline: none;
  box-sizing: border-box;
  color: #333;
}

.form-input:focus {
  border-color: #4a90d9;
  box-shadow: 0 0 0 2px rgba(74, 144, 217, 0.15);
}

.form-row {
  display: flex;
  gap: 20px;
  align-items: center;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #555;
  cursor: pointer;
}

.checkbox-label input[type="checkbox"] {
  width: 16px;
  height: 16px;
  cursor: pointer;
}

/* 详情网格 */
.detail-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px 24px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.detail-label {
  font-size: 12px;
  color: #999;
}

.detail-value {
  font-size: 14px;
  color: #333;
}

.detail-message {
  margin-top: 16px;
  padding-top: 12px;
  border-top: 1px solid #f0f2f5;
}

.detail-message pre {
  margin: 6px 0 0;
  padding: 10px;
  background: #f8f9fa;
  border-radius: 6px;
  font-size: 13px;
  white-space: pre-wrap;
  word-break: break-all;
}

/* 停止弹窗 */
.modal-sm {
  width: 480px;
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
