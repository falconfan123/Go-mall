package handler

import (
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/zeromicro/x/errors"
	xhttp "github.com/zeromicro/x/http"
	"net/http"

	"github.com/falconfan123/Go-mall/apis/payment/internal/logic"
	"github.com/falconfan123/Go-mall/apis/payment/internal/svc"
	"github.com/falconfan123/Go-mall/apis/payment/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func listPaymentsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PaymentListReq
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, errors.New(code.Fail, err.Error()))
			return
		}

		l := logic.NewListPaymentsLogic(r.Context(), svcCtx)
		resp, err := l.ListPayments(&req)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
