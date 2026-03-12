package handler

import (
	"net/http"

	"github.com/falconfan123/Go-mall/apis/product/internal/svc"
)

// UploadHandler handles HTTP requests.
func UploadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}
