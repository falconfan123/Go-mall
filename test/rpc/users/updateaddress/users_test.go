package updateaddress

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
	resp, err := users_client.UpdateAddress(context.Background(), &users.UpdateAddressRequest{

		RecipientName:   "djj",
		PhoneNumber:     "13800138000",
		Province:        "广东省",
		City:            "广州市",
		DetailedAddress: "天河区",
		IsDefault:       true,
		AddressId:       user.AddressID,
		UserId:          user.UserID,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("updatesuccess", resp)
}
