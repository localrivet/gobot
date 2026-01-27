package user

import (
	"context"
	"fmt"
	"time"

	"gobot/internal/auth"
	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCurrentUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCurrentUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCurrentUserLogic {
	return &GetCurrentUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCurrentUserLogic) GetCurrentUser() (resp *types.GetUserResponse, err error) {
	// Get email from JWT context
	email, err := auth.GetEmailFromContext(l.ctx)
	if err != nil {
		l.Errorf("Failed to get email from context: %v", err)
		return nil, err
	}

	// Handle admin user (not in database)
	if email == l.svcCtx.Config.Admin.Username {
		return &types.GetUserResponse{
			User: types.User{
				Id:            "admin",
				Email:         email,
				Name:          "Admin",
				EmailVerified: true,
				CreatedAt:     time.Now().Format("2006-01-02T15:04:05Z"),
				UpdatedAt:     time.Now().Format("2006-01-02T15:04:05Z"),
			},
		}, nil
	}

	// Use local auth when Levee is disabled
	if l.svcCtx.UseLocal() {
		return l.getCurrentUserLocal(email)
	}

	// Use Levee when enabled
	if l.svcCtx.Levee == nil {
		return nil, fmt.Errorf("auth service not configured")
	}

	// Get customer from Levee SDK
	customer, err := l.svcCtx.Levee.Customers.GetCustomerByEmail(l.ctx, email)
	if err != nil {
		l.Errorf("Failed to get customer %s: %v", email, err)
		return nil, err
	}

	return &types.GetUserResponse{
		User: types.User{
			Id:            customer.ID,
			Email:         customer.Email,
			Name:          customer.Name,
			EmailVerified: customer.Status == "active",
			CreatedAt:     customer.CreatedAt,
			UpdatedAt:     customer.CreatedAt, // SDK doesn't track UpdatedAt
		},
	}, nil
}

// getCurrentUserLocal gets user from local SQLite
func (l *GetCurrentUserLogic) getCurrentUserLocal(email string) (*types.GetUserResponse, error) {
	if l.svcCtx.Auth == nil {
		return nil, fmt.Errorf("local auth service not configured")
	}

	user, err := l.svcCtx.Auth.GetUserByEmail(l.ctx, email)
	if err != nil {
		l.Errorf("Failed to get user %s: %v", email, err)
		return nil, err
	}

	return &types.GetUserResponse{
		User: types.User{
			Id:            user.ID,
			Email:         user.Email,
			Name:          user.Name,
			EmailVerified: user.EmailVerified == 1,
			CreatedAt:     time.Unix(user.CreatedAt, 0).Format("2006-01-02T15:04:05Z"),
			UpdatedAt:     time.Unix(user.UpdatedAt, 0).Format("2006-01-02T15:04:05Z"),
		},
	}, nil
}
