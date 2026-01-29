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

type ListChatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// List chats - Single Bot Paradigm: returns only the companion chat
func NewListChatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListChatsLogic {
	return &ListChatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListChatsLogic) ListChats(req *types.ListChatsRequest) (resp *types.ListChatsResponse, err error) {
	// Single Bot Paradigm: Return only the companion chat
	// There is only ONE conversation with THE agent
	userID := companionUserID

	chat, err := l.svcCtx.DB.GetOrCreateCompanionChat(l.ctx, db.GetOrCreateCompanionChatParams{
		ID:     uuid.New().String(),
		UserID: sql.NullString{String: userID, Valid: true},
	})
	if err != nil {
		l.Errorf("Failed to get companion chat: %v", err)
		return nil, err
	}

	return &types.ListChatsResponse{
		Chats: []types.Chat{
			{
				Id:        chat.ID,
				Title:     chat.Title,
				CreatedAt: time.Unix(chat.CreatedAt, 0).Format(time.RFC3339),
				UpdatedAt: time.Unix(chat.UpdatedAt, 0).Format(time.RFC3339),
			},
		},
		Total: 1,
	}, nil
}
