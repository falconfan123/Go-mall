package main

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/zero-contrib/zrpc/registry/consul"

	checkout "github.com/falconfan123/Go-mall/services/checkout/pb"
	"github.com/falconfan123/Go-mall/services/checkout/internal/config"
	"github.com/falconfan123/Go-mall/services/checkout/internal/server"
	"github.com/falconfan123/Go-mall/services/checkout/internal/svc"

	"github.com/falconfan123/Go-mall/common/utils/ip"
	"strings"

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
