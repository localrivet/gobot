<script lang="ts">
	import type { Snippet } from 'svelte';
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { AppNav } from '$lib/components/navigation';
	import { auth, isAuthenticated, subscription, sessionExpiry, showSessionWarning, sessionSecondsRemaining } from '$lib/stores';
	import { getWebSocketClient } from '$lib/websocket/client';
	import SessionExpiryModal from '$lib/components/ui/SessionExpiryModal.svelte';

	let { children }: { children: Snippet } = $props();

	const authenticated = $derived($isAuthenticated);

	onMount(async () => {
		if (!authenticated) {
			goto('/auth/login');
			return;
		}

		getWebSocketClient().connect();

		// Start session expiry monitoring
		sessionExpiry.startMonitoring();

		await Promise.all([
			auth.fetchCurrentUser(),
			subscription.loadSubscription(),
			subscription.loadUsage()
		]);
	});

	onDestroy(() => {
		// Stop session monitoring when leaving authenticated area
		sessionExpiry.stopMonitoring();
	});

	function handleContinueSession() {
		sessionExpiry.continueSession();
	}

	function handleLogout() {
		sessionExpiry.dismiss();
		auth.logout();
		goto('/auth/login');
	}
</script>

<div class="layout-app">
	<AppNav />
	<main id="main-content" class="flex-1 p-6">
		<div class="max-w-[1400px] mx-auto">
			{@render children()}
		</div>
	</main>
</div>

<!-- Session Expiry Warning Modal -->
<SessionExpiryModal
	bind:show={$showSessionWarning}
	secondsRemaining={$sessionSecondsRemaining}
	onContinue={handleContinueSession}
	onLogout={handleLogout}
/>
