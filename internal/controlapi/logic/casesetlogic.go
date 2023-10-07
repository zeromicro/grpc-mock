package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zeromicro/grpc-mock/internal/controlapi/types"
	"github.com/zeromicro/grpc-mock/internal/svc"
)

type CaseSetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCaseSetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CaseSetLogic {
	return &CaseSetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CaseSetLogic) CaseSet(req *types.CaseSetRequest) (resp *types.CaseSetResponse, err error) {
	for _, _case := range req.Cases {
		if err = l.svcCtx.CaseManager.CaseAdd(l.ctx, _case); err != nil {
			return nil, err
		}
	}

	return &types.CaseSetResponse{}, nil
}
