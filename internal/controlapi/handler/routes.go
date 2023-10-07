// Code generated by goctl. DO NOT EDIT.
package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest"

	"github.com/zeromicro/grpc-mock/internal/svc"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/upstreams",
				Handler: UpstreamListHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/upstreams/set",
				Handler: UpstreamSetHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/upstreams/del",
				Handler: UpstreamDelHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/methods",
				Handler: MethodListHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/methods/detail",
				Handler: MethodDetailHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/cases",
				Handler: CaseListHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/cases/set",
				Handler: CaseSetHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/cases/del",
				Handler: CaseDelHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/cases/detail",
				Handler: CaseDetailHandler(serverCtx),
			},
		},
	)
}
