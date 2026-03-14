<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/utils/cn';

	interface Props {
		class?: string;
		children: Snippet;
	}

	let { class: className, children }: Props = $props();
</script>

<div class={cn('rounded-xl border border-zinc-800 bg-zinc-900 text-zinc-50 shadow', className)}>
	{@render children()}
</div>
