package config

import (
	controlConfig "github.com/zeromicro/grpc-mock/internal/controlapi/config"
	"github.com/zeromicro/grpc-mock/internal/match/config"
	proxyConfig "github.com/zeromicro/grpc-mock/internal/proxy/config"
)

type (
	Config struct {
		ControlService controlConfig.Config
		ProxyService   proxyConfig.Config
		MatchConf      config.MatchConfig
	}
)
