package handler

import (
	xhttp "github.com/zeromicro/x/http"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/falconfan123/Go-mall/apis/product/internal/logic"
	"github.com/falconfan123/Go-mall/apis/product/internal/svc"
	"github.com/falconfan123/Go-mall/apis/product/internal/types"
)

func GetProductByIDHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetProductByIDReq
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetProductByIDLogic(r.Context(), svcCtx)
		resp, err := l.GetProductByID(&req)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
