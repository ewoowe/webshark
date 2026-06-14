import type { ApiResponse, NetworkInterface } from '@/types';
import type { CaptureRequest, TaskInfo, UnifiedApiResponse } from '@/types';

const API_BASE = '/api/v1/webshark';

export class CaptureService {
  /**
   * 获取远程网卡列表
   */
  static async getInterfaces(hostId: number): Promise<UnifiedApiResponse<NetworkInterface[]>> {
    const response = await fetch(`${API_BASE}/interfaces?hostId=${hostId}`, {
      method: 'GET',
      headers: { 'Content-Type': 'application/json' },
    });
    return await response.json();
  }

  /**
   * 开始抓包
   */
  static async startCapture(req: CaptureRequest): Promise<UnifiedApiResponse<TaskInfo>> {
    const response = await fetch(`${API_BASE}/capture/start`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(req),
    });
    return await response.json();
  }

  /**
   * 停止抓包（按 taskId 或 taskGroupId）
   */
  static async stopCapture(taskGroupId?: number, taskId?: number): Promise<UnifiedApiResponse<null>> {
    const params = new URLSearchParams();
    if (taskGroupId) params.set('taskGroupId', String(taskGroupId));
    if (taskId) params.set('taskId', String(taskId));

    const response = await fetch(`${API_BASE}/capture/stop?${params}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
    });
    return await response.json();
  }

  /**
   * 获取数据包详情
   */
  static async getPacketDetail(taskId: number, frameNumber: number): Promise<UnifiedApiResponse<string>> {
    const params = new URLSearchParams();
    params.set('taskId', String(taskId));
    params.set('frameNumber', String(frameNumber));

    const response = await fetch(`${API_BASE}/capture/packet/detail?${params}`, {
      method: 'GET',
      headers: { 'Content-Type': 'application/json' },
    });
    return await response.json();
  }
}
