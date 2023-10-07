package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/zeromicro/grpc-mock/internal/controlapi/logic"
	"github.com/zeromicro/grpc-mock/internal/svc"
)

func UpstreamListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewUpstreamListLogic(r.Context(), svcCtx)
		resp, err := l.UpstreamList()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
