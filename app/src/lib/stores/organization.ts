import { writable, derived, get } from 'svelte/store';
import * as api from '$lib/api/gobot';
import type {
	Organization,
	OrganizationMember,
	OrganizationInvite
} from '$lib/api/gobotComponents';
import { logger } from '$lib/monitoring';

/**
 * Organization state interface
 */
export interface OrganizationState {
	organizations: Organization[];
	currentOrganization: Organization | null;
	members: OrganizationMember[];
	invites: OrganizationInvite[];
	role: string | null;
	isLoading: boolean;
	error: string | null;
}

const initialState: OrganizationState = {
	organizations: [],
	currentOrganization: null,
	members: [],
	invites: [],
	role: null,
	isLoading: false,
	error: null
};

// Storage key for current org
const CURRENT_ORG_KEY = 'gobot_current_org';

function getStoredOrgId(): string | null {
	if (typeof window === 'undefined') return null;
	return localStorage.getItem(CURRENT_ORG_KEY);
}

function storeOrgId(orgId: string | null): void {
	if (typeof window === 'undefined') return;
	if (orgId) {
		localStorage.setItem(CURRENT_ORG_KEY, orgId);
	} else {
		localStorage.removeItem(CURRENT_ORG_KEY);
	}
}

function createOrganizationStore() {
	const { subscribe, set, update } = writable<OrganizationState>(initialState);

	return {
		subscribe,

		/**
		 * Load user's organizations
		 */
		async loadOrganizations(): Promise<void> {
			update((state) => ({ ...state, isLoading: true, error: null }));

			try {
				const response = await api.listOrganizations();
				const storedOrgId = getStoredOrgId();

				// Find current org or default to first one
				let currentOrg = response.organizations.find((o) => o.id === storedOrgId) || null;
				if (!currentOrg && response.organizations.length > 0) {
					currentOrg = response.organizations[0];
					storeOrgId(currentOrg.id);
				}

				update((state) => ({
					...state,
					organizations: response.organizations,
					currentOrganization: currentOrg,
					isLoading: false
				}));

				// Load members and invites for current org
				if (currentOrg) {
					await this.loadMembers(currentOrg.id);
					await this.loadInvites(currentOrg.id);
				}

				logger.debug('Organizations loaded', { count: response.organizations.length });
			} catch (error) {
				const errorMessage =
					error instanceof Error ? error.message : 'Failed to load organizations';
				update((state) => ({ ...state, isLoading: false, error: errorMessage }));
				logger.error('Failed to load organizations', error);
			}
		},

		/**
		 * Create a new organization
		 */
		async createOrganization(data: {
			name: string;
			slug?: string;
		}): Promise<Organization | null> {
			update((state) => ({ ...state, isLoading: true, error: null }));

			try {
				const response = await api.createOrganization(data);
				const newOrg = response.organization;

				update((state) => ({
					...state,
					organizations: [...state.organizations, newOrg],
					currentOrganization: newOrg,
					isLoading: false
				}));

				storeOrgId(newOrg.id);
				logger.info('Organization created', { id: newOrg.id, name: newOrg.name });
				return newOrg;
			} catch (error) {
				const errorMessage =
					error instanceof Error ? error.message : 'Failed to create organization';
				update((state) => ({ ...state, isLoading: false, error: errorMessage }));
				logger.error('Failed to create organization', error);
				return null;
			}
		},

		/**
		 * Switch to a different organization
		 */
		async switchOrganization(orgId: string): Promise<boolean> {
			const state = get({ subscribe });
			const org = state.organizations.find((o) => o.id === orgId);
			if (!org) return false;

			try {
				await api.switchOrganization({ organizationId: orgId });

				update((s) => ({
					...s,
					currentOrganization: org,
					members: [],
					invites: []
				}));

				storeOrgId(orgId);

				// Load members and invites for new org
				await this.loadMembers(orgId);
				await this.loadInvites(orgId);

				logger.info('Switched organization', { id: orgId });
				return true;
			} catch (error) {
				logger.error('Failed to switch organization', error);
				return false;
			}
		},

		/**
		 * Load members for an organization
		 */
		async loadMembers(orgId: string): Promise<void> {
			try {
				const response = await api.listMembers({}, orgId);
				update((state) => ({ ...state, members: response.members }));
			} catch (error) {
				logger.error('Failed to load members', error);
			}
		},

		/**
		 * Load invites for an organization
		 */
		async loadInvites(orgId: string): Promise<void> {
			try {
				const response = await api.listInvites({}, orgId);
				update((state) => ({ ...state, invites: response.invites }));
			} catch (error) {
				logger.error('Failed to load invites', error);
			}
		},

		/**
		 * Invite a member to the organization
		 */
		async inviteMember(
			orgId: string,
			email: string,
			role: string = 'member'
		): Promise<OrganizationInvite | null> {
			try {
				const response = await api.inviteMember({}, { email, role }, orgId);
				update((state) => ({
					...state,
					invites: [...state.invites, response.invite]
				}));
				logger.info('Member invited', { email, role });
				return response.invite;
			} catch (error) {
				logger.error('Failed to invite member', error);
				throw error;
			}
		},

		/**
		 * Revoke an invite
		 */
		async revokeInvite(orgId: string, inviteId: string): Promise<boolean> {
			try {
				await api.revokeInvite({}, orgId, inviteId);
				update((state) => ({
					...state,
					invites: state.invites.filter((i) => i.id !== inviteId)
				}));
				logger.info('Invite revoked', { inviteId });
				return true;
			} catch (error) {
				logger.error('Failed to revoke invite', error);
				return false;
			}
		},

		/**
		 * Update a member's role
		 */
		async updateMemberRole(orgId: string, userId: string, role: string): Promise<boolean> {
			try {
				await api.updateMemberRole({}, { role }, orgId, userId);
				update((state) => ({
					...state,
					members: state.members.map((m) => (m.userId === userId ? { ...m, role } : m))
				}));
				logger.info('Member role updated', { userId, role });
				return true;
			} catch (error) {
				logger.error('Failed to update member role', error);
				return false;
			}
		},

		/**
		 * Remove a member from the organization
		 */
		async removeMember(orgId: string, userId: string): Promise<boolean> {
			try {
				await api.removeMember({}, orgId, userId);
				update((state) => ({
					...state,
					members: state.members.filter((m) => m.userId !== userId)
				}));
				logger.info('Member removed', { userId });
				return true;
			} catch (error) {
				logger.error('Failed to remove member', error);
				return false;
			}
		},

		/**
		 * Leave the current organization
		 */
		async leaveOrganization(orgId: string): Promise<boolean> {
			try {
				await api.leaveOrganization({}, orgId);
				update((state) => ({
					...state,
					organizations: state.organizations.filter((o) => o.id !== orgId),
					currentOrganization:
						state.currentOrganization?.id === orgId
							? state.organizations.find((o) => o.id !== orgId) || null
							: state.currentOrganization
				}));
				logger.info('Left organization', { orgId });
				return true;
			} catch (error) {
				logger.error('Failed to leave organization', error);
				return false;
			}
		},

		/**
		 * Accept an organization invite
		 */
		async acceptInvite(token: string): Promise<Organization | null> {
			try {
				const response = await api.acceptInvite({ token });
				// Reload organizations to include the new one
				await this.loadOrganizations();
				return response.organization;
			} catch (error) {
				logger.error('Failed to accept invite', error);
				throw error;
			}
		},

		/**
		 * Clear error
		 */
		clearError(): void {
			update((state) => ({ ...state, error: null }));
		},

		/**
		 * Reset store
		 */
		reset(): void {
			set(initialState);
			storeOrgId(null);
		}
	};
}

export const organization = createOrganizationStore();

// Derived stores
export const currentOrganization = derived(organization, ($org) => $org.currentOrganization);
export const organizations = derived(organization, ($org) => $org.organizations);
export const organizationMembers = derived(organization, ($org) => $org.members);
export const organizationInvites = derived(organization, ($org) => $org.invites);
export const organizationLoading = derived(organization, ($org) => $org.isLoading);
export const organizationError = derived(organization, ($org) => $org.error);
export const hasOrganization = derived(organization, ($org) => $org.organizations.length > 0);
