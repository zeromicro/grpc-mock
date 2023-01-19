package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/zeromicro/grpc-mock/internal/logic"
	"github.com/zeromicro/grpc-mock/internal/svc"
	"github.com/zeromicro/grpc-mock/internal/types"
)

func UpstreamDelHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpstreamDelRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := logic.NewUpstreamDelLogic(svcCtx)
		resp, err := l.UpstreamDel(r.Context(), &req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
