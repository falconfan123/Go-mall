package svc

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"strconv"

	_ "github.com/lib/pq"

	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/dal/model/inventory"
	"github.com/falconfan123/Go-mall/services/inventory/internal/config"
	"github.com/falconfan123/Go-mall/services/inventory/internal/decreaselua"
	"github.com/falconfan123/Go-mall/services/inventory/internal/returnlua"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config         config.Config
	Rdb            *redis.Redis
	InventoryModel inventory.InventoryModel
	// DtmDB barrier 专用连接（dtm BranchBarrier.Call 自管事务；与业务共用同一 DSN）
	DtmDB *sql.DB

	DecreaseInventoryShal string
	ReturnInventoryShal   string
}

func NewServiceContext(c config.Config) *ServiceContext {

	// 创建ServiceContext实例
	dtmDB, err := sql.Open("postgres", c.PostgresConfig.DataSource)
	if err != nil {
		logx.Errorf("open dtm barrier db failed: %v", err)
		panic(err)
	}
	dtmDB.SetMaxOpenConns(20)

	svcCtx := &ServiceContext{
		Config:         c,
		Rdb:            redis.MustNewRedis(c.RedisConf),
		InventoryModel: inventory.NewInventoryModel(sqlx.NewSqlConn("postgres", c.PostgresConfig.DataSource)),
		DtmDB:          dtmDB,
	}

	// 执行缓存预热，失败只记录日志，避免把短暂的数据库/缓存异常放大成服务不可用
	if err := svcCtx.PreheatInventoryCache(); err != nil {
		logx.Errorf("缓存预热失败: %v", err)
	}
	decreaseInventoryShashal, err := svcCtx.predecreaseloadScript()
	if err != nil {
		panic(fmt.Sprintf("加载Lua脚本失败: %v", err))
	}
	svcCtx.DecreaseInventoryShal = decreaseInventoryShashal
	returnInventoryShashal, err := svcCtx.prereturnloadScript()
	if err != nil {
		panic(fmt.Sprintf("加载Lua脚本失败: %v", err))
	}
	svcCtx.ReturnInventoryShal = returnInventoryShashal

	return svcCtx
}

// 预热：只回填**缺失**的缓存键，不覆盖已存在的值。
// 缓存键 inventory:product:{pid} 口径为可用库存（可售）；若键已存在（可能含在途预扣后的可用值），
// 用 DB total 覆盖会造成可用库存虚高（超卖面）。见 fix-inventory-cache-invalidation design D1b。
func (s *ServiceContext) PreheatInventoryCache() error {
	// 1. 从数据库读取所有库存数据（或指定商品）
	inventories, err := s.InventoryModel.FindAll(context.Background())
	if err != nil {
		return fmt.Errorf("读取库存数据失败: %v", err)
	}

	for _, inv := range inventories {
		productKey := fmt.Sprintf("%s:%d", biz.InventoryProductKey, inv.ProductId)
		exists, err := s.Rdb.ExistsCtx(context.Background(), productKey)
		if err != nil {
			return fmt.Errorf("检查库存缓存键失败: %v", err)
		}
		if exists {
			continue
		}
		if err := s.Rdb.SetCtx(context.Background(), productKey, strconv.Itoa(int(inv.Total))); err != nil {
			return fmt.Errorf("缓存库存数据失败: %v", err)
		}
	}
	return nil

}

func (s *ServiceContext) predecreaseloadScript() (string, error) {

	sha, err := s.Rdb.ScriptLoad(decreaselua.Decreaselua)

	if err != nil {
		logx.Errorf("Failed to decrease load script: %v", err)
		return "", err
	}
	return sha, nil
}
func (s *ServiceContext) prereturnloadScript() (string, error) {

	sha, err := s.Rdb.ScriptLoad(returnlua.Returnlua)

	if err != nil {
		logx.Errorf("Failed to load return script: %v", err)
		return "", err
	}
	return sha, nil
}
