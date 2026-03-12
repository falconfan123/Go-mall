package handler

import (
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/zeromicro/x/errors"
	xhttp "github.com/zeromicro/x/http"
	"net/http"

	"github.com/falconfan123/Go-mall/apis/checkout/internal/logic"
	"github.com/falconfan123/Go-mall/apis/checkout/internal/svc"
	"github.com/falconfan123/Go-mall/apis/checkout/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// DetailHandler handles HTTP requests.
func DetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CheckoutDetailReq
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, errors.New(code.Fail, err.Error()))
			return
		}

		l := logic.NewDetailLogic(r.Context(), svcCtx)
		resp, err := l.Detail(&req)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
