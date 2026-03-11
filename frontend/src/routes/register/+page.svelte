<script lang="ts">
	import { authStore } from '$lib/stores/auth.svelte';
	import { goto } from '$app/navigation';

	let username = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let localError = $state('');

	async function handleSubmit(e: Event) {
		e.preventDefault();
		localError = '';

		if (password !== confirmPassword) {
			localError = 'Passwords do not match';
			return;
		}

		if (password.length < 6) {
			localError = 'Password must be at least 6 characters';
			return;
		}

		const success = await authStore.register(username, password);
		if (success) {
			goto('/tasks');
		}
	}
</script>

<div class="w-full max-w-sm mx-auto px-6">
	<div class="text-center mb-10">
		<h1 class="text-3xl font-mono font-bold tracking-wider text-accent">TARS</h1>
		<p class="text-sm text-text-secondary mt-2">Create Account</p>
	</div>

	<form onsubmit={handleSubmit} class="space-y-5">
		{#if authStore.error || localError}
			<div class="px-4 py-3 bg-danger/10 border border-danger/30 rounded text-sm text-danger">
				{localError || authStore.error}
			</div>
		{/if}

		<div>
			<label for="username" class="block text-sm text-text-secondary mb-1.5">Username</label>
			<input
				id="username"
				type="text"
				bind:value={username}
				required
				autocomplete="username"
				class="w-full bg-bg-tertiary border border-border rounded px-4 py-2.5 text-sm text-text-primary
					placeholder:text-text-secondary focus:outline-none focus:border-accent transition-colors"
				placeholder="Choose username"
			/>
		</div>

		<div>
			<label for="password" class="block text-sm text-text-secondary mb-1.5">Password</label>
			<input
				id="password"
				type="password"
				bind:value={password}
				required
				autocomplete="new-password"
				class="w-full bg-bg-tertiary border border-border rounded px-4 py-2.5 text-sm text-text-primary
					placeholder:text-text-secondary focus:outline-none focus:border-accent transition-colors"
				placeholder="Choose password"
			/>
		</div>

		<div>
			<label for="confirm" class="block text-sm text-text-secondary mb-1.5">Confirm Password</label>
			<input
				id="confirm"
				type="password"
				bind:value={confirmPassword}
				required
				autocomplete="new-password"
				class="w-full bg-bg-tertiary border border-border rounded px-4 py-2.5 text-sm text-text-primary
					placeholder:text-text-secondary focus:outline-none focus:border-accent transition-colors"
				placeholder="Confirm password"
			/>
		</div>

		<button
			type="submit"
			disabled={authStore.loading}
			class="w-full py-2.5 bg-accent text-bg-primary font-medium text-sm rounded
				hover:bg-accent-hover disabled:opacity-50 transition-colors"
		>
			{authStore.loading ? 'Creating account...' : 'Create Account'}
		</button>
	</form>

	<p class="mt-6 text-center text-sm text-text-secondary">
		Already have an account?
		<a href="/login" class="text-accent hover:text-accent-hover transition-colors">Sign In</a>
	</p>
</div>
