package logic

import (
	"context"

	"github.com/zeromicro/grpc-mock/internal/svc"
	"github.com/zeromicro/grpc-mock/internal/types"
)

type UpstreamDelLogic struct {
	svcCtx *svc.ServiceContext
}

func NewUpstreamDelLogic(svcCtx *svc.ServiceContext) *UpstreamDelLogic {
	return &UpstreamDelLogic{
		svcCtx: svcCtx,
	}
}

func (l *UpstreamDelLogic) UpstreamDel(ctx context.Context, req *types.UpstreamDelRequest) (resp *types.UpstreamDelResponse, err error) {
	resp = &types.UpstreamDelResponse{BaseResponse: types.BaseResponse{
		ErrorCode: 0,
		ErrorMsg:  "success",
	}}

	if len(req.Upstreams) == 0 {
		return
	}

	err = l.svcCtx.Mgr.DelUpstreams(req.Upstreams)
	if err != nil {
		resp.ErrorCode = 500
		resp.ErrorMsg = err.Error()
		return
	}

	return
}
