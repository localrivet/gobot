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

type GetOrganizationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOrganizationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrganizationLogic {
	return &GetOrganizationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrganizationLogic) GetOrganization(req *types.GetOrganizationRequest) (resp *types.GetOrganizationResponse, err error) {
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

	// Get organization
	org, err := l.svcCtx.DB.Queries.GetOrganizationByID(l.ctx, req.Id)
	if err != nil {
		l.Errorf("Failed to get organization: %v", err)
		return nil, fmt.Errorf("organization not found")
	}

	// Check if user is a member
	member, err := l.svcCtx.DB.Queries.GetOrganizationMember(l.ctx, db.GetOrganizationMemberParams{
		OrganizationID: req.Id,
		UserID:         userID.String(),
	})
	if err != nil {
		l.Errorf("Failed to get member: %v", err)
		return nil, fmt.Errorf("you are not a member of this organization")
	}

	return &types.GetOrganizationResponse{
		Organization: types.Organization{
			Id:        org.ID,
			Name:      org.Name,
			Slug:      org.Slug,
			LogoUrl:   org.LogoUrl.String,
			OwnerId:   org.OwnerID,
			CreatedAt: time.Unix(org.CreatedAt, 0).Format(time.RFC3339),
			UpdatedAt: time.Unix(org.UpdatedAt, 0).Format(time.RFC3339),
		},
		Role: member.Role,
	}, nil
}
