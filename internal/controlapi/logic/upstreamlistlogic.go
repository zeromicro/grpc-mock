package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zeromicro/grpc-mock/internal/controlapi/types"
	"github.com/zeromicro/grpc-mock/internal/svc"
)

type UpstreamListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpstreamListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpstreamListLogic {
	return &UpstreamListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpstreamListLogic) UpstreamList() (resp *types.UpstreamListResponse, err error) {
	upstreams, err := l.svcCtx.DialManager.Upstreams(l.ctx)
	if err != nil {
		return nil, err
	}

	resp = &types.UpstreamListResponse{}
	for _, upstream := range upstreams {
		resp.Upstreams = append(resp.Upstreams, types.RpcClientConfig{
			Name: upstream.Name,
			Etcd: types.EtcdConf{
				Hosts:              upstream.Etcd.Hosts,
				Key:                upstream.Etcd.Key,
				ID:                 upstream.Etcd.ID,
				User:               upstream.Etcd.User,
				Pass:               upstream.Etcd.Pass,
				CertFile:           upstream.Etcd.CertFile,
				CertKeyFile:        upstream.Etcd.CertKeyFile,
				CACertFile:         upstream.Etcd.CACertFile,
				InsecureSkipVerify: upstream.Etcd.InsecureSkipVerify,
			},
			Endpoints: upstream.Endpoints,
			Target:    upstream.Target,
			App:       upstream.App,
			Token:     upstream.Token,
		})
	}

	return resp, nil
}
