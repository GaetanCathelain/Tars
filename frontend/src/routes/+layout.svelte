<script lang="ts">
	import '../app.css';
	import type { Snippet } from 'svelte';
	import { auth } from '$lib/stores/auth.svelte';
	import { wsClient } from '$lib/ws/client.svelte';
	import { startRouter, stopRouter } from '$lib/ws/router';
	import type { LayoutData } from './$types';

	interface Props {
		data: LayoutData;
		children: Snippet;
	}

	let { data, children }: Props = $props();

	// Sync server-loaded user into client auth store
	$effect(() => {
		auth.setUser(data.user);
	});

	// Connect WebSocket when user is authenticated
	$effect(() => {
		if (auth.user) {
			const wsUrl =
				(import.meta.env['PUBLIC_WS_URL'] as string | undefined) ??
				(typeof window !== 'undefined'
					? `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host.replace(/:\d+$/, ':8080')}/ws`
					: 'ws://localhost:8090/ws');

			startRouter();
			wsClient.connect(wsUrl);
		} else {
			wsClient.disconnect();
			stopRouter();
		}

		return () => {
			wsClient.disconnect();
			stopRouter();
		};
	});
</script>

{@render children()}
