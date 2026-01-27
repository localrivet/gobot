<script lang="ts">
	import { onMount } from 'svelte';
	import { Crown, Zap, CreditCard, ExternalLink } from 'lucide-svelte';
	import {
		subscription,
		currentPlan,
		usageStats,
		subscriptionLoading,
		subscriptionError
	} from '$lib/stores/subscription';
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Spinner from '$lib/components/ui/Spinner.svelte';

	let isLoadingPortal = $state(false);

	onMount(() => {
		subscription.loadSubscription();
		subscription.loadUsage();
	});

	// Computed values
	const planName = $derived($currentPlan?.displayName ?? 'Free');
	const status = $derived($subscription.subscription?.status ?? 'active');
	const billingCycle = $derived($subscription.subscription?.billingCycle ?? 'monthly');
	const cancelAtPeriodEnd = $derived($subscription.subscription?.cancelAtPeriodEnd ?? false);
	const periodEnd = $derived($subscription.subscription?.currentPeriodEnd);

	// Usage - read from meters if available
	const apiUsage = $derived($usageStats?.meters?.['api_calls'] ?? 0);

	// Status badge variant
	function getStatusVariant(status: string): 'success' | 'warning' | 'error' | 'default' {
		switch (status) {
			case 'active':
				return 'success';
			case 'trialing':
				return 'warning';
			case 'cancelled':
			case 'past_due':
			case 'expired':
				return 'error';
			default:
				return 'default';
		}
	}

	// Format date
	function formatDate(dateString?: string): string {
		if (!dateString) return '';
		try {
			return new Date(dateString).toLocaleDateString(undefined, {
				month: 'short',
				day: 'numeric',
				year: 'numeric'
			});
		} catch {
			return '';
		}
	}

	// Actions
	async function handleUpgrade() {
		const url = await subscription.createCheckout('pro', 'monthly');
		if (url) {
			window.location.href = url;
		}
	}

	async function handleManageBilling() {
		isLoadingPortal = true;
		try {
			const url = await subscription.openBillingPortal();
			if (url) {
				window.open(url, '_blank');
			}
		} finally {
			isLoadingPortal = false;
		}
	}

	const isPremium = $derived($currentPlan?.name !== 'free');
</script>

<Card>
	<section aria-labelledby="subscription-heading">
		<!-- Header -->
		<div class="flex items-center justify-between mb-6">
			<div class="flex items-center gap-3">
				<div class="w-10 h-10 rounded-xl flex items-center justify-center {isPremium ? 'bg-warning/10' : 'bg-primary/10'}">
					{#if isPremium}
						<Crown class="w-5 h-5 text-warning" />
					{:else}
						<Zap class="w-5 h-5 text-primary" />
					{/if}
				</div>
				<div>
					<h3 id="subscription-heading" class="text-lg font-semibold text-white">{planName} Plan</h3>
					<Badge variant={getStatusVariant(status)}>
						{#if cancelAtPeriodEnd}
							Cancelling
						{:else}
							{status.charAt(0).toUpperCase() + status.slice(1).replace('_', ' ')}
						{/if}
					</Badge>
				</div>
			</div>
		</div>

		{#if $subscriptionLoading}
			<div class="flex items-center justify-center gap-3 py-8">
				<Spinner size={24} />
				<span class="text-base-400">Loading subscription...</span>
			</div>
		{:else}
			<!-- Usage Summary -->
			{#if apiUsage > 0}
				<div class="mb-6 p-4 bg-base-800/50 rounded-lg">
					<div class="flex items-center justify-between text-sm">
						<span class="text-base-300">API Calls This Period</span>
						<span class="text-white font-medium">{apiUsage}</span>
					</div>
				</div>
			{/if}

			<!-- Actions -->
			<div class="flex flex-wrap gap-3">
				{#if !isPremium}
					<Button type="primary" onclick={handleUpgrade}>
						<Crown class="w-4 h-4" />
						Upgrade to Pro
					</Button>
				{/if}

				{#if isPremium}
					<Button type="secondary" onclick={handleManageBilling} disabled={isLoadingPortal}>
						{#if isLoadingPortal}
							<Spinner size={16} />
						{:else}
							<CreditCard class="w-4 h-4" />
						{/if}
						Manage Billing
						<ExternalLink class="w-3.5 h-3.5" />
					</Button>
				{/if}
			</div>

			<!-- Billing Info -->
			{#if periodEnd && isPremium}
				<div class="mt-6 pt-6 border-t border-base-700 text-sm text-base-400 space-y-1">
					{#if cancelAtPeriodEnd}
						<p>Your subscription will end on <span class="text-white font-medium">{formatDate(periodEnd)}</span></p>
					{:else}
						<p>Next billing date: <span class="text-white font-medium">{formatDate(periodEnd)}</span></p>
						<p>Billing cycle: <span class="text-white font-medium">{billingCycle === 'yearly' ? 'Annual' : 'Monthly'}</span></p>
					{/if}
				</div>
			{/if}
		{/if}

		{#if $subscriptionError}
			<div class="mt-4 p-3 bg-error/10 border border-error/30 rounded-lg text-sm text-error">
				{$subscriptionError}
			</div>
		{/if}
	</section>
</Card>
