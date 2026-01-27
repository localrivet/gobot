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

type ListMembersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMembersLogic {
	return &ListMembersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMembersLogic) ListMembers(req *types.ListMembersRequest) (resp *types.ListMembersResponse, err error) {
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

	// Check if user is a member of the organization
	_, err = l.svcCtx.DB.Queries.GetOrganizationMember(l.ctx, db.GetOrganizationMemberParams{
		OrganizationID: req.Id,
		UserID:         userID.String(),
	})
	if err != nil {
		l.Errorf("Failed to get member: %v", err)
		return nil, fmt.Errorf("you are not a member of this organization")
	}

	// List all members
	members, err := l.svcCtx.DB.Queries.ListOrganizationMembers(l.ctx, req.Id)
	if err != nil {
		l.Errorf("Failed to list members: %v", err)
		return nil, err
	}

	// Convert to response type
	result := make([]types.OrganizationMember, len(members))
	for i, m := range members {
		result[i] = types.OrganizationMember{
			Id:        m.ID,
			UserId:    m.UserID,
			Email:     m.Email,
			Name:      m.UserName,
			AvatarUrl: m.AvatarUrl.String,
			Role:      m.Role,
			JoinedAt:  time.Unix(m.JoinedAt, 0).Format(time.RFC3339),
		}
	}

	return &types.ListMembersResponse{
		Members: result,
	}, nil
}
