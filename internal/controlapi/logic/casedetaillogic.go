package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zeromicro/grpc-mock/internal/controlapi/types"
	"github.com/zeromicro/grpc-mock/internal/svc"
)

type CaseDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCaseDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CaseDetailLogic {
	return &CaseDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CaseDetailLogic) CaseDetail(req *types.CaseDetailRequest) (resp *types.CaseDetailResponse, err error) {
	_case, err := l.svcCtx.CaseManager.CaseGet(l.ctx, req.MethodName, req.Name)
	if err != nil {
		return nil, err
	}

	return &types.CaseDetailResponse{
		Detail: _case,
	}, nil
}
