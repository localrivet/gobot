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

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type AcceptInviteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAcceptInviteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AcceptInviteLogic {
	return &AcceptInviteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AcceptInviteLogic) AcceptInvite(req *types.AcceptInviteRequest) (resp *types.AcceptInviteResponse, err error) {
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

	// Get invite by token
	invite, err := l.svcCtx.DB.Queries.GetInviteByToken(l.ctx, req.Token)
	if err != nil {
		l.Errorf("Failed to get invite: %v", err)
		return nil, fmt.Errorf("invalid or expired invite")
	}

	// Check if already a member
	_, err = l.svcCtx.DB.Queries.GetOrganizationMember(l.ctx, db.GetOrganizationMemberParams{
		OrganizationID: invite.OrganizationID,
		UserID:         userID.String(),
	})
	if err == nil {
		// Already a member, delete invite and return success
		_ = l.svcCtx.DB.Queries.DeleteInviteByToken(l.ctx, req.Token)
		org, _ := l.svcCtx.DB.Queries.GetOrganizationByID(l.ctx, invite.OrganizationID)
		return &types.AcceptInviteResponse{
			Organization: types.Organization{
				Id:        org.ID,
				Name:      org.Name,
				Slug:      org.Slug,
				LogoUrl:   org.LogoUrl.String,
				OwnerId:   org.OwnerID,
				CreatedAt: time.Unix(org.CreatedAt, 0).Format(time.RFC3339),
				UpdatedAt: time.Unix(org.UpdatedAt, 0).Format(time.RFC3339),
			},
			Message: "You are already a member of this organization",
		}, nil
	}

	// Add user as member
	memberID := uuid.New().String()
	_, err = l.svcCtx.DB.Queries.AddOrganizationMember(l.ctx, db.AddOrganizationMemberParams{
		ID:             memberID,
		OrganizationID: invite.OrganizationID,
		UserID:         userID.String(),
		Role:           invite.Role,
	})
	if err != nil {
		l.Errorf("Failed to add member: %v", err)
		return nil, err
	}

	// Delete the invite
	err = l.svcCtx.DB.Queries.DeleteInviteByToken(l.ctx, req.Token)
	if err != nil {
		l.Errorf("Failed to delete invite: %v", err)
		// Non-fatal
	}

	// Set as current organization
	err = l.svcCtx.DB.Queries.SetCurrentOrganization(l.ctx, db.SetCurrentOrganizationParams{
		OrganizationID: sql.NullString{String: invite.OrganizationID, Valid: true},
		UserID:         userID.String(),
	})
	if err != nil {
		l.Errorf("Failed to set current organization: %v", err)
		// Non-fatal
	}

	// Get organization details
	org, err := l.svcCtx.DB.Queries.GetOrganizationByID(l.ctx, invite.OrganizationID)
	if err != nil {
		l.Errorf("Failed to get organization: %v", err)
		return nil, err
	}

	return &types.AcceptInviteResponse{
		Organization: types.Organization{
			Id:        org.ID,
			Name:      org.Name,
			Slug:      org.Slug,
			LogoUrl:   org.LogoUrl.String,
			OwnerId:   org.OwnerID,
			CreatedAt: time.Unix(org.CreatedAt, 0).Format(time.RFC3339),
			UpdatedAt: time.Unix(org.UpdatedAt, 0).Format(time.RFC3339),
		},
		Message: "Successfully joined " + org.Name,
	}, nil
}
