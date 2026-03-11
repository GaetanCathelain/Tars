<script lang="ts">
	import { onMount } from 'svelte';
	import { Terminal } from 'xterm';
	import { FitAddon } from '@xterm/addon-fit';
	import { WebLinksAddon } from '@xterm/addon-web-links';
	import 'xterm/css/xterm.css';

	let {
		sessionId,
		initialData = []
	}: {
		sessionId: string;
		initialData?: Uint8Array[];
	} = $props();

	let terminalEl: HTMLDivElement | undefined = $state();
	let terminal: Terminal | null = null;
	let fitAddon: FitAddon | null = null;

	export function write(data: Uint8Array) {
		if (terminal) {
			terminal.write(data);
		}
	}

	onMount(() => {
		if (!terminalEl) return;

		const term = new Terminal({
			cursorBlink: false,
			cursorStyle: 'bar',
			cursorInactiveStyle: 'none',
			disableStdin: true,
			fontSize: 13,
			fontFamily: "'JetBrains Mono', 'Fira Code', 'Cascadia Code', monospace",
			lineHeight: 1.4,
			scrollback: 5000,
			theme: {
				background: '#0a0a0f',
				foreground: '#e4e4e7',
				cursor: 'transparent',
				cursorAccent: 'transparent',
				selectionBackground: '#2a2a3e',
				selectionForeground: '#e4e4e7',
				black: '#1a1a2e',
				red: '#ff4466',
				green: '#00ff88',
				yellow: '#ffaa00',
				blue: '#00d4ff',
				magenta: '#d066ff',
				cyan: '#00d4ff',
				white: '#e4e4e7',
				brightBlack: '#8888a0',
				brightRed: '#ff6688',
				brightGreen: '#44ffaa',
				brightYellow: '#ffcc44',
				brightBlue: '#44ddff',
				brightMagenta: '#dd88ff',
				brightCyan: '#44ddff',
				brightWhite: '#ffffff'
			}
		});

		const fit = new FitAddon();
		const webLinks = new WebLinksAddon();

		term.loadAddon(fit);
		term.loadAddon(webLinks);

		term.open(terminalEl);
		fit.fit();

		terminal = term;
		fitAddon = fit;

		// Replay initial data
		for (const chunk of initialData) {
			term.write(chunk);
		}

		// Resize observer
		const observer = new ResizeObserver(() => {
			if (fitAddon) {
				try {
					fitAddon.fit();
				} catch {
					// ignore fit errors during transitions
				}
			}
		});
		observer.observe(terminalEl);

		return () => {
			observer.disconnect();
			term.dispose();
			terminal = null;
			fitAddon = null;
		};
	});
</script>

<div
	bind:this={terminalEl}
	class="terminal-container w-full"
	style="min-height: 200px; max-height: 400px;"
	data-session-id={sessionId}
></div>

<style>
	.terminal-container :global(.xterm) {
		padding: 8px;
	}
	.terminal-container :global(.xterm-viewport) {
		scrollbar-width: thin;
		scrollbar-color: #2a2a3e #0a0a0f;
	}
	.terminal-container :global(.xterm-viewport::-webkit-scrollbar) {
		width: 6px;
	}
	.terminal-container :global(.xterm-viewport::-webkit-scrollbar-track) {
		background: #0a0a0f;
	}
	.terminal-container :global(.xterm-viewport::-webkit-scrollbar-thumb) {
		background: #2a2a3e;
		border-radius: 3px;
	}
</style>
