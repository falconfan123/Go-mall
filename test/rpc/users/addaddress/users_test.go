package addaddress

import (
	"context"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	users "github.com/falconfan123/Go-mall/services/users/pb"
	"github.com/falconfan123/Go-mall/test/rpc/internal/seed"
	"github.com/falconfan123/Go-mall/test/rpc/internal/testenv"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var users_client users.UsersClient

func initusers() {

	conn, err := grpc.NewClient(testenv.ServiceAddr("users", biz.UsersRpcPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	users_client = users.NewUsersClient(conn)
}

func TestUsersRpc(t *testing.T) {
	initusers()
	user := seed.CreateUser(t, users_client)
	//这里可以从token中获取user——id
	resp, err := users_client.AddAddress(context.Background(), &users.AddAddressRequest{

		RecipientName:   "test2",
		PhoneNumber:     "13803423123",
		Province:        "山东省",
		City:            "威海市",
		DetailedAddress: "环翠区",
		IsDefault:       true,
		UserId:          user.UserID,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("addsuccess", resp)
}
