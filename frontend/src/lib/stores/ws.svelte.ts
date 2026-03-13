import { browser } from '$app/environment';

type MessageHandler = (data: unknown) => void;

class WebSocketStore {
	connected = $state(false);
	private ws: WebSocket | null = null;
	private handlers: MessageHandler[] = [];
	private reconnectTimeout: ReturnType<typeof setTimeout> | null = null;
	private reconnectDelay = 3000;
	private maxDelay = 30000;

	connect() {
		if (!browser) return;
		if (this.ws?.readyState === WebSocket.OPEN) return;

		const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
		const wsUrl = `${protocol}//${window.location.host}/ws`;

		try {
			this.ws = new WebSocket(wsUrl);

			this.ws.onopen = () => {
				this.connected = true;
				this.reconnectDelay = 3000;
			};

			this.ws.onclose = () => {
				this.connected = false;
				this.scheduleReconnect();
			};

			this.ws.onerror = () => {
				this.ws?.close();
			};

			this.ws.onmessage = (event: MessageEvent) => {
				try {
					const data = JSON.parse(event.data as string);
					for (const handler of this.handlers) {
						handler(data);
					}
				} catch {
					// ignore parse errors
				}
			};
		} catch {
			this.scheduleReconnect();
		}
	}

	private scheduleReconnect() {
		if (this.reconnectTimeout) return;
		this.reconnectTimeout = setTimeout(() => {
			this.reconnectTimeout = null;
			this.reconnectDelay = Math.min(this.reconnectDelay * 2, this.maxDelay);
			this.connect();
		}, this.reconnectDelay);
	}

	send(data: unknown) {
		if (this.ws?.readyState === WebSocket.OPEN) {
			this.ws.send(JSON.stringify(data));
		}
	}

	onMessage(handler: MessageHandler) {
		this.handlers.push(handler);
		return () => {
			this.handlers = this.handlers.filter((h) => h !== handler);
		};
	}

	disconnect() {
		if (this.reconnectTimeout) {
			clearTimeout(this.reconnectTimeout);
			this.reconnectTimeout = null;
		}
		this.ws?.close();
		this.ws = null;
		this.connected = false;
	}
}

export const ws = new WebSocketStore();
