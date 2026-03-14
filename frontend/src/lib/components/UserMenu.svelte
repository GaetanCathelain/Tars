<script lang="ts">
	import { Avatar } from '$lib/components/ui/avatar';
	import { auth } from '$lib/stores/auth.svelte';
</script>

{#if auth.user}
	<div class="flex items-center gap-3">
		<Avatar
			src={auth.user.avatar_url}
			alt={auth.user.name || auth.user.login}
			class="h-7 w-7"
		/>
		<span class="text-sm text-zinc-300">{auth.user.login}</span>
		<form method="POST" action="/auth/logout">
			<button
				type="submit"
				class="text-xs text-zinc-500 transition-colors hover:text-zinc-300"
			>
				Sign out
			</button>
		</form>
	</div>
{/if}
