package logic

import (
	"context"

	"github.com/zeromicro/grpc-mock/internal/svc"
	"github.com/zeromicro/grpc-mock/internal/types"
)

type CaseListLogic struct {
	svcCtx *svc.ServiceContext
}

func NewCaseListLogic(svcCtx *svc.ServiceContext) *CaseListLogic {
	return &CaseListLogic{
		svcCtx: svcCtx,
	}
}

func (l *CaseListLogic) CaseList(ctx context.Context, req *types.CaseListRequest) (resp *types.CaseListResponse, err error) {
	resp = &types.CaseListResponse{BaseResponse: types.BaseResponse{
		ErrorCode: 0,
		ErrorMsg:  "success",
	}}

	cases, err := l.svcCtx.Mgr.ListCases(ctx, req.TestedAppName)
	if err != nil {
		resp.ErrorCode = 500
		resp.ErrorMsg = err.Error()
		return
	}

	resp.Cases = cases
	return
}
