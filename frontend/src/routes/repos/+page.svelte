<script lang="ts">
	import type { PageData } from './$types';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import * as Dialog from '$lib/components/ui/dialog';
	import AddRepoDialog from '$lib/components/add-repo-dialog.svelte';
	import { del } from '$lib/api';
	import { invalidateAll } from '$app/navigation';
	import { Trash2, Database } from '@lucide/svelte';
	import type { Repo } from './+page.server';

	let { data }: { data: PageData } = $props();

	let deleteTarget = $state<Repo | null>(null);
	let deleteOpen = $state(false);
	let deleting = $state(false);

	async function handleDelete() {
		if (!deleteTarget) return;
		deleting = true;
		try {
			await del(`/api/v1/repos/${deleteTarget.id}`);
			deleteOpen = false;
			deleteTarget = null;
			await invalidateAll();
		} catch {
			// TODO: show error toast
		} finally {
			deleting = false;
		}
	}

	function formatDate(dateStr: string): string {
		try {
			return new Date(dateStr).toLocaleDateString('en-US', {
				year: 'numeric',
				month: 'short',
				day: 'numeric'
			});
		} catch {
			return dateStr;
		}
	}
</script>

<div class="flex flex-col gap-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight">Repositories</h1>
			<p class="text-sm text-muted-foreground">Manage repositories tracked by TARS.</p>
		</div>
		<AddRepoDialog onCreated={() => invalidateAll()} />
	</div>

	{#if data.repos.length === 0}
		<div class="flex flex-col items-center justify-center gap-4 rounded-lg border border-dashed border-border py-16">
			<Database class="size-12 text-muted-foreground" />
			<div class="text-center">
				<p class="text-lg font-medium">No repositories yet</p>
				<p class="text-sm text-muted-foreground">Add one to get started.</p>
			</div>
		</div>
	{:else}
		<div class="rounded-md border border-border">
			<Table.Root>
				<Table.Header>
					<Table.Row>
						<Table.Head>Name</Table.Head>
						<Table.Head>URL</Table.Head>
						<Table.Head>Branch</Table.Head>
						<Table.Head>Added</Table.Head>
						<Table.Head class="w-12"></Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each data.repos as repo}
						<Table.Row>
							<Table.Cell class="font-medium">{repo.name}</Table.Cell>
							<Table.Cell class="text-muted-foreground">{repo.url}</Table.Cell>
							<Table.Cell>{repo.default_branch}</Table.Cell>
							<Table.Cell class="text-muted-foreground">{formatDate(repo.created_at)}</Table.Cell>
							<Table.Cell>
								<Button
									variant="ghost"
									size="icon-sm"
									onclick={() => { deleteTarget = repo; deleteOpen = true; }}
								>
									<Trash2 class="size-4 text-muted-foreground" />
								</Button>
							</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</div>
	{/if}
</div>

<!-- Delete confirmation dialog -->
<Dialog.Root bind:open={deleteOpen}>
	<Dialog.Content class="sm:max-w-sm">
		<Dialog.Header>
			<Dialog.Title>Delete Repository</Dialog.Title>
			<Dialog.Description>
				Are you sure you want to remove <strong>{deleteTarget?.name}</strong>? This action cannot be undone.
			</Dialog.Description>
		</Dialog.Header>
		<div class="flex justify-end gap-2">
			<Button variant="outline" onclick={() => { deleteOpen = false; }}>Cancel</Button>
			<Button variant="destructive" onclick={handleDelete} disabled={deleting}>
				{deleting ? 'Deleting...' : 'Delete'}
			</Button>
		</div>
	</Dialog.Content>
</Dialog.Root>
