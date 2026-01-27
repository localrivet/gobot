import { writable, derived, get } from 'svelte/store';
import * as api from '$lib/api/gobot';
import type {
	SubscriptionPlan,
	UserSubscription,
	UsageStats,
	BillingRecord
} from '$lib/api/gobotComponents';
import { logger } from '$lib/monitoring';

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
	plan: null,
	usage: null,
	availablePlans: [],
	billingHistory: [],
	isLoading: false,
	error: null
};

/**
 * Create the subscription store
 */
function createSubscriptionStore() {
	const { subscribe, set, update } = writable<SubscriptionState>(initialState);

	return {
		subscribe,

		/**
		 * Load available subscription plans (public endpoint)
		 */
		async loadPlans(): Promise<void> {
			update((state) => ({ ...state, isLoading: true, error: null }));

			try {
				const response = await api.listPlans();
				update((state) => ({
					...state,
					availablePlans: response.plans,
					isLoading: false
				}));
				logger.debug('Loaded subscription plans');
			} catch (error) {
				const errorMessage = error instanceof Error ? error.message : 'Failed to load plans';
				update((state) => ({
					...state,
					isLoading: false,
					error: errorMessage
				}));
				logger.error('Failed to load subscription plans', error);
			}
		},

		/**
		 * Load current user's subscription
		 */
		async loadSubscription(): Promise<void> {
			update((state) => ({ ...state, isLoading: true, error: null }));

			try {
				const response = await api.getSubscription();
				update((state) => ({
					...state,
					subscription: response.subscription,
					plan: response.plan,
					isLoading: false
				}));
				logger.debug('Loaded user subscription');
			} catch (error) {
				const errorMessage = error instanceof Error ? error.message : 'Failed to load subscription';
				update((state) => ({
					...state,
					isLoading: false,
					error: errorMessage
				}));
				logger.error('Failed to load subscription', error);
			}
		},

		/**
		 * Load usage statistics
		 */
		async loadUsage(): Promise<void> {
			try {
				const response = await api.getUsageStats();
				update((state) => ({
					...state,
					usage: response.stats
				}));
				logger.debug('Loaded usage stats');
			} catch (error) {
				logger.error('Failed to load usage stats', error);
			}
		},

		/**
		 * Load billing history
		 */
		async loadBillingHistory(page = 1, pageSize = 10): Promise<void> {
			try {
				const response = await api.listBillingHistory({ page, pageSize });
				update((state) => ({
					...state,
					billingHistory: response.records
				}));
				logger.debug('Loaded billing history');
			} catch (error) {
				logger.error('Failed to load billing history', error);
			}
		},

		/**
		 * Create checkout session for subscribing to a plan
		 */
		async createCheckout(
			planName: string,
			billingCycle: 'monthly' | 'yearly' = 'monthly'
		): Promise<string | null> {
			update((state) => ({ ...state, isLoading: true, error: null }));

			try {
				const response = await api.createCheckout({ planName, billingCycle });
				update((state) => ({ ...state, isLoading: false }));
				logger.info('Created checkout session', { planName, billingCycle });
				return response.checkoutUrl;
			} catch (error) {
				const errorMessage = error instanceof Error ? error.message : 'Failed to create checkout';
				update((state) => ({
					...state,
					isLoading: false,
					error: errorMessage
				}));
				logger.error('Failed to create checkout session', error);
				return null;
			}
		},

		/**
		 * Open billing portal
		 */
		async openBillingPortal(): Promise<string | null> {
			update((state) => ({ ...state, isLoading: true, error: null }));

			try {
				const response = await api.createBillingPortal();
				update((state) => ({ ...state, isLoading: false }));
				logger.info('Opened billing portal');
				return response.portalUrl;
			} catch (error) {
				const errorMessage =
					error instanceof Error ? error.message : 'Failed to open billing portal';
				update((state) => ({
					...state,
					isLoading: false,
					error: errorMessage
				}));
				logger.error('Failed to open billing portal', error);
				return null;
			}
		},

		/**
		 * Cancel subscription
		 */
		async cancelSubscription(cancelAtPeriodEnd = true): Promise<boolean> {
			update((state) => ({ ...state, isLoading: true, error: null }));

			try {
				await api.cancelSubscription({ cancelAtPeriodEnd });
				// Reload subscription to get updated status
				await this.loadSubscription();
				logger.info('Subscription cancelled', { cancelAtPeriodEnd });
				return true;
			} catch (error) {
				const errorMessage =
					error instanceof Error ? error.message : 'Failed to cancel subscription';
				update((state) => ({
					...state,
					isLoading: false,
					error: errorMessage
				}));
				logger.error('Failed to cancel subscription', error);
				return false;
			}
		},

		/**
		 * Check if user has access to a specific feature
		 */
		async checkFeature(feature: string): Promise<boolean> {
			try {
				const response = await api.checkFeature({ feature });
				return response.hasAccess;
			} catch (error) {
				logger.error('Failed to check feature access', error);
				return false;
			}
		},

		/**
		 * Check if user is on a paid plan
		 */
		isPremium(): boolean {
			const state = get({ subscribe });
			if (!state.plan) return false;
			return state.plan.name !== 'free';
		},

		/**
		 * Get current plan name
		 */
		getPlanName(): string {
			const state = get({ subscribe });
			return state.plan?.name ?? 'free';
		},

		/**
		 * Get usage meter value
		 */
		getMeter(name: string): number {
			const state = get({ subscribe });
			return state.usage?.meters?.[name] ?? 0;
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
export const isPremium = derived(subscription, ($sub) => $sub.plan?.name !== 'free');
export const currentPlan = derived(subscription, ($sub) => $sub.plan);
export const usageStats = derived(subscription, ($sub) => $sub.usage);
export const subscriptionStatus = derived(
	subscription,
	($sub) => $sub.subscription?.status ?? 'none'
);
export const isTrialing = derived(subscription, ($sub) => $sub.subscription?.status === 'trialing');
export const isCancelled = derived(
	subscription,
	($sub) => $sub.subscription?.cancelAtPeriodEnd ?? false
);
export const subscriptionLoading = derived(subscription, ($sub) => $sub.isLoading);
export const subscriptionError = derived(subscription, ($sub) => $sub.error);
