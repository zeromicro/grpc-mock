package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zeromicro/grpc-mock/internal/controlapi/types"
	"github.com/zeromicro/grpc-mock/internal/dialmanager"
	"github.com/zeromicro/grpc-mock/internal/svc"
)

type UpstreamSetLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpstreamSetLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpstreamSetLogic {
	return &UpstreamSetLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpstreamSetLogic) UpstreamSet(req *types.UpstreamSetRequest) (resp *types.UpstreamSetResponse, err error) {
	var clients []dialmanager.RpcClientConf
	for _, upstream := range req.Upstreams {
		var rpcClient dialmanager.RpcClientConf
		conf.FillDefault(&rpcClient)
		rpcClient.Name = upstream.Name

		rpcClient.RpcClientConf.Etcd.Hosts = upstream.Etcd.Hosts
		rpcClient.RpcClientConf.Etcd.Key = upstream.Etcd.Key
		rpcClient.RpcClientConf.Etcd.ID = upstream.Etcd.ID
		rpcClient.RpcClientConf.Etcd.User = upstream.Etcd.User
		rpcClient.RpcClientConf.Etcd.Pass = upstream.Etcd.Pass
		rpcClient.RpcClientConf.Etcd.CertFile = upstream.Etcd.CertFile
		rpcClient.RpcClientConf.Etcd.CertKeyFile = upstream.Etcd.CertKeyFile
		rpcClient.RpcClientConf.Etcd.CACertFile = upstream.Etcd.CACertFile
		rpcClient.RpcClientConf.Etcd.InsecureSkipVerify = upstream.Etcd.InsecureSkipVerify
		rpcClient.RpcClientConf.Endpoints = upstream.Endpoints
		rpcClient.RpcClientConf.Target = upstream.Target
		rpcClient.RpcClientConf.App = upstream.App
		rpcClient.RpcClientConf.Token = upstream.Token

		clients = append(clients, rpcClient)
	}

	if err = l.svcCtx.DialManager.AddUpstream(l.ctx, clients); err != nil {
		return
	}

	return
}
