package logic

import (
	"context"

	"github.com/zeromicro/grpc-mock/internal/svc"
	"github.com/zeromicro/grpc-mock/internal/types"
)

type AppDelLogic struct {
	svcCtx *svc.ServiceContext
}

func NewAppDelLogic(svcCtx *svc.ServiceContext) *AppDelLogic {
	return &AppDelLogic{
		svcCtx: svcCtx,
	}
}

func (l *AppDelLogic) AppDel(ctx context.Context, req *types.AppDelRequest) (resp *types.AppDelResponse, err error) {
	resp = &types.AppDelResponse{BaseResponse: types.BaseResponse{
		ErrorCode: 0,
		ErrorMsg:  "success",
	}}

	if len(req.Apps) == 0 {
		return
	}

	err = l.svcCtx.Mgr.DelApp(req.Apps)
	if err != nil {
		resp.ErrorCode = 500
		resp.ErrorMsg = err.Error()
		return
	}

	return
}
