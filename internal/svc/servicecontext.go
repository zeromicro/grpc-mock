package svc

import (
	"github.com/zeromicro/grpc-mock/config"
	"github.com/zeromicro/grpc-mock/internal/casemanager"
	"github.com/zeromicro/grpc-mock/internal/dialmanager"
)

type ServiceContext struct {
	Config      config.Config
	DialManager *dialmanager.Manager
	CaseManager *casemanager.Manager
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:      c,
		DialManager: dialmanager.NewManager(),
		CaseManager: casemanager.NewManager(),
	}
}
