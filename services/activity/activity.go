package main

import (
	"flag"
	"fmt"

	"github.com/falconfan123/Go-mall/services/activity/internal/config"
	"github.com/falconfan123/Go-mall/services/activity/internal/server"
	"github.com/falconfan123/Go-mall/services/activity/internal/svc"
	activity "github.com/falconfan123/Go-mall/services/activity/pb"

	"github.com/zeromicro/go-zero/core/conf"
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
