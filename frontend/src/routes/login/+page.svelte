<script lang="ts">
	import { authStore } from '$lib/stores/auth.svelte';
	import { goto } from '$app/navigation';

	let username = $state('');
	let password = $state('');

	async function handleSubmit(e: Event) {
		e.preventDefault();
		const success = await authStore.login(username, password);
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
			<p class="text-sm text-zinc-400 mt-1.5">Sign in to continue</p>
		</div>

		<form onsubmit={handleSubmit} class="space-y-5">
			{#if authStore.error}
				<div class="px-3.5 py-2.5 bg-danger/10 border border-danger/20 rounded-lg text-sm text-danger">
					{authStore.error}
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
					placeholder="Enter username"
				/>
			</div>

			<div>
				<label for="password" class="block text-sm text-zinc-400 mb-2">Password</label>
				<input
					id="password"
					type="password"
					bind:value={password}
					required
					autocomplete="current-password"
					class="w-full h-10 bg-[#1c1c20] border border-zinc-800 rounded-lg px-3 text-sm text-text-primary
						placeholder:text-text-tertiary focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500/20 transition-colors duration-150"
					placeholder="Enter password"
				/>
			</div>

			<button
				type="submit"
				disabled={authStore.loading}
				class="w-full h-10 bg-indigo-500 text-white font-medium text-sm rounded-lg shadow-sm
					hover:bg-indigo-600 disabled:opacity-50 transition-colors duration-150"
			>
				{authStore.loading ? 'Signing in...' : 'Sign In'}
			</button>
		</form>

		<p class="mt-6 text-center text-sm text-zinc-400">
			No account?
			<a href="/register" class="text-indigo-400 hover:text-indigo-300 transition-colors duration-150">Register</a>
		</p>
	</div>
</div>
