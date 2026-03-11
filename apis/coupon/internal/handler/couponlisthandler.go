package handler

import (
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/zeromicro/x/errors"
	"net/http"

	"github.com/falconfan123/Go-mall/apis/coupon/internal/logic"
	"github.com/falconfan123/Go-mall/apis/coupon/internal/svc"
	"github.com/falconfan123/Go-mall/apis/coupon/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
	xhttp "github.com/zeromicro/x/http"
)

func CouponListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CouponListReq
		// parse 会简单的进行参数校验，详细见：https://go-zero.dev/docs/tutorials/http/server/request-body
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, errors.New(code.Fail, err.Error()))
			return
		}

		l := logic.NewCouponListLogic(r.Context(), svcCtx)
		resp, err := l.CouponList(&req)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
