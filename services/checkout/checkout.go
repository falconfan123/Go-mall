package main

import (
	"flag"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/falconfan123/Go-mall/services/checkout/internal/config"
	"github.com/falconfan123/Go-mall/services/checkout/internal/server"
	"github.com/falconfan123/Go-mall/services/checkout/internal/svc"
	checkout "github.com/falconfan123/Go-mall/services/checkout/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/checkout.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		checkout.RegisterCheckoutServiceServer(grpcServer, server.NewCheckoutServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
