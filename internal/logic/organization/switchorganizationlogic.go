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

type SwitchOrganizationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSwitchOrganizationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SwitchOrganizationLogic {
	return &SwitchOrganizationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SwitchOrganizationLogic) SwitchOrganization(req *types.SwitchOrganizationRequest) (resp *types.MessageResponse, err error) {
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

	// Verify user is a member of the target organization
	_, err = l.svcCtx.DB.Queries.GetOrganizationMember(l.ctx, db.GetOrganizationMemberParams{
		OrganizationID: req.OrganizationId,
		UserID:         userID.String(),
	})
	if err != nil {
		l.Errorf("Failed to get member: %v", err)
		return nil, fmt.Errorf("you are not a member of this organization")
	}

	// Set as current organization
	err = l.svcCtx.DB.Queries.SetCurrentOrganization(l.ctx, db.SetCurrentOrganizationParams{
		OrganizationID: sql.NullString{String: req.OrganizationId, Valid: true},
		UserID:         userID.String(),
	})
	if err != nil {
		l.Errorf("Failed to switch organization: %v", err)
		return nil, err
	}

	return &types.MessageResponse{
		Message: "Switched to organization",
	}, nil
}
