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

type LeaveOrganizationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLeaveOrganizationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LeaveOrganizationLogic {
	return &LeaveOrganizationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LeaveOrganizationLogic) LeaveOrganization(req *types.LeaveOrganizationRequest) (resp *types.MessageResponse, err error) {
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

	// Owner cannot leave - they must delete the org or transfer ownership
	if org.OwnerID == userID.String() {
		return nil, fmt.Errorf("owner cannot leave the organization; delete it or transfer ownership instead")
	}

	// Verify user is a member
	_, err = l.svcCtx.DB.Queries.GetOrganizationMember(l.ctx, db.GetOrganizationMemberParams{
		OrganizationID: req.Id,
		UserID:         userID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("you are not a member of this organization")
	}

	// Remove the member
	err = l.svcCtx.DB.Queries.RemoveOrganizationMember(l.ctx, db.RemoveOrganizationMemberParams{
		OrganizationID: req.Id,
		UserID:         userID.String(),
	})
	if err != nil {
		l.Errorf("Failed to leave organization: %v", err)
		return nil, err
	}

	// Clear current organization if this was it
	_ = l.svcCtx.DB.Queries.SetCurrentOrganization(l.ctx, db.SetCurrentOrganizationParams{
		OrganizationID: sql.NullString{Valid: false},
		UserID:         userID.String(),
	})

	return &types.MessageResponse{
		Message: "Successfully left the organization",
	}, nil
}
