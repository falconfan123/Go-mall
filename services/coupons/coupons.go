package main

import (
	"flag"
	"fmt"

	"github.com/falconfan123/Go-mall/services/coupons/internal/config"
	"github.com/falconfan123/Go-mall/services/coupons/internal/server"
	"github.com/falconfan123/Go-mall/services/coupons/internal/svc"
	coupons "github.com/falconfan123/Go-mall/services/coupons/pb"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/coupons.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		coupons.RegisterCouponsServer(grpcServer, server.NewCouponsServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})

	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
