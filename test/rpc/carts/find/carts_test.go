package find

import (
	"context"
	"fmt"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/services/carts/pb"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var carts_client carts.CartClient

func initCarts() {
	conn, err := grpc.NewClient(fmt.Sprintf("0.0.0.0:%d", biz.CartsRpcPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	carts_client = carts.NewCartClient(conn)
}

func TestCartsRpc(t *testing.T) {
	initCarts()
	req := &carts.UserInfo{
		Id: 3,
	}

	fmt.Printf("Sending RPC request: %+v\n", req)

	rsp, err := carts_client.CartItemList(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("CartItemList response:", rsp.StatusCode)
	t.Log("CartItemList success", rsp)
}
