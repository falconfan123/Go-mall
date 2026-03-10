package main

import (
	"flag"
	"fmt"

	"github.com/falconfan123/Go-mall/apis/coupon/internal/config"
	"github.com/falconfan123/Go-mall/apis/coupon/internal/handler"
	"github.com/falconfan123/Go-mall/apis/coupon/internal/svc"

	_ "github.com/zeromicro/zero-contrib/zrpc/registry/consul"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/coupon-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	server := rest.MustNewServer(c.RestConf, rest.WithCorsHeaders("Content-Type", "Access-Token", "Refresh-Token", "X-Real-IP", "X-Forward-For"))
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
