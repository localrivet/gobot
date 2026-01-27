package organization

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"gobot/internal/auth"
	"gobot/internal/db"
	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateOrganizationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateOrganizationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrganizationLogic {
	return &CreateOrganizationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateOrganizationLogic) CreateOrganization(req *types.CreateOrganizationRequest) (resp *types.CreateOrganizationResponse, err error) {
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

	// Generate slug from name if not provided
	slug := req.Slug
	if slug == "" {
		slug = generateSlug(req.Name)
	}

	// Check if slug already exists
	found, err := l.svcCtx.DB.Queries.CheckSlugExists(l.ctx, slug)
	if err != nil {
		l.Errorf("Failed to check slug: %v", err)
		return nil, err
	}
	if found > 0 {
		return nil, fmt.Errorf("organization with this slug already exists")
	}

	// Create organization
	orgID := uuid.New().String()
	org, err := l.svcCtx.DB.Queries.CreateOrganization(l.ctx, db.CreateOrganizationParams{
		ID:      orgID,
		Name:    req.Name,
		Slug:    slug,
		LogoUrl: sql.NullString{String: req.LogoUrl, Valid: req.LogoUrl != ""},
		OwnerID: userID.String(),
	})
	if err != nil {
		l.Errorf("Failed to create organization: %v", err)
		return nil, err
	}

	// Add owner as member with owner role
	memberID := uuid.New().String()
	_, err = l.svcCtx.DB.Queries.AddOrganizationMember(l.ctx, db.AddOrganizationMemberParams{
		ID:             memberID,
		OrganizationID: orgID,
		UserID:         userID.String(),
		Role:           "owner",
	})
	if err != nil {
		l.Errorf("Failed to add owner as member: %v", err)
		return nil, err
	}

	// Set as current organization for the user
	err = l.svcCtx.DB.Queries.SetCurrentOrganization(l.ctx, db.SetCurrentOrganizationParams{
		OrganizationID: sql.NullString{String: orgID, Valid: true},
		UserID:         userID.String(),
	})
	if err != nil {
		l.Errorf("Failed to set current organization: %v", err)
		// Non-fatal, continue
	}

	return &types.CreateOrganizationResponse{
		Organization: types.Organization{
			Id:        org.ID,
			Name:      org.Name,
			Slug:      org.Slug,
			LogoUrl:   org.LogoUrl.String,
			OwnerId:   org.OwnerID,
			CreatedAt: time.Unix(org.CreatedAt, 0).Format(time.RFC3339),
			UpdatedAt: time.Unix(org.UpdatedAt, 0).Format(time.RFC3339),
		},
	}, nil
}

func generateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)
	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove non-alphanumeric characters except hyphens
	reg := regexp.MustCompile("[^a-z0-9-]+")
	slug = reg.ReplaceAllString(slug, "")
	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile("-+")
	slug = reg.ReplaceAllString(slug, "-")
	// Trim hyphens from ends
	slug = strings.Trim(slug, "-")
	return slug
}
