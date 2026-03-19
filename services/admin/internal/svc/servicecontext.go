package svc

import (
	"github.com/falconfan123/Go-mall/services/admin/internal/config"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config       config.Config
	Redis        *redis.Redis
	DB           *gorm.DB
	ProductRpc   zrpc.Client
	InventoryRpc zrpc.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	r, _ := redis.NewRedis(c.RedisConf)

	// Initialize database
	db, err := gorm.Open(postgres.Open(c.PostgresConfig.DataSource), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	// Initialize RPC clients
	productRpc := zrpc.MustNewClient(c.ProductRpc)
	inventoryRpc := zrpc.MustNewClient(c.InventoryRpc)

	return &ServiceContext{
		Config:       c,
		Redis:        r,
		DB:           db,
		ProductRpc:   productRpc,
		InventoryRpc: inventoryRpc,
	}
}
