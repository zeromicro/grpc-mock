package logic

import (
	"context"

	"github.com/zeromicro/grpc-mock/internal/svc"
	"github.com/zeromicro/grpc-mock/internal/types"
)

type CaseDetailLogic struct {
	svcCtx *svc.ServiceContext
}

func NewCaseDetailLogic(svcCtx *svc.ServiceContext) *CaseDetailLogic {
	return &CaseDetailLogic{
		svcCtx: svcCtx,
	}
}

func (l *CaseDetailLogic) CaseDetail(ctx context.Context, req *types.CaseDetailRequest) (resp *types.CaseDetailResponse, err error) {
	resp = &types.CaseDetailResponse{BaseResponse: types.BaseResponse{
		ErrorCode: 0,
		ErrorMsg:  "success",
	}}

	cs, err := l.svcCtx.Mgr.CasesDetail(ctx, req.TestedAppName, req.MethodName, req.Name)
	if err != nil {
		resp.ErrorCode = 500
		resp.ErrorMsg = err.Error()
		return
	}

	resp.Detail = cs
	return
}
