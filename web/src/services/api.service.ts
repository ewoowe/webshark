import type { Ref } from 'vue';
import type { ApiResponse } from '@/types';

const API_BASE = '/api/v1/webshark';

/**
 * 安全解析 JSON 响应，处理空响应情况
 */
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

export class ApiService {
  /**
   * 获取网卡列表
   */
  static async getInterfaces(
    host: string,
    username: string,
    password: string
  ): Promise<ApiResponse> {
    const params = new URLSearchParams({
      host,
      username,
      password,
    });

    const response = await fetch(`${API_BASE}/interfaces?${params}`);
    return await safeJson<ApiResponse>(response);
  }

  /**
   * 开始抓包
   */
  static async startCapture(config: {
    host: string;
    username: string;
    password: string;
    interfaces: string[];
    bpf_filter: string;
    wireshark_filter: string;
  }): Promise<ApiResponse<string>> {
    const response = await fetch(`${API_BASE}/capture/start`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(config),
    });
    return await safeJson<ApiResponse<string>>(response);
  }

  /**
   * 停止抓包
   */
  static async stopCapture(sessionId: string): Promise<ApiResponse> {
    const response = await fetch(`${API_BASE}/capture/stop?session_id=${sessionId}`, {
      method: 'POST',
    });
    return await safeJson<ApiResponse>(response);
  }

  /**
   * 获取数据包详情
   */
  static async getPacketDetail(
    taskId: number,
    frameNumber: number
  ): Promise<ApiResponse<{ detail: string }>> {
    const params = new URLSearchParams({
      taskId: taskId.toString(),
      frameNumber: frameNumber.toString(),
    });

    const response = await fetch(`${API_BASE}/capture/packet/detail?${params}`);
    return await safeJson<ApiResponse<{ detail: string }>>(response);
  }
}