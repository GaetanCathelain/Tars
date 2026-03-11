<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { Terminal } from '@xterm/xterm';
	import { FitAddon } from '@xterm/addon-fit';
	import { WebLinksAddon } from '@xterm/addon-web-links';
	import '@xterm/xterm/css/xterm.css';
	import { api } from '$lib/api';
	import { wsStore } from '$lib/stores/websocket.svelte';

	interface OutputChunk {
		id: string;
		session_id: string;
		data: string;
		created_at: string;
	}

	let { sessionId, isLive }: { sessionId: string; isLive: boolean } = $props();

	let containerEl: HTMLDivElement | undefined = $state();
	let terminal: Terminal | undefined;
	let fitAddon: FitAddon | undefined;
	let resizeObserver: ResizeObserver | undefined;

	function writeData(data: Uint8Array) {
		terminal?.write(data);
	}

	function handleOutput(data: Uint8Array) {
		writeData(data);
	}

	onMount(async () => {
		if (!containerEl) return;

		terminal = new Terminal({
			cursorBlink: false,
			cursorStyle: 'underline',
			disableStdin: true,
			convertEol: true,
			scrollback: 10000,
			fontSize: 13,
			fontFamily: "'JetBrains Mono', 'Fira Code', 'Cascadia Code', Menlo, Monaco, monospace",
			theme: {
				background: '#09090b',
				foreground: '#fafafa',
				cursor: '#fafafa',
				selectionBackground: '#27272a',
				black: '#09090b',
				red: '#ef4444',
				green: '#22c55e',
				yellow: '#eab308',
				blue: '#3b82f6',
				magenta: '#a855f7',
				cyan: '#06b6d4',
				white: '#fafafa',
				brightBlack: '#71717a',
				brightRed: '#f87171',
				brightGreen: '#4ade80',
				brightYellow: '#facc15',
				brightBlue: '#60a5fa',
				brightMagenta: '#c084fc',
				brightCyan: '#22d3ee',
				brightWhite: '#ffffff'
			}
		});

		fitAddon = new FitAddon();
		terminal.loadAddon(fitAddon);
		terminal.loadAddon(new WebLinksAddon());

		terminal.open(containerEl);

		// Initial fit
		try { fitAddon.fit(); } catch { /* container may not be visible yet */ }

		// Auto-fit on resize
		resizeObserver = new ResizeObserver(() => {
			try { fitAddon?.fit(); } catch { /* ignore */ }
		});
		resizeObserver.observe(containerEl);

		// Load historical output for non-live workers
		if (!isLive) {
			try {
				const chunks = await api.get<OutputChunk[]>(`/workers/${sessionId}/output`);
				for (const chunk of chunks) {
					const decoded = atob(chunk.data);
					const bytes = Uint8Array.from(decoded, (c) => c.charCodeAt(0));
					terminal.write(bytes);
				}
			} catch {
				terminal.write('\x1b[90m(no output available)\x1b[0m');
			}
		}

		// Subscribe to live output
		if (isLive) {
			wsStore.subscribe(sessionId, handleOutput);
		}
	});

	onDestroy(() => {
		wsStore.unsubscribe(sessionId, handleOutput);
		resizeObserver?.disconnect();
		terminal?.dispose();
	});

	// React to isLive changes — subscribe/unsubscribe as status changes
	$effect(() => {
		if (isLive && terminal) {
			wsStore.subscribe(sessionId, handleOutput);
		}
		return () => {
			wsStore.unsubscribe(sessionId, handleOutput);
		};
	});
</script>

<div
	bind:this={containerEl}
	class="min-h-[200px] max-h-[400px] overflow-hidden bg-[#09090b] p-1"
></div>
