package agent

import (
	"context"

	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAgentStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get agent status
func NewGetAgentStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAgentStatusLogic {
	return &GetAgentStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAgentStatusLogic) GetAgentStatus(req *types.AgentStatusRequest) (resp *types.AgentStatusResponse, err error) {
	hub := l.svcCtx.AgentHub
	if hub == nil {
		return &types.AgentStatusResponse{
			AgentId:   req.AgentId,
			Connected: false,
		}, nil
	}

	// TODO: Get org ID from JWT context and look up agent
	// For now, return not connected
	return &types.AgentStatusResponse{
		AgentId:   req.AgentId,
		Connected: false,
		Uptime:    0,
	}, nil
}
