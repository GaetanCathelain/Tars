<script lang="ts">
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button';
	import { Separator } from '$lib/components/ui/separator';
	import * as Tooltip from '$lib/components/ui/tooltip';
	import { Home, Database, ListTodo } from '@lucide/svelte';

	interface NavItem {
		label: string;
		href: string;
		icon: typeof Home;
		disabled?: boolean;
		tooltip?: string;
	}

	const navItems: NavItem[] = [
		{ label: 'Dashboard', href: '/', icon: Home },
		{ label: 'Repos', href: '/repos', icon: Database },
		{ label: 'Tasks', href: '/tasks', icon: ListTodo, disabled: true, tooltip: 'Coming soon' }
	];

	function isActive(href: string): boolean {
		if (href === '/') return page.url.pathname === '/' || page.url.pathname === '/repos';
		return page.url.pathname.startsWith(href);
	}
</script>

<aside class="flex h-screen w-56 flex-col border-r border-sidebar-border bg-sidebar">
	<div class="flex h-14 items-center gap-2 px-4">
		<span class="text-lg font-bold tracking-tight text-sidebar-foreground">TARS</span>
		<span class="text-xs text-muted-foreground">v2</span>
	</div>

	<Separator />

	<nav class="flex flex-1 flex-col gap-1 p-2">
		{#each navItems as item}
			{#if item.disabled}
				<Tooltip.Root>
					<Tooltip.Trigger>
						<Button
							variant="ghost"
							class="w-full justify-start gap-2 text-muted-foreground opacity-50 cursor-not-allowed"
							disabled
						>
							<item.icon class="size-4" />
							{item.label}
						</Button>
					</Tooltip.Trigger>
					<Tooltip.Content side="right">
						<p>{item.tooltip}</p>
					</Tooltip.Content>
				</Tooltip.Root>
			{:else}
				<Button
					variant={isActive(item.href) ? 'secondary' : 'ghost'}
					class="w-full justify-start gap-2"
					href={item.href}
				>
					<item.icon class="size-4" />
					{item.label}
				</Button>
			{/if}
		{/each}
	</nav>

	<div class="p-2 text-xs text-muted-foreground text-center">
		Multiplayer Agent Conductor
	</div>
</aside>
