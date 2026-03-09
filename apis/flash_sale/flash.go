package main

import (
	"flag"
	"fmt"
	"jijizhazha1024/go-mall/apis/flash_sale/internal/config"
	"jijizhazha1024/go-mall/apis/flash_sale/internal/handler"
	"jijizhazha1024/go-mall/apis/flash_sale/internal/svc"

	_ "github.com/zeromicro/zero-contrib/zrpc/registry/consul"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/flash-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	server := rest.MustNewServer(c.RestConf, rest.WithCorsHeaders("Content-Type", "Access-Token", "Refresh-Token", "refresh-token", "X-Real-IP", "X-Forward-For"))
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
