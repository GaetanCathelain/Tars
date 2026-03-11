<script lang="ts">
	import { goto } from '$app/navigation';
	import { cn } from '$lib/utils';
	import { auth } from '$lib/stores/auth.svelte';
	import { tasksStore } from '$lib/stores/tasks.svelte';
	import { wsStore } from '$lib/stores/websocket.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Separator } from '$lib/components/ui/separator';
	import NewTaskDialog from './new-task-dialog.svelte';

	function statusColor(status: string): string {
		switch (status) {
			case 'completed': return 'bg-emerald-500';
			case 'running': return 'bg-blue-500 animate-pulse';
			case 'failed': return 'bg-red-500';
			default: return 'bg-zinc-500';
		}
	}

	function handleLogout() {
		auth.logout();
		goto('/login');
	}

	function handleSelectTask(id: string) {
		tasksStore.selectTask(id);
		goto(`/tasks/${id}`);
	}
</script>

<aside class="w-[280px] h-screen flex flex-col bg-zinc-950 border-r border-border">
	<!-- Header -->
	<div class="px-5 py-5">
		<div class="flex items-center gap-2">
			<span class="text-lg">🤖</span>
			<div>
				<h1 class="text-lg font-semibold text-foreground">TARS</h1>
				<p class="text-xs text-muted-foreground">Orchestrator</p>
			</div>
		</div>
	</div>

	<Separator />

	<!-- Tasks -->
	<div class="flex-1 overflow-y-auto px-3 py-4">
		<p class="px-2 mb-3 text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
			Tasks
		</p>
		<div class="space-y-1">
			{#each tasksStore.tasks as task (task.id)}
				<button
					class={cn(
						'flex items-center gap-3 px-3 py-2 rounded-md w-full text-left text-sm transition-colors',
						tasksStore.selectedTaskId === task.id
							? 'bg-accent text-accent-foreground'
							: 'text-muted-foreground hover:bg-accent/50 hover:text-accent-foreground'
					)}
					onclick={() => handleSelectTask(task.id)}
				>
					<div class={cn('h-2 w-2 rounded-full shrink-0', statusColor(task.status))}></div>
					<span class="truncate">{task.title}</span>
				</button>
			{/each}
		</div>
	</div>

	<!-- New Task -->
	<div class="px-3 pb-3">
		<NewTaskDialog />
	</div>

	<Separator />

	<!-- Footer -->
	<div class="px-5 py-4 flex items-center justify-between">
		<div class="flex items-center gap-2">
			<div class={cn(
				'h-2 w-2 rounded-full',
				wsStore.status === 'connected' ? 'bg-emerald-500' : 'bg-zinc-600'
			)}></div>
			<span class="text-xs text-muted-foreground">{auth.user?.username ?? 'user'}</span>
		</div>
		<Button variant="ghost" size="sm" class="text-xs text-muted-foreground h-7" onclick={handleLogout}>
			Logout
		</Button>
	</div>
</aside>
