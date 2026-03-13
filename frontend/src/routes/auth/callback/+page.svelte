<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	onMount(() => {
		// Backend already set the session cookie during the callback redirect.
		// Just redirect to the app root.
		goto('/', { replaceState: true });
	});
</script>

<div class="flex min-h-screen items-center justify-center bg-background">
	<div class="flex flex-col items-center gap-3">
		<div class="size-8 animate-spin rounded-full border-2 border-muted-foreground border-t-transparent"></div>
		<p class="text-sm text-muted-foreground">Signing in...</p>
	</div>
</div>
