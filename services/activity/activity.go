package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/falconfan123/Go-mall/services/activity/internal/config"
	"github.com/falconfan123/Go-mall/services/activity/internal/server"
	"github.com/falconfan123/Go-mall/services/activity/internal/svc"
	activity "github.com/falconfan123/Go-mall/services/activity/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/activity.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	ctx := svc.NewServiceContext(c)

	// 初始化活动开始时间到 Redis（设置为一个小时后，方便测试）
	initActivityStartTime(ctx)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		activity.RegisterActivityServer(grpcServer, server.NewActivityServer(ctx))
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})

	defer s.Stop()
	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}

// initActivityStartTime 初始化活动开始时间到 Redis
func initActivityStartTime(ctx *svc.ServiceContext) {
	// 设置活动开始时间为当前时间后 1 小时（方便测试）
	// 生产环境可以从配置或数据库读取
	startTime := time.Now().Add(1 * time.Hour).UnixMilli()

	// 为前端定义的秒杀商品初始化活动开始时间
	productIds := []int64{1, 2, 3, 4}
	for _, productId := range productIds {
		key := fmt.Sprintf("act_%d_start", productId)
		err := ctx.Redis.Set(key, fmt.Sprintf("%d", startTime))
		if err != nil {
			logx.Errorf("failed to set activity start time for product %d: %v", productId, err)
		} else {
			logx.Infof("initialized activity start time for product %d: %d", productId, startTime)
		}
	}
}
