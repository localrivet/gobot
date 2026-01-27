package user

import (
	"context"
	"errors"
	"fmt"

	"gobot/internal/auth"
	"gobot/internal/svc"
	"gobot/internal/types"

	levee "github.com/almatuck/levee-go"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteAccountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAccountLogic {
	return &DeleteAccountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteAccountLogic) DeleteAccount(req *types.DeleteAccountRequest) (resp *types.MessageResponse, err error) {
	if l.svcCtx.Levee == nil {
		return nil, fmt.Errorf("levee service not configured")
	}

	// Get email from JWT context
	email, err := auth.GetEmailFromContext(l.ctx)
	if err != nil {
		l.Errorf("Failed to get email from context: %v", err)
		return nil, err
	}

	// Verify password by attempting login via Levee SDK
	_, err = l.svcCtx.Levee.Auth.Login(l.ctx, &levee.SDKLoginRequest{
		Email:    email,
		Password: req.Password,
	})
	if err != nil {
		l.Errorf("Password verification failed for delete account: %v", err)
		return nil, errors.New("invalid password")
	}

	// Get customer ID
	customer, err := l.svcCtx.Levee.Customers.GetCustomerByEmail(l.ctx, email)
	if err != nil {
		l.Errorf("Failed to get customer %s: %v", email, err)
		return nil, err
	}

	// Delete customer via Levee SDK
	_, err = l.svcCtx.Levee.Customers.DeleteCustomer(l.ctx, customer.ID)
	if err != nil {
		l.Errorf("Failed to delete customer %s: %v", email, err)
		return nil, err
	}

	return &types.MessageResponse{
		Message: "Account deleted successfully.",
	}, nil
}
