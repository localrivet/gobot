package user

import (
	"context"

	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPreferencesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPreferencesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPreferencesLogic {
	return &GetPreferencesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPreferencesLogic) GetPreferences() (resp *types.GetPreferencesResponse, err error) {
	// Preferences are stored locally - for the boilerplate, return defaults
	// In a real app, you'd store these in your database or in Levee custom fields
	return &types.GetPreferencesResponse{
		Preferences: types.UserPreferences{
			EmailNotifications: true,
			MarketingEmails:    true,
			Timezone:           "UTC",
			Language:           "en",
			Theme:              "system",
		},
	}, nil
}
