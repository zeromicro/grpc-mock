package logic

import (
	"context"

	"github.com/zeromicro/grpc-mock/internal/svc"
	"github.com/zeromicro/grpc-mock/internal/types"
)

type CaseDelLogic struct {
	svcCtx *svc.ServiceContext
}

func NewCaseDelLogic(svcCtx *svc.ServiceContext) *CaseDelLogic {
	return &CaseDelLogic{
		svcCtx: svcCtx,
	}
}

func (l *CaseDelLogic) CaseDel(ctx context.Context, req *types.CaseDelRequest) (resp *types.CaseDelRespnse, err error) {
	var cases = []types.Case{{
		TestedAppName: req.TestedAppName,
		MethodName:    req.MethodName,
		Name:          req.Name,
	}}

	resp = &types.CaseDelRespnse{BaseResponse: types.BaseResponse{
		ErrorCode: 0,
		ErrorMsg:  "success",
	}}

	err = l.svcCtx.Mgr.DelCases(ctx, cases)
	if err != nil {
		resp.ErrorCode = 500
		resp.ErrorMsg = err.Error()
		return
	}
	return
}
