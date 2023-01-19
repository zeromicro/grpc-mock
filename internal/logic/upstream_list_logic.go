package logic

import (
	"context"

	"github.com/zeromicro/grpc-mock/internal/svc"
	"github.com/zeromicro/grpc-mock/internal/types"
)

type UpstreamListLogic struct {
	svcCtx *svc.ServiceContext
}

func NewUpstreamListLogic(svcCtx *svc.ServiceContext) *UpstreamListLogic {
	return &UpstreamListLogic{
		svcCtx: svcCtx,
	}
}

func (l *UpstreamListLogic) UpstreamList(ctx context.Context) (resp *types.UpstreamListResponse, err error) {
	resp = &types.UpstreamListResponse{BaseResponse: types.BaseResponse{
		ErrorCode: 0,
		ErrorMsg:  "success",
	}}

	upstreams, err := l.svcCtx.Mgr.GetUpstreams()
	if err != nil {
		resp.ErrorCode = 500
		resp.ErrorMsg = err.Error()
		return
	}

	var endpoints []string
	for _, upstream := range upstreams {
		endpoints = append(endpoints, upstream.Endpoints...)
	}

	resp.Upstreams = endpoints

	return
}
