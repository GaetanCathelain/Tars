<script lang="ts">
	import { goto } from '$app/navigation';
	import { auth } from '$lib/stores/auth.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Card from '$lib/components/ui/card';

	let username = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let error = $state('');
	let loading = $state(false);

	async function handleRegister() {
		if (!username || !password || !confirmPassword) {
			error = 'Please fill in all fields';
			return;
		}
		if (password !== confirmPassword) {
			error = 'Passwords do not match';
			return;
		}
		loading = true;
		error = '';
		try {
			await auth.register(username, password);
			goto('/tasks');
		} catch (e) {
			error = e instanceof Error ? e.message : 'Registration failed';
		} finally {
			loading = false;
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') handleRegister();
	}
</script>

<div class="min-h-screen flex items-center justify-center bg-background">
	<Card.Root class="w-[400px]">
		<Card.Header class="text-center">
			<div class="mx-auto mb-2 w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center">
				<span class="text-lg">🤖</span>
			</div>
			<Card.Title class="text-xl">TARS</Card.Title>
			<Card.Description>Create an account</Card.Description>
		</Card.Header>
		<Card.Content class="space-y-4">
			{#if error}
				<div class="text-sm text-destructive text-center">{error}</div>
			{/if}
			<div class="space-y-2">
				<Label for="username">Username</Label>
				<Input
					id="username"
					placeholder="Choose a username"
					bind:value={username}
					onkeydown={handleKeydown}
				/>
			</div>
			<div class="space-y-2">
				<Label for="password">Password</Label>
				<Input
					id="password"
					type="password"
					placeholder="Choose a password"
					bind:value={password}
					onkeydown={handleKeydown}
				/>
			</div>
			<div class="space-y-2">
				<Label for="confirm-password">Confirm Password</Label>
				<Input
					id="confirm-password"
					type="password"
					placeholder="Confirm password"
					bind:value={confirmPassword}
					onkeydown={handleKeydown}
				/>
			</div>
			<Button class="w-full" onclick={handleRegister} disabled={loading}>
				{loading ? 'Creating account...' : 'Register'}
			</Button>
		</Card.Content>
		<Card.Footer class="justify-center">
			<p class="text-sm text-muted-foreground">
				Already have an account? <a href="/login" class="text-primary hover:underline">Sign in</a>
			</p>
		</Card.Footer>
	</Card.Root>
</div>
