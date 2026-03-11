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

<div class="w-full max-w-sm mx-auto px-6">
	<div class="bg-bg-secondary border border-border rounded-xl p-8 shadow-[0_1px_3px_rgba(0,0,0,0.4)]">
		<div class="text-center mb-8">
			<h1 class="text-base font-semibold tracking-wide text-zinc-200">TARS</h1>
			<p class="text-[12px] text-text-tertiary mt-1">Sign in to continue</p>
		</div>

		<form onsubmit={handleSubmit} class="space-y-4">
			{#if authStore.error}
				<div class="px-3 py-2 bg-danger/10 border border-danger/20 rounded-md text-[13px] text-danger">
					{authStore.error}
				</div>
			{/if}

			<div>
				<label for="username" class="block text-[12px] font-medium text-text-secondary mb-1">Username</label>
				<input
					id="username"
					type="text"
					bind:value={username}
					required
					autocomplete="username"
					class="w-full bg-bg-primary border border-border rounded-md px-3 py-2 text-[13px] text-text-primary
						placeholder:text-text-tertiary focus:outline-none focus:border-accent transition-all duration-150"
					placeholder="Enter username"
				/>
			</div>

			<div>
				<label for="password" class="block text-[12px] font-medium text-text-secondary mb-1">Password</label>
				<input
					id="password"
					type="password"
					bind:value={password}
					required
					autocomplete="current-password"
					class="w-full bg-bg-primary border border-border rounded-md px-3 py-2 text-[13px] text-text-primary
						placeholder:text-text-tertiary focus:outline-none focus:border-accent transition-all duration-150"
					placeholder="Enter password"
				/>
			</div>

			<button
				type="submit"
				disabled={authStore.loading}
				class="w-full py-2 bg-accent text-white font-medium text-[13px] rounded-md
					hover:bg-accent-hover disabled:opacity-50 transition-all duration-150"
			>
				{authStore.loading ? 'Signing in...' : 'Sign In'}
			</button>
		</form>

		<p class="mt-5 text-center text-[12px] text-text-tertiary">
			No account?
			<a href="/register" class="text-accent hover:text-accent-hover transition-all duration-150">Register</a>
		</p>
	</div>
</div>
