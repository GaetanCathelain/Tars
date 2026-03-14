<script lang="ts">
	import { page } from '$app/stores';
	import { Separator } from '$lib/components/ui/separator';
	import UserMenu from '$lib/components/UserMenu.svelte';
	import {
		LayoutDashboard,
		GitFork,
		CheckSquare,
		Bot,
		ChevronRight,
		X
	} from 'lucide-svelte';
	import { cn } from '$lib/utils/cn';

	interface Props {
		mobileOpen?: boolean;
		onclose?: () => void;
	}

	let { mobileOpen = false, onclose }: Props = $props();

	const navItems = [
		{ href: '/dashboard', label: 'Overview', icon: LayoutDashboard },
		{ href: '/repos', label: 'Repositories', icon: GitFork },
		{ href: '/tasks', label: 'Tasks', icon: CheckSquare },
		{ href: '/agents', label: 'Agents', icon: Bot }
	];

	const currentPath = $derived($page.url.pathname);

	function isActive(href: string): boolean {
		if (href === '/dashboard') return currentPath === '/dashboard';
		return currentPath.startsWith(href);
	}

	function handleNavClick() {
		onclose?.();
	}
</script>

<!-- Mobile backdrop -->
{#if mobileOpen}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div
		class="fixed inset-0 z-30 bg-black/60 backdrop-blur-sm md:hidden"
		onclick={onclose}
	></div>
{/if}

<aside
	class={cn(
		'flex w-56 shrink-0 flex-col border-r border-zinc-800 bg-zinc-900 transition-transform duration-200',
		// Desktop: always visible
		'md:relative md:translate-x-0',
		// Mobile: slide in/out as overlay
		'fixed inset-y-0 left-0 z-40 md:static',
		mobileOpen ? 'translate-x-0' : '-translate-x-full md:translate-x-0'
	)}
>
	<!-- Logo + mobile close -->
	<div class="flex h-14 items-center gap-2 px-4">
		<div class="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-zinc-50">
			<span class="text-xs font-black text-zinc-900">T</span>
		</div>
		<span class="flex-1 font-semibold tracking-tight text-zinc-50">TARS</span>
		<button
			onclick={onclose}
			class="rounded-md p-1 text-zinc-500 hover:text-zinc-200 md:hidden"
			aria-label="Close sidebar"
		>
			<X class="h-4 w-4" />
		</button>
	</div>

	<Separator />

	<!-- Navigation -->
	<nav class="flex-1 space-y-1 p-2 pt-3" aria-label="Main navigation">
		{#each navItems as item (item.href)}
			{@const active = isActive(item.href)}
			<a
				href={item.href}
				onclick={handleNavClick}
				class={cn(
					'group flex items-center gap-2.5 rounded-md px-3 py-2 text-sm font-medium transition-colors',
					active
						? 'bg-zinc-800 text-zinc-50'
						: 'text-zinc-400 hover:bg-zinc-800/50 hover:text-zinc-200'
				)}
				aria-current={active ? 'page' : undefined}
			>
				<item.icon
					class={cn(
						'h-4 w-4 shrink-0',
						active ? 'text-zinc-50' : 'text-zinc-500 group-hover:text-zinc-300'
					)}
				/>
				{item.label}
				{#if active}
					<ChevronRight class="ml-auto h-3 w-3 text-zinc-500" />
				{/if}
			</a>
		{/each}
	</nav>

	<Separator />

	<!-- User -->
	<div class="p-3">
		<UserMenu />
	</div>
</aside>
