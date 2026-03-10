package handler

import (
	"fmt"
	// "github.com/zeromicro/x/errors"
	// xhttp "github.com/zeromicro/x/http"
	// "github.com/falconfan123/Go-mall/common/consts/code"
	"net/http"

	// "github.com/zeromicro/go-zero/rest/httpx"
	// "github.com/falconfan123/Go-mall/apis/order/internal/logic"
	"github.com/falconfan123/Go-mall/apis/order/internal/svc"
	// "github.com/falconfan123/Go-mall/apis/order/internal/types"
)

func CreateOrderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("CreateOrderHandler: REACHED!")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "handler_reached_raw"}`))
		return 

		// httpx.OkJsonCtx(r.Context(), w, map[string]string{"status": "handler_reached"})
		// return 

		// var req types.CreateOrderReq
		// if err := httpx.Parse(r, &req); err != nil {
		// 	httpx.ErrorCtx(r.Context(), w, errors.New(code.Fail, err.Error()))
		// 	return
		// }

		// l := logic.NewCreateOrderLogic(r.Context(), svcCtx)
		// resp, err := l.CreateOrder(&req)
		// if err != nil {
		// 	// xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		// 	httpx.ErrorCtx(r.Context(), w, err)
		// } else {
		// 	// xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		// 	httpx.OkJsonCtx(r.Context(), w, resp)
		// }
	}
}
