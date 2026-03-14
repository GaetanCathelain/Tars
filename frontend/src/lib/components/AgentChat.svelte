<script lang="ts">
	import { tick } from 'svelte';
	import { wsClient } from '$lib/ws/client.svelte';
	import { chat } from '$lib/stores/chat.svelte';
	import { api, ApiError } from '$lib/utils/api';
	import { Avatar } from '$lib/components/ui/avatar';
	import { Button } from '$lib/components/ui/button';
	import type { AgentLogLine } from '$shared/types/models';
	import { SendHorizontal, Bot, User, AlertCircle } from 'lucide-svelte';

	interface Props {
		agentId: string;
		repoId: string;
		agentName: string;
		agentStatus: string;
		/** Live output lines from the agents store — used to show agent responses */
		outputLines?: AgentLogLine[];
	}

	let {
		agentId,
		repoId,
		agentName,
		agentStatus,
		outputLines = []
	}: Props = $props();

	let inputText = $state('');
	let messagesEl: HTMLDivElement;
	let inputEl: HTMLTextAreaElement;

	const isRunning = $derived(agentStatus === 'running' || agentStatus === 'starting');
	const userMessages = $derived(chat.getMessages(agentId));

	// Build a unified timeline of user messages + agent output chunks
	// Each entry has a timestamp so we can interleave them chronologically
	interface UserEntry {
		kind: 'user';
		id: string;
		text: string;
		ts: string;
	}
	interface AgentEntry {
		kind: 'agent';
		seq: number;
		text: string;
		stream: 'stdout' | 'stderr';
		ts: string;
	}
	type TimelineEntry = UserEntry | AgentEntry;

	const timeline = $derived.by<TimelineEntry[]>(() => {
		const entries: TimelineEntry[] = [
			...userMessages.map((m) => ({
				kind: 'user' as const,
				id: m.id,
				text: m.text,
				ts: m.sentAt
			})),
			// Only include stdout/stderr lines that look like agent responses
			// (not raw tool calls / progress indicators — include all for full transparency)
			...outputLines.map((l) => ({
				kind: 'agent' as const,
				seq: l.seq,
				text: l.text,
				stream: l.stream,
				ts: l.ts
			}))
		];
		// Sort by timestamp, then by seq for same-timestamp agent lines
		entries.sort((a, b) => {
			const tsDiff = a.ts.localeCompare(b.ts);
			if (tsDiff !== 0) return tsDiff;
			if (a.kind === 'agent' && b.kind === 'agent') return a.seq - b.seq;
			return 0;
		});
		return entries;
	});

	// Auto-scroll to bottom when timeline grows
	$effect(() => {
		// Track timeline length reactively
		const _len = timeline.length;
		void _len;
		tick().then(() => {
			if (messagesEl) {
				messagesEl.scrollTop = messagesEl.scrollHeight;
			}
		});
	});

	async function sendMessage() {
		const text = inputText.trim();
		if (!text || !isRunning || chat.sending) return;

		inputText = '';
		chat.clearError();

		const msg = {
			id: crypto.randomUUID(),
			agentId,
			role: 'user' as const,
			text,
			sentAt: new Date().toISOString()
		};
		chat.addMessage(msg);
		chat.setSending(true);

		try {
			// Prefer WS for lower latency; fall back to REST
			if (wsClient.isConnected) {
				wsClient.agentInput({ agent_id: agentId, text });
			} else {
				await api.post(`/repos/${repoId}/agents/${agentId}/input`, { text });
			}
		} catch (err) {
			const message = err instanceof ApiError ? err.message : 'Failed to send message.';
			chat.setError(message);
		} finally {
			chat.setSending(false);
		}

		// Refocus input
		await tick();
		inputEl?.focus();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			sendMessage();
		}
	}

	function formatTime(iso: string): string {
		try {
			return new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
		} catch {
			return '';
		}
	}

	// Group consecutive agent lines into blocks for cleaner rendering
	interface AgentBlock {
		kind: 'agent-block';
		lines: AgentEntry[];
		ts: string;
	}
	interface UserBlock {
		kind: 'user-block';
		entry: UserEntry;
	}
	type Block = AgentBlock | UserBlock;

	const blocks = $derived.by<Block[]>(() => {
		const result: Block[] = [];
		for (const entry of timeline) {
			if (entry.kind === 'user') {
				result.push({ kind: 'user-block', entry });
			} else {
				const last = result[result.length - 1];
				if (last?.kind === 'agent-block') {
					last.lines.push(entry);
				} else {
					result.push({ kind: 'agent-block', lines: [entry], ts: entry.ts });
				}
			}
		}
		return result;
	});
