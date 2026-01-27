<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { PasswordResetForm } from '$lib/components/auth';

	const token = $derived($page.url.searchParams.get('token') || '');
	const mode = $derived(token ? 'reset' : 'request') as 'request' | 'reset';

	function handleSuccess() {
		if (mode === 'reset') {
			goto('/auth/login');
		}
	}

	function handleBackToLogin() {
		goto('/auth/login');
	}
</script>

<svelte:head>
	<title>Reset Password - Gobot</title>
	<meta name="description" content="Reset your Gobot account password." />
</svelte:head>

<PasswordResetForm {mode} {token} onSuccess={handleSuccess} onBackToLogin={handleBackToLogin} />
