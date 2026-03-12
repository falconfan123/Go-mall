package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/common/utils/token"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/gateway"
)

var configFile = flag.String("f", "etc/gateway.yaml", "the config file")

func main() {
	flag.Parse()

	var c gateway.GatewayConf
	conf.MustLoad(*configFile, &c)

	gw := gateway.MustNewServer(c)

	// Add CORS and Token Injection middleware
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

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// 提取并验证长短令牌
			shortToken := r.Header.Get("Short-Token")
			longToken := r.Header.Get("Long-Token")

			var userID uint32
			var needRefresh bool

			// 1. 首先尝试验证短令牌
			if shortToken != "" {
				uid, _, _, err := token.VerifyShortToken(shortToken, biz.TokenSignSecret)
				if err == nil {
					// 短令牌验证成功
					userID = uid
					needRefresh = false
					fmt.Printf("Short token validated, user_id: %d\n", userID)
				} else {
					// 短令牌过期或无效，尝试长令牌
					fmt.Printf("Short token validation failed: %v, trying long token\n", err)
					if longToken != "" {
						sessionID, err := token.VerifyLongToken(longToken, biz.TokenSignSecret)
						if err == nil {
							fmt.Printf("Long token validated, session_id: %s\n", sessionID)
							// 长令牌验证成功，但需要刷新短令牌
							// 注意：这里需要调用 auths 服务来获取完整的用户信息和刷新短令牌
							// 为简化处理，我们这里先设置一个标记，在实际生产环境中应该调用 RPC
							needRefresh = true
						}
					}
				}
			} else if longToken != "" {
				// 没有短令牌，直接验证长令牌
				sessionID, err := token.VerifyLongToken(longToken, biz.TokenSignSecret)
				if err == nil {
					fmt.Printf("Long token validated, session_id: %s\n", sessionID)
					needRefresh = true
				}
			}

			// 如果验证成功，注入 user_id header
			if userID > 0 {
				r.Header.Set("user_id", fmt.Sprintf("%d", userID))
				r.Header.Set("Grpc-Metadata-User-Id", fmt.Sprintf("%d", userID))
			}

			// 如果需要刷新短令牌，在响应头中设置标记
			// 实际生产环境中应该调用 auths 服务的 ValidateToken 接口来获取新的短令牌
			if needRefresh {
				w.Header().Set("X-Need-Token-Refresh", "true")
			}

			next(w, r)
		}
	})

	defer gw.Stop()
	fmt.Printf("Starting gateway at %s:%d...\n", c.Host, c.Port)
	gw.Start()
}
