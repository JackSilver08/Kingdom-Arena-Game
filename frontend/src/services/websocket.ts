import type { ClientMessage, ServerMessage } from '../types/messages';

type MessageHandler = (msg: ServerMessage) => void;

export class WebSocketService {
  private ws: WebSocket | null = null;
  private url: string;
  private handlers: MessageHandler[] = [];
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;

  constructor(url: string) {
    this.url = url;
  }

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(this.url);

        this.ws.onopen = () => {
          console.log('[WS] Connected to Kingdom Arena server');
          this.reconnectAttempts = 0;
          resolve();
        };

        this.ws.onmessage = (event) => {
          try {
            const msg: ServerMessage = JSON.parse(event.data);
            this.handlers.forEach((h) => h(msg));
          } catch (e) {
            console.error('[WS] Parse error:', e);
          }
        };

        this.ws.onclose = () => {
          console.log('[WS] Disconnected');
          this.tryReconnect();
        };

        this.ws.onerror = (err) => {
          console.warn('[WS] Connection issue (backend offline or connecting):', err);
          reject(err);
        };
      } catch (err) {
        reject(err);
      }
    });
  }

  send(msg: ClientMessage) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(msg));
    }
  }

  onMessage(handler: MessageHandler) {
    this.handlers.push(handler);
  }

  private tryReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++;
      const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 8000);
      setTimeout(() => this.connect().catch(() => {}), delay);
    }
  }

  disconnect() {
    this.ws?.close();
    this.ws = null;
  }
}
