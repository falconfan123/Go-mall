package main
import (
  "fmt"
  ordercfg "github.com/falconfan123/Go-mall/services/order/internal/config"
  gzconf "github.com/zeromicro/go-zero/core/conf"
)
func main(){
  var c ordercfg.Config
  gzconf.MustLoad("/Users/fan/.superset/projects/Go-mall/.artifacts/ci-rpc-stack/configs/order.yaml", &c, gzconf.UseEnv())
  fmt.Printf("dns=%s\n", c.RabbitMQConfig.Dns())
  fmt.Printf("cfg=%+v\n", c.RabbitMQConfig)
}
