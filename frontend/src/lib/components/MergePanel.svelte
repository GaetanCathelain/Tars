<script lang="ts">
	import type { Agent, RepoDiff, MergeResult } from '$shared/types/models';
	import { enhance } from '$app/forms';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import DiffViewer from './DiffViewer.svelte';
	import { GitMerge, AlertCircle, CheckCircle2, Loader2 } from 'lucide-svelte';

	interface Props {
		agent: Agent;
		repoId: string;
		diff: RepoDiff | null;
		class?: string;
	}

	let { agent, repoId, diff, class: className = '' }: Props = $props();

	let targetBranch = $state('main');
	$effect(() => {
		targetBranch = agent.branch.replace('tars/', '').replace(/^agent-[^/]+$/, '') || 'main';
	});
	let strategy = $state<'squash' | 'merge' | 'rebase'>('squash');
	let commitMessage = $state('');
	let merging = $state(false);
	let mergeError = $state<string | null>(null);
	let mergeResult = $state<MergeResult | null>(null);

	// Default commit message
	$effect(() => {
		if (!commitMessage) {
			commitMessage = `feat: ${agent.name} changes\n\nCo-authored-by: TARS Agent ${agent.name}`;
		}
	});

	const canMerge = $derived(
		(agent.status === 'stopped' || agent.status === 'crashed') &&
		diff !== null &&
		(diff.stats.insertions > 0 || diff.stats.deletions > 0)
	);
</script>

<div class="flex h-full flex-col gap-4 overflow-y-auto p-4 {className}">
	<!-- Diff section -->
	{#if diff}
		<DiffViewer {diff} />
	{:else}
		<div class="flex items-center justify-center rounded-lg border border-zinc-800 py-12">
			<p class="text-sm text-zinc-600">No diff available. Agent may still be running.</p>
		</div>
	{/if}

	<!-- Merge form -->
	<div class="rounded-lg border border-zinc-800 bg-zinc-900/60 p-4">
		<div class="mb-4 flex items-center gap-2">
			<GitMerge class="h-4 w-4 text-zinc-400" />
			<h3 class="text-sm font-semibold text-zinc-200">Merge Agent Branch</h3>
		</div>

		{#if mergeResult}
			<div class="flex items-start gap-2 rounded-md border border-green-900 bg-green-950/50 p-3 text-sm text-green-400">
				<CheckCircle2 class="mt-0.5 h-4 w-4 shrink-0" />
				<div>
					<p class="font-medium">Merged successfully!</p>
					<p class="mt-0.5 text-xs text-green-600">
						{mergeResult.agent_branch} → {mergeResult.target_branch}
						{#if mergeResult.commit_sha}
							<span class="font-mono"> ({mergeResult.commit_sha.slice(0, 8)})</span>
						{/if}
					</p>
				</div>
			</div>
		{:else}
			{#if mergeError}
				<div class="mb-3 flex items-start gap-2 rounded-md border border-red-900 bg-red-950/50 p-3 text-xs text-red-400">
					<AlertCircle class="mt-0.5 h-3.5 w-3.5 shrink-0" />
					<span>{mergeError}</span>
					<button onclick={() => (mergeError = null)} class="ml-auto text-red-700 hover:text-red-400">✕</button>
				</div>
			{/if}

			{#if !canMerge && agent.status !== 'stopped' && agent.status !== 'crashed'}
				<p class="mb-3 text-xs text-zinc-500">Stop the agent before merging.</p>
			{/if}

			<form
				method="POST"
				action="?/merge"
				use:enhance={() => {
					merging = true;
					mergeError = null;
					return async ({ result, update }) => {
						merging = false;
						if (result.type === 'success' && result.data?.mergeResult) {
							mergeResult = result.data.mergeResult as MergeResult;
						} else if (result.type === 'failure') {
							mergeError = (result.data?.error as string) ?? 'Merge failed.';
						}
						await update({ reset: false });
					};
				}}
				class="flex flex-col gap-3"
			>
				<input type="hidden" name="repoId" value={repoId} />

				<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
					<div>
						<label class="mb-1 block text-xs text-zinc-500" for="targetBranch">Target branch</label>
						<Input
							id="targetBranch"
							name="targetBranch"
							bind:value={targetBranch}
							class="h-8 font-mono text-sm"
							placeholder="main"
						/>
					</div>
					<div>
						<label class="mb-1 block text-xs text-zinc-500" for="strategy">Strategy</label>
						<select
							id="strategy"
							name="strategy"
							bind:value={strategy}
							class="h-8 w-full rounded-md border border-zinc-700 bg-zinc-800 px-3 text-sm text-zinc-200 focus:outline-none focus:ring-1 focus:ring-zinc-500"
						>
							<option value="squash">Squash</option>
							<option value="merge">Merge</option>
							<option value="rebase">Rebase</option>
						</select>
					</div>
				</div>

				{#if strategy === 'squash'}
					<div>
						<label class="mb-1 block text-xs text-zinc-500" for="commitMessage">Commit message</label>
						<textarea
							id="commitMessage"
							name="commitMessage"
							bind:value={commitMessage}
							rows={3}
							class="w-full rounded-md border border-zinc-700 bg-zinc-800 px-3 py-2 font-mono text-xs text-zinc-200 focus:outline-none focus:ring-1 focus:ring-zinc-500"
						></textarea>
					</div>
				{/if}

				<Button
					type="submit"
					disabled={!canMerge || merging}
					class="w-full gap-2"
				>
					{#if merging}
						<Loader2 class="h-4 w-4 animate-spin" />
						Merging…
					{:else}
						<GitMerge class="h-4 w-4" />
						Merge into {targetBranch}
					{/if}
				</Button>
			</form>
		{/if}
	</div>
</div>
