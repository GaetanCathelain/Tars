<script lang="ts">
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import type { PageData, ActionData } from './$types';
	import type { Repo } from '$shared/types/models';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Card, CardHeader, CardTitle, CardContent } from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { GitFork, Plus, Trash2, Settings, X, Check, ExternalLink } from 'lucide-svelte';

	interface Props {
		data: PageData;
		form: ActionData;
	}

	let { data, form }: Props = $props();

	let showAddForm = $state(false);
	let editingRepo = $state<Repo | null>(null);
	let deletingRepoId = $state<string | null>(null);
	let submitting = $state(false);

	// Reset form state on successful action
	$effect(() => {
		if (form?.success) {
			showAddForm = false;
			editingRepo = null;
			deletingRepoId = null;
		}
	});
</script>

<div class="p-6">
	<!-- Header -->
	<div class="mb-6 flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight text-zinc-50">Repositories</h1>
			<p class="mt-1 text-sm text-zinc-400">
				{data.repos.length} {data.repos.length === 1 ? 'repository' : 'repositories'}
			</p>
		</div>
		<Button
			onclick={() => {
				showAddForm = !showAddForm;
				editingRepo = null;
			}}
			size="sm"
		>
			{#if showAddForm}
				<X class="h-4 w-4" />
				Cancel
			{:else}
				<Plus class="h-4 w-4" />
				Add Repository
			{/if}
		</Button>
	</div>

	<!-- Error banner -->
	{#if form?.error}
		<div class="mb-4 rounded-md border border-red-800 bg-red-950/50 px-4 py-3 text-sm text-red-400">
			{form.error}
		</div>
	{/if}

	<!-- Add Repository Form -->
	{#if showAddForm}
		<Card class="mb-6">
			<CardHeader>
				<CardTitle>Add Repository</CardTitle>
			</CardHeader>
			<CardContent>
				<form
					method="POST"
					action="?/create"
					use:enhance={() => {
						submitting = true;
						return async ({ update }) => {
							await update();
							submitting = false;
						};
					}}
					class="space-y-4"
				>
					<div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
						<div class="space-y-1.5">
							<label for="name" class="text-xs font-medium text-zinc-400">Name</label>
							<Input id="name" name="name" placeholder="my-project" required />
						</div>
						<div class="space-y-1.5">
							<label for="github_url" class="text-xs font-medium text-zinc-400">GitHub URL</label>
							<Input
								id="github_url"
								name="github_url"
								type="url"
								placeholder="https://github.com/org/repo"
								required
							/>
						</div>
					</div>
					<div class="space-y-1.5">
						<label for="path" class="text-xs font-medium text-zinc-400">Server Path</label>
						<Input id="path" name="path" placeholder="/workspaces/my-project" required />
					</div>
					<div class="flex justify-end gap-2">
						<Button
							type="button"
							variant="ghost"
							size="sm"
							onclick={() => (showAddForm = false)}
						>
							Cancel
						</Button>
						<Button type="submit" size="sm" disabled={submitting}>
							{submitting ? 'Adding...' : 'Add Repository'}
						</Button>
					</div>
				</form>
			</CardContent>
		</Card>
	{/if}

	<!-- Repository List -->
	{#if data.repos.length === 0}
		<div class="flex flex-col items-center justify-center rounded-xl border border-dashed border-zinc-800 py-16 text-center">
			<GitFork class="mb-3 h-10 w-10 text-zinc-700" />
			<p class="text-sm font-medium text-zinc-400">No repositories yet</p>
			<p class="mt-1 text-xs text-zinc-600">Add a repository to get started with agents.</p>
			<Button
				class="mt-4"
				size="sm"
				onclick={() => (showAddForm = true)}
			>
				<Plus class="h-4 w-4" />
				Add Repository
			</Button>
		</div>
	{:else}
		<div class="space-y-3">
			{#each data.repos as repo (repo.id)}
				<div class="rounded-xl border border-zinc-800 bg-zinc-900 p-4">
					{#if editingRepo?.id === repo.id}
						<!-- Edit form -->
						<form
							method="POST"
							action="?/update"
							use:enhance={() => {
								submitting = true;
								return async ({ update }) => {
									await update();
									submitting = false;
								};
							}}
							class="space-y-3"
						>
							<input type="hidden" name="repoId" value={repo.id} />
							<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
								<div class="space-y-1">
									<label for="edit-name-{repo.id}" class="text-xs font-medium text-zinc-400">Name</label>
									<Input
										id="edit-name-{repo.id}"
										name="name"
										value={repo.name}
									/>
								</div>
								<div class="space-y-1">
									<label for="edit-branch-{repo.id}" class="text-xs font-medium text-zinc-400">Default Branch</label>
									<Input
										id="edit-branch-{repo.id}"
										name="default_branch"
										value={repo.default_branch}
									/>
								</div>
							</div>
							<div class="flex justify-end gap-2">
								<Button
									type="button"
									variant="ghost"
									size="sm"
									onclick={() => (editingRepo = null)}
								>
									Cancel
								</Button>
								<Button type="submit" size="sm" disabled={submitting}>
									<Check class="h-4 w-4" />
									{submitting ? 'Saving...' : 'Save'}
								</Button>
							</div>
						</form>
					{:else}
						<!-- View row -->
						<div class="flex items-start justify-between gap-4">
							<div class="min-w-0 flex-1">
								<div class="flex items-center gap-2">
									<GitFork class="h-4 w-4 shrink-0 text-zinc-500" />
									<span class="font-medium text-zinc-50">{repo.name}</span>
									<Badge variant="outline" class="text-xs">{repo.default_branch}</Badge>
								</div>
								<div class="mt-1 flex items-center gap-2">
									<a
										href={repo.github_url}
										target="_blank"
										rel="noopener noreferrer"
										class="flex items-center gap-1 text-xs text-zinc-500 transition-colors hover:text-zinc-300"
									>
										<ExternalLink class="h-3 w-3" />
										{repo.github_url.replace('https://github.com/', '')}
									</a>
								</div>
								<p class="mt-0.5 text-xs text-zinc-600">{repo.path}</p>
							</div>
							<div class="flex shrink-0 items-center gap-1">
								<Button
									variant="ghost"
									size="icon"
									onclick={() => (editingRepo = repo)}
									class="h-7 w-7"
								>
									<Settings class="h-3.5 w-3.5" />
								</Button>

								{#if deletingRepoId === repo.id}
									<form
										method="POST"
										action="?/delete"
										use:enhance={() => {
											return async ({ update }) => {
												deletingRepoId = null;
												await update();
											};
										}}
									>
										<input type="hidden" name="repoId" value={repo.id} />
										<Button type="submit" variant="destructive" size="sm" class="h-7 text-xs">
											Confirm delete
										</Button>
									</form>
									<Button
										variant="ghost"
										size="icon"
										onclick={() => (deletingRepoId = null)}
										class="h-7 w-7"
									>
										<X class="h-3.5 w-3.5" />
									</Button>
								{:else}
									<Button
										variant="ghost"
										size="icon"
										onclick={() => (deletingRepoId = repo.id)}
										class="h-7 w-7 text-zinc-600 hover:text-red-400"
									>
										<Trash2 class="h-3.5 w-3.5" />
									</Button>
								{/if}
							</div>
						</div>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
</div>
