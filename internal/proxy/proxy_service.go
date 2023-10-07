package proxy

import (
	"github.com/zeromicro/go-zero/zrpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/zeromicro/grpc-mock/internal/match"
	internal2 "github.com/zeromicro/grpc-mock/internal/proxy/internal"
	"github.com/zeromicro/grpc-mock/internal/svc"
)

type Proxy struct {
	s       *zrpc.RpcServer
	svcCtx  *svc.ServiceContext
	matcher *match.Matcher
}

func NewProxy(svcCtx *svc.ServiceContext) *Proxy {
	s := zrpc.MustNewServer(svcCtx.Config.ProxyService.RpcServerConf, func(server *grpc.Server) {})
	matcher := match.NewMatcher(svcCtx)

	s.AddOptions(grpc.UnknownServiceHandler(
		internal2.TransparentHandler(func(ctx context.Context, fullMethodName string) (grpc.ClientConnInterface, error) {
			return svcCtx.DialManager.UpstreamClient(ctx, fullMethodName)
		}, matcher.Match)))
	return &Proxy{
		s:       s,
		svcCtx:  svcCtx,
		matcher: matcher,
	}
}

func (p *Proxy) Start() {
	p.s.Start()
}

func (p *Proxy) Stop() {
	p.s.Stop()
}
