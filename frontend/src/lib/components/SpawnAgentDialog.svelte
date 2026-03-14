<script lang="ts">
	import type { AgentPersona } from '$shared/types/models';
	import type { Task } from '$shared/types/models';
	import { enhance } from '$app/forms';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Badge } from '$lib/components/ui/badge';
	import { X, Bot, Loader2, CheckCircle2 } from 'lucide-svelte';

	interface Props {
		repoId: string;
		tasks?: Task[];
		open?: boolean;
		onclose?: () => void;
		onsuccess?: () => void;
	}

	let { repoId, tasks = [], open = $bindable(false), onclose, onsuccess }: Props = $props();

	interface PersonaDef {
		id: AgentPersona;
		label: string;
		description: string;
		suggestedPrompt: string;
		color: string;
		badge: string;
	}

	const PERSONAS: PersonaDef[] = [
		{
			id: 'backend',
			label: 'Backend Architect',
			description: 'Server-side systems, APIs, databases, and infrastructure. Go/Python/Rust expert.',
			suggestedPrompt: 'You are a senior backend engineer. Focus on correctness, performance, and security.',
			color: 'border-blue-900 hover:border-blue-700 bg-blue-950/20',
			badge: 'text-blue-400 border-blue-800'
		},
		{
			id: 'frontend',
			label: 'Frontend Developer',
			description: 'UI/UX, React/Svelte, accessibility, and performance optimization.',
			suggestedPrompt: 'You are a senior frontend engineer. Focus on user experience, accessibility, and clean component design.',
			color: 'border-purple-900 hover:border-purple-700 bg-purple-950/20',
			badge: 'text-purple-400 border-purple-800'
		},
		{
			id: 'devops',
			label: 'DevOps Engineer',
			description: 'CI/CD, containers, cloud infrastructure, and build tooling.',
			suggestedPrompt: 'You are a senior DevOps engineer. Focus on reliability, automation, and reproducible builds.',
			color: 'border-orange-900 hover:border-orange-700 bg-orange-950/20',
			badge: 'text-orange-400 border-orange-800'
		},
		{
			id: 'qa',
			label: 'QA Engineer',
			description: 'Testing strategy, coverage, automation, and quality assurance.',
			suggestedPrompt: 'You are a senior QA engineer. Focus on comprehensive test coverage and catching edge cases.',
			color: 'border-green-900 hover:border-green-700 bg-green-950/20',
			badge: 'text-green-400 border-green-800'
		},
		{
			id: 'general',
			label: 'General Agent',
			description: 'Versatile assistant for any task — analysis, planning, writing, or coding.',
			suggestedPrompt: 'You are a capable software engineering assistant. Adapt to whatever the task requires.',
			color: 'border-zinc-700 hover:border-zinc-500 bg-zinc-900',
			badge: 'text-zinc-400 border-zinc-700'
		}
	];

	let selectedPersona = $state<AgentPersona>('general');
	let agentName = $state('');
	let selectedTaskId = $state('');
	let model = $state('claude-opus-4-6');
	let systemPrompt = $state('');
	let spawning = $state(false);
	let spawnError = $state<string | null>(null);

	// Auto-fill system prompt when persona changes
	$effect(() => {
		const persona = PERSONAS.find((p) => p.id === selectedPersona);
		if (persona) systemPrompt = persona.suggestedPrompt;
	});

	// Auto-generate name from persona + timestamp
	$effect(() => {
		if (!agentName) {
			agentName = `${selectedPersona}-${Date.now().toString(36).slice(-4)}`;
		}
	});

	function close() {
		open = false;
		onclose?.();
	}

	function selectPersona(id: AgentPersona) {
		selectedPersona = id;
		// Reset name so auto-gen fires
		agentName = `${id}-${Date.now().toString(36).slice(-4)}`;
	}

	const selectedPersonaDef = $derived(PERSONAS.find((p) => p.id === selectedPersona));

	const MODELS = [
		{ id: 'claude-opus-4-6', label: 'Claude Opus 4.6' },
		{ id: 'claude-sonnet-4-6', label: 'Claude Sonnet 4.6' },
		{ id: 'claude-haiku-4-5-20251001', label: 'Claude Haiku 4.5' }
	];
</script>

