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

<div class="w-[380px] mx-auto px-4">
	<div class="bg-[#18181b] border border-zinc-800/50 rounded-xl p-8 shadow-lg surface-gradient">
		<div class="text-center mb-8">
			<div class="text-3xl mb-3">🤖</div>
			<h1 class="text-xl font-semibold text-zinc-100">TARS</h1>
			<p class="text-sm text-zinc-400 mt-1.5">Create your account</p>
		</div>

		<form onsubmit={handleSubmit} class="space-y-5">
			{#if authStore.error || localError}
				<div class="px-3.5 py-2.5 bg-danger/10 border border-danger/20 rounded-lg text-sm text-danger">
					{localError || authStore.error}
				</div>
			{/if}

			<div>
				<label for="username" class="block text-sm text-zinc-400 mb-2">Username</label>
				<input
					id="username"
					type="text"
					bind:value={username}
					required
					autocomplete="username"
					class="w-full h-10 bg-[#1c1c20] border border-zinc-800 rounded-lg px-3 text-sm text-text-primary
						placeholder:text-text-tertiary focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500/20 transition-colors duration-150"
					placeholder="Choose username"
				/>
			</div>

			<div>
				<label for="password" class="block text-sm text-zinc-400 mb-2">Password</label>
				<input
					id="password"
					type="password"
					bind:value={password}
					required
					autocomplete="new-password"
					class="w-full h-10 bg-[#1c1c20] border border-zinc-800 rounded-lg px-3 text-sm text-text-primary
						placeholder:text-text-tertiary focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500/20 transition-colors duration-150"
					placeholder="Choose password"
				/>
			</div>

			<div>
				<label for="confirm" class="block text-sm text-zinc-400 mb-2">Confirm Password</label>
				<input
					id="confirm"
					type="password"
					bind:value={confirmPassword}
					required
					autocomplete="new-password"
					class="w-full h-10 bg-[#1c1c20] border border-zinc-800 rounded-lg px-3 text-sm text-text-primary
						placeholder:text-text-tertiary focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500/20 transition-colors duration-150"
					placeholder="Confirm password"
				/>
			</div>

			<button
				type="submit"
				disabled={authStore.loading}
				class="w-full h-10 bg-indigo-500 text-white font-medium text-sm rounded-lg shadow-sm
					hover:bg-indigo-600 disabled:opacity-50 transition-colors duration-150"
			>
				{authStore.loading ? 'Creating account...' : 'Create Account'}
			</button>
		</form>

		<p class="mt-6 text-center text-sm text-zinc-400">
			Already have an account?
			<a href="/login" class="text-indigo-400 hover:text-indigo-300 transition-colors duration-150">Sign In</a>
		</p>
	</div>
</div>
