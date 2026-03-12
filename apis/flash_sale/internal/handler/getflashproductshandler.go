package handler

import (
	xhttp "github.com/zeromicro/x/http"
	"net/http"

	"github.com/falconfan123/Go-mall/apis/flash_sale/internal/logic"
	"github.com/falconfan123/Go-mall/apis/flash_sale/internal/svc"
	"github.com/falconfan123/Go-mall/apis/flash_sale/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// GetFlashProductsHandler handles HTTP requests.
func GetFlashProductsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetFlashProductsReq
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetFlashProductsLogic(r.Context(), svcCtx)
		resp, err := l.GetFlashProducts(&req)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
