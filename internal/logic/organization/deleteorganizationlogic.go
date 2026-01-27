package organization

import (
	"context"
	"database/sql"
	"fmt"

	"gobot/internal/auth"
	"gobot/internal/db"
	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteOrganizationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteOrganizationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteOrganizationLogic {
	return &DeleteOrganizationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteOrganizationLogic) DeleteOrganization(req *types.DeleteOrganizationRequest) (resp *types.MessageResponse, err error) {
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
	org, err := l.svcCtx.DB.Queries.GetOrganizationByID(l.ctx, req.Id)
	if err != nil {
		l.Errorf("Failed to get organization: %v", err)
		return nil, fmt.Errorf("organization not found")
	}

	// Only the owner can delete an organization
	if org.OwnerID != userID.String() {
		return nil, fmt.Errorf("only the owner can delete this organization")
	}

	// Delete the organization (cascades to members and invites via foreign keys)
	err = l.svcCtx.DB.Queries.DeleteOrganization(l.ctx, req.Id)
	if err != nil {
		l.Errorf("Failed to delete organization: %v", err)
		return nil, err
	}

	// Clear current organization for the user if this was their current org
	_ = l.svcCtx.DB.Queries.SetCurrentOrganization(l.ctx, db.SetCurrentOrganizationParams{
		OrganizationID: sql.NullString{Valid: false},
		UserID:         userID.String(),
	})

	return &types.MessageResponse{
		Message: "Organization deleted successfully",
	}, nil
}
