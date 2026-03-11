<script lang="ts">
	import { goto } from '$app/navigation';
	import { auth } from '$lib/stores/auth.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Card from '$lib/components/ui/card';

	let username = $state('');
	let password = $state('');
	let error = $state('');
	let loading = $state(false);

	async function handleLogin() {
		if (!username || !password) {
			error = 'Please fill in all fields';
			return;
		}
		loading = true;
		error = '';
		try {
			await auth.login(username, password);
			goto('/tasks');
		} catch (e) {
			error = e instanceof Error ? e.message : 'Login failed';
		} finally {
			loading = false;
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') handleLogin();
	}
</script>

<div class="min-h-screen flex items-center justify-center bg-background">
	<Card.Root class="w-[400px]">
		<Card.Header class="text-center">
			<div class="mx-auto mb-2 w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center">
				<span class="text-lg">🤖</span>
			</div>
			<Card.Title class="text-xl">TARS</Card.Title>
			<Card.Description>Sign in to continue</Card.Description>
		</Card.Header>
		<Card.Content class="space-y-4">
			{#if error}
				<div class="text-sm text-destructive text-center">{error}</div>
			{/if}
			<div class="space-y-2">
				<Label for="username">Username</Label>
				<Input
					id="username"
					placeholder="Enter username"
					bind:value={username}
					onkeydown={handleKeydown}
				/>
			</div>
			<div class="space-y-2">
				<Label for="password">Password</Label>
				<Input
					id="password"
					type="password"
					placeholder="Enter password"
					bind:value={password}
					onkeydown={handleKeydown}
				/>
			</div>
			<Button class="w-full" onclick={handleLogin} disabled={loading}>
				{loading ? 'Signing in...' : 'Sign In'}
			</Button>
		</Card.Content>
		<Card.Footer class="justify-center">
			<p class="text-sm text-muted-foreground">
				No account? <a href="/register" class="text-primary hover:underline">Register</a>
			</p>
		</Card.Footer>
	</Card.Root>
</div>
