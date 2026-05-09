package main

import (
	"flag"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/falconfan123/Go-mall/services/audit/internal/config"
	"github.com/falconfan123/Go-mall/services/audit/internal/server"
	"github.com/falconfan123/Go-mall/services/audit/internal/svc"
	audit "github.com/falconfan123/Go-mall/services/audit/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/audit.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	ctx := svc.NewServiceContext(c)
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		audit.RegisterAuditServer(grpcServer, server.NewAuditServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
