package handler

import (
	xhttp "github.com/zeromicro/x/http"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"jijizhazha1024/go-mall/apis/flash_sale/internal/logic"
	"jijizhazha1024/go-mall/apis/flash_sale/internal/svc"
	"jijizhazha1024/go-mall/apis/flash_sale/internal/types"
)

func FlashBuyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FlashBuyReq
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		l := logic.NewFlashBuyLogic(r.Context(), svcCtx)
		resp, err := l.FlashBuy(&req)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
