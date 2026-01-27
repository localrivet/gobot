<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { auth } from '$lib/stores/auth';

	let status: 'loading' | 'success' | 'error' = 'loading';
	let errorMessage = '';
	let isNewUser = false;

	onMount(async () => {
		const params = $page.url.searchParams;

		// Check for error from OAuth provider
		const error = params.get('error');
		if (error) {
			status = 'error';
			errorMessage = decodeURIComponent(error);
			return;
		}

		// Get tokens from URL params
		const token = params.get('token');
		const refresh = params.get('refresh');
		const expires = params.get('expires');
		const newUser = params.get('new');

		if (!token || !refresh || !expires) {
			status = 'error';
			errorMessage = 'Missing authentication tokens';
			return;
		}

		isNewUser = newUser === 'true';

		try {
			const expiresAt = parseInt(expires, 10);
			const success = await auth.setOAuthTokens(token, refresh, expiresAt);

			if (success) {
				status = 'success';
				// Clear the URL params for security (don't leave tokens in browser history)
				window.history.replaceState({}, '', '/auth/callback');

				// Redirect after a brief moment
				setTimeout(() => {
					goto('/app');
				}, 1500);
			} else {
				status = 'error';
				errorMessage = 'Failed to complete authentication';
			}
		} catch (err) {
			status = 'error';
			errorMessage = err instanceof Error ? err.message : 'Authentication failed';
		}
	});

	function goToLogin() {
		goto('/auth/login');
	}
</script>

<svelte:head>
	<title>Signing in... - Gobot</title>
</svelte:head>

<div class="flex min-h-screen items-center justify-center">
	<div class="card bg-base-100 w-full max-w-md shadow-xl">
		<div class="card-body items-center text-center">
			{#if status === 'loading'}
				<span class="loading loading-spinner loading-lg text-primary"></span>
				<h2 class="card-title mt-4">Completing sign in...</h2>
				<p class="text-base-content/60">Please wait while we set up your session.</p>
			{:else if status === 'success'}
				<div class="text-success">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
				</div>
				<h2 class="card-title mt-4">
					{isNewUser ? 'Welcome!' : 'Welcome back!'}
				</h2>
				<p class="text-base-content/60">
					{isNewUser ? 'Your account has been created.' : 'You have been signed in.'}
					Redirecting...
				</p>
			{:else if status === 'error'}
				<div class="text-error">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
					</svg>
				</div>
				<h2 class="card-title mt-4">Authentication Failed</h2>
				<p class="text-base-content/60">{errorMessage}</p>
				<div class="card-actions mt-4">
					<button class="btn btn-primary" onclick={goToLogin}>
						Back to Login
					</button>
				</div>
			{/if}
		</div>
	</div>
</div>
