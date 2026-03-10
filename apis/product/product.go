package main

import (
	"flag"
	"fmt"
	"strings"
	"github.com/falconfan123/Go-mall/apis/product/internal/config"
	"github.com/falconfan123/Go-mall/apis/product/internal/handler"
	"github.com/falconfan123/Go-mall/apis/product/internal/svc"
	"github.com/falconfan123/Go-mall/common/utils/ip"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/zero-contrib/zrpc/registry/consul"
)

var configFile = flag.String("f", "etc/product-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	server := rest.MustNewServer(c.RestConf, rest.WithCorsHeaders("Content-Type", "Access-Token", "Refresh-Token", "X-Real-IP", "X-Forward-For"))
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	// 注册服务到Consul
	registerOn := fmt.Sprintf("%s:%d", c.Host, c.Port)
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

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
