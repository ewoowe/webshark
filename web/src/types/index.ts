// 接口定义
export interface NetworkInterface {
  name: string;
  ip: string;
}

export interface Packet {
  id: number;
  taskId: number;
  no: number;
  frameNumber: number;
  timestamp: number;
  ethSrc: string;
  ethDst: string;
  ip6Src: string;
  ip6Dst: string;
  ip4Src: string;
  ip4Dst: string;
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

// 抓包相关类型
export interface CaptureHost {
  id: number;
  hostName: string;
  ip: string;
  userName: string;
}

export interface CaptureTask {
  streamId: number;
  interfaces: string[];
  bpfFilter: string;
  wiresharkFilter: string;
}

export interface HostCaptureConfig {
  hostId: number;
  captures: CaptureTask[];
}

export interface CaptureRequest {
  taskName: string;
  onlyCapture: boolean;
  parseDetail: boolean;
  detailFormat: string;
  hostCaptures: HostCaptureConfig[];
}

export interface TaskInfo {
  taskGroupId: number;
  taskIds: Record<number, number>;
}

// 任务管理相关类型
export interface Task {
  id: number;
  taskName: string;
  streamId: number;
  hostId: number;
  interfaces: string[];
  onlyCapture: boolean;
  parseDetail: boolean;
  detailFormat: string;
  filePath: string;
  bpfFilter: string;
  wiresharkFilter: string;
  fullCommand: string;
  createdAt: string;
  stopAt: string | null;
  taskGroupId: number;
  status: string;
  message: string;
}

export interface CreateTaskRequest {
  taskName: string;
  hostId: number;
  interfaces?: string[];
  onlyCapture?: boolean;
  parseDetail?: boolean;
  detailFormat?: string;
  bpfFilter?: string;
  wiresharkFilter?: string;
}

export interface UpdateTaskRequest {
  id: number;
  taskName?: string;
  interfaces?: string[];
  onlyCapture?: boolean;
  parseDetail?: boolean;
  detailFormat?: string;
  bpfFilter?: string;
  wiresharkFilter?: string;
}

export interface TaskListResponse {
  items: Task[];
  total: number;
  page: number;
  pageSize: number;
  totalPage: number;
}