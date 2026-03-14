<script lang="ts">
	import { onMount } from 'svelte';
	import type { AgentLogLine } from '$shared/types/models';

	interface Props {
		agentId: string;
		lines?: AgentLogLine[];
		class?: string;
	}

	let { agentId, lines = [], class: className = '' }: Props = $props();

	let container: HTMLDivElement;
	let terminal: import('@xterm/xterm').Terminal | null = null;
	let fitAddon: import('@xterm/addon-fit').FitAddon | null = null;
	// Track highest seq written to avoid duplicates
	let lastSeq = $state(-1);
	let resizeObserver: ResizeObserver | null = null;
	let ready = $state(false);

	function writeLine(line: AgentLogLine) {
		if (!terminal) return;
		const text = line.stream === 'stderr'
			? `\x1b[31m${line.text}\x1b[0m`
			: line.text;
		// xterm.js needs \r\n for proper line endings
		terminal.write(text.replace(/(?<!\r)\n/g, '\r\n'));
	}

	onMount(() => {
		let disposed = false;

		(async () => {
			const [{ Terminal }, { FitAddon }, { WebLinksAddon }] = await Promise.all([
				import('@xterm/xterm'),
				import('@xterm/addon-fit'),
				import('@xterm/addon-web-links')
			]);

			if (disposed) return;

			terminal = new Terminal({
				theme: {
					background: '#09090b',
					foreground: '#fafafa',
					cursor: '#a1a1aa',
					selectionBackground: '#3f3f46',
					black: '#18181b',
					red: '#ef4444',
					green: '#22c55e',
					yellow: '#eab308',
					blue: '#3b82f6',
					magenta: '#a855f7',
					cyan: '#06b6d4',
					white: '#d4d4d8',
					brightBlack: '#3f3f46',
					brightRed: '#f87171',
					brightGreen: '#4ade80',
					brightYellow: '#fbbf24',
					brightBlue: '#60a5fa',
					brightMagenta: '#c084fc',
					brightCyan: '#22d3ee',
					brightWhite: '#fafafa'
				},
				fontFamily: '"JetBrains Mono", "Fira Code", "Cascadia Code", Menlo, monospace',
				fontSize: 13,
				lineHeight: 1.4,
				cursorBlink: false,
				cursorStyle: 'bar',
				scrollback: 5000,
				convertEol: false,
				disableStdin: true,
				allowTransparency: true,
				drawBoldTextInBrightColors: true
			});

			fitAddon = new FitAddon();
			terminal.loadAddon(fitAddon);
			terminal.loadAddon(new WebLinksAddon());
			terminal.open(container);
			fitAddon.fit();

			// Write historical lines from props
			for (const line of lines) {
				writeLine(line);
				lastSeq = Math.max(lastSeq, line.seq);
			}
			terminal.scrollToBottom();

			resizeObserver = new ResizeObserver(() => fitAddon?.fit());
			resizeObserver.observe(container);

			ready = true;
		})();

		return () => {
			disposed = true;
			resizeObserver?.disconnect();
			terminal?.dispose();
			terminal = null;
		};
	});

	// Write new lines streamed via WS (reactive on lines array changes)
	$effect(() => {
		if (!ready || !terminal) return;
		const newLines = lines.filter((l) => l.seq > lastSeq);
		if (newLines.length === 0) return;
		newLines.sort((a, b) => a.seq - b.seq);
		for (const line of newLines) {
			writeLine(line);
		}
		lastSeq = newLines[newLines.length - 1]?.seq ?? lastSeq;
		terminal.scrollToBottom();
	});
</script>

<div
	bind:this={container}
	class="h-full w-full overflow-hidden rounded-lg bg-zinc-950 {className}"
	data-agent-id={agentId}
></div>

<style>
	:global(.xterm) {
		height: 100%;
		padding: 8px;
	}
	:global(.xterm-viewport) {
		background-color: transparent !important;
	}
</style>
