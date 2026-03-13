<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { post } from '$lib/api';
	import { Plus } from '@lucide/svelte';

	interface Props {
		onCreated: () => void;
	}

	let { onCreated }: Props = $props();

	let open = $state(false);
	let loading = $state(false);
	let error = $state('');
	let name = $state('');
	let url = $state('');
	let localPath = $state('');
	let defaultBranch = $state('main');

	function reset() {
		name = '';
		url = '';
		localPath = '';
		defaultBranch = 'main';
		error = '';
		loading = false;
	}

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		if (!name || !url) return;

		loading = true;
		error = '';

		try {
			await post('/api/v1/repos', {
				name,
				url,
				local_path: localPath || undefined,
				default_branch: defaultBranch || 'main'
			});
			open = false;
			reset();
			onCreated();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to add repository';
		} finally {
			loading = false;
		}
	}
</script>

<Dialog.Root bind:open onOpenChange={(v) => { if (!v) reset(); }}>
	<Dialog.Trigger>
		{#snippet child({ props })}
			<Button {...props} class="gap-2">
				<Plus class="size-4" />
				Add Repository
			</Button>
		{/snippet}
	</Dialog.Trigger>
	<Dialog.Content class="sm:max-w-md">
		<Dialog.Header>
			<Dialog.Title>Add Repository</Dialog.Title>
			<Dialog.Description>
				Add a new repository to track with TARS.
			</Dialog.Description>
		</Dialog.Header>

		<form onsubmit={handleSubmit} class="flex flex-col gap-4">
			{#if error}
				<p class="text-sm text-destructive">{error}</p>
			{/if}

			<div class="flex flex-col gap-2">
				<Label for="name">Name</Label>
				<Input id="name" bind:value={name} placeholder="my-project" required />
			</div>

			<div class="flex flex-col gap-2">
				<Label for="url">URL</Label>
				<Input id="url" bind:value={url} placeholder="https://github.com/user/repo" required />
			</div>

			<div class="flex flex-col gap-2">
				<Label for="localPath">Local Path <span class="text-muted-foreground">(optional)</span></Label>
				<Input id="localPath" bind:value={localPath} placeholder="/home/user/projects/repo" />
			</div>

			<div class="flex flex-col gap-2">
				<Label for="branch">Default Branch</Label>
				<Input id="branch" bind:value={defaultBranch} placeholder="main" />
			</div>

			<div class="flex justify-end gap-2">
				<Button variant="outline" type="button" onclick={() => { open = false; }}>
					Cancel
				</Button>
				<Button type="submit" disabled={loading || !name || !url}>
					{loading ? 'Adding...' : 'Add Repository'}
				</Button>
			</div>
		</form>
	</Dialog.Content>
</Dialog.Root>
