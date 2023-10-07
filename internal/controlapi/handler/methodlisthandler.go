package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/zeromicro/grpc-mock/internal/controlapi/logic"
	"github.com/zeromicro/grpc-mock/internal/controlapi/types"
	"github.com/zeromicro/grpc-mock/internal/svc"
)

func MethodListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MethodListRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewMethodListLogic(r.Context(), svcCtx)
		resp, err := l.MethodList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
