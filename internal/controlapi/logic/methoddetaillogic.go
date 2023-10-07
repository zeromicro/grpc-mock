package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zeromicro/grpc-mock/internal/controlapi/types"
	"github.com/zeromicro/grpc-mock/internal/svc"
)

type MethodDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMethodDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MethodDetailLogic {
	return &MethodDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MethodDetailLogic) MethodDetail(req *types.MethodDetailRequest) (resp *types.MethodDetailResponse, err error) {
	desc, err := l.svcCtx.DialManager.MethodDetail(l.ctx, req.FullMethodName)
	if err != nil {
		return nil, err
	}

	return &types.MethodDetailResponse{
		MethodName: desc.FullName,
		ProtoDesc:  desc.ProtoDesc,
		In: types.FieldItem{
			Name:      desc.In.Name,
			ProtoDesc: desc.In.ProtoDesc,
			JsonDesc:  desc.In.JsonDesc,
		},
		Out: types.FieldItem{
			Name:      desc.Out.Name,
			ProtoDesc: desc.Out.ProtoDesc,
			JsonDesc:  desc.Out.JsonDesc,
		},
	}, nil
}
