package logic

import (
	"context"
	"fmt"

	"github.com/falconfan123/Go-mall/services/product/internal/svc"
	"github.com/falconfan123/Go-mall/services/product/product"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListProductsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListProductsLogic {
	return &ListProductsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 高性能游标分页接口
func (l *ListProductsLogic) ListProducts(in *product.ListProductsReq) (*product.ListProductsResp, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// Call model to get products
	products, err := l.svcCtx.ProductModel.FindListByCursor(l.ctx, in.Cursor, limit)
	if err != nil {
		l.Logger.Errorw("query products failed", logx.Field("err", err))
		return &product.ListProductsResp{
			StatusCode: 500,
			StatusMsg:  "query failed",
		}, nil
	}

	var respProducts []*product.Product
	var nextCursor int64
	var hasMore bool

	// Base URL for MinIO
	protocol := "http"
	if l.svcCtx.Config.Minio.UseSSL {
		protocol = "https"
	}
	minioBase := fmt.Sprintf("%s://%s/%s/", protocol, l.svcCtx.Config.Minio.Endpoint, l.svcCtx.Config.Minio.Bucket)

	respProducts = make([]*product.Product, 0)
	for _, p := range products {
		thumbnailUrl := p.Picture.String
		// If picture is a relative path (not starting with http) and not empty, prepend MinIO domain
		if len(thumbnailUrl) > 0 && (len(thumbnailUrl) < 4 || thumbnailUrl[:4] != "http") {
			thumbnailUrl = minioBase + thumbnailUrl
		}

		respProducts = append(respProducts, &product.Product{
			Id:           uint32(p.Id),
			Name:         p.Name,
			Description:  p.Description.String,
			Picture:      p.Picture.String,
			Price:        p.Price,
			ThumbnailUrl: thumbnailUrl,
			CratedAt:     p.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    p.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
		nextCursor = p.Id
	}

	if len(products) == int(limit) {
		hasMore = true
	}

	return &product.ListProductsResp{
		Products:   respProducts,
		NextCursor: nextCursor,
		HasMore:    hasMore,
		StatusCode: 0,
		StatusMsg:  "success",
	}, nil
}
