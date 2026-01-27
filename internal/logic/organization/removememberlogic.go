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

type RemoveMemberLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRemoveMemberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RemoveMemberLogic {
	return &RemoveMemberLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RemoveMemberLogic) RemoveMember(req *types.RemoveMemberRequest) (resp *types.MessageResponse, err error) {
	if !l.svcCtx.Config.IsOrganizationsEnabled() {
		return nil, fmt.Errorf("organizations feature is not enabled")
	}

	if !l.svcCtx.UseLocal() {
		return nil, fmt.Errorf("organizations not available in this mode")
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

	// Cannot remove the owner
	if req.UserId == org.OwnerID {
		return nil, fmt.Errorf("cannot remove the organization owner")
	}

	// Check if user has permission to remove members (owner or admin)
	currentMember, err := l.svcCtx.DB.Queries.GetOrganizationMember(l.ctx, db.GetOrganizationMemberParams{
		OrganizationID: req.OrgId,
		UserID:         userID.String(),
	})
	if err != nil {
		l.Errorf("Failed to get member: %v", err)
		return nil, fmt.Errorf("you are not a member of this organization")
	}
	if currentMember.Role != "owner" && currentMember.Role != "admin" {
		return nil, fmt.Errorf("you do not have permission to remove members")
	}

	// Admins cannot remove other admins or the owner
	if currentMember.Role == "admin" {
		targetMember, err := l.svcCtx.DB.Queries.GetOrganizationMember(l.ctx, db.GetOrganizationMemberParams{
			OrganizationID: req.OrgId,
			UserID:         req.UserId,
		})
		if err != nil {
			return nil, fmt.Errorf("user is not a member of this organization")
		}
		if targetMember.Role == "owner" || targetMember.Role == "admin" {
			return nil, fmt.Errorf("admins cannot remove other admins or the owner")
		}
	}

	// Remove the member
	err = l.svcCtx.DB.Queries.RemoveOrganizationMember(l.ctx, db.RemoveOrganizationMemberParams{
		OrganizationID: req.OrgId,
		UserID:         req.UserId,
	})
	if err != nil {
		l.Errorf("Failed to remove member: %v", err)
		return nil, err
	}

	return &types.MessageResponse{
		Message: "Member removed successfully",
	}, nil
}
