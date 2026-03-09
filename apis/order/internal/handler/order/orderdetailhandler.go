package order

import (
	"net/http"

	"jijizhazha1024/go-mall/apis/order/internal/logic/order"
	"jijizhazha1024/go-mall/apis/order/internal/svc"
	"jijizhazha1024/go-mall/apis/order/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
	xhttp "github.com/zeromicro/x/http"
)

func OrderDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetOrderReq
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		l := order.NewOrderDetailLogic(r.Context(), svcCtx)
		resp, err := l.OrderDetail(&req)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
