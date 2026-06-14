<template>
  <div class="host-management-page">
    <!-- 操作栏 -->
    <div class="card">
      <div class="card-body">
        <div class="toolbar">
          <div class="toolbar-left">
            <div class="search-box">
              <input
                type="text"
                class="search-input"
                v-model="searchKeyword"
                placeholder="搜索主机名称、IP地址或用户名"
                @keyup.enter="handleSearch"
              />
              <button class="search-btn" @click="handleSearch">🔍</button>
            </div>
          </div>
          <div class="toolbar-right">
            <!-- 批量操作按钮组 - 用 v-show 保持布局稳定 -->
            <div class="batch-actions" v-show="selectedRows.length > 0">
              <span class="batch-count">已选择 {{ selectedRows.length }} 项</span>
              <button class="btn btn-primary" @click="addToCaptureHosts">
                <span class="btn-icon">🎯</span>
                <span class="btn-text">+抓包主机</span>
              </button>
              <button class="btn btn-danger" @click="showBatchDeleteConfirm">
                <span class="btn-icon">🗑️</span>
                <span class="btn-text">批量删除</span>
              </button>
            </div>
            <button class="btn btn-success" @click="goToCapture" v-if="captureHostCount > 0">
              <span class="btn-icon">🚀</span>
              <span class="btn-text">去抓包 ({{ captureHostCount }})</span>
            </button>
            <button class="btn btn-primary" @click="showCreateDialog">
              <span class="btn-icon">➕</span>
              <span class="btn-text">新增主机</span>
            </button>
            <button class="btn btn-secondary" @click="refreshList">
              <span class="btn-icon">🔄</span>
              <span class="btn-text">刷新</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 主机列表 -->
    <div class="card">
      <div class="card-body">
        <div class="table-container">
          <vxe-grid
            ref="gridRef"
            :data="hosts"
            :columns="columns"
            :row-config="{ isCurrent: true, isHover: true }"
            :show-header="true"
            :stripe="true"
            :border="true"
            :show-overflow="true"
            :loading="loading"
            :checkbox-config="{ range: true }"
            @checkbox-change="handleCheckboxChange"
            @checkbox-all="handleCheckboxAll"
          >
            <!-- 创建时间列模板 -->
            <template #createdAt="{ row }">
              {{ new Date(row.createdAt).toLocaleString('zh-CN') }}
            </template>

            <!-- 更新时间列模板 -->
            <template #updatedAt="{ row }">
              {{ new Date(row.updatedAt).toLocaleString('zh-CN') }}
            </template>

            <!-- 操作列模板 -->
            <template #actions="{ row }">
              <div class="action-buttons">
                <button class="action-btn" @click="editHost(row)">编辑</button>
                <button class="action-btn danger" @click="showDeleteConfirm(row)">删除</button>
              </div>
            </template>

            <!-- 无数据模板 -->
            <template #noData>
              <div class="no-data">
                <div class="no-data-icon">🖥️</div>
                <div class="no-data-text">暂无主机数据</div>
                <button class="btn btn-primary btn-sm" @click="showCreateDialog">
                  新增主机
                </button>
              </div>
            </template>
          </vxe-grid>
        </div>

        <!-- 分页 -->
        <div class="pagination">
          <div class="pagination-info">
            共 {{ total }} 条记录，第 {{ currentPage }} / {{ totalPages || 1 }} 页
          </div>
          <div class="pagination-controls">
            <button
              class="page-btn"
              :disabled="currentPage === 1"
              @click="goToPage(1)"
            >
              首页
            </button>
            <button
              class="page-btn"
              :disabled="currentPage === 1"
              @click="goToPage(currentPage - 1)"
            >
              上一页
            </button>
            <span class="page-info">{{ currentPage }} / {{ totalPages }}</span>
            <button
              class="page-btn"
              :disabled="currentPage === totalPages"
              @click="goToPage(currentPage + 1)"
            >
              下一页
            </button>
            <button
              class="page-btn"
              :disabled="currentPage === totalPages"
              @click="goToPage(totalPages)"
            >
              末页
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 新增/编辑主机对话框 -->
    <div class="modal-overlay" v-if="showDialog" @click.self="closeDialog">
      <div class="modal">
        <div class="modal-header">
          <h3 class="modal-title">{{ isEditMode ? '编辑主机' : '新增主机' }}</h3>
          <button class="modal-close" @click="closeDialog">✕</button>
        </div>
        <div class="modal-body">
          <div class="form-grid">
            <div class="form-item">
              <label class="form-label">主机名称 <span class="required">*</span></label>
              <input
                type="text"
                class="form-input"
                v-model="hostForm.hostName"
                placeholder="例如: 生产服务器-1"
              />
            </div>
            <div class="form-item">
              <label class="form-label">IP地址 <span class="required">*</span></label>
              <input
                type="text"
                class="form-input"
                v-model="hostForm.ip"
                placeholder="例如: 192.168.1.100"
              />
            </div>
            <div class="form-item">
              <label class="form-label">用户名 <span class="required">*</span></label>
              <input
                type="text"
                class="form-input"
                v-model="hostForm.userName"
                placeholder="例如: root"
              />
            </div>
            <div class="form-item">
              <label class="form-label">密码 <span class="required">*</span></label>
              <div class="password-input-wrapper">
                <input
                  :type="showPassword ? 'text' : 'password'"
                  class="form-input password-input"
                  v-model="hostForm.password"
                  placeholder="输入SSH密码"
                />
                <button
                  type="button"
                  class="password-toggle"
                  @click="showPassword = !showPassword"
                  :title="showPassword ? '隐藏密码' : '显示密码'"
                >
                  {{ showPassword ? '🙈' : '👁️' }}
                </button>
              </div>
            </div>
            <div class="form-item full-width">
              <label class="form-label">操作系统</label>
              <select class="form-select" v-model="hostForm.os">
                <option value="">请选择操作系统</option>
                <option value="linux">Linux</option>
                <option value="macos">macOS</option>
                <option value="windows">Windows</option>
                <option value="other">其他</option>
              </select>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeDialog">取消</button>
          <button class="btn btn-primary" @click="saveHost" :disabled="saving">
            {{ saving ? '保存中...' : (isEditMode ? '更新' : '创建') }}
          </button>
        </div>
      </div>
    </div>

    <!-- 删除确认对话框 -->
    <div class="modal-overlay" v-if="showDeleteDialog" @click.self="closeDeleteDialog">
      <div class="modal modal-sm">
        <div class="modal-header">
          <h3 class="modal-title">确认删除</h3>
          <button class="modal-close" @click="closeDeleteDialog">✕</button>
        </div>
        <div class="modal-body">
          <div class="delete-warning">
            <span class="warning-icon">⚠️</span>
            <p>确定要删除主机 <strong>{{ deletingHost?.hostName }}</strong> 吗？</p>
            <p class="warning-text">此操作无法撤销！</p>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeDeleteDialog">取消</button>
          <button class="btn btn-danger" @click="confirmDelete" :disabled="deleting">
            {{ deleting ? '删除中...' : '确认删除' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 批量删除确认对话框 -->
    <div class="modal-overlay" v-if="showBatchDeleteDialog" @click.self="closeBatchDeleteDialog">
      <div class="modal modal-sm">
        <div class="modal-header">
          <h3 class="modal-title">确认批量删除</h3>
          <button class="modal-close" @click="closeBatchDeleteDialog">✕</button>
        </div>
        <div class="modal-body">
          <div class="delete-warning">
            <span class="warning-icon">⚠️</span>
            <p>确定要删除选中的 <strong>{{ selectedRows.length }}</strong> 台主机吗？</p>
            <p class="selected-hosts-list" v-if="selectedRows.length <= 5">
              <span v-for="(host, index) in selectedRows" :key="host.id">
                {{ host.hostName }}{{ index < selectedRows.length - 1 ? '、' : '' }}
              </span>
            </p>
            <p class="selected-hosts-list" v-else>
              {{ selectedRows.slice(0, 5).map(h => h.hostName).join('、') }} ... 等
            </p>
            <p class="warning-text">此操作无法撤销！</p>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeBatchDeleteDialog">取消</button>
          <button class="btn btn-danger" @click="confirmBatchDelete" :disabled="batchDeleting">
            {{ batchDeleting ? '删除中...' : '确认批量删除' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import type { VxeGridInstance, VxeGridPropTypes } from 'vxe-table';
import { HostService } from '@/services/host.service';
import type { Host, CreateHostRequest, UpdateHostRequest } from '@/types';
import { useCaptureHosts } from '@/composables/useCaptureHosts';

const router = useRouter();

// 抓包主机状态
const { store: captureStore, addHosts, hostCount } = useCaptureHosts();
const captureHostCount = computed(() => hostCount());

// 状态管理
const hosts = ref<Host[]>([]);
const total = ref(0);
const currentPage = ref(1);
const pageSize = ref(10);
const totalPages = ref(0);
const loading = ref(false);
const searchKeyword = ref('');

// 批量选择
const selectedRows = ref<Host[]>([]);

// 表格引用
const gridRef = ref<VxeGridInstance>();

// 对话框状态
const showDialog = ref(false);
const showDeleteDialog = ref(false);
const showBatchDeleteDialog = ref(false);
const isEditMode = ref(false);
const saving = ref(false);
const deleting = ref(false);
const showPassword = ref(false);
const batchDeleting = ref(false);

// 表单数据
const hostForm = reactive<CreateHostRequest & { id?: number }>({
  hostName: '',
  ip: '',
  userName: '',
  password: '',
  os: '',
});

// 删除的主机
const deletingHost = ref<Host | null>(null);

// 表格列配置 - 使用 min-width 代替 width，让表格自动铺满容器宽度
const columns = reactive<VxeGridPropTypes.Columns<Host>>([
  {
    type: 'checkbox',
    width: 50,
    align: 'center',
    fixed: 'left',
  },
  {
    field: 'id',
    title: 'ID',
    minWidth: 60,
    align: 'center',
  },
  {
    field: 'hostName',
    title: '主机名称',
    minWidth: 120,
    align: 'center',
  },
  {
    field: 'ip',
    title: 'IP地址',
    minWidth: 120,
    align: 'center',
  },
  {
    field: 'userName',
    title: '用户名',
    minWidth: 100,
    align: 'center',
  },
  {
    field: 'os',
    title: '操作系统',
    minWidth: 90,
    align: 'center',
  },
  {
    field: 'createdAt',
    title: '创建时间',
    minWidth: 150,
    align: 'center',
    slots: {
      default: 'createdAt',
    },
  },
  {
    field: 'updatedAt',
    title: '更新时间',
    minWidth: 150,
    align: 'center',
    slots: {
      default: 'updatedAt',
    },
  },
  {
    title: '操作',
    minWidth: 130,
    align: 'center',
    fixed: 'right',
    slots: {
      default: 'actions',
    },
  },
]);

// 获取主机列表
const fetchHostList = async () => {
  loading.value = true;
  try {
    const response = await HostService.getHostList({
      page: currentPage.value,
      pageSize: pageSize.value,
    });

    console.log('API Response:', response);

    if (response.code === 0 && response.data) {
      console.log('Hosts data:', response.data.items);
      hosts.value = response.data.items || [];
      total.value = Number(response.data.total) || 0;
      totalPages.value = Number(response.data.totalPage) || 1;
      console.log('Total:', total.value, 'TotalPages:', totalPages.value);
      // 清空选择
      selectedRows.value = [];
    } else {
      alert('获取主机列表失败: ' + response.message);
    }
  } catch (error) {
    console.error('获取主机列表失败:', error);
    alert('网络错误: ' + (error instanceof Error ? error.message : 'Unknown error'));
  } finally {
    loading.value = false;
  }
};

// 搜索主机
const handleSearch = async () => {
  if (!searchKeyword.value.trim()) {
    fetchHostList();
    return;
  }

  loading.value = true;
  try {
    const response = await HostService.searchHosts(searchKeyword.value, {
      page: currentPage.value,
      pageSize: pageSize.value,
    });

    if (response.code === 0 && response.data) {
      hosts.value = response.data.items || [];
      total.value = response.data.total || 0;
      totalPages.value = response.data.totalPage || 1;
      selectedRows.value = [];
    } else {
      alert('搜索失败: ' + response.message);
    }
  } catch (error) {
    console.error('搜索失败:', error);
    alert('网络错误: ' + (error instanceof Error ? error.message : 'Unknown error'));
  } finally {
    loading.value = false;
  }
};

// 刷新列表
const refreshList = () => {
  searchKeyword.value = '';
  currentPage.value = 1;
  fetchHostList();
};

// 分页导航
const goToPage = (page: number) => {
  currentPage.value = page;
  if (searchKeyword.value.trim()) {
    handleSearch();
  } else {
    fetchHostList();
  }
};

// 复选框选择变化
const handleCheckboxChange = ({ checked, row }: { checked: boolean; row: Host }) => {
  if (checked) {
    selectedRows.value.push(row);
  } else {
    const index = selectedRows.value.findIndex(item => item.id === row.id);
    if (index > -1) {
      selectedRows.value.splice(index, 1);
    }
  }
};

// 全选/取消全选
const handleCheckboxAll = ({ checked }: { checked: boolean }) => {
  if (checked) {
    selectedRows.value = [...hosts.value];
  } else {
    selectedRows.value = [];
  }
};

// 显示创建对话框
const showCreateDialog = () => {
  isEditMode.value = false;
  Object.assign(hostForm, {
    hostName: '',
    ip: '',
    userName: '',
    password: '',
    os: '',
  });
  showDialog.value = true;
};

// 编辑主机
const editHost = (host: Host) => {
  isEditMode.value = true;
  Object.assign(hostForm, {
    id: host.id,
    hostName: host.hostName,
    ip: host.ip,
    userName: host.userName,
    password: host.password,
    os: host.os,
  });
  showDialog.value = true;
};

// 关闭对话框
const closeDialog = () => {
  showDialog.value = false;
  showPassword.value = false;
  Object.assign(hostForm, {
    hostName: '',
    ip: '',
    userName: '',
    password: '',
    os: '',
  });
};

// 保存主机
const saveHost = async () => {
  if (!hostForm.hostName || !hostForm.ip || !hostForm.userName || !hostForm.password) {
    alert('请填写所有必填字段');
    return;
  }

  saving.value = true;
  try {
    let response;
    if (isEditMode.value && hostForm.id) {
      // 更新主机
      const updateData: UpdateHostRequest = {
        id: hostForm.id,
        hostName: hostForm.hostName,
        ip: hostForm.ip,
        userName: hostForm.userName,
        password: hostForm.password,
        os: hostForm.os,
      };
      response = await HostService.updateHost(hostForm.id, updateData);
    } else {
      // 创建主机
      const createData: CreateHostRequest = {
        hostName: hostForm.hostName,
        ip: hostForm.ip,
        userName: hostForm.userName,
        password: hostForm.password,
        os: hostForm.os,
      };
      response = await HostService.createHost(createData);
    }

    if (response.code === 0) {
      alert(isEditMode.value ? '更新成功' : '创建成功');
      closeDialog();
      fetchHostList();
    } else {
      alert((isEditMode.value ? '更新失败: ' : '创建失败: ') + response.message);
    }
  } catch (error) {
    console.error('保存主机失败:', error);
    alert('网络错误: ' + (error instanceof Error ? error.message : 'Unknown error'));
  } finally {
    saving.value = false;
  }
};

// 显示删除确认对话框
const showDeleteConfirm = (host: Host) => {
  deletingHost.value = host;
  showDeleteDialog.value = true;
};

// 关闭删除对话框
const closeDeleteDialog = () => {
  showDeleteDialog.value = false;
  deletingHost.value = null;
};

// 确认删除
const confirmDelete = async () => {
  if (!deletingHost.value) return;

  deleting.value = true;
  try {
    const response = await HostService.deleteHost(deletingHost.value.id);
    if (response.code === 0) {
      alert('删除成功');
      closeDeleteDialog();
      fetchHostList();
    } else {
      alert('删除失败: ' + response.message);
    }
  } catch (error) {
    console.error('删除主机失败:', error);
    alert('网络错误: ' + (error instanceof Error ? error.message : 'Unknown error'));
  } finally {
    deleting.value = false;
  }
};

// 显示批量删除确认对话框
const showBatchDeleteConfirm = () => {
  if (selectedRows.value.length === 0) {
    alert('请先选择要删除的主机');
    return;
  }
  showBatchDeleteDialog.value = true;
};

// 关闭批量删除对话框
const closeBatchDeleteDialog = () => {
  showBatchDeleteDialog.value = false;
};

// 添加选中主机到抓包列表
const addToCaptureHosts = () => {
  if (selectedRows.value.length === 0) {
    alert('请先选择要抓包的主机');
    return;
  }
  addHosts(selectedRows.value.map(h => ({
    id: h.id,
    hostName: h.hostName,
    ip: h.ip,
    userName: h.userName,
  })));
  selectedRows.value = [];
  alert('已添加到抓包列表');
};

// 跳转到抓包页面
const goToCapture = () => {
  router.push('/capture');
};

// 确认批量删除
const confirmBatchDelete = async () => {
  if (selectedRows.value.length === 0) return;

  batchDeleting.value = true;
  try {
    // 批量删除 - 串行删除
    let successCount = 0;
    let failCount = 0;
    const failedHosts: string[] = [];

    for (const host of selectedRows.value) {
      try {
        const response = await HostService.deleteHost(host.id);
        if (response.code === 0) {
          successCount++;
        } else {
          failCount++;
          failedHosts.push(host.hostName);
        }
      } catch (error) {
        failCount++;
        failedHosts.push(host.hostName);
      }
    }

    // 显示结果
    if (failCount === 0) {
      alert(`批量删除成功！共删除 ${successCount} 台主机`);
    } else {
      alert(`批量删除完成：成功 ${successCount} 台，失败 ${failCount} 台\n失败的主机：${failedHosts.join('、')}`);
    }

    closeBatchDeleteDialog();
    selectedRows.value = [];
    fetchHostList();
  } catch (error) {
    console.error('批量删除失败:', error);
    alert('批量删除过程中发生错误');
  } finally {
    batchDeleting.value = false;
  }
};

// 组件挂载时加载数据
onMounted(() => {
  fetchHostList();
});
</script>

<style scoped>
.host-management-page {
  animation: fadeIn 0.3s ease-in-out;
  height: 100%;
  display: flex;
  flex-direction: column;
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
  margin-bottom: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  overflow: hidden;
  flex-shrink: 0;
}

.card:last-child {
  margin-bottom: 0;
  /* 不再强制 flex:1 撑满，让内容自然决定高度 */
}

.card-body {
  padding: 20px;
}

/* 工具栏 */
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.toolbar-left,
.toolbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 400px;
}

.search-input {
  flex: 1;
  padding: 10px 12px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  font-size: 14px;
  transition: border-color 0.3s;
}

.search-input:focus {
  outline: none;
  border-color: #667eea;
}

.search-btn {
  padding: 10px 16px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.3s;
}

.search-btn:hover {
  opacity: 0.85;
}

/* 批量操作 - 始终占据空间，避免选中时布局跳动 */
.batch-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 16px;
  background: #fff1f0;
  border: 1px solid #ffccc7;
  border-radius: 4px;
  /* 关键：v-show 隐藏时用 visibility 而非 display，保持占位 */
  transition: visibility 0.15s, opacity 0.15s;
}

.batch-actions[style*="display: none"] {
  display: flex !important;
  visibility: hidden;
  opacity: 0;
  pointer-events: none;
}

.batch-count {
  font-size: 14px;
  color: #ff4d4f;
  font-weight: 500;
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
  display: flex;
  align-items: center;
  gap: 8px;
  white-space: nowrap;
}

.btn:hover {
  opacity: 0.85;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.btn-secondary {
  background: #f0f2f5;
  color: #333;
}

.btn-danger {
  background: linear-gradient(135deg, #eb3349 0%, #f45c43 100%);
  color: white;
}

.btn-sm {
  padding: 6px 12px;
  font-size: 12px;
}

.btn-icon {
  font-size: 16px;
}

/* 表格容器 - 铺满宽度 */
.table-container {
  border: 1px solid #f0f2f5;
  border-radius: 8px;
  overflow: hidden;
  margin-bottom: 16px;
}

/* 确保 vxe-table 及其内部元素铺满容器宽度 */
.table-container :deep(.vxe-grid),
.table-container :deep(.vxe-table) {
  width: 100% !important;
}

.table-container :deep(.vxe-table) {
  font-size: 13px;
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

.table-container :deep(.vxe-body--row) {
  cursor: pointer;
}

.table-container :deep(.vxe-body--row:hover) {
  background-color: #f9f9f9 !important;
}

/* 复选框样式 */
.table-container :deep(.vxe-checkbox) {
  cursor: pointer;
}

.table-container :deep(.vxe-table--fixed-left-wrapper) {
  box-shadow: 2px 0 6px rgba(0, 0, 0, 0.1);
}

.table-container :deep(.vxe-table--fixed-right-wrapper) {
  box-shadow: -2px 0 6px rgba(0, 0, 0, 0.1);
}

.table-container :deep(.no-data) {
  padding: 60px 20px;
  text-align: center;
}

.no-data-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.no-data-text {
  font-size: 16px;
  color: #999;
  margin-bottom: 20px;
}

/* 操作按钮 */
.action-buttons {
  display: flex;
  gap: 8px;
  justify-content: center;
}

.action-btn {
  padding: 6px 12px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  background: white;
  color: #333;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.3s;
}

.action-btn:hover {
  background: #f0f2f5;
  border-color: #667eea;
}

.action-btn.danger {
  color: #ff4d4f;
  border-color: #ffccc7;
}

.action-btn.danger:hover {
  background: #fff1f0;
  border-color: #ff4d4f;
}

/* 分页 */
.pagination {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 16px;
  border-top: 1px solid #f0f2f5;
  flex-shrink: 0;
}

.pagination-info {
  font-size: 14px;
  color: #666;
}

.pagination-controls {
  display: flex;
  align-items: center;
  gap: 8px;
}

.page-btn {
  padding: 6px 12px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  background: white;
  color: #333;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s;
}

.page-btn:hover:not(:disabled) {
  background: #f0f2f5;
  border-color: #667eea;
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-info {
  padding: 6px 16px;
  font-size: 14px;
  color: #333;
}

/* 模态框 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: white;
  border-radius: 8px;
  width: 600px;
  max-width: 90vw;
  max-height: 90vh;
  overflow: auto;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
}

.modal-sm {
  width: 400px;
}

.modal-header {
  padding: 20px 24px;
  border-bottom: 1px solid #f0f2f5;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #333;
}

.modal-close {
  background: none;
  border: none;
  font-size: 20px;
  cursor: pointer;
  color: #999;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.3s;
}

.modal-close:hover {
  background: #f0f2f5;
  color: #333;
}

.modal-body {
  padding: 24px;
}

.modal-footer {
  padding: 16px 24px;
  border-top: 1px solid #f0f2f5;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
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

.form-label {
  font-size: 14px;
  color: #333;
  font-weight: 500;
}

.required {
  color: #ff4d4f;
  margin-left: 4px;
}

.form-input,
.form-select {
  padding: 10px 12px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  font-size: 14px;
  transition: border-color 0.3s;
}

.form-input:focus,
.form-select:focus {
  outline: none;
  border-color: #667eea;
}

/* 密码输入框容器 */
.password-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.password-input {
  width: 100%;
  padding-right: 40px;
}

.password-toggle {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  cursor: pointer;
  font-size: 18px;
  padding: 4px;
  line-height: 1;
  color: #999;
  transition: color 0.2s;
}

.password-toggle:hover {
  color: #333;
}

/* 删除警告 */
.delete-warning {
  text-align: center;
  padding: 20px 0;
}

.warning-icon {
  font-size: 48px;
  margin-bottom: 16px;
  display: block;
}

.delete-warning p {
  margin: 8px 0;
  color: #333;
}

.selected-hosts-list {
  font-size: 12px;
  color: #666;
  max-width: 300px;
  margin: 8px auto !important;
  line-height: 1.6;
}

.warning-text {
  color: #ff4d4f !important;
  font-size: 12px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .search-box {
    width: 100%;
  }

  .toolbar-left,
  .toolbar-right {
    width: 100%;
  }

  .toolbar-right {
    flex-wrap: wrap;
  }

  .batch-actions {
    width: 100%;
    justify-content: space-between;
  }

  .modal {
    width: 95vw;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }
}
</style>