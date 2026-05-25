package find

import (
	"context"
	"fmt"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	carts "github.com/falconfan123/Go-mall/services/carts/pb"
	"github.com/falconfan123/Go-mall/test/rpc/internal/testenv"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var carts_client carts.CartClient

func initCarts() {
	conn, err := grpc.NewClient(testenv.ServiceAddr("carts", biz.CartsRpcPort),
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
