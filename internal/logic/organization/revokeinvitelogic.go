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

type RevokeInviteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRevokeInviteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RevokeInviteLogic {
	return &RevokeInviteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RevokeInviteLogic) RevokeInvite(req *types.RevokeInviteRequest) (resp *types.MessageResponse, err error) {
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

	// Check if user has permission to revoke invites (owner or admin)
	member, err := l.svcCtx.DB.Queries.GetOrganizationMember(l.ctx, db.GetOrganizationMemberParams{
		OrganizationID: req.OrgId,
		UserID:         userID.String(),
	})
	if err != nil {
		l.Errorf("Failed to get member: %v", err)
		return nil, fmt.Errorf("you are not a member of this organization")
	}
	if member.Role != "owner" && member.Role != "admin" {
		return nil, fmt.Errorf("you do not have permission to revoke invites")
	}

	// Delete the invite
	err = l.svcCtx.DB.Queries.DeleteInvite(l.ctx, req.InviteId)
	if err != nil {
		l.Errorf("Failed to revoke invite: %v", err)
		return nil, err
	}

	return &types.MessageResponse{
		Message: "Invite revoked successfully",
	}, nil
}
