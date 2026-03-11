<script lang="ts">
	import type { Message } from '$lib/types';
	import { cn } from '$lib/utils';
	import * as Avatar from '$lib/components/ui/avatar';

	let { message }: { message: Message } = $props();

	const isSystem = $derived(message.sender_type === 'system');

	const senderName = $derived(
		message.sender_type === 'user' ? 'You' :
		message.sender_type === 'tars' ? 'TARS' : 'System'
	);

	const initial = $derived(
		message.sender_type === 'user' ? 'U' :
		message.sender_type === 'tars' ? 'T' : 'S'
	);

	const avatarClass = $derived(
		message.sender_type === 'user' ? 'bg-blue-600 text-white' :
		message.sender_type === 'tars' ? 'bg-primary text-primary-foreground' :
		'bg-muted text-muted-foreground'
	);

	function formatTime(dateStr: string): string {
		const d = new Date(dateStr);
		return d.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' });
	}
</script>

{#if isSystem}
	<div class="flex items-start gap-3 py-1">
		<div class="border-l-2 border-border pl-3">
			<p class="text-xs text-muted-foreground italic">{message.content}</p>
			<span class="text-[10px] text-muted-foreground/60">{formatTime(message.created_at)}</span>
		</div>
	</div>
{:else}
	<div class="flex gap-4">
		<Avatar.Root class="h-8 w-8 shrink-0">
			<Avatar.Fallback class={avatarClass}>{initial}</Avatar.Fallback>
		</Avatar.Root>
		<div class="flex-1 space-y-1 min-w-0">
			<div class="flex items-baseline gap-2">
				<span class="text-sm font-medium text-foreground">{senderName}</span>
				<span class="text-xs text-muted-foreground">{formatTime(message.created_at)}</span>
			</div>
			<p class="text-sm leading-relaxed text-foreground/90">{message.content}</p>
		</div>
	</div>
{/if}