{#if open}
	<!-- Backdrop -->
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-40 bg-black/60 backdrop-blur-sm"
		onclick={close}
	></div>

	<!-- Dialog -->
	<div class="fixed inset-0 z-50 flex items-end justify-center p-0 sm:items-center sm:p-4">
		<div class="flex w-full max-w-2xl flex-col rounded-t-xl border border-zinc-800 bg-zinc-950 shadow-2xl sm:max-h-[90vh] sm:rounded-xl">
			<!-- Header -->
			<div class="flex items-center justify-between border-b border-zinc-800 px-6 py-4">
				<div class="flex items-center gap-2">
					<Bot class="h-5 w-5 text-zinc-400" />
					<h2 class="text-base font-semibold text-zinc-50">Spawn Agent</h2>
				</div>
				<button onclick={close} class="rounded-md p-1 text-zinc-500 hover:text-zinc-200">
					<X class="h-4 w-4" />
				</button>
			</div>

			<!-- Body -->
			<div class="max-h-[70vh] overflow-y-auto p-4 sm:max-h-none sm:p-6">
				{#if spawnError}
					<div class="mb-4 flex items-center gap-2 rounded-md border border-red-900 bg-red-950/50 px-3 py-2 text-xs text-red-400">
						{spawnError}
						<button onclick={() => (spawnError = null)} class="ml-auto text-red-700 hover:text-red-400">✕</button>
					</div>
				{/if}

				<!-- Persona cards -->
				<p class="mb-3 text-xs font-medium uppercase tracking-wider text-zinc-500">Persona</p>
				<div class="mb-5 grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">
					{#each PERSONAS as persona (persona.id)}
						<button
							type="button"
							onclick={() => selectPersona(persona.id)}
							class="flex flex-col items-start gap-1.5 rounded-lg border p-3 text-left transition-all {persona.color} {selectedPersona === persona.id ? 'ring-2 ring-zinc-300' : ''}"
						>
							<div class="flex w-full items-center justify-between">
								<Badge variant="outline" class="text-[10px] {persona.badge}">{persona.label}</Badge>
								{#if selectedPersona === persona.id}
									<CheckCircle2 class="h-3.5 w-3.5 text-zinc-300" />
								{/if}
							</div>
							<p class="text-xs text-zinc-500">{persona.description}</p>
						</button>
					{/each}
				</div>

				<!-- Form fields -->
				<form
					method="POST"
					action="?/spawn"
					use:enhance={() => {
						spawning = true;
						spawnError = null;
						return async ({ result, update }) => {
							spawning = false;
							if (result.type === 'success') {
								close();
								onsuccess?.();
							} else if (result.type === 'failure') {
								spawnError = (result.data?.error as string) ?? 'Failed to spawn agent.';
							}
							await update({ reset: false });
						};
					}}
					class="flex flex-col gap-4"
				>
					<input type="hidden" name="repoId" value={repoId} />
					<input type="hidden" name="persona" value={selectedPersona} />

					<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
						<div>
							<label class="mb-1 block text-xs text-zinc-500" for="agentName">Agent name</label>
							<Input
								id="agentName"
								name="name"
								bind:value={agentName}
								placeholder="my-agent"
								class="h-8 font-mono text-sm"
								required
							/>
						</div>
						<div>
							<label class="mb-1 block text-xs text-zinc-500" for="model">Model</label>
							<select
								id="model"
								name="model"
								bind:value={model}
								class="h-8 w-full rounded-md border border-zinc-700 bg-zinc-800 px-3 text-sm text-zinc-200 focus:outline-none focus:ring-1 focus:ring-zinc-500"
							>
								{#each MODELS as m (m.id)}
									<option value={m.id}>{m.label}</option>
								{/each}
							</select>
						</div>
					</div>

					{#if tasks.length > 0}
						<div>
							<label class="mb-1 block text-xs text-zinc-500" for="taskId">Assign to task (optional)</label>
							<select
								id="taskId"
								name="task_id"
								bind:value={selectedTaskId}
								class="h-8 w-full rounded-md border border-zinc-700 bg-zinc-800 px-3 text-sm text-zinc-200 focus:outline-none focus:ring-1 focus:ring-zinc-500"
							>
								<option value="">No task</option>
								{#each tasks.filter((t) => t.status === 'pending' || t.status === 'in_progress') as task (task.id)}
									<option value={task.id}>[{task.status}] {task.title}</option>
								{/each}
							</select>
						</div>
					{/if}

					<div>
						<label class="mb-1 block text-xs text-zinc-500" for="systemPrompt">
							System prompt
							{#if selectedPersonaDef}
								<span class="text-zinc-600">— {selectedPersonaDef.label} default</span>
							{/if}
						</label>
						<textarea
							id="systemPrompt"
							name="system_prompt"
							bind:value={systemPrompt}
							rows={3}
							class="w-full rounded-md border border-zinc-700 bg-zinc-800 px-3 py-2 text-xs text-zinc-200 focus:outline-none focus:ring-1 focus:ring-zinc-500"
						></textarea>
					</div>

					<div class="flex justify-end gap-2 pt-1">
						<Button type="button" variant="ghost" size="sm" onclick={close}>Cancel</Button>
						<Button type="submit" size="sm" disabled={!agentName.trim() || spawning} class="gap-2">
							{#if spawning}
								<Loader2 class="h-3.5 w-3.5 animate-spin" />
								Spawning…
							{:else}
								<Bot class="h-3.5 w-3.5" />
								Spawn {selectedPersonaDef?.label ?? 'Agent'}
							{/if}
						</Button>
					</div>
				</form>
			</div>
		</div>
	</div>
{/if}
