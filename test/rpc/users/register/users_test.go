package register

import (
	"context"
	"fmt"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	users "github.com/falconfan123/Go-mall/services/users/pb"
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
	name := testenv.UniqueName("register")
	// 测试用户名注册
	resp, err := users_client.Register(context.Background(), &users.RegisterRequest{
		Username:        name,
		Email:           fmt.Sprintf("%s@example.com", name),
		Password:        "password123",
		ConfirmPassword: "password123",
	})
	if err != nil {
		t.Error(err)
	}
	fmt.Println("register success", resp)
	t.Log("register success", resp)
}
