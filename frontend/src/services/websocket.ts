import type { ClientMessage, ServerMessage } from '../types/messages';

type MessageHandler = (msg: ServerMessage) => void;

function defaultWebSocketUrl(): string {
  if (import.meta.env.VITE_WS_URL) return import.meta.env.VITE_WS_URL;
  if (import.meta.env.DEV) return 'ws://localhost:8080/ws';
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  return `${protocol}//${window.location.host}/ws`;
}

export class WebSocketService {
  private ws: WebSocket | null = null;
  private url: string;
  private handlers: MessageHandler[] = [];
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private shouldReconnect = true;

  constructor(url: string = defaultWebSocketUrl()) {
    this.url = url;
  }

  connect(): Promise<void> {
    this.shouldReconnect = true;
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
            this.handlers.forEach((handler) => handler(msg));
          } catch (error) {
            console.error('[WS] Parse error:', error);
          }
        };
        this.ws.onclose = () => {
          console.log('[WS] Disconnected');
          if (this.shouldReconnect) this.tryReconnect();
        };
        this.ws.onerror = (error) => {
          console.warn('[WS] Connection issue (backend offline or connecting):', error);
          reject(error);
        };
      } catch (error) {
        reject(error);
      }
    });
  }

  send(msg: ClientMessage): boolean {
    if (this.ws?.readyState !== WebSocket.OPEN) return false;
    this.ws.send(JSON.stringify(msg));
    return true;
  }

  onMessage(handler: MessageHandler) {
    this.handlers.push(handler);
  }

  private tryReconnect() {
    if (!this.shouldReconnect || this.reconnectAttempts >= this.maxReconnectAttempts) return;
    this.reconnectAttempts++;
    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 8000);
    window.setTimeout(() => {
      if (this.shouldReconnect) this.connect().catch(() => {});
    }, delay);
  }

  disconnect() {
    this.shouldReconnect = false;
    this.ws?.close();
    this.ws = null;
  }
}
