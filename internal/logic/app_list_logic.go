package logic

import (
	"context"

	"github.com/zeromicro/grpc-mock/internal/svc"
	"github.com/zeromicro/grpc-mock/internal/types"
)

type AppListLogic struct {
	svcCtx *svc.ServiceContext
}

func NewAppListLogic(svcCtx *svc.ServiceContext) *AppListLogic {
	return &AppListLogic{
		svcCtx: svcCtx,
	}
}

func (l *AppListLogic) AppList(ctx context.Context) (resp *types.AppListResponse, err error) {
	resp = &types.AppListResponse{BaseResponse: types.BaseResponse{
		ErrorCode: 0,
		ErrorMsg:  "success",
	}}

	apps, err := l.svcCtx.Mgr.GetApps()
	if err != nil {
		resp.ErrorCode = 500
		resp.ErrorMsg = err.Error()
		return
	}

	resp.Apps = apps

	return
}
