package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zeromicro/grpc-mock/internal/controlapi/types"
	"github.com/zeromicro/grpc-mock/internal/svc"
)

type CaseListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCaseListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CaseListLogic {
	return &CaseListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CaseListLogic) CaseList(req *types.CaseListRequest) (resp *types.CaseListResponse, err error) {
	resp = &types.CaseListResponse{}
	for _, method := range req.MethodNames {
		cases, err := l.svcCtx.CaseManager.CaseList(l.ctx, method)
		if err != nil {
			return nil, err
		}
		resp.Cases = append(resp.Cases, cases...)
	}

	return resp, nil
}
