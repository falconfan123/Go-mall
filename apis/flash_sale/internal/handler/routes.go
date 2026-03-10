package handler

import (
	"net/http"

	"github.com/falconfan123/Go-mall/apis/flash_sale/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{serverCtx.WithClientMiddleware, serverCtx.WrapperAuthMiddleware},
			[]rest.Route{
				{
					Method:  http.MethodGet,
					Path:    "/products",
					Handler: GetFlashProductsHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/buy",
					Handler: FlashBuyHandler(serverCtx),
				},
			}...,
		),
		rest.WithPrefix("/douyin/flash"),
	)
}
