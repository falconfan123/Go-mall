package order

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"jijizhazha1024/go-mall/apis/order/internal/logic/order"
	"jijizhazha1024/go-mall/apis/order/internal/svc"
	"jijizhazha1024/go-mall/apis/order/internal/types"
)

func CancelOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CancelOrderReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := order.NewCancelOrderLogic(r.Context(), svcCtx)
		resp, err := l.CancelOrder(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
