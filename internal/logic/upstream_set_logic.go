package logic

import (
	"context"

	"github.com/zeromicro/grpc-mock/internal/svc"
	"github.com/zeromicro/grpc-mock/internal/types"
)

type UpstreamSetLogic struct {
	svcCtx *svc.ServiceContext
}

func NewUpstreamSetLogic(svcCtx *svc.ServiceContext) *UpstreamSetLogic {
	return &UpstreamSetLogic{
		svcCtx: svcCtx,
	}
}

func (l *UpstreamSetLogic) UpstreamSet(ctx context.Context, req *types.UpstreamSetRequest) (resp *types.UpstreamSetResponse, err error) {
	resp = &types.UpstreamSetResponse{BaseResponse: types.BaseResponse{
		ErrorCode: 0,
		ErrorMsg:  "success",
	}}

	if len(req.Upstreams) == 0 {
		return
	}

	err = l.svcCtx.Mgr.SetUpstreams(req.Upstreams)
	if err != nil {
		resp.ErrorCode = 500
		resp.ErrorMsg = err.Error()
		return
	}

	return
}
