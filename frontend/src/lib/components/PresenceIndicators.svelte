<script lang="ts">
	import { presence } from '$lib/stores/presence.svelte';
	import type { PresenceUser } from '$shared/types/models';

	interface Props {
		repoId: string;
		class?: string;
	}

	let { repoId, class: className = '' }: Props = $props();

	const users: PresenceUser[] = $derived(presence.getUsersForRepo(repoId));

	function initials(login: string): string {
		return login.slice(0, 2).toUpperCase();
	}
</script>

{#if users.length > 0}
	<div class="flex items-center gap-1 {className}">
		<span class="mr-1 text-xs text-zinc-600">{users.length} online</span>
		<div class="flex -space-x-1.5">
			{#each users.slice(0, 5) as user (user.user_id)}
				<div
					title="{user.login}{user.viewing_agent_id ? ' (viewing agent)' : ''}"
					class="relative flex h-6 w-6 shrink-0 items-center justify-center rounded-full border-2 border-zinc-900 bg-zinc-700 text-[9px] font-semibold text-zinc-200 ring-0"
				>
					{#if user.avatar_url}
						<img
							src={user.avatar_url}
							alt={user.login}
							class="h-full w-full rounded-full object-cover"
						/>
					{:else}
						{initials(user.login)}
					{/if}
					{#if user.viewing_agent_id}
						<span
							class="absolute -bottom-0.5 -right-0.5 h-2 w-2 rounded-full border border-zinc-900 bg-green-400"
							title="Viewing agent"
						></span>
					{/if}
				</div>
			{/each}
			{#if users.length > 5}
				<div
					class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full border-2 border-zinc-900 bg-zinc-800 text-[9px] font-semibold text-zinc-400"
				>
					+{users.length - 5}
				</div>
			{/if}
		</div>
	</div>
{/if}
