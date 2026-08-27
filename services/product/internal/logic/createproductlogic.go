package logic

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	product2 "github.com/falconfan123/Go-mall/dal/model/products/product"
	pc "github.com/falconfan123/Go-mall/dal/model/products/product_categories"
	inventoryclient "github.com/falconfan123/Go-mall/services/inventory/inventoryclient"
	"github.com/falconfan123/Go-mall/services/product/internal/svc"
	product "github.com/falconfan123/Go-mall/services/product/pb"
	"github.com/olivere/elastic/v7"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateProductLogic {
	return &CreateProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateProduct 添加新商品
func (l *CreateProductLogic) CreateProduct(in *product.CreateProductReq) (*product.CreateProductResp, error) {
	// 1. 敏感词校验
	if err := l.checkSensitiveWords(in.Name); err != nil {
		l.Logger.Errorw("product sensitive word failed",
			logx.Field("err", err))
		return nil, err
	}
	var productId int64
	picture := string(in.Picture)

	// 创建 Products 结构体实例
	productRes := &product2.Products{
		Name:        in.Name,
		Description: sql.NullString{String: in.Description, Valid: in.Description != ""},
		Picture:     sql.NullString{String: picture, Valid: picture != ""},
		Price:       in.Price, // 注意类型转换
		Stock:       in.Stock,
	}
	res := &product.CreateProductResp{}
	// 2. 使用 Transact 开启事务
	if err := l.svcCtx.Postgres.Transact(func(session sqlx.Session) error {
		// 通过 withSession 生成支持事务的 productModel 实例
		productModel := product2.NewProductsModel(l.svcCtx.Postgres).WithSession(session)
		// 通过 withSession 生成支持事务的 productCategoriesModel 实例
		productCategoriesModel := pc.NewProductCategoriesModel(l.svcCtx.Postgres).WithSession(session)
		// 得到图片对应url
		var err error
		productId, err = productModel.InsertReturningID(l.ctx, productRes)
		if err != nil {
			return err
		}
		// 3. 插入商品分类关联信息
		for _, categoryId := range in.Categories {
			categoryId, err := strconv.ParseInt(categoryId, 10, 64)
			if err != nil {
				return err
			}
			p := &pc.ProductCategories{
				ProductId:  sql.NullInt64{Int64: productId, Valid: productId != 0},
				CategoryId: sql.NullInt64{Int64: categoryId, Valid: categoryId != 0},
			}
			if _, err := productCategoriesModel.Insert(l.ctx, p); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		l.Logger.Errorw("product creation failed",
			logx.Field("err", err))
		return nil, err
	}
	// 创建文档（自动JSON序列化）
	// 4. 初始化库存，失败时补偿删除商品，避免留下无法结算的脏数据
	if _, err := l.svcCtx.InventoryRpc.UpdateInventory(l.ctx, &inventoryclient.UpdateInventoryReq{
		Items: []*inventoryclient.UpdateInventoryReq_Items{
			{
				ProductId: int32(productId),
				Quantity:  int32(in.Stock),
			},
		},
	}); err != nil {
		l.Logger.Errorw("product inventory update failed",
			logx.Field("err", err),
			logx.Field("product_id", productId))
		if rollbackErr := l.compensateProductCreation(productId); rollbackErr != nil {
			return nil, fmt.Errorf("initialize inventory for product %d: %w (compensation failed: %v)", productId, err, rollbackErr)
		}
		return nil, fmt.Errorf("initialize inventory for product %d: %w", productId, err)
	}

	res.ProductId = productId

	// 创建文档（自动JSON序列化）
	if _, err := l.svcCtx.EsClient.Index().
		Index(biz.ProductEsIndexName).
		Id(strconv.FormatInt(productId, 10)).
		BodyJson(productRes).
		Refresh("true").
		Do(l.ctx); err != nil {
		l.Logger.Errorw("product es creation failed",
			logx.Field("err", err),
			logx.Field("product_id", productId))
		return res, nil
	}
	return res, nil
}

func (l *CreateProductLogic) checkSensitiveWords(text string) error {
	// 敏感词过滤逻辑
	// 目前仅使用简单的字符串匹配

	if text == "敏感词" {
		return fmt.Errorf("包含敏感词")
	}
	return nil
}

func (l *CreateProductLogic) compensateProductCreation(productID int64) error {
	if err := l.svcCtx.Postgres.Transact(func(session sqlx.Session) error {
		productModel := product2.NewProductsModel(l.svcCtx.Postgres).WithSession(session)
		productCategoriesModel := pc.NewProductCategoriesModel(l.svcCtx.Postgres).WithSession(session)

		if err := productCategoriesModel.DeleteByProductId(l.ctx, productID); err != nil {
			return err
		}

		if err := productModel.Delete(l.ctx, productID); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	if _, err := l.svcCtx.EsClient.Delete().
		Index(biz.ProductEsIndexName).
		Id(strconv.FormatInt(productID, 10)).
		Refresh("true").
		Do(l.ctx); err != nil && !elastic.IsNotFound(err) {
		return err
	}

	return nil
}
