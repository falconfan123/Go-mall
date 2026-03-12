package handler

import (
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/zeromicro/x/errors"
	xhttp "github.com/zeromicro/x/http"
	"net/http"

	"github.com/falconfan123/Go-mall/apis/order/internal/logic"
	"github.com/falconfan123/Go-mall/apis/order/internal/svc"
	"github.com/falconfan123/Go-mall/apis/order/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetOrderReq
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, errors.New(code.Fail, err.Error()))
			return
		}

		l := logic.NewGetOrderLogic(r.Context(), svcCtx)
		resp, err := l.GetOrder(&req)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)

		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
