<script lang="ts">
	import '../app.css';
	import type { LayoutData } from './$types';
	import { auth } from '$lib/stores/auth.svelte';
	import { ws } from '$lib/stores/ws.svelte';
	import { page } from '$app/state';
	import AppSidebar from '$lib/components/app-sidebar.svelte';
	import { Button } from '$lib/components/ui/button';
	import * as Avatar from '$lib/components/ui/avatar';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import { LogOut } from '@lucide/svelte';

	let { data, children }: { data: LayoutData; children: any } = $props();

	const isPublicPage = $derived(
		page.url.pathname.startsWith('/login') || page.url.pathname.startsWith('/auth')
	);

	$effect(() => {
		if (data.user) {
			auth.setUser(data.user);
			ws.connect();
		}
		return () => {
			ws.disconnect();
		};
	});
</script>

{#if isPublicPage}
	{@render children()}
{:else}
	<div class="flex h-screen">
		<AppSidebar />

		<div class="flex flex-1 flex-col overflow-hidden">
			<!-- Top bar -->
			<header class="flex h-14 items-center justify-end border-b border-border px-4">
				{#if auth.user}
					<DropdownMenu.Root>
						<DropdownMenu.Trigger>
							<button class="flex items-center gap-2 rounded-md px-2 py-1 hover:bg-accent transition-colors">
								<Avatar.Root class="size-7">
									<Avatar.Image src={auth.user.avatar_url} alt={auth.user.username} />
									<Avatar.Fallback>{auth.user.username.slice(0, 2).toUpperCase()}</Avatar.Fallback>
								</Avatar.Root>
								<span class="text-sm text-foreground">{auth.user.username}</span>
							</button>
						</DropdownMenu.Trigger>
						<DropdownMenu.Content align="end">
							<DropdownMenu.Item onclick={() => auth.logout()}>
								<LogOut class="mr-2 size-4" />
								Sign out
							</DropdownMenu.Item>
						</DropdownMenu.Content>
					</DropdownMenu.Root>
				{/if}
			</header>

			<!-- Main content -->
			<main class="flex-1 overflow-auto p-6">
				{@render children()}
			</main>
		</div>
	</div>
{/if}
