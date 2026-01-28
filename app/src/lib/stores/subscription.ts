import { writable, derived, get } from 'svelte/store';
import { logger } from '$lib/monitoring';

/**
 * Subscription plan type (stubbed for open source)
 */
export interface SubscriptionPlan {
	name: string;
	displayName: string;
	monthlyPrice: number;
	yearlyPrice: number;
	features: string[];
}

/**
 * User subscription type (stubbed for open source)
 */
export interface UserSubscription {
	id: string;
	planId: string;
	status: string;
	currentPeriodEnd: string;
	cancelAtPeriodEnd: boolean;
}

/**
 * Usage stats type (stubbed for open source)
 */
export interface UsageStats {
	meters: Record<string, number>;
}

/**
 * Billing record type (stubbed for open source)
 */
export interface BillingRecord {
	id: string;
	amount: number;
	date: string;
	status: string;
}

/**
 * Subscription state interface
 */
export interface SubscriptionState {
	subscription: UserSubscription | null;
	plan: SubscriptionPlan | null;
	usage: UsageStats | null;
	availablePlans: SubscriptionPlan[];
	billingHistory: BillingRecord[];
	isLoading: boolean;
	error: string | null;
}

/**
 * Initial subscription state
 */
const initialState: SubscriptionState = {
	subscription: null,
	plan: {
		name: 'free',
		displayName: 'Free',
		monthlyPrice: 0,
		yearlyPrice: 0,
		features: ['Unlimited use']
	},
	usage: { meters: {} },
	availablePlans: [],
	billingHistory: [],
	isLoading: false,
	error: null
};

/**
 * Create the subscription store (stubbed for open source - no billing)
 */
function createSubscriptionStore() {
	const { subscribe, set, update } = writable<SubscriptionState>(initialState);

	return {
		subscribe,

		/**
		 * Load available subscription plans (stubbed)
		 */
		async loadPlans(): Promise<void> {
			logger.debug('Subscription plans not available (open source version)');
		},

		/**
		 * Load current user's subscription (stubbed)
		 */
		async loadSubscription(): Promise<void> {
			logger.debug('Subscription not available (open source version)');
		},

		/**
		 * Load usage statistics (stubbed)
		 */
		async loadUsage(): Promise<void> {
			logger.debug('Usage stats not available (open source version)');
		},

		/**
		 * Load billing history (stubbed)
		 */
		async loadBillingHistory(page = 1, pageSize = 10): Promise<void> {
			logger.debug('Billing history not available (open source version)');
		},

		/**
		 * Create checkout session (stubbed)
		 */
		async createCheckout(
			planName: string,
			billingCycle: 'monthly' | 'yearly' = 'monthly'
		): Promise<string | null> {
			logger.debug('Checkout not available (open source version)');
			return null;
		},

		/**
		 * Open billing portal (stubbed)
		 */
		async openBillingPortal(): Promise<string | null> {
			logger.debug('Billing portal not available (open source version)');
			return null;
		},

		/**
		 * Cancel subscription (stubbed)
		 */
		async cancelSubscription(cancelAtPeriodEnd = true): Promise<boolean> {
			logger.debug('Cancel subscription not available (open source version)');
			return false;
		},

		/**
		 * Check if user has access to a specific feature (stubbed - always true)
		 */
		async checkFeature(feature: string): Promise<boolean> {
			return true;
		},

		/**
		 * Check if user is on a paid plan
		 */
		isPremium(): boolean {
			return false;
		},

		/**
		 * Get current plan name
		 */
		getPlanName(): string {
			return 'free';
		},

		/**
		 * Get usage meter value
		 */
		getMeter(name: string): number {
			return 0;
		},

		/**
		 * Reset store state
		 */
		reset(): void {
			set(initialState);
		},

		/**
		 * Clear error
		 */
		clearError(): void {
			update((state) => ({ ...state, error: null }));
		}
	};
}

// Export the subscription store singleton
export const subscription = createSubscriptionStore();

// Derived stores for convenience
export const isPremium = derived(subscription, ($sub) => false);
export const currentPlan = derived(subscription, ($sub) => $sub.plan);
export const usageStats = derived(subscription, ($sub) => $sub.usage);
export const subscriptionStatus = derived(
	subscription,
	($sub) => $sub.subscription?.status ?? 'active'
);
export const isTrialing = derived(subscription, ($sub) => false);
export const isCancelled = derived(
	subscription,
	($sub) => false
);
export const subscriptionLoading = derived(subscription, ($sub) => $sub.isLoading);
export const subscriptionError = derived(subscription, ($sub) => $sub.error);
