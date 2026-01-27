<script lang="ts">
	import { goto } from '$app/navigation';
	import { LoginForm } from '$lib/components/auth';

	function handleSuccess() {
		goto('/app');
	}

	function handleRegisterClick() {
		goto('/auth/register');
	}

	function handleForgotPasswordClick() {
		goto('/auth/reset-password');
	}
</script>

<svelte:head>
	<title>Sign In - Gobot</title>
	<meta
		name="description"
		content="Sign in to your Gobot account to access brand analysis tools and insights."
	/>
</svelte:head>

<LoginForm
	onSuccess={handleSuccess}
	onRegisterClick={handleRegisterClick}
	onForgotPasswordClick={handleForgotPasswordClick}
/>
