// Based on https://github.com/trusch/grpc-proxy
// Copyright Michal Witkowski. Licensed under Apache2 license: https://github.com/trusch/grpc-proxy/blob/master/LICENSE.txt

package internal

import (
	"errors"
	"sync"

	"github.com/zeromicro/go-zero/zrpc"
	"go.uber.org/atomic"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/zeromicro/grpc-mock/internal/proxy/internal/codec"
)

// StreamDirector returns a gRPC ClientConn to be used to forward the call to.
//
// The presence of the `Context` allows for rich filtering, e.g. based on Metadata (headers).
// If no handling is meant to be done, a `codes.NotImplemented` gRPC error should be returned.
//
// The context returned from this function should be the context for the *outgoing* (to backend) call. In case you want
// to forward any Metadata between the inbound request and outbound requests, you should do it manually. However, you
// *must* propagate the cancel function (`context.WithCancel`) of the inbound context to the one returned.
//
// It is worth noting that the StreamDirector will be fired *after* all server-side stream interceptors
// are invoked. So decisions around authorization, monitoring etc. are better to be handled there.
//
// See the rather rich example.

var grpcMgr = &grpcManager{}

type grpcManager struct {
	mu               sync.RWMutex
	method2CliAtomic atomic.Value
}

func NewClient(conf *zrpc.RpcClientConf) (zrpc.Client, error) {
	return zrpc.NewClient(*conf,
		zrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		zrpc.WithDialOption(grpc.WithDefaultCallOptions(grpc.CallContentSubtype((&codec.Proxy{}).Name()))))
}

func SetMethod2CliAtomic(mp map[string]zrpc.Client) {
	grpcMgr.method2CliAtomic.Store(mp)
}

type StreamDirector func(ctx context.Context, fullMethodName string) (context.Context, grpc.ClientConnInterface, func(), error)

func GetStreamDirector() StreamDirector {
	return grpcMgr.StreamDirector
}

func (m *grpcManager) StreamDirector(ctx context.Context, fullMethodName string) (context.Context, grpc.ClientConnInterface, func(), error) {
	m.mu.RLock()
	cli := grpcMgr.method2CliAtomic.Load().(map[string]zrpc.Client)[fullMethodName]
	m.mu.RUnlock()

	if cli == nil {
		return ctx, nil, func() {}, errors.New("unregistered client")
	}

	return ctx, cli.Conn(), func() {}, nil
}
