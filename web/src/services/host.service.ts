import type {
  Host,
  CreateHostRequest,
  UpdateHostRequest,
  HostListResponse,
  UnifiedApiResponse,
  PaginationParams,
} from '@/types';

const API_BASE = '/api/v1/webshark';
const HOSTS_BASE = `${API_BASE}/hosts`;

export class HostService {
  /**
   * 获取主机列表
   */
  static async getHostList(params: PaginationParams): Promise<UnifiedApiResponse<HostListResponse>> {
    const queryParams = new URLSearchParams({
      page: params.page.toString(),
      pageSize: params.pageSize.toString(),
    });

    const response = await fetch(`${HOSTS_BASE}?${queryParams}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });
    return await response.json();
  }

  /**
   * 搜索主机
   */
  static async searchHosts(
    keyword: string,
    params: PaginationParams
  ): Promise<UnifiedApiResponse<HostListResponse>> {
    const queryParams = new URLSearchParams({
      keyword,
      page: params.page.toString(),
      pageSize: params.pageSize.toString(),
    });

    const response = await fetch(`${HOSTS_BASE}/search?${queryParams}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });
    return await response.json();
  }

  /**
   * 获取单个主机
   */
  static async getHost(id: number): Promise<UnifiedApiResponse<Host>> {
    const response = await fetch(`${HOSTS_BASE}/${id}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
    });
    return await response.json();
  }

  /**
   * 创建主机
   */
  static async createHost(host: CreateHostRequest): Promise<UnifiedApiResponse<Host>> {
    const response = await fetch(`${HOSTS_BASE}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(host),
    });
    return await response.json();
  }

  /**
   * 更新主机
   */
  static async updateHost(id: number, host: Partial<UpdateHostRequest>): Promise<UnifiedApiResponse<Host>> {
    const response = await fetch(`${HOSTS_BASE}/${id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ id, ...host }),
    });
    return await response.json();
  }

  /**
   * 删除主机
   */
  static async deleteHost(id: number): Promise<UnifiedApiResponse<any>> {
    const response = await fetch(`${HOSTS_BASE}/${id}`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
      },
    });
    return await response.json();
  }
}