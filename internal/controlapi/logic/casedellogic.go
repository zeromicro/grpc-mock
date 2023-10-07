package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zeromicro/grpc-mock/internal/controlapi/types"
	"github.com/zeromicro/grpc-mock/internal/svc"
)

type CaseDelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCaseDelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CaseDelLogic {
	return &CaseDelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CaseDelLogic) CaseDel(req *types.CaseDelRequest) (resp *types.CaseDelRespnse, err error) {
	if err = l.svcCtx.CaseManager.CaseDel(l.ctx, req.MethodName, req.Name); err != nil {
		return nil, err
	}

	return &types.CaseDelRespnse{}, nil
}
