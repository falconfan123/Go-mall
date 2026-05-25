package rpc

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	gorse "github.com/falconfan123/Go-mall/common/utils/gorse"
	"github.com/falconfan123/Go-mall/dal/model/products/categories"
	product2 "github.com/falconfan123/Go-mall/dal/model/products/product"
	inventory "github.com/falconfan123/Go-mall/services/inventory/pb"
	product "github.com/falconfan123/Go-mall/services/product/pb"
	"github.com/falconfan123/Go-mall/test/rpc/internal/testenv"
	"github.com/olivere/elastic/v7"
	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
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

func hasQiniuConfig() bool {
	return os.Getenv("QINIU_ACCESSKEY") != "" &&
		os.Getenv("QINIU_SECRETKEY") != "" &&
		os.Getenv("QINIU_BUCKET") != "" &&
		os.Getenv("QINIU_DOMAIN") != ""
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
	if !hasQiniuConfig() {
		t.Skip("qiniu not configured")
	}
	initproduct()
	resp, err := product_client.CreateProduct(context.Background(), &product.CreateProductReq{
		Name:        testenv.UniqueName("product-qiniu"),
		Description: "手机信息2",
		Price:       122,
		Stock:       5000,
		Picture:     []byte("hello"),
		Categories:  []string{"10", "220", "330"},
	})
	if err != nil {
		t.Fatal(err)

	}
	t.Log(" success", resp)
}
func TestProductsGetRpc(t *testing.T) {
	id := createProductFixture(t, nil, []string{"get"})
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
	id := createProductFixture(t, nil, []string{"update"})
	initproduct()
	resp, err := product_client.UpdateProduct(context.Background(), &product.UpdateProductReq{
		Id:          id,
		Name:        "we1",
		Description: "dsd",
		Price:       21,
		Stock:       32,
		Categories:  []string{"1111", "2222", "33333"},
	})
	if err != nil {
		t.Fatal(err)

	}
	t.Log(" success", resp)
}
func TestProductsDeleteRpc(t *testing.T) {
	id := createProductFixture(t, nil, []string{"delete"})
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
	createProductFixture(t, nil, []string{"智能手机"})
	initproduct()
	resp, err := product_client.QueryProduct(context.Background(), &product.QueryProductReq{
		New:      true,
		Hot:      true,
		Category: []string{"智能手机"},
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

// 七牛云配置
const (
	accessKey = ""
	secretKey = ""
	bucket    = ""
	domain    = "" // 七牛云存储空间绑定的域名
)

func TestPicture(t *testing.T) {
	imagePath := "a.jpg" // 替换为实际的图片文件路径
	if _, err := os.Stat(imagePath); err != nil {
		t.Skip("image fixture not found")
	}
	base64Str, err := imageToBase64(imagePath)
	if err != nil {
		fmt.Printf("转换失败: %v\n", err)
		return
	}
	fmt.Printf("Base64 编码字符串: %s\n", base64Str)
	image, err := uploadImage(base64Str)
	if err != nil {
		fmt.Println(err)

	}
	fmt.Println(image)
}
func imageToBase64(imagePath string) (string, error) {
	// 打开图片文件
	file, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	// 读取图片文件内容
	imageData, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	// 进行 Base64 编码
	base64Str := base64.StdEncoding.EncodeToString(imageData)
	return base64Str, nil
}
func uploadImage(image string) (url string, err error) {

	// 1. Base64 解码
	decodedData, err := base64.StdEncoding.DecodeString(image)
	if err != nil {
		return "", fmt.Errorf("Base64 解码失败: %v", err)
	}
	// 2. 初始化七牛云认证信息
	mac := qbox.NewMac(accessKey, secretKey)
	putPolicy := storage.PutPolicy{
		Scope: bucket,
	}
	upToken := putPolicy.UploadToken(mac)
	cfg := storage.Config{}
	// 空间对应的机房
	cfg.Zone = &storage.ZoneHuabei
	// 是否使用 https 域名
	cfg.UseHTTPS = false
	// 上传是否使用 CDN 上传加速
	cfg.UseCdnDomains = false
	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}
	putExtra := storage.PutExtra{}
	// 生成一个唯一的文件名，这里简单使用时间戳
	filename := fmt.Sprintf("%d.jpg", time.Now().UnixNano())
	// 将 []byte 转换为 io.Reader
	reader := bytes.NewReader(decodedData)
	err = formUploader.Put(context.Background(), &ret, upToken, filename, reader, int64(len(decodedData)), &putExtra)
	if err != nil {
		return "", fmt.Errorf("上传到七牛云失败: %v", err)
	}
	// 3. 生成七牛云 URL
	return fmt.Sprintf("http://%s/%s", domain, ret.Key), nil
}
