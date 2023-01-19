package logic

import (
	"context"
	"errors"

	"github.com/zeromicro/grpc-mock/internal/svc"
	"github.com/zeromicro/grpc-mock/internal/types"
)

type MethodDetailLogic struct {
	svcCtx *svc.ServiceContext
}

func NewMethodDetailLogic(svcCtx *svc.ServiceContext) *MethodDetailLogic {
	return &MethodDetailLogic{
		svcCtx: svcCtx,
	}
}

func (l *MethodDetailLogic) MethodDetail(ctx context.Context, req *types.MethodDetailRequest) (resp *types.MethodDetailResponse, err error) {
	baseResponse := types.BaseResponse{
		ErrorCode: 0,
		ErrorMsg:  "success",
	}
	resp = &types.MethodDetailResponse{BaseResponse: baseResponse}

	methodDesc := l.svcCtx.Mgr.GetMethodDesc(ctx, req.FullMethodName)
	if methodDesc == nil {
		resp.ErrorCode = 400
		resp.ErrorMsg = "illegal method name"
		return nil, errors.New("illegal method name")
	}

	return &types.MethodDetailResponse{
		BaseResponse: baseResponse,
		MethodName:   methodDesc.FullName,
		ProtoDesc:    methodDesc.ProtoDesc,
		In: types.FieldItem{
			Name:      methodDesc.In.Name,
			ProtoDesc: methodDesc.In.ProtoDesc,
		},
		Out: types.FieldItem{
			Name:      methodDesc.Out.Name,
			ProtoDesc: methodDesc.Out.ProtoDesc,
		},
	}, nil
}
