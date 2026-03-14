<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { api, ApiError } from '$lib/utils/api';
	import Button from '$lib/components/ui/button/button.svelte';
	import Input from '$lib/components/ui/input/input.svelte';
	import Separator from '$lib/components/ui/separator/separator.svelte';

	const apiBase: string = (import.meta.env['PUBLIC_API_URL'] as string) ?? 'http://localhost:8090';
	const githubLoginUrl = `${apiBase}/api/v1/auth/github/login`;

	const oauthError = $derived($page.url.searchParams.get('error'));

	// 'signin' | 'register'
	let mode = $state<'signin' | 'register'>('signin');

	// Sign-in fields
	let signinEmail = $state('');
	let signinPassword = $state('');

	// Register fields
	let registerName = $state('');
	let registerEmail = $state('');
	let registerPassword = $state('');
	let registerConfirm = $state('');

	let loading = $state(false);
	let errorMsg = $state('');

	function clearError() {
		errorMsg = '';
	}

	function switchMode(next: 'signin' | 'register') {
		mode = next;
		clearError();
	}

	async function handleSignIn(e: SubmitEvent) {
		e.preventDefault();
		clearError();

		if (!signinEmail || !signinPassword) {
			errorMsg = 'Email and password are required.';
			return;
		}

		loading = true;
		try {
			await api.post('/auth/login', { email: signinEmail, password: signinPassword });
			goto('/dashboard');
		} catch (err) {
			if (err instanceof ApiError) {
				if (err.status === 401) {
					errorMsg = 'Invalid email or password.';
				} else {
					errorMsg = err.message;
				}
			} else {
				errorMsg = 'An unexpected error occurred.';
			}
		} finally {
			loading = false;
		}
	}

	async function handleRegister(e: SubmitEvent) {
		e.preventDefault();
		clearError();

		if (!registerName || !registerEmail || !registerPassword || !registerConfirm) {
			errorMsg = 'All fields are required.';
			return;
		}

		if (registerPassword.length < 8) {
			errorMsg = 'Password must be at least 8 characters.';
			return;
		}

		if (registerPassword !== registerConfirm) {
			errorMsg = 'Passwords do not match.';
			return;
		}

		const emailRe = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
		if (!emailRe.test(registerEmail)) {
			errorMsg = 'Please enter a valid email address.';
			return;
		}

		loading = true;
		try {
			await api.post('/auth/register', {
				name: registerName,
				email: registerEmail,
				password: registerPassword
			});
			goto('/dashboard');
		} catch (err) {
			if (err instanceof ApiError) {
				if (err.status === 409) {
					errorMsg = 'An account with that email already exists.';
				} else if (err.status === 400) {
					errorMsg = err.message;
				} else {
					errorMsg = err.message;
				}
			} else {
				errorMsg = 'An unexpected error occurred.';
			}
		} finally {
			loading = false;
		}
	}
</script>

<div class="flex min-h-screen items-center justify-center bg-zinc-950">
	<div class="w-full max-w-sm space-y-6 rounded-xl border border-zinc-800 bg-zinc-900 p-8 shadow-xl">
		<div class="space-y-2 text-center">
			<h1 class="text-2xl font-bold tracking-tight text-zinc-50">TARS</h1>
			<p class="text-sm text-zinc-400">Multiplayer agent conductor</p>
		</div>

		{#if oauthError}
			<div class="rounded-md border border-red-800 bg-red-950/50 px-4 py-3 text-sm text-red-400">
				Authentication failed: {oauthError}
			</div>
		{/if}

		{#if errorMsg}
			<div class="rounded-md border border-red-800 bg-red-950/50 px-4 py-3 text-sm text-red-400">
				{errorMsg}
			</div>
		{/if}

		<!-- Tab switcher -->
		<div class="flex rounded-lg border border-zinc-800 bg-zinc-950 p-1">
			<button
				type="button"
				onclick={() => switchMode('signin')}
				class="flex-1 rounded-md py-1.5 text-sm font-medium transition-colors {mode === 'signin'
					? 'bg-zinc-800 text-zinc-50'
					: 'text-zinc-500 hover:text-zinc-300'}"
			>
				Sign In
			</button>
			<button
				type="button"
				onclick={() => switchMode('register')}
				class="flex-1 rounded-md py-1.5 text-sm font-medium transition-colors {mode === 'register'
					? 'bg-zinc-800 text-zinc-50'
					: 'text-zinc-500 hover:text-zinc-300'}"
			>
				Create Account
			</button>
		</div>

		{#if mode === 'signin'}
			<form onsubmit={handleSignIn} class="space-y-4">
				<div class="space-y-2">
					<label for="signin-email" class="text-sm font-medium text-zinc-300">Email</label>
					<Input
						id="signin-email"
						type="email"
						placeholder="you@example.com"
						bind:value={signinEmail}
						required
						disabled={loading}
					/>
				</div>
				<div class="space-y-2">
					<label for="signin-password" class="text-sm font-medium text-zinc-300">Password</label>
					<Input
						id="signin-password"
						type="password"
						placeholder="••••••••"
						bind:value={signinPassword}
						required
						disabled={loading}
					/>
				</div>
				<Button type="submit" class="w-full" disabled={loading}>
					{loading ? 'Signing in…' : 'Sign In'}
				</Button>
			</form>
		{:else}
			<form onsubmit={handleRegister} class="space-y-4">
				<div class="space-y-2">
					<label for="reg-name" class="text-sm font-medium text-zinc-300">Display Name</label>
					<Input
						id="reg-name"
						type="text"
						placeholder="Your name"
						bind:value={registerName}
						required
						disabled={loading}
					/>
				</div>
				<div class="space-y-2">
					<label for="reg-email" class="text-sm font-medium text-zinc-300">Email</label>
					<Input
						id="reg-email"
						type="email"
						placeholder="you@example.com"
						bind:value={registerEmail}
						required
						disabled={loading}
					/>
				</div>
				<div class="space-y-2">
					<label for="reg-password" class="text-sm font-medium text-zinc-300">Password</label>
					<Input
						id="reg-password"
						type="password"
						placeholder="Min. 8 characters"
						bind:value={registerPassword}
						required
						disabled={loading}
					/>
				</div>
				<div class="space-y-2">
					<label for="reg-confirm" class="text-sm font-medium text-zinc-300">Confirm Password</label>
					<Input
						id="reg-confirm"
						type="password"
						placeholder="••••••••"
						bind:value={registerConfirm}
						required
						disabled={loading}
					/>
				</div>
				<Button type="submit" class="w-full" disabled={loading}>
					{loading ? 'Creating account…' : 'Create Account'}
				</Button>
			</form>
		{/if}

		<div class="flex items-center gap-3">
			<Separator />
			<span class="text-xs whitespace-nowrap text-zinc-500">Or continue with</span>
			<Separator />
		</div>

		<a
			href={githubLoginUrl}
			class="flex w-full items-center justify-center gap-3 rounded-md bg-zinc-50 px-4 py-2.5 text-sm font-semibold text-zinc-900 shadow-sm transition-colors hover:bg-zinc-200 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-zinc-50"
		>
			<svg viewBox="0 0 24 24" class="h-5 w-5" fill="currentColor" aria-hidden="true">
				<path
					d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0 1 12 6.844a9.59 9.59 0 0 1 2.504.337c1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.02 10.02 0 0 0 22 12.017C22 6.484 17.522 2 12 2Z"
				/>
			</svg>
			GitHub
		</a>

		<p class="text-center text-xs text-zinc-600">Access is restricted to authorized users.</p>
	</div>
</div>
