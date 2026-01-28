import webapi from "./gocliRequest"
import * as components from "./gobotComponents"
export * from "./gobotComponents"

/**
 * @description "Health check endpoint"
 */
export function healthCheck() {
	return webapi.get<components.HealthResponse>(`/health`)
}

/**
 * @description "List connected agents for organization"
 * @param params
 */
export function listAgents(params: components.ListAgentsRequestParams, orgId: string) {
	return webapi.get<components.ListAgentsResponse>(`/api/v1/agents`, params)
}

/**
 * @description "Get agent status"
 * @param params
 */
export function getAgentStatus(params: components.AgentStatusRequestParams, agentId: string) {
	return webapi.get<components.AgentStatusResponse>(`/api/v1/agents/${agentId}/status`, params)
}

/**
 * @description "Get auth configuration (OAuth providers enabled)"
 */
export function getAuthConfig() {
	return webapi.get<components.AuthConfigResponse>(`/api/v1/auth/config`)
}

/**
 * @description "Request password reset"
 * @param req
 */
export function forgotPassword(req: components.ForgotPasswordRequest) {
	return webapi.post<components.MessageResponse>(`/api/v1/auth/forgot-password`, req)
}

/**
 * @description "User login"
 * @param req
 */
export function login(req: components.LoginRequest) {
	return webapi.post<components.LoginResponse>(`/api/v1/auth/login`, req)
}

/**
 * @description "Refresh authentication token"
 * @param req
 */
export function refreshToken(req: components.RefreshTokenRequest) {
	return webapi.post<components.RefreshTokenResponse>(`/api/v1/auth/refresh`, req)
}

/**
 * @description "Register new user"
 * @param req
 */
export function register(req: components.RegisterRequest) {
	return webapi.post<components.LoginResponse>(`/api/v1/auth/register`, req)
}

/**
 * @description "Resend email verification"
 * @param req
 */
export function resendVerification(req: components.ResendVerificationRequest) {
	return webapi.post<components.MessageResponse>(`/api/v1/auth/resend-verification`, req)
}

/**
 * @description "Reset password with token"
 * @param req
 */
export function resetPassword(req: components.ResetPasswordRequest) {
	return webapi.post<components.MessageResponse>(`/api/v1/auth/reset-password`, req)
}

/**
 * @description "Verify email address with token"
 * @param req
 */
export function verifyEmail(req: components.EmailVerificationRequest) {
	return webapi.post<components.MessageResponse>(`/api/v1/auth/verify-email`, req)
}

/**
 * @description "List user notifications"
 * @param params
 */
export function listNotifications(params: components.ListNotificationsRequestParams) {
	return webapi.get<components.ListNotificationsResponse>(`/api/v1/notifications`, params)
}

/**
 * @description "Delete notification"
 * @param params
 */
export function deleteNotification(params: components.DeleteNotificationRequestParams, id: string) {
	return webapi.delete<components.MessageResponse>(`/api/v1/notifications/${id}`, params)
}

/**
 * @description "Mark notification as read"
 * @param params
 */
export function markNotificationRead(params: components.MarkNotificationReadRequestParams, id: string) {
	return webapi.put<components.MessageResponse>(`/api/v1/notifications/${id}/read`, params)
}

/**
 * @description "Mark all notifications as read"
 */
export function markAllNotificationsRead() {
	return webapi.put<components.MessageResponse>(`/api/v1/notifications/read-all`)
}

/**
 * @description "Get unread notification count"
 */
export function getUnreadCount() {
	return webapi.get<components.GetUnreadCountResponse>(`/api/v1/notifications/unread-count`)
}

/**
 * @description "OAuth callback - exchange code for tokens"
 * @param params
 * @param req
 */
export function oAuthCallback(params: components.OAuthLoginRequestParams, req: components.OAuthLoginRequest, provider: string) {
	return webapi.post<components.OAuthLoginResponse>(`/api/v1/oauth/${provider}/callback`, params, req)
}

/**
 * @description "Get OAuth authorization URL"
 * @param params
 */
export function getOAuthUrl(params: components.GetOAuthUrlRequestParams, provider: string) {
	return webapi.get<components.GetOAuthUrlResponse>(`/api/v1/oauth/${provider}/url`, params)
}

/**
 * @description "Disconnect OAuth provider"
 * @param params
 */
export function disconnectOAuth(params: components.DisconnectOAuthRequestParams, provider: string) {
	return webapi.delete<components.MessageResponse>(`/api/v1/oauth/${provider}`, params)
}

/**
 * @description "List connected OAuth providers"
 */
export function listOAuthProviders() {
	return webapi.get<components.ListOAuthProvidersResponse>(`/api/v1/oauth/providers`)
}

/**
 * @description "Create a new organization"
 * @param req
 */
export function createOrganization(req: components.CreateOrganizationRequest) {
	return webapi.post<components.CreateOrganizationResponse>(`/api/v1/organizations`, req)
}

/**
 * @description "List user's organizations"
 */
export function listOrganizations() {
	return webapi.get<components.ListOrganizationsResponse>(`/api/v1/organizations`)
}

/**
 * @description "Get organization by ID"
 * @param params
 */
