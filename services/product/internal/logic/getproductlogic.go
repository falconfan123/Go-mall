package logic

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/common/consts/code"
	gorse "github.com/falconfan123/Go-mall/common/utils/gorse"
	product2 "github.com/falconfan123/Go-mall/dal/model/products/product"
	"github.com/falconfan123/Go-mall/services/inventory/pb"
	"github.com/falconfan123/Go-mall/services/product/internal/svc"
	"github.com/falconfan123/Go-mall/services/product/pb"
	"github.com/zeromicro/go-zero/core/logx"
	"strconv"
	"time"
)

type GetProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProductLogic {
	return &GetProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetProduct 根据商品id得到商品详细信息
func (l *GetProductLogic) GetProduct(in *product.GetProductReq) (*product.GetProductResp, error) {

	// 在redis中维护商品的访问频率次数 PV
	// 检查商品 ID 是否存在
	redisKey := biz.ProductRedisPVName
	cacheKey := fmt.Sprintf(biz.ProductIDKey, in.Id)
	_, err := l.svcCtx.RedisClient.Zincrby(redisKey, 1, cacheKey)
	if err != nil {
		// 这里可以只进行记录即可，可以无需返回,还是可以正常的进行执行的，不影响返回结果
		l.Logger.Errorw("自增商品的访问次数失败",
			logx.Field("err", err),
			logx.Field("product_id", in.Id))
	}
	// 从Redis中获取数据
	cacheData, err := l.svcCtx.RedisClient.Get(cacheKey)
	if err != nil {
		// ...
		l.Logger.Errorw("get product from cache failed",
			logx.Field("err", err),
			logx.Field("product_id", in.Id))
	}

	// 如果Redis中有数据且没有错误，直接反序列化并返回
	if err == nil && cacheData != "" {
		var productRes product.Product
		if err := json.Unmarshal([]byte(cacheData), &productRes); err == nil {
			// 序列化成功返回，查询库存，我们进行返回动态库存
			stock, sold, err := l.getRealTimeStockAndSold(int64(in.Id))
			if err == nil {
				productRes.Stock = stock
				productRes.Sold = sold
			}
			return &product.GetProductResp{
				Product: &productRes,
			}, nil
		}
		// 序列失败 也是一样进行记录日志，因为在后面还可以从mysql查询，这样用用户体验感好点
		logx.Errorw("Failed to unmarshal data",
			logx.Field("err", err),
			logx.Field("product_id", in.Id))
	}

	// 如果Redis中没有数据，从数据库中获取

	productModel := product2.NewProductsModel(l.svcCtx.Mysql)
	// 4. 从数据库获取
	productData, err := productModel.FindOne(l.ctx, int64(in.Id))
	if err != nil {
		if err == product2.ErrNotFound {
			return &product.GetProductResp{
				StatusCode: uint32(code.ProductNotFound),
				StatusMsg:  code.ProductNotFoundMsg,
			}, nil
		}
		return nil, err
	}

	resp := &product.GetProductResp{
		Product: &product.Product{
			Id:          uint32(productData.Id),
			Name:        productData.Name,
			Description: productData.Description.String,
			Picture:     productData.Picture.String,
			Price:       productData.Price,
			Stock:       productData.Stock,
			CratedAt:    productData.CreatedAt.Format(time.DateTime),
			UpdatedAt:   productData.CreatedAt.Format(time.DateTime),
		},
	}

	// 在这里创建连接，懒惰创建连接。
	categories, err := l.svcCtx.CategoriesModel.FindCategoryNameByProductID(l.ctx, int64(in.Id))
	if err != nil {
		l.Logger.Errorw("Failed to find product_category from database",
			logx.Field("err", err),
			logx.Field("product_id", in.Id))
		// 因为查询不完整，所以不需要写入缓存了，直接返回
		return resp, nil
	}

	resp.Product.Categories = categories
	// 到这里就说明数据是完整的，将数据缓存到Redis中
	data, err := json.Marshal(resp.Product)
	cacheData = string(data)
	if err != nil {
		l.Logger.Errorw("Failed to unmarshal data",
			logx.Field("err", err),
			logx.Field("product_id", in.Id))
		return resp, nil
	}
	// 设置合理的过期时间
	if err = l.svcCtx.RedisClient.SetexCtx(l.ctx, cacheKey, cacheData, biz.ProductIDKeyExpire); err != nil {
		l.Logger.Errorw("Failed to save redis data",
			logx.Field("err", err),
			logx.Field("product_id", in.Id))
		return resp, nil
	}

	// 查询实时库存
	stock, sold, err := l.getRealTimeStockAndSold(productData.Id)
	if err == nil {
		resp.Product.Stock = stock
		resp.Product.Sold = sold
	}
	if in.UserId != 0 {
		go func() {
			// 插入反馈
			if _, err := l.svcCtx.GorseClient.InsertFeedback(l.ctx, []gorse.Feedback{
				{
					ItemId:       strconv.Itoa(int(productData.Id)),
					UserId:       strconv.Itoa(int(in.UserId)),
					Timestamp:    time.Now().Format(time.DateTime),
					FeedbackType: biz.ReadFeedBackType,
				},
			}); err != nil {
				l.Logger.Infow("Failed to insert feedback", logx.Field("err", err), logx.Field("product_id", productData.Id))
				return
			}
		}()
	}

	return resp, nil
}

// 抽取重复的库存查询逻辑
func (l *GetProductLogic) getRealTimeStockAndSold(productId int64) (int64, int64, error) {
	ctx, cancel := context.WithTimeout(l.ctx, 500*time.Millisecond)
	defer cancel()
	inventoryResp, err := l.svcCtx.InventoryRpc.GetInventory(ctx, &inventory.GetInventoryReq{
		ProductId: int32(productId),
	})
	if err != nil {
		l.Logger.Errorw("call rpc InventoryRpc.GetInventory failed", logx.Field("err", err), logx.Field("product_id", productId))
		return 0, 0, err
	}
	return inventoryResp.Inventory, inventoryResp.SoldCount, nil
}
