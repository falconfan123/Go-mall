package handler

import (
	"github.com/zeromicro/x/errors"
	xhttp "github.com/zeromicro/x/http"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/falconfan123/Go-mall/apis/checkout/internal/logic"
	"github.com/falconfan123/Go-mall/apis/checkout/internal/svc"
	"github.com/falconfan123/Go-mall/apis/checkout/internal/types"
)

func ListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CheckoutListReq
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, errors.New(code.Fail, err.Error()))
			return
		}

		l := logic.NewListLogic(r.Context(), svcCtx)
		resp, err := l.List(&req)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
