package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/common/consts/code"
	authsclient "github.com/falconfan123/Go-mall/services/auths/authsclient"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/gateway"
	"github.com/zeromicro/go-zero/zrpc"
)

var configFile = flag.String("f", "etc/gateway.yaml", "the config file")

var publicPaths = map[string]struct{}{
	"/api/v1/users/login":     {},
	"/api/v1/users/register":  {},
	"/api/v1/auth/refresh":    {},
	"/api/v1/products":        {},
	"/api/v1/products/detail": {},
	"/api/v1/system/time":     {},
	"/api/v1/payments/status": {},
	"/douyin/user/login":      {},
	"/douyin/user/register":   {},
	"/douyin/product/list":    {},
	"/douyin/product/detail":  {},
}

func main() {
	flag.Parse()

	var c gateway.GatewayConf
	conf.MustLoad(*configFile, &c)

	authRpcConf, err := findAuthRPC(c.Upstreams)
	if err != nil {
		panic(err)
	}
	authClient := authsclient.NewAuths(zrpc.MustNewClient(*authRpcConf))

	gw := gateway.MustNewServer(c)

	// 错误响应结构化中间件（最外层注册，包裹 auth 与上游 RPC 的一切输出）。
	// 上游/认证错误（HTTP >=400 且 body 非合法 JSON）→ 结构化 JSON（fix-gateway-error-response-format）。
	gw.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
			next(rec, r)
			rec.flush()
		}
	})

	gw.Use(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Short-Token, Long-Token, user_id, x-requested-with")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			if _, ok := publicPaths[r.URL.Path]; ok {
				next(w, r)
				return
			}

			shortToken := r.Header.Get(biz.TokenKey)
			longToken := r.Header.Get(biz.LongTokenKey)
			clientIP := resolveClientIP(r)
			if clientIP == "" {
				writeAuthError(w, code.IllegalProxyAddress, code.IllegalProxyAddressMsg)
				return
			}

			validateRes, err := authClient.ValidateToken(context.Background(), &authsclient.AuthValidateReq{
				ShortToken: shortToken,
				LongToken:  longToken,
				ClientIp:   clientIP,
			})
			if err != nil {
				writeAuthError(w, code.ServerError, code.ServerErrorMsg)
				return
			}

			if validateRes.StatusCode != code.Success {
				writeAuthError(w, int(validateRes.StatusCode), validateRes.StatusMsg)
				return
			}

			r.Header.Set("user_id", fmt.Sprintf("%d", validateRes.UserId))
			r.Header.Set("Grpc-Metadata-User-Id", fmt.Sprintf("%d", validateRes.UserId))
			next(w, r)
		}
	})

	defer gw.Stop()
	fmt.Printf("Starting gateway at %s:%d...\n", c.Host, c.Port)
	gw.Start()
}

func findAuthRPC(upstreams []gateway.Upstream) (*zrpc.RpcClientConf, error) {
	for _, upstream := range upstreams {
		if upstream.Name == "auths" && upstream.Grpc != nil {
			return upstream.Grpc, nil
		}
	}
	return nil, errors.New("auths upstream not configured")
}

func resolveClientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func writeAuthError(w http.ResponseWriter, statusCode int, statusMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	// 用 json.Marshal 生成，避免 statusMsg 含引号/换行时产出非法 JSON（fix-gateway-error-response-format）
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status_code": statusCode,
		"status_msg":  statusMsg,
	})
}

// responseRecorder 缓冲响应 status 与 body，在 flush 时统一决策：错误且非 JSON → 结构化包装；否则原样透出。
type responseRecorder struct {
	http.ResponseWriter
	status  int
	body    bytes.Buffer
	written bool
}

func (r *responseRecorder) WriteHeader(status int) {
	if r.written {
		return
	}
	r.written = true
	r.status = status
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	return r.body.Write(b)
}

const maxStatusMsgLen = 1024

func (r *responseRecorder) flush() {
	body := r.body.Bytes()
	w := r.ResponseWriter
	if r.status >= http.StatusBadRequest && !json.Valid(body) {
		msg := strings.TrimSpace(string(body))
		if msg == "" {
			msg = http.StatusText(r.status)
		}
		if len(msg) > maxStatusMsgLen {
			msg = msg[:maxStatusMsgLen]
		}
		w.Header().Del("Content-Length")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(r.status)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status_code": r.status,
			"status_msg":  msg,
			"trace_id":    newTraceID(),
		})
		return
	}
	w.WriteHeader(r.status)
	if len(body) > 0 {
		_, _ = w.Write(body)
	}
}

// newTraceID 生成请求级 32 位十六进制标识（时间 + 随机兜底）。
func newTraceID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
