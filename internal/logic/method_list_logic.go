package logic

import (
	"context"

	"github.com/zeromicro/grpc-mock/internal/svc"
	"github.com/zeromicro/grpc-mock/internal/types"
)

type MethodListLogic struct {
	svcCtx *svc.ServiceContext
}

func NewMethodListLogic(svcCtx *svc.ServiceContext) *MethodListLogic {
	return &MethodListLogic{
		svcCtx: svcCtx,
	}
}

func (l *MethodListLogic) MethodList(ctx context.Context, req *types.MethodListRequest) (resp *types.MethodListResponse, err error) {
	resp = &types.MethodListResponse{BaseResponse: types.BaseResponse{
		ErrorCode: 0,
		ErrorMsg:  "success",
	}}

	for _, s := range l.svcCtx.Mgr.GetServiceDesc(ctx, req.Upstream) {
		for _, m := range s.Methods {
			resp.List = append(resp.List, types.Method{
				Service: s.FullName,
				Name:    m.FullName,
			})
		}
	}

	return resp, nil
}
