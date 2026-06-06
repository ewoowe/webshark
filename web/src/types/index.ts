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