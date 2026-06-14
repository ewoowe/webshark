import type { UnifiedApiResponse, Task, TaskListResponse, CreateTaskRequest, UpdateTaskRequest } from '@/types';

const API_BASE = '/api/v1/webshark/tasks';

async function safeJson<T>(response: Response): Promise<T> {
  const text = await response.text();
  if (!text || text.trim() === '') {
    throw new Error(`服务器返回空响应 (HTTP ${response.status})`);
  }
  try {
    return JSON.parse(text);
  } catch {
    throw new Error(`服务器返回非 JSON 格式数据: ${text.substring(0, 200)}`);
  }
}

export class TaskService {
  /**
   * 获取任务列表（分页）
   */
  static async listTasks(page = 1, pageSize = 10): Promise<UnifiedApiResponse<TaskListResponse>> {
    const params = new URLSearchParams({ page: String(page), pageSize: String(pageSize) });
    const response = await fetch(`${API_BASE}?${params}`);
    return safeJson<UnifiedApiResponse<TaskListResponse>>(response);
  }

  /**
   * 按主机 ID 搜索任务
   */
  static async listTasksByHostId(hostId: number, page = 1, pageSize = 10): Promise<UnifiedApiResponse<TaskListResponse>> {
    const params = new URLSearchParams({ hostId: String(hostId), page: String(page), pageSize: String(pageSize) });
    const response = await fetch(`${API_BASE}/search?${params}`);
    return safeJson<UnifiedApiResponse<TaskListResponse>>(response);
  }

  /**
   * 按任务组 ID 获取任务
   */
  static async listTasksByGroupId(taskGroupId: number, page = 1, pageSize = 10): Promise<UnifiedApiResponse<TaskListResponse>> {
    const params = new URLSearchParams({ taskGroupId: String(taskGroupId), page: String(page), pageSize: String(pageSize) });
    const response = await fetch(`${API_BASE}/group?${params}`);
    return safeJson<UnifiedApiResponse<TaskListResponse>>(response);
  }

  /**
   * 获取单个任务
   */
  static async getTask(id: number): Promise<UnifiedApiResponse<Task>> {
    const response = await fetch(`${API_BASE}/${id}`);
    return safeJson<UnifiedApiResponse<Task>>(response);
  }

  /**
   * 创建任务
   */
  static async createTask(data: CreateTaskRequest): Promise<UnifiedApiResponse<Task>> {
    const response = await fetch(API_BASE, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return safeJson<UnifiedApiResponse<Task>>(response);
  }

  /**
   * 更新任务
   */
  static async updateTask(data: UpdateTaskRequest): Promise<UnifiedApiResponse<Task>> {
    const response = await fetch(`${API_BASE}/${data.id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return safeJson<UnifiedApiResponse<Task>>(response);
  }

  /**
   * 停止任务
   */
  static async stopTask(id: number): Promise<UnifiedApiResponse<null>> {
    const response = await fetch(`${API_BASE}/${id}/stop`, { method: 'POST' });
    return safeJson<UnifiedApiResponse<null>>(response);
  }

  /**
   * 删除任务
   */
  static async deleteTask(id: number): Promise<UnifiedApiResponse<null>> {
    const response = await fetch(`${API_BASE}/${id}`, { method: 'DELETE' });
    return safeJson<UnifiedApiResponse<null>>(response);
  }
}
