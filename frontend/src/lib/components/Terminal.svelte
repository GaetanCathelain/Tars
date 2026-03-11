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
				background: '#0a0a0a',
				foreground: '#e4e4e7',
				cursor: 'transparent',
				cursorAccent: 'transparent',
				selectionBackground: '#27272a',
				selectionForeground: '#e4e4e7',
				black: '#18181b',
				red: '#f87171',
				green: '#34d399',
				yellow: '#fbbf24',
				blue: '#818cf8',
				magenta: '#c084fc',
				cyan: '#67e8f9',
				white: '#e4e4e7',
				brightBlack: '#71717a',
				brightRed: '#fca5a5',
				brightGreen: '#6ee7b7',
				brightYellow: '#fde68a',
				brightBlue: '#a5b4fc',
				brightMagenta: '#d8b4fe',
				brightCyan: '#a5f3fc',
				brightWhite: '#fafafa'
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
		scrollbar-color: #27272a #0a0a0a;
	}
	.terminal-container :global(.xterm-viewport::-webkit-scrollbar) {
		width: 4px;
	}
	.terminal-container :global(.xterm-viewport::-webkit-scrollbar-track) {
		background: #0a0a0a;
	}
	.terminal-container :global(.xterm-viewport::-webkit-scrollbar-thumb) {
		background: #27272a;
		border-radius: 2px;
	}
</style>
