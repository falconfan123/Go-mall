package main

import (
	"flag"
	"fmt"

	"jijizhazha1024/go-mall/services/order/internal/config"
	"jijizhazha1024/go-mall/services/order/internal/server"
	"jijizhazha1024/go-mall/services/order/internal/svc"
	"jijizhazha1024/go-mall/services/order/order"

	"jijizhazha1024/go-mall/common/utils/ip"
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

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	ctx := svc.NewServiceContext(c)

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