</script>

<div class="flex h-full flex-col bg-zinc-950">
	<!-- Message history -->
	<div bind:this={messagesEl} class="flex-1 space-y-4 overflow-y-auto p-4">
		{#if blocks.length === 0}
			<div class="flex flex-col items-center justify-center py-12 text-center">
				<Bot class="mb-3 h-10 w-10 text-zinc-700" />
				<p class="text-sm text-zinc-500">No messages yet.</p>
				<p class="mt-1 text-xs text-zinc-600">
					{isRunning ? 'Send a message to guide the agent.' : 'Agent is not running.'}
				</p>
			</div>
		{:else}
			{#each blocks as block (block.kind === 'user-block' ? block.entry.id : block.ts + block.lines[0]?.seq)}
				{#if block.kind === 'user-block'}
					<!-- User message -->
					<div class="flex items-start justify-end gap-3">
						<div class="max-w-[75%]">
							<div class="rounded-2xl rounded-tr-sm bg-zinc-700 px-4 py-2.5 text-sm text-zinc-50">
								{block.entry.text}
							</div>
							<p class="mt-1 text-right text-xs text-zinc-600">
								{formatTime(block.entry.ts)}
							</p>
						</div>
						<div class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-zinc-700">
							<User class="h-4 w-4 text-zinc-300" />
						</div>
					</div>
				{:else}
					<!-- Agent output block -->
					<div class="flex items-start gap-3">
						<div class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-zinc-800">
							<Bot class="h-4 w-4 text-zinc-400" />
						</div>
						<div class="min-w-0 flex-1">
							<p class="mb-1 text-xs font-medium text-zinc-500">{agentName}</p>
							<div class="rounded-2xl rounded-tl-sm bg-zinc-900 px-4 py-2.5">
								<pre class="whitespace-pre-wrap break-words font-mono text-xs leading-relaxed text-zinc-200">{block.lines.map((l) => l.text).join('')}</pre>
							</div>
							<p class="mt-1 text-xs text-zinc-600">{formatTime(block.ts)}</p>
						</div>
					</div>
				{/if}
			{/each}
		{/if}
	</div>

	<!-- Error banner -->
	{#if chat.error}
		<div class="mx-4 mb-2 flex items-center gap-2 rounded-md border border-red-900 bg-red-950/50 px-3 py-2 text-xs text-red-400">
			<AlertCircle class="h-3.5 w-3.5 shrink-0" />
			{chat.error}
		</div>
	{/if}

	<!-- Input -->
	<div class="border-t border-zinc-800 p-3">
		<div class="flex items-end gap-2 rounded-xl border border-zinc-700 bg-zinc-900 px-3 py-2 focus-within:border-zinc-500 focus-within:ring-1 focus-within:ring-zinc-500">
			<textarea
				bind:this={inputEl}
				bind:value={inputText}
				onkeydown={handleKeydown}
				disabled={!isRunning || chat.sending}
				placeholder={isRunning ? 'Send a message to the agent… (Enter to send, Shift+Enter for newline)' : 'Agent is not running'}
				rows={1}
				class="max-h-32 flex-1 resize-none bg-transparent text-sm text-zinc-50 placeholder:text-zinc-600 focus:outline-none disabled:cursor-not-allowed disabled:opacity-50"
				style="field-sizing: content"
			></textarea>
			<Button
				onclick={sendMessage}
				disabled={!inputText.trim() || !isRunning || chat.sending}
				size="icon"
				class="h-7 w-7 shrink-0"
			>
				<SendHorizontal class="h-3.5 w-3.5" />
			</Button>
		</div>
		<p class="mt-1.5 text-center text-xs text-zinc-700">
			Enter to send · Shift+Enter for new line
		</p>
	</div>
</div>
