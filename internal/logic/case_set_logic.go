package logic

import (
	"context"

	"github.com/zeromicro/grpc-mock/internal/svc"
	"github.com/zeromicro/grpc-mock/internal/types"
)

type CaseSetLogic struct {
	svcCtx *svc.ServiceContext
}

func NewCaseSetLogic(svcCtx *svc.ServiceContext) *CaseSetLogic {
	return &CaseSetLogic{
		svcCtx: svcCtx,
	}
}

func (l *CaseSetLogic) CaseSet(ctx context.Context, req *types.CaseSetRequest) (resp *types.CaseSetResponse, err error) {
	resp = &types.CaseSetResponse{BaseResponse: types.BaseResponse{
		ErrorCode: 0,
		ErrorMsg:  "success",
	}}

	err = l.svcCtx.Mgr.SetCases(ctx, req.Cases)
	if err != nil {
		resp.ErrorCode = 500
		resp.ErrorMsg = err.Error()
		return
	}

	return
}
