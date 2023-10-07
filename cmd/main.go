package main

import (
	"flag"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"

	"github.com/zeromicro/grpc-mock/config"
	"github.com/zeromicro/grpc-mock/internal/controlapi"
	"github.com/zeromicro/grpc-mock/internal/proxy"
	"github.com/zeromicro/grpc-mock/internal/svc"
)

var configFile = flag.String("f", "etc/config.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	sg := service.NewServiceGroup()

	svcCtx := svc.NewServiceContext(c)

	sg.Add(controlapi.NewControl(svcCtx))
	sg.Add(proxy.NewProxy(svcCtx))

	defer sg.Stop()

	sg.Start()
}
