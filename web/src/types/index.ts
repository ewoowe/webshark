// 接口定义
export interface NetworkInterface {
  name: string;
  ip: string;
}

export interface Packet {
  frame: number;
  timestamp: string;
  source: string;
  dest: string;
  protocol: string;
  length: number;
  info: string;
  [key: string]: any;
}

export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  message?: string;
  error?: string;
}

export interface CaptureConfig {
  host: string;
  username: string;
  password: string;
  interfaces: string[];
  bpf_filter: string;
  wireshark_filter: string;
}

// 主机管理相关类型
export interface Host {
  id: number;
  hostName: string;
  ip: string;
  userName: string;
  password: string;
  os: string;
  updatedAt: string;
  createdAt: string;
}

export interface CreateHostRequest {
  hostName: string;
  ip: string;
  userName: string;
  password: string;
  os?: string;
}

export interface UpdateHostRequest {
  id: number;
  hostName?: string;
  ip?: string;
  userName?: string;
  password?: string;
  os?: string;
}

export interface HostListResponse {
  items: Host[]; // 后端返回的是 items 不是 hosts
  total: number;
  page: number;
  pageSize: number;
  totalPage: number;
}

export interface UnifiedApiResponse<T = any> {
  code: number; // 0: 成功, 1: 失败
  message: string;
  msg: string;
  data: T;
}

export interface PaginationParams {
  page: number;
  pageSize: number;
}