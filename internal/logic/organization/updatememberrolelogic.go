package organization

import (
	"context"
	"fmt"

	"gobot/internal/auth"
	"gobot/internal/db"
	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateMemberRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateMemberRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateMemberRoleLogic {
	return &UpdateMemberRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateMemberRoleLogic) UpdateMemberRole(req *types.UpdateMemberRoleRequest) (resp *types.MessageResponse, err error) {
	if !l.svcCtx.Config.IsOrganizationsEnabled() {
		return nil, fmt.Errorf("organizations feature is not enabled")
	}

	if !l.svcCtx.UseLocal() {
		return nil, fmt.Errorf("organizations not available in this mode")
	}

	// Validate role
	validRoles := map[string]bool{"owner": true, "admin": true, "member": true}
	if !validRoles[req.Role] {
		return nil, fmt.Errorf("invalid role: must be owner, admin, or member")
	}

	// Get user ID from context
	userID, err := auth.GetUserIDFromContext(l.ctx)
	if err != nil {
		l.Errorf("Failed to get user ID: %v", err)
		return nil, err
	}

	// Get organization to check ownership
	org, err := l.svcCtx.DB.Queries.GetOrganizationByID(l.ctx, req.OrgId)
	if err != nil {
		l.Errorf("Failed to get organization: %v", err)
		return nil, fmt.Errorf("organization not found")
	}

	// Check if user is owner or admin
	currentMember, err := l.svcCtx.DB.Queries.GetOrganizationMember(l.ctx, db.GetOrganizationMemberParams{
		OrganizationID: req.OrgId,
		UserID:         userID.String(),
	})
	if err != nil {
		l.Errorf("Failed to get member: %v", err)
		return nil, fmt.Errorf("you are not a member of this organization")
	}

	// Only owner can change roles
	if currentMember.Role != "owner" {
		return nil, fmt.Errorf("only the owner can change member roles")
	}

	// Cannot change owner's role (transfer ownership is a separate operation)
	if req.UserId == org.OwnerID && req.Role != "owner" {
		return nil, fmt.Errorf("cannot change the owner's role")
	}

	// Cannot make someone else an owner through this endpoint
	if req.Role == "owner" && req.UserId != org.OwnerID {
		return nil, fmt.Errorf("cannot assign owner role; use transfer ownership instead")
	}

	// Verify target user is a member
	_, err = l.svcCtx.DB.Queries.GetOrganizationMember(l.ctx, db.GetOrganizationMemberParams{
		OrganizationID: req.OrgId,
		UserID:         req.UserId,
	})
	if err != nil {
		return nil, fmt.Errorf("user is not a member of this organization")
	}

	// Update the role
	err = l.svcCtx.DB.Queries.UpdateMemberRole(l.ctx, db.UpdateMemberRoleParams{
		Role:           req.Role,
		OrganizationID: req.OrgId,
		UserID:         req.UserId,
	})
	if err != nil {
		l.Errorf("Failed to update member role: %v", err)
		return nil, err
	}

	return &types.MessageResponse{
		Message: "Member role updated successfully",
	}, nil
}