export function getOrganization(params: components.GetOrganizationRequestParams, id: string) {
	return webapi.get<components.GetOrganizationResponse>(`/api/v1/organizations/${id}`, params)
}

/**
 * @description "Update organization"
 * @param params
 * @param req
 */
export function updateOrganization(params: components.UpdateOrganizationRequestParams, req: components.UpdateOrganizationRequest, id: string) {
	return webapi.put<components.GetOrganizationResponse>(`/api/v1/organizations/${id}`, params, req)
}

/**
 * @description "Delete organization"
 * @param params
 */
export function deleteOrganization(params: components.DeleteOrganizationRequestParams, id: string) {
	return webapi.delete<components.MessageResponse>(`/api/v1/organizations/${id}`, params)
}

/**
 * @description "Invite member to organization"
 * @param params
 * @param req
 */
export function inviteMember(params: components.InviteMemberRequestParams, req: components.InviteMemberRequest, id: string) {
	return webapi.post<components.InviteMemberResponse>(`/api/v1/organizations/${id}/invites`, params, req)
}

/**
 * @description "List pending invites"
 * @param params
 */
export function listInvites(params: components.ListInvitesRequestParams, id: string) {
	return webapi.get<components.ListInvitesResponse>(`/api/v1/organizations/${id}/invites`, params)
}

/**
 * @description "Leave organization"
 * @param params
 */
export function leaveOrganization(params: components.LeaveOrganizationRequestParams, id: string) {
	return webapi.post<components.MessageResponse>(`/api/v1/organizations/${id}/leave`, params)
}

/**
 * @description "List organization members"
 * @param params
 */
export function listMembers(params: components.ListMembersRequestParams, id: string) {
	return webapi.get<components.ListMembersResponse>(`/api/v1/organizations/${id}/members`, params)
}

/**
 * @description "Revoke invite"
 * @param params
 */
export function revokeInvite(params: components.RevokeInviteRequestParams, orgId: string, inviteId: string) {
	return webapi.delete<components.MessageResponse>(`/api/v1/organizations/${orgId}/invites/${inviteId}`, params)
}

/**
 * @description "Update member role"
 * @param params
 * @param req
 */
export function updateMemberRole(params: components.UpdateMemberRoleRequestParams, req: components.UpdateMemberRoleRequest, orgId: string, userId: string) {
	return webapi.put<components.MessageResponse>(`/api/v1/organizations/${orgId}/members/${userId}`, params, req)
}

/**
 * @description "Remove member from organization"
 * @param params
 */
export function removeMember(params: components.RemoveMemberRequestParams, orgId: string, userId: string) {
	return webapi.delete<components.MessageResponse>(`/api/v1/organizations/${orgId}/members/${userId}`, params)
}

/**
 * @description "Accept organization invite"
 * @param req
 */
export function acceptInvite(req: components.AcceptInviteRequest) {
	return webapi.post<components.AcceptInviteResponse>(`/api/v1/organizations/invites/accept`, req)
}

/**
 * @description "Switch current organization"
 * @param req
 */
export function switchOrganization(req: components.SwitchOrganizationRequest) {
	return webapi.post<components.MessageResponse>(`/api/v1/organizations/switch`, req)
}

/**
 * @description "Get invite details by token"
 * @param params
 */
export function getInviteByToken(params: components.GetInviteByTokenRequestParams, token: string) {
	return webapi.get<components.GetInviteByTokenResponse>(`/api/v1/organizations/invites/${token}`, params)
}

/**
 * @description "Create the first admin user (only works when no admin exists)"
 * @param req
 */
export function createAdmin(req: components.CreateAdminRequest) {
	return webapi.post<components.CreateAdminResponse>(`/api/v1/setup/admin`, req)
}

/**
 * @description "Check if setup is required (no admin exists)"
 */
export function setupStatus() {
	return webapi.get<components.SetupStatusResponse>(`/api/v1/setup/status`)
}

/**
 * @description "Get current user profile"
 */
export function getCurrentUser() {
	return webapi.get<components.GetUserResponse>(`/api/v1/user/me`)
}

/**
 * @description "Update current user profile"
 * @param req
 */
export function updateCurrentUser(req: components.UpdateUserRequest) {
	return webapi.put<components.GetUserResponse>(`/api/v1/user/me`, req)
}

/**
 * @description "Delete current user account"
 * @param req
 */
export function deleteAccount(req: components.DeleteAccountRequest) {
	return webapi.delete<components.MessageResponse>(`/api/v1/user/me`, req)
}

/**
 * @description "Change password for authenticated user"
 * @param req
 */
export function changePassword(req: components.ChangePasswordRequest) {
	return webapi.post<components.MessageResponse>(`/api/v1/user/me/change-password`, req)
}

/**
 * @description "Get user preferences"
 */
export function getPreferences() {
	return webapi.get<components.GetPreferencesResponse>(`/api/v1/user/me/preferences`)
}

/**
 * @description "Update user preferences"
 * @param req
 */
export function updatePreferences(req: components.UpdatePreferencesRequest) {
	return webapi.put<components.GetPreferencesResponse>(`/api/v1/user/me/preferences`, req)
}
