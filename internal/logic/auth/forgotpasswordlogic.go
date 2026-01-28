package auth

import (
	"context"
	"fmt"

	"gobot/internal/local"
	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ForgotPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewForgotPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ForgotPasswordLogic {
	return &ForgotPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ForgotPasswordLogic) ForgotPassword(req *types.ForgotPasswordRequest) (resp *types.MessageResponse, err error) {
	if l.svcCtx.Auth == nil {
		return nil, fmt.Errorf("auth service not configured")
	}

	token, err := l.svcCtx.Auth.CreatePasswordResetToken(l.ctx, req.Email)
	if err != nil {
		l.Errorf("Failed to create password reset token: %v", err)
	}

	if token != "" && l.svcCtx.Email != nil {
		baseURL := l.svcCtx.Config.App.BaseURL
		resetURL := fmt.Sprintf("%s/auth/reset-password?token=%s", baseURL, token)

		_, emailErr := l.svcCtx.Email.SendEmail(l.ctx, local.SendEmailRequest{
			To:      req.Email,
			Subject: "Reset your password",
			Body: fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"></head>
<body style="font-family: sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
	<h1 style="color: #333;">Reset Your Password</h1>
	<p>You requested to reset your password. Click the button below to set a new password:</p>
	<p style="margin: 30px 0;">
		<a href="%s" style="background-color: #4F46E5; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px;">
			Reset Password
		</a>
	</p>
	<p style="color: #666; font-size: 14px;">This link will expire in 1 hour.</p>
	<p style="color: #666; font-size: 14px;">If you didn't request this, you can safely ignore this email.</p>
</body>
</html>`, resetURL),
			TextBody: fmt.Sprintf("Reset your password by visiting: %s\n\nThis link will expire in 1 hour.\n\nIf you didn't request this, you can safely ignore this email.", resetURL),
		})

		if emailErr != nil {
			l.Errorf("Failed to send password reset email: %v", emailErr)
		} else {
			l.Infof("Password reset email sent to %s", req.Email)
		}
	}

	return &types.MessageResponse{
		Message: "If an account with that email exists, a password reset link has been sent.",
	}, nil
}
