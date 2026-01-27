package organization

import (
	"context"
	"fmt"
	"time"

	"gobot/internal/auth"
	"gobot/internal/db"
	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListInvitesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListInvitesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListInvitesLogic {
	return &ListInvitesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListInvitesLogic) ListInvites(req *types.ListInvitesRequest) (resp *types.ListInvitesResponse, err error) {
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

	// Check if user has permission to view invites (owner or admin)
	member, err := l.svcCtx.DB.Queries.GetOrganizationMember(l.ctx, db.GetOrganizationMemberParams{
		OrganizationID: req.Id,
		UserID:         userID.String(),
	})
	if err != nil {
		l.Errorf("Failed to get member: %v", err)
		return nil, fmt.Errorf("you are not a member of this organization")
	}
	if member.Role != "owner" && member.Role != "admin" {
		return nil, fmt.Errorf("you do not have permission to view invites")
	}

	// List pending invites
	invites, err := l.svcCtx.DB.Queries.ListOrganizationInvites(l.ctx, req.Id)
	if err != nil {
		l.Errorf("Failed to list invites: %v", err)
		return nil, err
	}

	// Convert to response type
	result := make([]types.OrganizationInvite, len(invites))
	for i, inv := range invites {
		result[i] = types.OrganizationInvite{
			Id:           inv.ID,
			Email:        inv.Email,
			Role:         inv.Role,
			InviterName:  inv.InviterName,
			InviterEmail: inv.InviterEmail,
			ExpiresAt:    time.Unix(inv.ExpiresAt, 0).Format(time.RFC3339),
			CreatedAt:    time.Unix(inv.CreatedAt, 0).Format(time.RFC3339),
		}
	}

	return &types.ListInvitesResponse{
		Invites: result,
	}, nil
}
