package handler

import (
	"net/http"

	"github.com/falconfan123/Go-mall/apis/user/internal/logic"
	"github.com/falconfan123/Go-mall/apis/user/internal/svc"
	"github.com/falconfan123/Go-mall/apis/user/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
	xhttp "github.com/zeromicro/x/http"
)

func AddAddressHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AddAddressRequest
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		l := logic.NewAddAddressLogic(r.Context(), svcCtx)
		resp, err := l.AddAddress(&req)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
