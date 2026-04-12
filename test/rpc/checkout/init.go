package checkout

import (
	"fmt"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/services/checkout/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var checkoutClient checkout.CheckoutServiceClient

func init() {
	conn, err := grpc.NewClient(fmt.Sprintf("0.0.0.0:%d", biz.CheckoutRpcPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	checkoutClient = checkout.NewCheckoutServiceClient(conn)
}
