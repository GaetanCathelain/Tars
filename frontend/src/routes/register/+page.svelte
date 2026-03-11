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
	<div class="bg-bg-secondary border border-border rounded-xl p-8 shadow-[0_1px_3px_rgba(0,0,0,0.4)]">
		<div class="text-center mb-8">
			<h1 class="text-base font-semibold tracking-wide text-zinc-200">TARS</h1>
			<p class="text-[12px] text-text-tertiary mt-1">Create your account</p>
		</div>

		<form onsubmit={handleSubmit} class="space-y-4">
			{#if authStore.error || localError}
				<div class="px-3 py-2 bg-danger/10 border border-danger/20 rounded-md text-[13px] text-danger">
					{localError || authStore.error}
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
					placeholder="Choose username"
				/>
			</div>

			<div>
				<label for="password" class="block text-[12px] font-medium text-text-secondary mb-1">Password</label>
				<input
					id="password"
					type="password"
					bind:value={password}
					required
					autocomplete="new-password"
					class="w-full bg-bg-primary border border-border rounded-md px-3 py-2 text-[13px] text-text-primary
						placeholder:text-text-tertiary focus:outline-none focus:border-accent transition-all duration-150"
					placeholder="Choose password"
				/>
			</div>

			<div>
				<label for="confirm" class="block text-[12px] font-medium text-text-secondary mb-1">Confirm Password</label>
				<input
					id="confirm"
					type="password"
					bind:value={confirmPassword}
					required
					autocomplete="new-password"
					class="w-full bg-bg-primary border border-border rounded-md px-3 py-2 text-[13px] text-text-primary
						placeholder:text-text-tertiary focus:outline-none focus:border-accent transition-all duration-150"
					placeholder="Confirm password"
				/>
			</div>

			<button
				type="submit"
				disabled={authStore.loading}
				class="w-full py-2 bg-accent text-white font-medium text-[13px] rounded-md
					hover:bg-accent-hover disabled:opacity-50 transition-all duration-150"
			>
				{authStore.loading ? 'Creating account...' : 'Create Account'}
			</button>
		</form>

		<p class="mt-5 text-center text-[12px] text-text-tertiary">
			Already have an account?
			<a href="/login" class="text-accent hover:text-accent-hover transition-all duration-150">Sign In</a>
		</p>
	</div>
</div>
