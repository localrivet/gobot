package organization

import (
	"context"
	"time"

	"gobot/internal/auth"
	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListOrganizationsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListOrganizationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListOrganizationsLogic {
	return &ListOrganizationsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListOrganizationsLogic) ListOrganizations() (resp *types.ListOrganizationsResponse, err error) {
	if !l.svcCtx.Config.IsOrganizationsEnabled() {
		return &types.ListOrganizationsResponse{Organizations: []types.Organization{}}, nil
	}

	if !l.svcCtx.UseLocal() {
		return &types.ListOrganizationsResponse{Organizations: []types.Organization{}}, nil
	}

	// Get user ID from context
	userID, err := auth.GetUserIDFromContext(l.ctx)
	if err != nil {
		l.Errorf("Failed to get user ID: %v", err)
		return nil, err
	}

	// List organizations for user
	orgs, err := l.svcCtx.DB.Queries.ListUserOrganizations(l.ctx, userID.String())
	if err != nil {
		l.Errorf("Failed to list organizations: %v", err)
		return nil, err
	}

	// Get current organization
	currentOrgID, _ := l.svcCtx.DB.Queries.GetCurrentOrganization(l.ctx, userID.String())

	// Convert to response type
	result := make([]types.Organization, len(orgs))
	for i, org := range orgs {
		result[i] = types.Organization{
			Id:        org.ID,
			Name:      org.Name,
			Slug:      org.Slug,
			LogoUrl:   org.LogoUrl.String,
			OwnerId:   org.OwnerID,
			CreatedAt: time.Unix(org.CreatedAt, 0).Format(time.RFC3339),
			UpdatedAt: time.Unix(org.UpdatedAt, 0).Format(time.RFC3339),
		}
	}

	// Note: currentOrgID is available but not returned in the response type
	_ = currentOrgID

	return &types.ListOrganizationsResponse{
		Organizations: result,
	}, nil
}
