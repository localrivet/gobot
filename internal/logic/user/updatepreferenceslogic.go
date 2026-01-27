package user

import (
	"context"

	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdatePreferencesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdatePreferencesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePreferencesLogic {
	return &UpdatePreferencesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdatePreferencesLogic) UpdatePreferences(req *types.UpdatePreferencesRequest) (resp *types.GetPreferencesResponse, err error) {
	// Preferences would be stored in your database or Levee custom fields
	// For the boilerplate, return the updated values
	return &types.GetPreferencesResponse{
		Preferences: types.UserPreferences{
			EmailNotifications: req.EmailNotifications,
			MarketingEmails:    req.MarketingEmails,
			Timezone:           req.Timezone,
			Language:           req.Language,
			Theme:              req.Theme,
		},
	}, nil
}
