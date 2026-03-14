<script lang="ts">
	import type { RepoDiff, FileDiff } from '$shared/types/models';
	import { ChevronDown, ChevronRight, FilePlus, FileMinus, FileEdit, File } from 'lucide-svelte';
	import { Badge } from '$lib/components/ui/badge';

	interface Props {
		diff: RepoDiff;
		class?: string;
	}

	let { diff, class: className = '' }: Props = $props();

	// Track which files are expanded — initialise reactively so it updates if diff prop changes
	let expanded = $state<Record<string, boolean>>({});
	$effect(() => {
		expanded = Object.fromEntries(diff.files.map((f) => [f.path, true]));
	});

	function toggle(path: string) {
		expanded[path] = !expanded[path];
	}

	interface DiffLine {
		type: 'added' | 'removed' | 'context' | 'hunk';
		content: string;
		oldLine: number | null;
		newLine: number | null;
	}

	function parsePatch(patch: string): DiffLine[] {
		const lines: DiffLine[] = [];
		let oldLine = 0;
		let newLine = 0;

		for (const raw of patch.split('\n')) {
			if (raw.startsWith('@@')) {
				// Parse hunk header: @@ -old,count +new,count @@
				const m = raw.match(/@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@/);
				if (m) {
					oldLine = parseInt(m[1] ?? '0', 10);
					newLine = parseInt(m[2] ?? '0', 10);
				}
				lines.push({ type: 'hunk', content: raw, oldLine: null, newLine: null });
			} else if (raw.startsWith('+')) {
				lines.push({ type: 'added', content: raw.slice(1), oldLine: null, newLine: newLine++ });
			} else if (raw.startsWith('-')) {
				lines.push({ type: 'removed', content: raw.slice(1), oldLine: oldLine++, newLine: null });
			} else if (raw.startsWith(' ') || raw === '') {
				lines.push({ type: 'context', content: raw.slice(1), oldLine: oldLine++, newLine: newLine++ });
			}
		}
		return lines;
	}

	function fileStatusIcon(status: FileDiff['status']) {
		switch (status) {
			case 'added': return FilePlus;
			case 'deleted': return FileMinus;
			case 'modified': return FileEdit;
			default: return File;
		}
	}

	function fileStatusColor(status: FileDiff['status']): string {
		switch (status) {
			case 'added': return 'text-green-400';
			case 'deleted': return 'text-red-400';
			case 'modified': return 'text-blue-400';
			default: return 'text-zinc-400';
		}
	}

	function lineClass(type: DiffLine['type']): string {
		switch (type) {
			case 'added': return 'bg-green-950/40 text-green-300';
			case 'removed': return 'bg-red-950/40 text-red-300';
			case 'hunk': return 'bg-blue-950/30 text-blue-400 font-mono text-xs';
			default: return 'text-zinc-400';
		}
	}

	function lineNumClass(type: DiffLine['type']): string {
		switch (type) {
			case 'added': return 'bg-green-950/60 text-green-700 select-none';
			case 'removed': return 'bg-red-950/60 text-red-700 select-none';
			case 'hunk': return 'bg-blue-950/50 select-none';
			default: return 'bg-zinc-900 text-zinc-700 select-none';
		}
	}
</script>

<div class="flex flex-col gap-1 {className}">
	<!-- Stats bar -->
	<div class="flex items-center gap-4 rounded-lg border border-zinc-800 bg-zinc-900/60 px-4 py-2.5 text-sm">
		<span class="text-zinc-400">{diff.base_ref} → <span class="font-mono text-zinc-200">{diff.head_ref}</span></span>
		<div class="ml-auto flex items-center gap-3 text-xs">
			<span class="text-zinc-400">{diff.stats.files_changed} file{diff.stats.files_changed !== 1 ? 's' : ''}</span>
			<span class="text-green-400">+{diff.stats.insertions}</span>
			<span class="text-red-400">-{diff.stats.deletions}</span>
		</div>
	</div>

	<!-- File list -->
	{#each diff.files as file (file.path)}
		{@const lines = parsePatch(file.patch)}
		{@const isExpanded = expanded[file.path] ?? true}
		{@const StatusIcon = fileStatusIcon(file.status)}

		<div class="overflow-hidden rounded-lg border border-zinc-800">
			<!-- File header -->
			<button
				onclick={() => toggle(file.path)}
				class="flex w-full items-center gap-2 border-b border-zinc-800 bg-zinc-900 px-3 py-2 text-left transition-colors hover:bg-zinc-800/60"
			>
				{#if isExpanded}
					<ChevronDown class="h-3.5 w-3.5 shrink-0 text-zinc-500" />
				{:else}
					<ChevronRight class="h-3.5 w-3.5 shrink-0 text-zinc-500" />
				{/if}
				<StatusIcon class="h-3.5 w-3.5 shrink-0 {fileStatusColor(file.status)}" />
				<span class="min-w-0 flex-1 truncate font-mono text-xs text-zinc-200">{file.path}</span>
				<div class="ml-auto flex shrink-0 items-center gap-2">
					{#if file.additions > 0}
						<span class="text-xs text-green-400">+{file.additions}</span>
					{/if}
					{#if file.deletions > 0}
						<span class="text-xs text-red-400">-{file.deletions}</span>
					{/if}
					<Badge variant="outline" class="h-4 px-1 text-[10px] {fileStatusColor(file.status)} border-current">
						{file.status}
					</Badge>
				</div>
			</button>

			<!-- Diff lines -->
			{#if isExpanded && file.patch}
				<div class="overflow-x-auto bg-zinc-950">
					<table class="w-full font-mono text-xs">
						<tbody>
							{#each lines as line, i (i)}
								{#if line.type === 'hunk'}
									<tr>
										<td colspan="3" class="px-4 py-0.5 {lineClass(line.type)}">{line.content}</td>
									</tr>
								{:else}
									<tr class="group {lineClass(line.type)}">
										<td class="w-10 px-2 py-0 text-right text-[10px] {lineNumClass(line.type)}">
											{line.oldLine ?? ''}
										</td>
										<td class="w-10 px-2 py-0 text-right text-[10px] {lineNumClass(line.type)}">
											{line.newLine ?? ''}
										</td>
										<td class="whitespace-pre px-3 py-0 leading-5">{line.type === 'added' ? '+' : line.type === 'removed' ? '-' : ' '}{line.content}</td>
									</tr>
								{/if}
							{/each}
						</tbody>
					</table>
				</div>
			{:else if isExpanded && !file.patch}
				<div class="px-4 py-3 text-xs text-zinc-600 bg-zinc-950">No patch data available.</div>
			{/if}
		</div>
	{/each}

	{#if diff.files.length === 0}
		<div class="flex items-center justify-center rounded-lg border border-zinc-800 py-8">
			<p class="text-sm text-zinc-600">No changes from {diff.base_ref}</p>
		</div>
	{/if}
</div>
