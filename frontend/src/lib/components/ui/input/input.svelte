<script lang="ts">
	import { cn } from '$lib/utils/cn';

	interface Props {
		class?: string;
		type?: string;
		placeholder?: string;
		value?: string;
		disabled?: boolean;
		id?: string;
		name?: string;
		required?: boolean;
		oninput?: (e: Event) => void;
		onkeydown?: (e: KeyboardEvent) => void;
	}

	let {
		class: className,
		type = 'text',
		placeholder,
		value = $bindable(''),
		disabled = false,
		id,
		name,
		required = false,
		oninput,
		onkeydown
	}: Props = $props();
</script>

<input
	{type}
	{placeholder}
	{disabled}
	{id}
	{name}
	{required}
	{oninput}
	{onkeydown}
	bind:value
	class={cn(
		'flex h-9 w-full rounded-md border border-zinc-700 bg-zinc-800 px-3 py-1 text-sm text-zinc-50 shadow-sm transition-colors placeholder:text-zinc-500 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-zinc-400 disabled:cursor-not-allowed disabled:opacity-50',
		className
	)}
/>
