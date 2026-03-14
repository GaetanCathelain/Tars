<script lang="ts">
	import { cn } from '$lib/utils/cn';

	interface Props {
		src?: string;
		alt?: string;
		fallback?: string;
		class?: string;
	}

	let { src, alt = '', fallback, class: className }: Props = $props();

	let imgError = $state(false);

	const initials = $derived(
		fallback ??
			alt
				.split(' ')
				.map((n) => n[0])
				.join('')
				.toUpperCase()
				.slice(0, 2)
	);
</script>

<div
	class={cn(
		'relative flex h-8 w-8 shrink-0 overflow-hidden rounded-full bg-zinc-800',
		className
	)}
>
	{#if src && !imgError}
		<img
			{src}
			{alt}
			class="aspect-square h-full w-full object-cover"
			onerror={() => (imgError = true)}
		/>
	{:else}
		<span class="flex h-full w-full items-center justify-center text-xs font-medium text-zinc-400">
			{initials}
		</span>
	{/if}
</div>
