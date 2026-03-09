package main

import (
	"flag"
	"fmt"
	"jijizhazha1024/go-mall/services/auths/auths"
	"jijizhazha1024/go-mall/services/auths/internal/config"
	"jijizhazha1024/go-mall/services/auths/internal/server"
	"jijizhazha1024/go-mall/services/auths/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/zero-contrib/zrpc/registry/consul"

	"jijizhazha1024/go-mall/common/utils/ip"
	"strings"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/auths.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	ctx := svc.NewServiceContext(c)
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		auths.RegisterAuthsServer(grpcServer, server.NewAuthsServer(ctx))
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})

	// 注册服务
	registerOn := c.ListenOn
	// 如果是0.0.0.0，替换为本机IP以支持mac上的docker-consul和本地服务发现
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
