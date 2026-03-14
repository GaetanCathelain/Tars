<script lang="ts">
	import type { Snippet } from 'svelte';
	import AppLayout from '$lib/components/AppLayout.svelte';

	interface Props {
		children: Snippet;
	}

	let { children }: Props = $props();
</script>

<AppLayout>{@render children()}</AppLayout>
