package svc

import (
	"log"
	"net"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"google.golang.org/grpc"

	"github.com/zeromicro/grpc-mock/internal/config"
	"github.com/zeromicro/grpc-mock/internal/logic/mockmgr"
	proxy2 "github.com/zeromicro/grpc-mock/internal/logic/proxy"
)

type ServiceContext struct {
	Config config.Config
	Mgr    *mockmgr.CaseMgr

	BizRedis *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	bizRedis := c.BizRedis.NewRedis()
	caseMgr, err := mockmgr.NewCaseMgr(&c, bizRedis)
	if err != nil {
		log.Fatal(err)
	}

	// start proxy rpc
	lis, err := net.Listen("tcp", c.ProxyPort)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer(grpc.UnknownServiceHandler(
		proxy2.TransparentHandler(&c, proxy2.GetStreamDirector(), caseMgr.GetMockResponseByMeta, caseMgr.GetMockResponseByBody)))
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	return &ServiceContext{
		Config:   c,
		Mgr:      caseMgr,
		BizRedis: c.BizRedis.NewRedis(),
	}
}
