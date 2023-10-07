package dialmanager

import (
	"github.com/zeromicro/go-zero/zrpc"

	"github.com/zeromicro/grpc-mock/internal/dialmanager/parser"
)

type (
	RpcClientConf struct {
		Name string
		zrpc.RpcClientConf
	}

	RpcClient struct {
		RpcClientConf
		zrpc.Client

		ServicesDesc []parser.ServiceDesc
	}
)
