<script lang="ts">
	import { goto } from '$app/navigation';
	import { createAdmin, setupStatus } from '$lib/api';
	import { auth } from '$lib/stores/auth';
	import { onMount } from 'svelte';

	let email = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let name = $state('');
	let loading = $state(false);
	let error = $state('');
	let checkingStatus = $state(true);

	onMount(async () => {
		try {
			const status = await setupStatus();
			if (!status.setupRequired) {
				goto('/auth/login');
			}
		} catch (e) {
			console.error('Failed to check setup status', e);
		} finally {
			checkingStatus = false;
		}
	});

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = '';

		if (password !== confirmPassword) {
			error = 'Passwords do not match';
			return;
		}

		if (password.length < 8) {
			error = 'Password must be at least 8 characters';
			return;
		}

		loading = true;

		try {
			const response = await createAdmin({ email, password, name });
			await auth.setOAuthTokens(response.token, response.refreshToken, response.expiresAt);
			goto('/app');
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Failed to create admin account';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Setup - Gobot</title>
	<meta name="description" content="Set up your Gobot instance by creating the first admin account." />
</svelte:head>

{#if checkingStatus}
	<div class="flex items-center justify-center p-8">
		<span class="loading loading-spinner loading-lg"></span>
	</div>
{:else}
	<div class="card bg-base-100 border border-base-300">
		<div class="card-body">
			<h2 class="card-title text-2xl mb-2">Welcome to Gobot</h2>
			<p class="text-base-content/70 mb-6">
				Create the first admin account to get started.
			</p>

			{#if error}
				<div class="alert alert-error mb-4">
					<span>{error}</span>
				</div>
			{/if}

			<form onsubmit={handleSubmit}>
				<div class="form-control mb-4">
					<label class="label" for="name">
						<span class="label-text">Full Name</span>
					</label>
					<input
						type="text"
						id="name"
						bind:value={name}
						class="input input-bordered"
						placeholder="Admin User"
						required
					/>
				</div>

				<div class="form-control mb-4">
					<label class="label" for="email">
						<span class="label-text">Email</span>
					</label>
					<input
						type="email"
						id="email"
						bind:value={email}
						class="input input-bordered"
						placeholder="admin@example.com"
						required
					/>
				</div>

				<div class="form-control mb-4">
					<label class="label" for="password">
						<span class="label-text">Password</span>
					</label>
					<input
						type="password"
						id="password"
						bind:value={password}
						class="input input-bordered"
						placeholder="Min 8 characters"
						minlength="8"
						required
					/>
				</div>

				<div class="form-control mb-6">
					<label class="label" for="confirmPassword">
						<span class="label-text">Confirm Password</span>
					</label>
					<input
						type="password"
						id="confirmPassword"
						bind:value={confirmPassword}
						class="input input-bordered"
						placeholder="Repeat password"
						minlength="8"
						required
					/>
				</div>

				<button
					type="submit"
					class="btn btn-primary w-full"
					disabled={loading}
				>
					{#if loading}
						<span class="loading loading-spinner loading-sm"></span>
						Creating Account...
					{:else}
						Create Admin Account
					{/if}
				</button>
			</form>
		</div>
	</div>
{/if}
