package coupons

import (
	"github.com/falconfan123/Go-mall/common/consts/biz"
	coupons "github.com/falconfan123/Go-mall/services/coupons/pb"
	"github.com/falconfan123/Go-mall/test/rpc/internal/testenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var couponsClient coupons.CouponsClient

func init() {
	conn, err := grpc.NewClient(testenv.ServiceAddr("coupons", biz.CouponsRpcPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	couponsClient = coupons.NewCouponsClient(conn)
}
