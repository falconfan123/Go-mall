package rpc

import (
	"context"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	gorse "github.com/falconfan123/Go-mall/common/utils/gorse"
	"github.com/falconfan123/Go-mall/dal/model/products/categories"
	product2 "github.com/falconfan123/Go-mall/dal/model/products/product"
	inventory "github.com/falconfan123/Go-mall/services/inventory/pb"
	product "github.com/falconfan123/Go-mall/services/product/pb"
	"github.com/falconfan123/Go-mall/test/rpc/internal/testenv"
	"github.com/olivere/elastic/v7"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"
)

var product_client product.ProductCatalogServiceClient

func initproduct() {
	conn, err := grpc.NewClient(testenv.ServiceAddr("product", biz.ProductRpcPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	product_client = product.NewProductCatalogServiceClient(conn)
}

func hasGorseConfig() bool {
	return os.Getenv("GORSE_HOST") != "" && os.Getenv("GORSE_APIKEY") != ""
}

func createProductFixture(t *testing.T, picture []byte, categories []string) int64 {
	t.Helper()
	initproduct()
	resp, err := product_client.CreateProduct(context.Background(), &product.CreateProductReq{
		Name:        testenv.UniqueName("product"),
		Description: "integration test product",
		Price:       122,
		Stock:       5000,
		Picture:     picture,
		Categories:  categories,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.GetStatusCode() != 0 {
		t.Fatalf("create product rejected: %s", resp.GetStatusMsg())
	}
	return resp.GetProductId()
}
func TestGetallProduct(t *testing.T) {
	initproduct()
	resp, err := product_client.GetAllProduct(context.Background(), &product.GetAllProductsReq{
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(" success", resp)
}
func TestProductsCreateRpc(t *testing.T) {
	initproduct()
	resp, err := product_client.CreateProduct(context.Background(), &product.CreateProductReq{
		Name:        testenv.UniqueName("product-create"),
		Description: "手机信息2",
		Price:       122,
		Stock:       5000,
		Picture:     []byte("hello"),
		Categories:  []string{"1", "2", "3"},
	})
	if err != nil {
		t.Fatal(err)

	}
	t.Log(" success", resp)
}
func TestProductsGetRpc(t *testing.T) {
	id := createProductFixture(t, nil, []string{"1"})
	initproduct()
	resp, err := product_client.GetProduct(context.Background(), &product.GetProductReq{
		Id: uint32(id),
	})
	if err != nil {
		t.Fatal(err)

	}
	t.Log(" success", resp)
}
func TestProductsUpdateRpc(t *testing.T) {
	id := createProductFixture(t, nil, []string{"1"})
	initproduct()
	resp, err := product_client.UpdateProduct(context.Background(), &product.UpdateProductReq{
		Id:          id,
		Name:        "we1",
		Description: "dsd",
		Price:       21,
		Stock:       32,
		Categories:  []string{"2", "3", "4"},
	})
	if err != nil {
		t.Fatal(err)

	}
	t.Log(" success", resp)
}
func TestProductsDeleteRpc(t *testing.T) {
	id := createProductFixture(t, nil, []string{"1"})
	initproduct()
	resp, err := product_client.DeleteProduct(context.Background(), &product.DeleteProductReq{
		Id: id,
	})
	if err != nil {
		t.Fatal(err)

	}
	t.Log(" success", resp)
}

func TestQueryProduct(t *testing.T) {
	createProductFixture(t, nil, []string{"1"})
	initproduct()
	resp, err := product_client.QueryProduct(context.Background(), &product.QueryProductReq{
		New:      true,
		Hot:      true,
		Category: []string{"Electronics"},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(" success", resp)
}

func TestLoadProduct2EsAndGorse(t *testing.T) {
	esAddress := os.Getenv("ELASTICSEARCH_HOST")
	mysqlAddress := os.Getenv("MYSQL_DATA_SOURCE")
	gorseAddr := os.Getenv("GORSE_HOST")
	gorseApikey := os.Getenv("GORSE_APIKEY")
	if esAddress == "" || mysqlAddress == "" || gorseAddr == "" || gorseApikey == "" {
		t.Skip("elasticsearch/mysql/gorse not configured")
	}

	ctx := context.TODO()
	client, err := elastic.NewClient(elastic.SetURL(esAddress),
		elastic.SetSniff(false),
		elastic.SetHealthcheckTimeoutStartup(30*time.Second))
	if err != nil {
		t.Fatal("elasticsearch init error", logx.Field("err", err))
	}
	productsModel := product2.NewProductsModel(sqlx.NewMysql(mysqlAddress))
	categoryModel := categories.NewCategoriesModel(sqlx.NewMysql(mysqlAddress))
	products, err := productsModel.QueryAllProducts(ctx)
	gorseClient := gorse.NewGorseClient(gorseAddr, gorseApikey)
	if err != nil {
		t.Fatal(err)
	}
	items := make([]gorse.Item, len(products))
	for i, p := range products {
		category, err := categoryModel.FindCategoryNameByProductID(ctx, p.Id)
		if err != nil {
			t.Fatal("query category failed", logx.Field("err", err))
			return
		}
		// 创建文档（自动JSON序列化）
		if _, err = client.Index().
			Index(biz.ProductEsIndexName).
			Id(strconv.FormatInt(p.Id, 10)).
			BodyJson(map[string]interface{}{
				"id":          p.Id,
				"name":        p.Name,
				"description": p.Description.String,
				"picture":     p.Picture.String,
				"price":       p.Price,
				"created_at":  p.CreatedAt.Format(time.DateTime),
				"updated_at":  p.UpdatedAt.Format(time.DateTime),
				"category":    category,
			}).
			Refresh("true").
			Do(ctx); err != nil {
			t.Fatal("product es creation failed", logx.Field("err", err))
			return
		}
		items[i] = gorse.Item{
			ItemId:     strconv.FormatInt(p.Id, 10),
			IsHidden:   false,
			Categories: category,
			Labels:     category,
			Comment:    p.Description.String,
			Timestamp:  p.CreatedAt.Format(time.DateTime),
		}
	}
	if _, err = gorseClient.InsertItems(ctx, items); err != nil {
		t.Fatal("gorse insert items failed", logx.Field("err", err))
		return
	}
}
func TestProductRecommend(t *testing.T) {
	if !hasGorseConfig() {
		t.Skip("gorse not configured")
	}
	initproduct()
	recommendProduct, err := product_client.RecommendProduct(context.Background(), &product.RecommendProductReq{
		UserId:   93,
		Category: []string{"手机"},
		Paginator: &product.RecommendProductReq_Paginator{
			Page:     1,
			PageSize: 10,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range recommendProduct.Products {
		t.Log(" success", p)
	}
}

func TestLoad2Inventory(t *testing.T) {
	conn, err := grpc.NewClient(testenv.ServiceAddr("inventory", biz.InventoryRpcPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	client := inventory.NewInventoryClient(conn)
	mysqlAddress := os.Getenv("MYSQL_DATA_SOURCE")
	if mysqlAddress == "" {
		t.Skip("MYSQL_DATA_SOURCE not configured")
	}
	productsModel := product2.NewProductsModel(sqlx.NewMysql(mysqlAddress))
	products, err := productsModel.QueryAllProducts(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range products {
		_, err := client.UpdateInventory(context.Background(), &inventory.UpdateInventoryReq{
			Items: []*inventory.UpdateInventoryReq_Items{
				{
					ProductId: int32(p.Id),
					Quantity:  rand.Int31n(1000),
				},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
	}

}
