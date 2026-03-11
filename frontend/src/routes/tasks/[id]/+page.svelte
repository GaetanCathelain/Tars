<script lang="ts">
	import { page } from '$app/state';
	import { tasksStore } from '$lib/stores/tasks.svelte';
	import { messagesStore, mockWorkerSession } from '$lib/stores/messages.svelte';
	import type { WorkerEvent } from '$lib/stores/messages.svelte';
	import { wsStore } from '$lib/stores/websocket.svelte';
	import { workersStore } from '$lib/stores/workers.svelte';
	import type { Message, WorkerSession } from '$lib/types';
	import WorkerCard from '$lib/components/WorkerCard.svelte';
	import { onMount, tick } from 'svelte';

	let messageInput = $state('');
	let messagesContainer: HTMLDivElement | undefined = $state();

	const taskId = $derived(page.params.id);
	const task = $derived(tasksStore.tasks.find((t) => t.id === taskId) ?? null);

	// Track previous taskId for cleanup
	let prevTaskId = $state<string | null>(null);

	$effect(() => {
		if (taskId) {
			// Unsubscribe from previous task
			if (prevTaskId && prevTaskId !== taskId) {
				wsStore.unsubscribeFromTask(prevTaskId);
			}
			prevTaskId = taskId;

			tasksStore.selectTask(taskId);
			messagesStore.fetchMessages(taskId);
			wsStore.subscribeToTask(taskId);
		}
	});

	// Auto-scroll to bottom when timeline changes
	$effect(() => {
		if (messagesStore.timeline.length && messagesContainer) {
			tick().then(() => {
				if (messagesContainer) {
					messagesContainer.scrollTop = messagesContainer.scrollHeight;
				}
			});
		}
	});

	function isWorkerEvent(entry: unknown): entry is WorkerEvent {
		return (entry as WorkerEvent).type === 'worker_event';
	}

	function getWorkerSession(sessionId: string): WorkerSession {
		const session = workersStore.getWorker(sessionId);
		if (session) return session;
		// Fallback to mock worker for mock data
		if (sessionId === mockWorkerSession.id) return mockWorkerSession;
		// Return a placeholder
		return {
			id: sessionId,
			task_id: taskId,
			status: 'running',
			command: 'claude-code',
			started_at: new Date().toISOString()
		};
	}

	function senderInitial(type: string): string {
		switch (type) {
			case 'user': return 'U';
			case 'tars': return 'T';
			case 'system': return 'S';
			default: return '?';
		}
	}

	function senderLabel(type: string): string {
		switch (type) {
			case 'user': return 'You';
			case 'tars': return 'TARS';
			case 'system': return 'System';
			default: return type;
		}
	}

	function senderAvatarClass(type: string): string {
		switch (type) {
			case 'tars': return 'bg-accent-muted text-accent';
			case 'user': return 'bg-bg-elevated text-text-secondary';
			default: return 'bg-bg-tertiary text-text-tertiary';
		}
	}

	function statusBadgeClass(status: string): string {
		switch (status) {
			case 'open': return 'text-text-tertiary border-border bg-bg-tertiary';
			case 'running': return 'text-running border-running/30 bg-running/10';
			case 'completed': return 'text-success border-success/30 bg-success/10';
			case 'failed': return 'text-danger border-danger/30 bg-danger/10';
			default: return 'text-text-tertiary border-border bg-bg-tertiary';
		}
	}

	function formatTime(iso: string): string {
		try {
			const d = new Date(iso);
			return d.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', hour12: false });
		} catch {
			return '';
		}
	}

	async function handleSend(e: Event) {
		e.preventDefault();
		if (!messageInput.trim() || !taskId) return;
		const content = messageInput.trim();
		messageInput = '';
		await messagesStore.sendMessage(taskId, content);
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			handleSend(e);
		}
	}
</script>

{#if !task}
	<div class="flex-1 flex items-center justify-center">
		<p class="text-text-tertiary text-[13px]">Task not found</p>
	</div>
{:else}
	<!-- Header -->
	<header class="flex items-center gap-3 px-6 py-3 border-b border-border shrink-0">
		<h2 class="text-[14px] font-medium text-text-primary truncate tracking-[-0.01em]">{task.title}</h2>
		<span class="px-2 py-0.5 text-[11px] font-medium border rounded-md shrink-0 {statusBadgeClass(task.status)}">
			{task.status}
		</span>
	</header>

	<!-- Messages / Timeline -->
	<div
		bind:this={messagesContainer}
		class="flex-1 overflow-y-auto px-6 py-4"
	>
		{#if messagesStore.loading}
			<div class="flex justify-center py-8">
				<p class="text-text-tertiary text-[13px]">Loading messages...</p>
			</div>
		{:else if messagesStore.timeline.length === 0}
			<div class="flex flex-col items-center justify-center py-16 text-center">
				<p class="text-text-tertiary text-[13px]">No messages yet. Send a message to start.</p>
			</div>
		{:else}
			<div class="space-y-5">
				{#each messagesStore.timeline as entry (entry.id)}
					{#if isWorkerEvent(entry) && entry.event === 'start'}
						<!-- Worker Card inline in the timeline -->
						<WorkerCard session={getWorkerSession(entry.session_id)} />
					{:else if !isWorkerEvent(entry)}
						{@const message = entry as Message}
						{#if message.sender_type === 'system'}
							<div class="py-1">
								<p class="text-[12px] italic text-text-tertiary">{message.content}</p>
							</div>
						{:else}
							<div class="flex gap-3">
								<!-- Avatar -->
								<div class="w-6 h-6 rounded-full flex items-center justify-center shrink-0 text-[10px] font-medium {senderAvatarClass(message.sender_type)}">
									{senderInitial(message.sender_type)}
								</div>

								<!-- Content -->
								<div class="flex-1 min-w-0">
									<div class="flex items-baseline gap-2 mb-0.5">
										<span class="text-[13px] font-medium text-zinc-300">
											{senderLabel(message.sender_type)}
										</span>
										<span class="text-[11px] text-text-tertiary">
											{formatTime(message.created_at)}
										</span>
									</div>
									<div class="text-[13px] text-zinc-200 leading-[1.5] whitespace-pre-wrap">
										{message.content}
									</div>
								</div>
							</div>
						{/if}
					{/if}
				{/each}
			</div>
		{/if}
	</div>

	<!-- Message input -->
	<form onsubmit={handleSend} class="shrink-0 border-t border-border px-6 py-3">
		<div class="flex gap-2">
			<input
				type="text"
				bind:value={messageInput}
				onkeydown={handleKeydown}
				placeholder="Send a message..."
				class="flex-1 bg-bg-tertiary border border-border rounded-md px-3.5 py-2 text-[13px] text-text-primary
					placeholder:text-text-tertiary focus:outline-none focus:border-accent transition-all duration-150"
			/>
			<button
				type="submit"
				disabled={!messageInput.trim()}
				class="px-4 py-2 bg-accent text-white text-[13px] font-medium rounded-md
					hover:bg-accent-hover disabled:opacity-30 disabled:cursor-not-allowed transition-all duration-150"
			>
				Send
			</button>
		</div>
	</form>
{/if}
