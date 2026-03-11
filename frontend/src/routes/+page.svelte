<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/stores/auth.svelte';

	onMount(() => {
		auth.init();
		if (auth.isAuthenticated) {
			goto('/tasks');
		} else {
			goto('/login');
		}
	});
</script>

<div class="min-h-screen flex items-center justify-center bg-background">
	<p class="text-muted-foreground text-sm">Loading...</p>
</div>
