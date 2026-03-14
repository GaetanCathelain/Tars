<script lang="ts">
	import type { Snippet } from 'svelte';
	import AppSidebar from './AppSidebar.svelte';
	import { Menu } from 'lucide-svelte';

	interface Props {
		children: Snippet;
	}

	let { children }: Props = $props();

	let sidebarOpen = $state(false);
</script>

<div class="flex h-screen overflow-hidden bg-zinc-950">
	<AppSidebar mobileOpen={sidebarOpen} onclose={() => (sidebarOpen = false)} />

	<main class="flex min-w-0 flex-1 flex-col overflow-hidden">
		<!-- Mobile top bar -->
		<div class="flex h-12 shrink-0 items-center gap-3 border-b border-zinc-800 bg-zinc-900 px-4 md:hidden">
			<button
				onclick={() => (sidebarOpen = true)}
				class="rounded-md p-1.5 text-zinc-400 hover:bg-zinc-800 hover:text-zinc-200"
				aria-label="Open navigation"
			>
				<Menu class="h-5 w-5" />
			</button>
			<div class="flex h-6 w-6 items-center justify-center rounded-md bg-zinc-50">
				<span class="text-xs font-black text-zinc-900">T</span>
			</div>
			<span class="font-semibold tracking-tight text-zinc-50">TARS</span>
		</div>

		<div class="flex-1 overflow-y-auto">
			{@render children()}
		</div>
	</main>
</div>
