package organization

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gobot/internal/auth"
	"gobot/internal/db"
	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateOrganizationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateOrganizationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateOrganizationLogic {
	return &UpdateOrganizationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateOrganizationLogic) UpdateOrganization(req *types.UpdateOrganizationRequest) (resp *types.GetOrganizationResponse, err error) {
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

	// Check if user is owner or admin
	member, err := l.svcCtx.DB.Queries.GetOrganizationMember(l.ctx, db.GetOrganizationMemberParams{
		OrganizationID: req.Id,
		UserID:         userID.String(),
	})
	if err != nil {
		l.Errorf("Failed to get member: %v", err)
		return nil, fmt.Errorf("you are not a member of this organization")
	}
	if member.Role != "owner" && member.Role != "admin" {
		return nil, fmt.Errorf("you do not have permission to update this organization")
	}

	// Check if new slug is unique (if changing)
	if req.Slug != "" {
		org, _ := l.svcCtx.DB.Queries.GetOrganizationByID(l.ctx, req.Id)
		if req.Slug != org.Slug {
			found, err := l.svcCtx.DB.Queries.CheckSlugExists(l.ctx, req.Slug)
			if err != nil {
				l.Errorf("Failed to check slug: %v", err)
				return nil, err
			}
			if found > 0 {
				return nil, fmt.Errorf("organization with this slug already exists")
			}
		}
	}

	// Update organization
	err = l.svcCtx.DB.Queries.UpdateOrganization(l.ctx, db.UpdateOrganizationParams{
		Name:    sql.NullString{String: req.Name, Valid: req.Name != ""},
		Slug:    sql.NullString{String: req.Slug, Valid: req.Slug != ""},
		LogoUrl: sql.NullString{String: req.LogoUrl, Valid: req.LogoUrl != ""},
		ID:      req.Id,
	})
	if err != nil {
		l.Errorf("Failed to update organization: %v", err)
		return nil, err
	}

	// Get updated organization
	org, err := l.svcCtx.DB.Queries.GetOrganizationByID(l.ctx, req.Id)
	if err != nil {
		l.Errorf("Failed to get organization: %v", err)
		return nil, err
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
