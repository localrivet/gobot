<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { RegistrationForm } from '$lib/components/auth';

	// Get plan from URL query param (e.g., /auth/register?plan=pro)
	const plan = $derived($page.url.searchParams.get('plan') || 'free');

	function handleSuccess(checkoutUrl?: string) {
		if (checkoutUrl) {
			// Redirect to Stripe checkout for paid plans
			window.location.href = checkoutUrl;
		} else {
			// Free plan - go directly to app
			goto('/app');
		}
	}

	function handleLoginClick() {
		goto('/auth/login');
	}
</script>

<svelte:head>
	<title>Create Account - Gobot</title>
	<meta name="description" content="Create a Gobot account to get started." />
</svelte:head>

<RegistrationForm {plan} onSuccess={handleSuccess} onLoginClick={handleLoginClick} />
