package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zeromicro/grpc-mock/internal/controlapi/types"
	"github.com/zeromicro/grpc-mock/internal/svc"
)

type MethodListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMethodListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MethodListLogic {
	return &MethodListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MethodListLogic) MethodList(req *types.MethodListRequest) (resp *types.MethodListResponse, err error) {
	methods, err := l.svcCtx.DialManager.Methods(l.ctx)
	if err != nil {
		return nil, err
	}

	resp = &types.MethodListResponse{}

	for s, ms := range methods {
		for _, m := range ms {
			resp.List = append(resp.List, types.Method{
				Service: s,
				Name:    m,
			})
		}
	}

	return resp, nil
}
