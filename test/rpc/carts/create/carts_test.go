package create

import (
	"context"
	"fmt"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/services/carts/carts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
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
	req := &carts.CartItemRequest{
		UserId:    8,
		ProductId: 8,
	}

	fmt.Printf("Sending RPC request: %+v\n", req)

	rsp, err := carts_client.CreateCartItem(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("CreateCartItem response:", rsp.StatusCode)
	t.Log("CreateCartItem success", rsp)
}
