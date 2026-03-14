<script lang="ts">
	import type { PresenceUser } from '$shared/types/models';

	interface Props {
		users: PresenceUser[];
		/** If set, highlight users viewing this agent */
		agentId?: string | null;
		max?: number;
		class?: string;
	}

	let { users, agentId = null, max = 5, class: className = '' }: Props = $props();

	const visible = $derived(users.slice(0, max));
	const overflow = $derived(users.length - max);

	function initials(login: string): string {
		return login.slice(0, 2).toUpperCase();
	}

	function tooltip(u: PresenceUser): string {
		if (agentId && u.viewing_agent_id === agentId) return `${u.login} (viewing)`;
		return u.login;
	}
</script>

<div class="flex items-center gap-1 {className}">
	{#each visible as user (user.user_id)}
		<div
			class="relative shrink-0"
			title={tooltip(user)}
		>
			{#if user.avatar_url}
				<img
					src={user.avatar_url}
					alt={user.login}
					class="h-6 w-6 rounded-full ring-2 {agentId && user.viewing_agent_id === agentId ? 'ring-blue-500' : 'ring-zinc-800'}"
				/>
			{:else}
				<div
					class="flex h-6 w-6 items-center justify-center rounded-full bg-zinc-700 text-[9px] font-semibold text-zinc-300 ring-2 {agentId && user.viewing_agent_id === agentId ? 'ring-blue-500' : 'ring-zinc-800'}"
				>
					{initials(user.login)}
				</div>
			{/if}
			<!-- Online dot -->
			<span class="absolute -right-0.5 -top-0.5 h-2 w-2 rounded-full bg-green-500 ring-1 ring-zinc-900"></span>
		</div>
	{/each}

	{#if overflow > 0}
		<div
			class="flex h-6 w-6 items-center justify-center rounded-full bg-zinc-700 text-[9px] font-semibold text-zinc-400 ring-2 ring-zinc-800"
			title="{overflow} more"
		>
			+{overflow}
		</div>
	{/if}
</div>
