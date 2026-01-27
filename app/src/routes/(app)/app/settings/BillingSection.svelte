<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Crown, CreditCard, Receipt, ExternalLink, Zap } from 'lucide-svelte';
	import {
		subscription,
		currentPlan,
		subscriptionLoading,
		subscriptionError
	} from '$lib/stores/subscription';
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Spinner from '$lib/components/ui/Spinner.svelte';
	import Alert from '$lib/components/ui/Alert.svelte';

	let isLoadingPortal = $state(false);
	let isLoadingCheckout = $state(false);

	onMount(() => {
		subscription.loadSubscription();
	});

	// Computed values
	const planName = $derived($currentPlan?.displayName ?? 'Free');
	const status = $derived($subscription.subscription?.status ?? 'active');
	const billingCycle = $derived($subscription.subscription?.billingCycle ?? 'monthly');
	const cancelAtPeriodEnd = $derived($subscription.subscription?.cancelAtPeriodEnd ?? false);
	const periodEnd = $derived($subscription.subscription?.currentPeriodEnd);
	const isPremium = $derived($currentPlan?.name !== 'free');

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

	async function handleUpgrade() {
		isLoadingCheckout = true;
		try {
			const url = await subscription.createCheckout('pro', 'monthly');
			if (url) {
				window.location.href = url;
			}
		} finally {
			isLoadingCheckout = false;
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

	function handleViewPlans() {
		goto('/pricing');
	}
</script>

<div class="space-y-6">
	{#if $subscriptionLoading}
		<Card>
			<div class="flex flex-col items-center justify-center gap-4 py-8">
				<Spinner size={32} />
				<p class="text-sm text-base-content/60">Loading billing information...</p>
			</div>
		</Card>
	{:else}
		<!-- Current Plan -->
		<Card>
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
						<h2 class="text-lg font-semibold text-base-content">Current Plan</h2>
						<p class="text-sm text-base-content/60">Manage your subscription</p>
					</div>
				</div>
				<Badge variant={getStatusVariant(status)}>
					{#if cancelAtPeriodEnd}
						Cancelling
					{:else}
						{status.charAt(0).toUpperCase() + status.slice(1).replace('_', ' ')}
					{/if}
				</Badge>
			</div>

			<div class="p-4 bg-base-800/50 rounded-lg mb-6">
				<div class="flex items-center justify-between">
					<div>
						<p class="text-2xl font-bold text-base-content">{planName}</p>
						{#if isPremium && periodEnd}
							<p class="text-sm text-base-content/60 mt-1">
								{#if cancelAtPeriodEnd}
									Ends on {formatDate(periodEnd)}
								{:else}
									Renews {formatDate(periodEnd)} ({billingCycle === 'yearly' ? 'Annual' : 'Monthly'})
								{/if}
							</p>
						{:else if !isPremium}
							<p class="text-sm text-base-content/60 mt-1">Upgrade to unlock premium features</p>
						{/if}
					</div>
					{#if isPremium}
						<p class="text-2xl font-bold text-base-content">
							{billingCycle === 'yearly' ? '$290' : '$29'}<span class="text-sm text-base-content/60 font-normal">/{billingCycle === 'yearly' ? 'year' : 'mo'}</span>
						</p>
					{:else}
						<p class="text-2xl font-bold text-base-content">$0</p>
					{/if}
				</div>
			</div>

			<div class="flex flex-wrap gap-3">
				{#if !isPremium}
					<Button type="primary" onclick={handleUpgrade} disabled={isLoadingCheckout}>
						{#if isLoadingCheckout}
							<Spinner size={16} />
						{:else}
							<Crown class="w-4 h-4" />
						{/if}
						Upgrade to Pro
					</Button>
					<Button type="secondary" onclick={handleViewPlans}>
						View All Plans
					</Button>
				{:else}
					<Button type="secondary" onclick={handleViewPlans}>
						Change Plan
					</Button>
				{/if}
			</div>
		</Card>

		<!-- Payment Method -->
		{#if isPremium}
			<Card>
				<div class="flex items-center gap-3 mb-6">
					<div class="w-10 h-10 rounded-xl bg-secondary/10 flex items-center justify-center">
						<CreditCard class="w-5 h-5 text-secondary" />
					</div>
					<div>
						<h2 class="text-lg font-semibold text-base-content">Payment Method</h2>
						<p class="text-sm text-base-content/60">Manage your payment details</p>
					</div>
				</div>

				<p class="text-sm text-base-content/60 mb-4">
					Update your payment method, view billing history, or download invoices from the Stripe billing portal.
				</p>

				<Button type="secondary" onclick={handleManageBilling} disabled={isLoadingPortal}>
					{#if isLoadingPortal}
						<Spinner size={16} />
					{:else}
						<CreditCard class="w-4 h-4" />
					{/if}
					Manage Payment
					<ExternalLink class="w-3.5 h-3.5" />
				</Button>
			</Card>
		{/if}

		<!-- Invoices -->
		{#if isPremium}
			<Card>
				<div class="flex items-center gap-3 mb-6">
					<div class="w-10 h-10 rounded-xl bg-tertiary/10 flex items-center justify-center">
						<Receipt class="w-5 h-5 text-tertiary" />
					</div>
					<div>
						<h2 class="text-lg font-semibold text-base-content">Invoices</h2>
						<p class="text-sm text-base-content/60">View and download your invoices</p>
					</div>
				</div>

				<p class="text-sm text-base-content/60 mb-4">
					Access your complete billing history and download invoices from the billing portal.
				</p>

				<Button type="secondary" onclick={handleManageBilling} disabled={isLoadingPortal}>
					{#if isLoadingPortal}
						<Spinner size={16} />
					{:else}
						<Receipt class="w-4 h-4" />
					{/if}
					View Invoices
					<ExternalLink class="w-3.5 h-3.5" />
				</Button>
			</Card>
		{/if}

		{#if $subscriptionError}
			<Alert type="error" title="Error">{$subscriptionError}</Alert>
		{/if}
	{/if}
</div>
