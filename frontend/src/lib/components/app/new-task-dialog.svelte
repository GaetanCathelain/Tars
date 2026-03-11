<script lang="ts">
	import { goto } from '$app/navigation';
	import { tasksStore } from '$lib/stores/tasks.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import * as Dialog from '$lib/components/ui/dialog';

	let open = $state(false);
	let title = $state('');
	let loading = $state(false);

	async function handleCreate() {
		if (!title.trim()) return;
		loading = true;
		try {
			const task = await tasksStore.createTask(title.trim());
			title = '';
			open = false;
			tasksStore.selectTask(task.id);
			goto(`/tasks/${task.id}`);
		} finally {
			loading = false;
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') handleCreate();
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Trigger>
		{#snippet children({ props })}
			<Button {...props} variant="outline" class="w-full border-dashed">
				+ New Task
			</Button>
		{/snippet}
	</Dialog.Trigger>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>Create Task</Dialog.Title>
			<Dialog.Description>Describe what you want TARS to work on.</Dialog.Description>
		</Dialog.Header>
		<div class="space-y-4 py-4">
			<Input
				placeholder="Task description..."
				bind:value={title}
				onkeydown={handleKeydown}
			/>
		</div>
		<Dialog.Footer>
			<Button onclick={handleCreate} disabled={loading || !title.trim()}>
				{loading ? 'Creating...' : 'Create'}
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
