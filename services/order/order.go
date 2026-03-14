package main

import (
	"flag"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/falconfan123/Go-mall/services/order/internal/config"
	"github.com/falconfan123/Go-mall/services/order/internal/server"
	"github.com/falconfan123/Go-mall/services/order/internal/svc"
	order "github.com/falconfan123/Go-mall/services/order/pb"

	"github.com/falconfan123/Go-mall/common/utils/ip"
	"strings"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"github.com/zeromicro/zero-contrib/zrpc/registry/consul"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/order.yaml", "the config file")

// initSeckillData 初始化秒杀活动数据到 Redis
func initSeckillData(ctx *svc.ServiceContext) {
	// 设置秒杀活动开始时间限制（当前时间，允许立即抢购）
	startTime := time.Now().UnixMilli()
	err := ctx.RedisClient.Set("act_start_limit", fmt.Sprintf("%d", startTime))
	if err != nil {
		logx.Errorf("failed to set act_start_limit: %v", err)
	} else {
		logx.Infof("initialized act_start_limit: %d", startTime)
	}

	// 初始化秒杀商品库存
	productIds := []int64{1, 2, 3, 4}
	for _, productId := range productIds {
		stockKey := fmt.Sprintf("act_%d_stock", productId)
		err := ctx.RedisClient.Set(stockKey, "10")
		if err != nil {
			logx.Errorf("failed to set stock for product %d: %v", productId, err)
		} else {
			logx.Infof("initialized stock for product %d: 10", productId)
		}
	}
}

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	ctx := svc.NewServiceContext(c)

	// 初始化秒杀活动数据
	initSeckillData(ctx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		order.RegisterOrderServiceServer(grpcServer, server.NewOrderServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})

	registerOn := c.ListenOn
	if strings.Contains(registerOn, "0.0.0.0") {
		localIP, err := ip.GetLocalIP()
		if err == nil && localIP != "" {
			registerOn = strings.Replace(registerOn, "0.0.0.0", localIP, 1)
		} else {
			registerOn = strings.Replace(registerOn, "0.0.0.0", "host.docker.internal", 1)
		}
	}
	if err := consul.RegisterService(registerOn, c.Consul); err != nil {
		logx.Errorw("register service error", logx.Field("err", err))
		panic(err)
	}
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
