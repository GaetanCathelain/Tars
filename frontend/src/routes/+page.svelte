<script lang="ts">
	import { authStore } from '$lib/stores/auth.svelte';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	onMount(() => {
		if (authStore.isAuthenticated) {
			goto('/tasks');
		} else {
			goto('/login');
		}
	});
</script>

<div class="flex items-center justify-center h-full">
	<p class="text-text-secondary font-mono text-sm">Loading...</p>
</div>
