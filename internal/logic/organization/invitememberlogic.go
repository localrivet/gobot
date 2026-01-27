package organization

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"gobot/internal/auth"
	"gobot/internal/db"
	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type InviteMemberLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInviteMemberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InviteMemberLogic {
	return &InviteMemberLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InviteMemberLogic) InviteMember(req *types.InviteMemberRequest) (resp *types.InviteMemberResponse, err error) {
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

	// Check if user has permission to invite (owner or admin)
	member, err := l.svcCtx.DB.Queries.GetOrganizationMember(l.ctx, db.GetOrganizationMemberParams{
		OrganizationID: req.Id,
		UserID:         userID.String(),
	})
	if err != nil {
		l.Errorf("Failed to get member: %v", err)
		return nil, fmt.Errorf("you are not a member of this organization")
	}
	if member.Role != "owner" && member.Role != "admin" {
		return nil, fmt.Errorf("you do not have permission to invite members")
	}

	// Check if email is already a member
	existingMember, err := l.svcCtx.DB.Queries.GetInviteByEmail(l.ctx, db.GetInviteByEmailParams{
		OrganizationID: req.Id,
		Email:          req.Email,
	})
	if err == nil && existingMember.ID != "" {
		return nil, fmt.Errorf("an invite has already been sent to this email")
	}

	// Generate invite token
	token, err := generateToken(32)
	if err != nil {
		l.Errorf("Failed to generate token: %v", err)
		return nil, err
	}

	// Set default role
	role := req.Role
	if role == "" {
		role = "member"
	}

	// Set expiration (7 days from now)
	expiresAt := time.Now().Add(7 * 24 * time.Hour).Unix()

	// Create invite
	inviteID := uuid.New().String()
	invite, err := l.svcCtx.DB.Queries.CreateOrganizationInvite(l.ctx, db.CreateOrganizationInviteParams{
		ID:             inviteID,
		OrganizationID: req.Id,
		Email:          req.Email,
		Role:           role,
		Token:          token,
		InvitedBy:      userID.String(),
		ExpiresAt:      expiresAt,
	})
	if err != nil {
		l.Errorf("Failed to create invite: %v", err)
		return nil, err
	}

	// Get organization name for email
	org, _ := l.svcCtx.DB.Queries.GetOrganizationByID(l.ctx, req.Id)

	// Send invite email (if email service is configured)
	if l.svcCtx.Email != nil {
		inviteURL := fmt.Sprintf("%s/invite/%s", l.svcCtx.Config.App.BaseURL, token)
		_, err = l.svcCtx.Email.SendSimpleEmail(l.ctx,
			req.Email,
			"You've been invited to join "+org.Name,
			fmt.Sprintf(
				"<p>You've been invited to join <strong>%s</strong>.</p><p><a href=\"%s\">Click here to accept</a></p><p>This invite expires in 7 days.</p>",
				org.Name, inviteURL,
			),
			fmt.Sprintf(
				"You've been invited to join %s.\n\nClick here to accept: %s\n\nThis invite expires in 7 days.",
				org.Name, inviteURL,
			),
		)
		if err != nil {
			l.Errorf("Failed to send invite email: %v", err)
			// Continue anyway - invite is created
		}
	}

	return &types.InviteMemberResponse{
		Invite: types.OrganizationInvite{
			Id:               invite.ID,
			OrganizationName: org.Name,
			Email:            invite.Email,
			Role:             invite.Role,
			ExpiresAt:        time.Unix(invite.ExpiresAt, 0).Format(time.RFC3339),
			CreatedAt:        time.Unix(invite.CreatedAt, 0).Format(time.RFC3339),
		},
	}, nil
}

func generateToken(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
