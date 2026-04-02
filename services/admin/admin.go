package main

import (
	"flag"
	"fmt"

	"github.com/falconfan123/Go-mall/services/admin/internal/config"
	"github.com/falconfan123/Go-mall/services/admin/internal/db"
	"github.com/falconfan123/Go-mall/services/admin/internal/server"
	"github.com/falconfan123/Go-mall/services/admin/internal/svc"
	adminpb "github.com/falconfan123/Go-mall/services/admin/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/admin.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	ctx := svc.NewServiceContext(c)

	// Auto migrate database
	err := ctx.DB.AutoMigrate(&db.Activity{})
	if err != nil {
		fmt.Printf("failed to migrate database: %v\n", err)
	}

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		adminpb.RegisterAdminProductServiceServer(grpcServer, server.NewAdminProductServiceServer(ctx))
		adminpb.RegisterAdminCategoryServiceServer(grpcServer, server.NewAdminCategoryServiceServer(ctx))
		adminpb.RegisterAdminSeckillServiceServer(grpcServer, server.NewAdminSeckillServiceServer(ctx))
		adminpb.RegisterAdminInventoryServiceServer(grpcServer, server.NewAdminInventoryServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})

	defer s.Stop()
	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
