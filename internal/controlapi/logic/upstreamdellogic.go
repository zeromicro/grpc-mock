package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zeromicro/grpc-mock/internal/controlapi/types"
	"github.com/zeromicro/grpc-mock/internal/svc"
)

type UpstreamDelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpstreamDelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpstreamDelLogic {
	return &UpstreamDelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpstreamDelLogic) UpstreamDel(req *types.UpstreamDelRequest) (resp *types.UpstreamDelResponse, err error) {
	for _, name := range req.Upstreams {
		err := l.svcCtx.DialManager.DelUpstream(l.ctx, name)
		if err != nil {
			return nil, err
		}
	}

	return &types.UpstreamDelResponse{}, nil
}
