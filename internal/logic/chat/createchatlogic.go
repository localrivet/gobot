package chat

import (
	"context"
	"database/sql"
	"time"

	"gobot/internal/db"
	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateChatLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Create new chat - Single Bot Paradigm: returns the companion chat instead of creating new ones
func NewCreateChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateChatLogic {
	return &CreateChatLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateChatLogic) CreateChat(req *types.CreateChatRequest) (resp *types.CreateChatResponse, err error) {
	// Single Bot Paradigm: Always return the companion chat
	// We don't create new chats - there is only ONE conversation with THE agent
	userID := companionUserID

	chat, err := l.svcCtx.DB.GetOrCreateCompanionChat(l.ctx, db.GetOrCreateCompanionChatParams{
		ID:     uuid.New().String(),
		UserID: sql.NullString{String: userID, Valid: true},
	})
	if err != nil {
		l.Errorf("Failed to get companion chat: %v", err)
		return nil, err
	}

	return &types.CreateChatResponse{
		Chat: types.Chat{
			Id:        chat.ID,
			Title:     chat.Title,
			CreatedAt: time.Unix(chat.CreatedAt, 0).Format(time.RFC3339),
			UpdatedAt: time.Unix(chat.UpdatedAt, 0).Format(time.RFC3339),
		},
	}, nil
}
