import type {
	Envelope,
	InboundMessage,
	PresenceUpdatePayload,
	AgentInputPayload
} from './types';

// ---------------------------------------------------------------------------
// Constants (per ws-protocol.md)
// ---------------------------------------------------------------------------

const PING_INTERVAL_MS = 25_000;
const RECONNECT_INITIAL_MS = 1_000;
const RECONNECT_MAX_MS = 30_000;
const CLOSE_AUTH_FAILED = 4001;
const CLOSE_PROTOCOL_ERROR = 4002;

// ---------------------------------------------------------------------------
// Message handler type
// ---------------------------------------------------------------------------

export type MessageHandler = (msg: InboundMessage) => void;

// ---------------------------------------------------------------------------
// WebSocket client state
// ---------------------------------------------------------------------------

type ConnectionStatus = 'disconnected' | 'connecting' | 'connected' | 'failed';

interface WsState {
	status: ConnectionStatus;
	error: string | null;
}

// ---------------------------------------------------------------------------
// WsClient
// ---------------------------------------------------------------------------

class WsClient {
	private ws: WebSocket | null = null;
	private url = '';

	// Svelte 5 reactive state
	private _state = $state<WsState>({ status: 'disconnected', error: null });

	// Active subscriptions — channel → true; re-sent on reconnect
	private subscriptions = new Set<string>();

	// Message handlers registered by consumers
	private handlers = new Set<MessageHandler>();

	// Reconnect state
	private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
	private reconnectDelay = RECONNECT_INITIAL_MS;
	private manualClose = false;

	// Heartbeat
	private pingTimer: ReturnType<typeof setInterval> | null = null;

	// Pending outbound messages queued while disconnected
	private queue: Envelope[] = [];

	// Request ID counter
	private reqCounter = 0;

	// ---------------------------------------------------------------------------
	// Reactive getters
	// ---------------------------------------------------------------------------

	get status(): ConnectionStatus {
		return this._state.status;
	}

	get error(): string | null {
		return this._state.error;
	}

	get isConnected(): boolean {
		return this._state.status === 'connected';
	}

	// ---------------------------------------------------------------------------
	// Connect / disconnect
	// ---------------------------------------------------------------------------

	connect(wsUrl: string): void {
		if (this.ws && (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING)) {
			return;
		}
		this.url = wsUrl;
		this.manualClose = false;
		this._open();
	}

	disconnect(): void {
		this.manualClose = true;
		this._clearTimers();
		this.ws?.close(1000);
		this.ws = null;
		this._state = { status: 'disconnected', error: null };
	}

	// ---------------------------------------------------------------------------
	// Subscriptions
	// ---------------------------------------------------------------------------

	subscribe(channel: string): void {
		this.subscriptions.add(channel);
		this._send({ type: 'subscribe', id: this._reqId(), payload: { channel } });
	}

	unsubscribe(channel: string): void {
		this.subscriptions.delete(channel);
		this._send({ type: 'unsubscribe', id: this._reqId(), payload: { channel } });
	}

	// ---------------------------------------------------------------------------
	// Outbound messages
	// ---------------------------------------------------------------------------

	presenceUpdate(payload: PresenceUpdatePayload): void {
		this._send({ type: 'presence.update', payload: payload as unknown as object });
	}

	agentInput(payload: AgentInputPayload): void {
		this._send({ type: 'agent.input', payload: payload as unknown as object });
	}

	// ---------------------------------------------------------------------------
	// Handler registration
	// ---------------------------------------------------------------------------

	onMessage(handler: MessageHandler): () => void {
		this.handlers.add(handler);
		return () => this.handlers.delete(handler);
	}

	// ---------------------------------------------------------------------------
	// Private — WebSocket lifecycle
	// ---------------------------------------------------------------------------

	private _open(): void {
		this._state = { status: 'connecting', error: null };
		try {
			this.ws = new WebSocket(this.url);
		} catch (err) {
			this._state = { status: 'failed', error: String(err) };
			return;
		}

		this.ws.onopen = () => {
			this._state = { status: 'connected', error: null };
			this.reconnectDelay = RECONNECT_INITIAL_MS;

			// Flush queued messages
			const queued = this.queue.splice(0);
			for (const msg of queued) {
				this._sendRaw(msg);
			}

			// Re-subscribe to all active channels
			for (const channel of this.subscriptions) {
				this._sendRaw({ type: 'subscribe', id: this._reqId(), payload: { channel } });
			}

			// Start heartbeat
			this._startPing();
		};

		this.ws.onmessage = (event: MessageEvent) => {
			let msg: InboundMessage;
			try {
				msg = JSON.parse(event.data as string) as InboundMessage;
			} catch {
				return; // Ignore malformed frames
			}

			// Dispatch to all handlers
			for (const handler of this.handlers) {
				try {
					handler(msg);
				} catch {
					// Isolate handler errors
				}
			}
		};

		this.ws.onerror = () => {
			// onerror is always followed by onclose — let onclose handle reconnect
		};

		this.ws.onclose = (event: CloseEvent) => {
			this._stopPing();
			this.ws = null;

			if (this.manualClose) return;

			// Non-recoverable close codes
			if (event.code === CLOSE_AUTH_FAILED) {
				this._state = { status: 'failed', error: 'Authentication failed. Please refresh and log in again.' };
				return;
			}
			if (event.code === CLOSE_PROTOCOL_ERROR) {
				this._state = { status: 'failed', error: 'WebSocket protocol error.' };
				return;
			}

			// Exponential backoff reconnect
			this._state = { status: 'disconnected', error: null };
			this._scheduleReconnect();
		};
	}

	private _scheduleReconnect(): void {
		this.reconnectTimer = setTimeout(() => {
			this.reconnectTimer = null;
			if (!this.manualClose) {
				this._open();
			}
		}, this.reconnectDelay);

		// Double delay, cap at max
		this.reconnectDelay = Math.min(this.reconnectDelay * 2, RECONNECT_MAX_MS);
	}

	// ---------------------------------------------------------------------------
	// Private — heartbeat
	// ---------------------------------------------------------------------------

	private _startPing(): void {
		this._stopPing();
		this.pingTimer = setInterval(() => {
			this._sendRaw({ type: 'ping', payload: {} });
		}, PING_INTERVAL_MS);
	}

	private _stopPing(): void {
		if (this.pingTimer !== null) {
			clearInterval(this.pingTimer);
			this.pingTimer = null;
		}
	}

	private _clearTimers(): void {
		this._stopPing();
		if (this.reconnectTimer !== null) {
			clearTimeout(this.reconnectTimer);
			this.reconnectTimer = null;
		}
	}

	// ---------------------------------------------------------------------------
	// Private — send helpers
	// ---------------------------------------------------------------------------

	private _send(msg: Envelope): void {
		if (this.ws?.readyState === WebSocket.OPEN) {
			this._sendRaw(msg);
		} else {
			// Queue for when connection is restored
			this.queue.push(msg);
		}
	}

	private _sendRaw(msg: Envelope): void {
		try {
			this.ws?.send(JSON.stringify(msg));
		} catch {
			// Socket closed mid-send — will reconnect
		}
	}

	private _reqId(): string {
		return `req_${++this.reqCounter}`;
	}
}

// ---------------------------------------------------------------------------
// Singleton export
// ---------------------------------------------------------------------------

export const wsClient = new WsClient();
