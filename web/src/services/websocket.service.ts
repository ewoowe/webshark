import type { Packet } from '../types';

export class WebSocketService {
  private ws: WebSocket | null = null;
  private onMessageCallback: ((packet: Packet) => void) | null = null;
  private onErrorCallback: ((error: Event) => void) | null = null;
  private onCloseCallback: (() => void) | null = null;
  private onOpenCallback: (() => void) | null = null;

  /**
   * 连接 WebSocket
   */
  connect(sessionId: string): void {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws/capture?session_id=${sessionId}`;

    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => {
      console.log('WebSocket 连接已建立');
      this.onOpenCallback?.();
    };

    this.ws.onmessage = (event) => {
      try {
        const packet = JSON.parse(event.data) as Packet;
        this.onMessageCallback?.(packet);
      } catch (error) {
        console.error('解析数据包失败:', error);
      }
    };

    this.ws.onerror = (error) => {
      console.error('WebSocket 错误:', error);
      this.onErrorCallback?.(error);
    };

    this.ws.onclose = () => {
      console.log('WebSocket 连接已关闭');
      this.onCloseCallback?.();
    };
  }

  /**
   * 关闭连接
   */
  disconnect(): void {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  /**
   * 设置消息回调
   */
  onMessage(callback: (packet: Packet) => void): void {
    this.onMessageCallback = callback;
  }

  /**
   * 设置错误回调
   */
  onError(callback: (error: Event) => void): void {
    this.onErrorCallback = callback;
  }

  /**
   * 设置关闭回调
   */
  onClose(callback: () => void): void {
    this.onCloseCallback = callback;
  }

  /**
   * 设置打开回调
   */
  onOpen(callback: () => void): void {
    this.onOpenCallback = callback;
  }

  /**
   * 检查连接状态
   */
  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
  }
}