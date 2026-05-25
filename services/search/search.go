package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"

	"github.com/falconfan123/Go-mall/services/search/internal/config"
	"github.com/falconfan123/Go-mall/services/search/internal/server"
	"github.com/falconfan123/Go-mall/services/search/internal/svc"
	search "github.com/falconfan123/Go-mall/services/search/pb"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/search.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	ctx := svc.NewServiceContext(c)
	if err := ctx.RAGStats.EnsureSchema(context.Background()); err != nil {
		log.Fatalf("failed to ensure rag stats schema: %v", err)
	}
	if err := ctx.StatsHTTPServer.Start(); err != nil {
		log.Fatalf("failed to start stats http server: %v", err)
	}

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		search.RegisterSearchServiceServer(grpcServer, server.NewSearchServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := ctx.StatsHTTPServer.Stop(shutdownCtx); err != nil {
			log.Printf("failed to stop stats http server: %v", err)
		}
	}()
	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
