package coupons

import (
	"fmt"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/services/coupons/coupons"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var couponsClient coupons.CouponsClient

func init() {
	conn, err := grpc.NewClient(fmt.Sprintf("0.0.0.0:%d", biz.CouponsRpcPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	couponsClient = coupons.NewCouponsClient(conn)
}
