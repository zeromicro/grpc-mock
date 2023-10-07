package controlapi

import (
	"github.com/zeromicro/go-zero/rest"

	"github.com/zeromicro/grpc-mock/internal/controlapi/handler"
	"github.com/zeromicro/grpc-mock/internal/svc"
)

type Control struct {
	s      *rest.Server
	svcCtx *svc.ServiceContext
}

func NewControl(ctx *svc.ServiceContext) *Control {
	server := rest.MustNewServer(ctx.Config.ControlService.RestConf)
	handler.RegisterHandlers(server, ctx)

	return &Control{
		s:      server,
		svcCtx: ctx,
	}
}

func (c *Control) Start() {
	c.s.Start()
}

func (c *Control) Stop() {
	c.s.Stop()
}
