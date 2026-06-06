import type { Ref } from 'vue';
import type { ApiResponse } from '../types';

const API_BASE = '/api';

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
    return await response.json();
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
    return await response.json();
  }

  /**
   * 停止抓包
   */
  static async stopCapture(sessionId: string): Promise<ApiResponse> {
    const response = await fetch(`${API_BASE}/capture/stop?session_id=${sessionId}`, {
      method: 'POST',
    });
    return await response.json();
  }
}