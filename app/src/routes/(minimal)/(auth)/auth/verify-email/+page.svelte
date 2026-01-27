<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { Card, Alert, Spinner } from '$lib/components/ui';
	import * as api from '$lib/api';

	let status = $state<'loading' | 'success' | 'error'>('loading');
	let errorMessage = $state('');

	const token = $derived($page.url.searchParams.get('token'));

	onMount(async () => {
		if (!token) {
			status = 'error';
			errorMessage = 'No verification token provided';
			return;
		}

		try {
			await api.verifyEmail({ token });
			status = 'success';
			setTimeout(() => {
				goto('/app');
			}, 3000);
		} catch (error) {
			status = 'error';
			errorMessage = error instanceof Error ? error.message : 'Verification failed';
		}
	});
</script>

<svelte:head>
	<title>Verify Email - Gobot</title>
</svelte:head>

<Card class="w-full text-center">
	<a href="/" class="text-2xl font-bold text-white font-display"> Gobot </a>

	{#if status === 'loading'}
		<div class="py-8">
			<Spinner size={32} class="mx-auto mb-4" />
			<p class="text-base-content/60">Verifying your email...</p>
		</div>
	{:else if status === 'success'}
		<div class="py-8">
			<Alert type="success" title="Success"
				>Email verified successfully! Redirecting to your dashboard...</Alert
			>
		</div>
	{:else}
		<div class="py-8">
			<Alert type="error" title="Verification Failed">
				{errorMessage}
			</Alert>
			<p class="text-base-content/40 text-sm mt-4">
				<a href="/auth/login" class="text-primary hover:text-primary-light"> Return to login </a>
			</p>
		</div>
	{/if}
</Card>
