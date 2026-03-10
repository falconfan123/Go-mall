package handler

import (
	"github.com/zeromicro/x/errors"
	xhttp "github.com/zeromicro/x/http"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/falconfan123/Go-mall/apis/coupon/internal/logic"
	"github.com/falconfan123/Go-mall/apis/coupon/internal/svc"
	"github.com/falconfan123/Go-mall/apis/coupon/internal/types"
)

func MyCouponListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CouponListReq
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, errors.New(code.Fail, err.Error()))
			return
		}

		l := logic.NewMyCouponListLogic(r.Context(), svcCtx)
		resp, err := l.MyCouponList(&req)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)

		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
