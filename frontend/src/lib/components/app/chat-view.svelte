<script lang="ts">
	import { tick } from 'svelte';
	import { tasksStore } from '$lib/stores/tasks.svelte';
	import { messagesStore } from '$lib/stores/messages.svelte';
	import { workersStore } from '$lib/stores/workers.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Separator } from '$lib/components/ui/separator';
	import * as ScrollArea from '$lib/components/ui/scroll-area';
	import MessageItem from './message-item.svelte';
	import WorkerCard from './worker-card.svelte';

	let inputValue = $state('');
	let sending = $state(false);
	let messagesEnd: HTMLDivElement | undefined = $state();

	const task = $derived(tasksStore.selectedTask);
	const messages = $derived(messagesStore.messages);
	const workers = $derived(task ? workersStore.getWorkers(task.id) : []);

	function statusVariant(status: string): 'default' | 'secondary' | 'destructive' | 'outline' {
		switch (status) {
			case 'completed': return 'default';
			case 'running': return 'secondary';
			case 'failed': return 'destructive';
			default: return 'outline';
		}
	}

	async function handleSend() {
		if (!inputValue.trim() || !task) return;
		sending = true;
		try {
			await messagesStore.sendMessage(task.id, inputValue.trim());
			inputValue = '';
			await tick();
			messagesEnd?.scrollIntoView({ behavior: 'smooth' });
		} finally {
			sending = false;
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && !e.shiftKey) {
			e.preventDefault();
			handleSend();
		}
	}
</script>

<main class="flex-1 flex flex-col h-screen min-w-0">
	{#if task}
		<!-- Header -->
		<div class="px-6 py-4 border-b border-border flex items-center justify-between shrink-0">
			<h2 class="text-lg font-semibold text-foreground truncate">{task.title}</h2>
			<Badge variant={statusVariant(task.status)}>
				{task.status}
			</Badge>
		</div>

		<!-- Messages Area -->
		<div class="flex-1 overflow-y-auto">
			<div class="px-6 py-6 space-y-6">
				{#each messages as message (message.id)}
					<MessageItem {message} />
				{/each}

				{#each workers as worker (worker.id)}
					<WorkerCard {worker} />
				{/each}

				<div bind:this={messagesEnd}></div>
			</div>
		</div>

		<!-- Input Area -->
		<div class="border-t border-border px-6 py-4 shrink-0">
			<div class="flex gap-3">
				<Input
					placeholder="Type a message..."
					bind:value={inputValue}
					onkeydown={handleKeydown}
					class="flex-1"
				/>
				<Button onclick={handleSend} disabled={sending || !inputValue.trim()}>
					Send
				</Button>
			</div>
		</div>
	{:else}
		<div class="flex-1 flex items-center justify-center">
			<p class="text-muted-foreground">Select a task to view messages</p>
		</div>
	{/if}
</main>
