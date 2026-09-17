package svc

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	gorse "github.com/falconfan123/Go-mall/common/utils/gorse"
	"github.com/falconfan123/Go-mall/dal/es/product"
	"github.com/falconfan123/Go-mall/dal/model/products/categories"
	product2 "github.com/falconfan123/Go-mall/dal/model/products/product"
	"github.com/falconfan123/Go-mall/services/inventory/inventoryclient"
	"github.com/falconfan123/Go-mall/services/product/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/olivere/elastic/v7"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
	"net/http"
	"strconv"
	"time"
)

type ServiceContext struct {
	Config          config.Config
	Postgres        sqlx.SqlConn
	RedisClient     *redis.Redis
	CategoriesModel categories.CategoriesModel
	EsClient        *elastic.Client
	InventoryRpc    inventoryclient.Inventory
	GorseClient     *gorse.GorseClient
	ProductModel    product2.ProductsModel
	MinioClient     *minio.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化 Redis 配置
	redisClient, err := redis.NewRedis(c.RedisConf)
	if err != nil {
		logx.Errorw("redis init error", logx.Field("err", err))
		panic(err)
	}

	var minioClient *minio.Client
	minioEnabled := c.Minio.Enabled && c.Minio.Endpoint != "" && c.Minio.Bucket != ""
	if minioEnabled {
		minioClient, err = minio.New(c.Minio.Endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(c.Minio.AccessKey, c.Minio.SecretKey, ""),
			Secure: c.Minio.UseSSL,
		})
		if err != nil {
			logx.Errorw("minio init error", logx.Field("err", err))
		} else {
			ctx := context.Background()
			exists, bucketErr := minioClient.BucketExists(ctx, c.Minio.Bucket)
			if bucketErr != nil {
				logx.Errorw("minio check bucket error", logx.Field("err", bucketErr))
			} else if !exists {
				bucketErr = minioClient.MakeBucket(ctx, c.Minio.Bucket, minio.MakeBucketOptions{})
				if bucketErr != nil {
					logx.Errorw("minio make bucket error", logx.Field("err", bucketErr))
				}
			}
		}
	} else if c.Minio.Enabled {
		logx.Infow("minio disabled due to incomplete config",
			logx.Field("endpoint", c.Minio.Endpoint),
			logx.Field("bucket", c.Minio.Bucket),
		)
	}

	// 初始化 ES 客户端
	var client *elastic.Client
	client, err = elastic.NewClient(elastic.SetURL(c.ElasticSearch.Addr),
		elastic.SetSniff(false),
		elastic.SetHealthcheckTimeoutStartup(30*time.Second))
	if err != nil {
		logx.Errorw("elasticsearch init error", logx.Field("err", err))
		// 不panic，继续运行服务
	}
	gorseClient := gorse.NewGorseClient(c.GorseConfig.GorseAddr, c.GorseConfig.GorseApikey)
	svcCtx := &ServiceContext{
		Config:          c,
		Postgres:        sqlx.NewSqlConn("postgres", c.PostgresConfig.DataSource),
		RedisClient:     redisClient,
		EsClient:        client,
		GorseClient:     gorseClient,
		ProductModel:    product2.NewProductsModel(sqlx.NewSqlConn("postgres", c.PostgresConfig.DataSource)),
		InventoryRpc:    inventoryclient.NewInventory(zrpc.MustNewClient(c.InventoryRpc)),
		CategoriesModel: categories.NewCategoriesModel(sqlx.NewSqlConn("postgres", c.PostgresConfig.DataSource)),
		MinioClient:     minioClient,
	}
	// 确保 products 索引就绪且映射正确（fix-product-search-index-readiness）。
	// ES 未就绪时有限重试（60s 窗口）；重建失败（删除索引后无法恢复）→ panic 阻断启动，不静默带空索引运行。
	if err := svcCtx.ensureProductIndex(context.TODO()); err != nil {
		logx.Errorw("elasticsearch ensure product index error", logx.Field("err", err))
		panic(err)
	}
	return svcCtx
}

