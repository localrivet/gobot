<script lang="ts">
	import { goto } from '$app/navigation';
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import { Settings, CreditCard, User, Zap } from 'lucide-svelte';
	import { subscription } from '$lib/stores/subscription';
	import { auth } from '$lib/stores/auth';

	// Get subscription and user data from stores
	let currentSubscription = $derived($subscription);
	let currentUser = $derived($auth.user);
</script>

<svelte:head>
	<title>Dashboard - Gobot</title>
	<meta name="description" content="Your application dashboard." />
</svelte:head>

<div class="mb-8">
	<h1 class="font-display text-2xl font-bold text-base-content mb-1">Dashboard</h1>
	<p class="text-sm text-base-content/60">
		Welcome back{currentUser?.name ? `, ${currentUser.name}` : ''}!
	</p>
</div>

<div class="grid gap-6">
	<!-- Quick Stats Grid -->
	<div class="grid sm:grid-cols-2 lg:grid-cols-4 gap-6">
		<Card>
			<div class="flex items-center justify-between mb-4">
				<div class="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center">
					<User class="w-5 h-5 text-primary" />
				</div>
			</div>
			<p class="text-sm font-medium text-base-content/60 mb-1">Account Status</p>
			<p class="font-display text-2xl font-bold text-base-content">Active</p>
			<div class="mt-3 pt-3 border-t border-base-content/20 text-xs text-base-content/40">
				Member since {currentUser?.createdAt
					? new Date(currentUser.createdAt).toLocaleDateString()
					: 'today'}
			</div>
		</Card>

		<Card>
			<div class="flex items-center justify-between mb-4">
				<div class="w-10 h-10 rounded-xl bg-secondary/10 flex items-center justify-center">
					<CreditCard class="w-5 h-5 text-secondary" />
				</div>
			</div>
			<p class="text-sm font-medium text-base-content/60 mb-1">Current Plan</p>
			<p class="font-display text-2xl font-bold text-base-content capitalize">
				{currentSubscription?.plan?.name || 'Free'}
			</p>
			<div class="mt-3 pt-3 border-t border-base-content/20 text-xs text-base-content/40">
				{currentSubscription?.subscription?.status === 'active'
					? 'Active subscription'
					: 'No active subscription'}
			</div>
		</Card>

		<Card>
			<div class="flex items-center justify-between mb-4">
				<div class="w-10 h-10 rounded-xl bg-tertiary/10 flex items-center justify-center">
					<Zap class="w-5 h-5 text-tertiary" />
				</div>
			</div>
			<p class="text-sm font-medium text-base-content/60 mb-1">API Usage</p>
			<p class="font-display text-2xl font-bold text-base-content">0</p>
			<div class="mt-3 pt-3 border-t border-base-content/20 text-xs text-base-content/40">
				Requests this month
			</div>
		</Card>

		<Card>
			<div class="flex items-center justify-between mb-4">
				<div class="w-10 h-10 rounded-xl bg-success/10 flex items-center justify-center">
					<Settings class="w-5 h-5 text-success" />
				</div>
			</div>
			<p class="text-sm font-medium text-base-content/60 mb-1">Quick Actions</p>
			<p class="font-display text-2xl font-bold text-base-content">3</p>
			<div class="mt-3 pt-3 border-t border-base-content/20 text-xs text-base-content/40">Available below</div>
		</Card>
	</div>

	<!-- Getting Started Section -->
	<Card>
		<h2 class="font-display text-lg font-bold text-base-content mb-6">Getting Started</h2>
		<div class="grid sm:grid-cols-3 gap-4">
			<button
				onclick={() => goto('/app/settings')}
				class="p-4 rounded-xl border border-base-content/20 hover:border-primary/50 hover:bg-base-200 transition-all text-left group"
			>
				<div
					class="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center mb-3 group-hover:bg-primary/20 transition-colors"
				>
					<User class="w-5 h-5 text-primary" />
				</div>
				<h3 class="font-display font-bold text-base-content mb-1">Complete Profile</h3>
				<p class="text-sm text-base-content/60">Add your details and preferences</p>
			</button>

			<button
				onclick={() => goto('/pricing')}
				class="p-4 rounded-xl border border-base-content/20 hover:border-secondary/50 hover:bg-base-200 transition-all text-left group"
			>
				<div
					class="w-10 h-10 rounded-xl bg-secondary/10 flex items-center justify-center mb-3 group-hover:bg-secondary/20 transition-colors"
				>
					<CreditCard class="w-5 h-5 text-secondary" />
				</div>
				<h3 class="font-display font-bold text-base-content mb-1">View Plans</h3>
				<p class="text-sm text-base-content/60">Explore available subscription options</p>
			</button>

			<button
				onclick={() => goto('/app/settings')}
				class="p-4 rounded-xl border border-base-content/20 hover:border-tertiary/50 hover:bg-base-200 transition-all text-left group"
			>
				<div
					class="w-10 h-10 rounded-xl bg-tertiary/10 flex items-center justify-center mb-3 group-hover:bg-tertiary/20 transition-colors"
				>
					<Settings class="w-5 h-5 text-tertiary" />
				</div>
				<h3 class="font-display font-bold text-base-content mb-1">Settings</h3>
				<p class="text-sm text-base-content/60">Configure your account preferences</p>
			</button>
		</div>
	</Card>

	<!-- Placeholder for Your App Content -->
	<Card>
		<h2 class="font-display text-lg font-bold text-base-content mb-4">Your App Content</h2>
		<div class="py-12 text-center">
			<div
				class="w-16 h-16 rounded-2xl bg-base-300/50 flex items-center justify-center mx-auto mb-4"
			>
				<Zap class="w-8 h-8 text-base-content/40" />
			</div>
			<h3 class="font-display font-bold text-base-content mb-2">Build Something Amazing</h3>
			<p class="text-base-content/60 max-w-md mx-auto mb-6">
				This is your dashboard. Add your app's main functionality here. The boilerplate includes
				authentication, billing, and user management ready to go.
			</p>
			<Button
				type="secondary"
				onclick={() => window.open('https://github.com/almatuck/gobot', '_blank')}
			>
				View Documentation
			</Button>
		</div>
	</Card>
</div>
