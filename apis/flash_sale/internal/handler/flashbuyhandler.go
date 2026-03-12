package handler

import (
	xhttp "github.com/zeromicro/x/http"
	"net/http"

	"github.com/falconfan123/Go-mall/apis/flash_sale/internal/logic"
	"github.com/falconfan123/Go-mall/apis/flash_sale/internal/svc"
	"github.com/falconfan123/Go-mall/apis/flash_sale/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// FlashBuyHandler handles HTTP requests.
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
