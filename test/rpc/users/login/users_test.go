package login

import (
	"context"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	auths "github.com/falconfan123/Go-mall/services/auths/pb"
	users "github.com/falconfan123/Go-mall/services/users/pb"
	"github.com/falconfan123/Go-mall/test/rpc/internal/seed"
	"github.com/falconfan123/Go-mall/test/rpc/internal/testenv"
	"sync"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var users_client users.UsersClient
var auths_client auths.AuthsClient
var once1 sync.Once

func initusers() {
	once1.Do(func() {
		conn, err := grpc.NewClient(testenv.ServiceAddr("users", biz.UsersRpcPort),
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			panic(err)
		}
		conn1, err := grpc.NewClient(testenv.ServiceAddr("auths", biz.AuthsRpcPort),
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			panic(err)

		}
		users_client = users.NewUsersClient(conn)
		auths_client = auths.NewAuthsClient(conn1)
	})
}

func TestUsersRpc(t *testing.T) {
	initusers()
	user := seed.CreateUser(t, users_client)

	// 测试用户名登录（推荐方式）
	resp, err := users_client.Login(context.Background(), &users.LoginRequest{
		Username: user.Username,
		Password: user.Password,
	})
	if err != nil {

		t.Fatal(err)
	}

	if resp.StatusCode == 0 {
		auths_res, err := auths_client.GenerateToken(context.Background(), &auths.AuthGenReq{
			UserId:   resp.UserId,
			Username: resp.UserName,
			ClientIp: "127.0.0.1",
		})
		if err != nil {
			t.Fatal(err)
		}
		t.Log("login success", resp)
		t.Log("auths success", auths_res)
	} else {
		t.Log("login failed", resp)
	}

}

// TestLoginWithEmail 测试邮箱登录（兼容模式）
func TestLoginWithEmail(t *testing.T) {
	initusers()
	user := seed.CreateUser(t, users_client)
	resp, err := users_client.Login(context.Background(), &users.LoginRequest{
		Email:    user.Email,
		Password: user.Password,
	})
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode == 0 {
		t.Log("email login success", resp)
	} else {
		t.Log("email login failed", resp)
	}
}
