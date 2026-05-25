package logout

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
	resp, err := users_client.Logout(context.Background(), &users.LogoutRequest{
		UserId: user.UserID,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("logout success", resp)
}
