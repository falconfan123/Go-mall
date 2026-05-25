package auths

import (
	"context"
	"sync"
	"testing"

	"github.com/falconfan123/Go-mall/common/consts/biz"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	auths "github.com/falconfan123/Go-mall/services/auths/pb"
	"github.com/falconfan123/Go-mall/test/rpc/internal/testenv"

	"github.com/stretchr/testify/assert"
)

var client auths.AuthsClient
var once1 sync.Once
var clientIP string

func init() {
	// 获取客户端IP
	clientIP = "127.0.0.1"
}
func setupGRPCConnection(t *testing.T) {
	once1.Do(func() {
		conn, err := grpc.NewClient(testenv.ServiceAddr("auths", biz.AuthsRpcPort),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			t.Fatalf("Failed to connect to RPC server: %v", err)
		}
		client = auths.NewAuthsClient(conn)
	})
}

// 验证token
func TestAuthenticationLogic_Authentication(t *testing.T) {
	setupGRPCConnection(t)

	tokenResp, err := client.GenerateToken(context.Background(), &auths.AuthGenReq{
		UserId:   4,
		Username: "test",
		ClientIp: clientIP,
	})
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	assert.Equal(t, uint32(0), tokenResp.StatusCode)

	res, err := client.Authentication(context.Background(), &auths.AuthReq{
		Token: tokenResp.GetShortToken(), ClientIp: clientIP,
	})
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}
	assert.Equal(t, uint32(0), res.StatusCode)
	t.Log(res)
}

// 签发token
func TestAuthenticationLogic_GenerateToken(t *testing.T) {
	setupGRPCConnection(t)

	resp, err := client.GenerateToken(context.Background(), &auths.AuthGenReq{
		UserId:   4,
		Username: "test",
		ClientIp: clientIP,
	})
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	assert.Equal(t, uint32(0), resp.StatusCode)
	assert.NotEmpty(t, resp.GetShortToken())
	assert.NotEmpty(t, resp.GetLongToken())

	t.Log(resp)
}

// 续期token
func TestAuthenticationLogic_RenewToken(t *testing.T) {
	setupGRPCConnection(t)

	resp, err := client.GenerateToken(context.Background(), &auths.AuthGenReq{
		UserId:   4,
		Username: "test",
		ClientIp: clientIP,
	})
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	assert.Equal(t, uint32(0), resp.StatusCode)

	renewResp, err := client.RenewToken(context.Background(), &auths.AuthRenewalReq{
		LongToken:  resp.GetLongToken(),
		ShortToken: resp.GetShortToken(),
		ClientIp:   clientIP,
	})
	if err != nil {
		t.Fatalf("RenewToken failed: %v", err)
	}
	assert.Equal(t, uint32(0), renewResp.StatusCode)
	t.Logf("renew token is %s", renewResp.GetShortToken())
}