// ensureProductIndex 确保 products 索引存在且映射为 snake_case：
//  1. 幂等 Put 索引模板（任何 auto-create 的 products* 索引都应用正确映射）；
//  2. 索引缺失 → 用 EsMapping 创建；
//  3. 索引已存在但映射错误（无 snake_case 关键字段或含 PascalCase 字段）→ 重建（删→建→分页重索引）。
//
// ES 未就绪时有限重试（60s 窗口、2s 间隔）；重建过程失败必须阻断启动（调用方 panic），防止静默空索引。
func (s *ServiceContext) ensureProductIndex(ctx context.Context) error {
	var lastErr error
	for i := 0; i < 30; i++ {
		lastErr = s.ensureProductIndexOnce(ctx)
		if lastErr == nil {
			return nil
		}
		logx.Errorw("ensure product index retry", logx.Field("attempt", i+1), logx.Field("err", lastErr))
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	return fmt.Errorf("ensure product index failed after retries: %w", lastErr)
}

func (s *ServiceContext) ensureProductIndexOnce(ctx context.Context) error {
	// 幂等 Put 模板
	if _, err := s.EsClient.IndexPutTemplate(productIndexTemplateName).BodyString(product.EsTemplate).Do(ctx); err != nil {
		return fmt.Errorf("put product index template failed: %w", err)
	}

	exists, err := s.EsClient.IndexExists(biz.ProductEsIndexName).Do(ctx)
	if err != nil {
		return err
	}
	if !exists {
		ci, err := s.EsClient.CreateIndex(biz.ProductEsIndexName).Body(product.EsMapping).Do(ctx)
		if err != nil {
			return err
		}
		if !ci.Acknowledged {
			return fmt.Errorf("create product index not acknowledged")
		}
		return nil
	}

	wrong, err := s.productIndexMappingWrong(ctx)
	if err != nil {
		return err
	}
	if !wrong {
		return nil
	}
	// 错误/混合映射 → 重建（删→建→分页重索引）；失败向上返回（调用方 panic 阻断启动）
	return s.rebuildProductIndex(ctx)
}

const productIndexTemplateName = "go-mall-products"

// productIndexMappingWrong 判定 products 映射是否与 snake_case 期望不一致：
// 缺少任一 snake_case 关键字段（created_at/price/category）或存在 PascalCase 字段（CreatedAt/Price）即视为错误/混合映射。
func (s *ServiceContext) productIndexMappingWrong(ctx context.Context) (bool, error) {
	// olivere v7 的 GetMapping 会带 ES7 专属 include_type_name=true，ES 8.x 拒绝（400）；
	// 改用裸 GET /products/_mapping 解析。
	resp, err := s.EsClient.PerformRequest(ctx, elastic.PerformRequestOptions{
		Method: "GET",
		Path:   "/" + biz.ProductEsIndexName + "/_mapping",
	})
	if err != nil {
		return false, err
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("get product mapping status=%d", resp.StatusCode)
	}
	var raw struct {
		Products struct {
			Mappings struct {
				Properties map[string]interface{} `json:"properties"`
			} `json:"mappings"`
		} `json:"products"`
	}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return false, err
	}
	fields := raw.Products.Mappings.Properties
	for _, want := range []string{"created_at", "price", "category"} {
		if _, ok := fields[want]; !ok {
			return true, nil
		}
	}
	for _, bad := range []string{"CreatedAt", "Price", "Id"} {
		if _, ok := fields[bad]; ok {
			return true, nil
		}
	}
	return false, nil
}

// rebuildProductIndex 删除错误映射索引 → 用 EsMapping 重建 → 分页从 DB 全量重索引（snake_case DTO）。
// 单实例部署前提下执行；重建前复查映射（幂等）。
func (s *ServiceContext) rebuildProductIndex(ctx context.Context) error {
	wrong, err := s.productIndexMappingWrong(ctx)
	if err != nil {
		return err
	}
	if !wrong {
		return nil
	}
	if _, err := s.EsClient.DeleteIndex(biz.ProductEsIndexName).Do(ctx); err != nil {
		return fmt.Errorf("rebuild: delete index failed: %w", err)
	}
	ci, err := s.EsClient.CreateIndex(biz.ProductEsIndexName).Body(product.EsMapping).Do(ctx)
	if err != nil {
		return fmt.Errorf("rebuild: create index failed: %w", err)
	}
	if !ci.Acknowledged {
		return fmt.Errorf("rebuild: create index not acknowledged")
	}
	const batch = 500
	for offset := 0; ; offset += batch {
		items, err := s.ProductModel.FindPage(ctx, offset, batch)
		if err != nil {
			return fmt.Errorf("rebuild: find page failed: %w", err)
		}
		if len(items) == 0 {
			return nil
		}
		for _, it := range items {
			if _, err := s.EsClient.Index().
				Index(biz.ProductEsIndexName).
				Id(strconv.FormatInt(it.Id, 10)).
				BodyJson(product.NewProductDocument(it)).
				Refresh("false").
				Do(ctx); err != nil {
				return fmt.Errorf("rebuild: index doc failed: %w", err)
			}
		}
	}
}
