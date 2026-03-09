package main

import (
	"flag"
	"fmt"
	"net/http"

	"strings"

	"jijizhazha1024/go-mall/common/utils/token"

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
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Access-Token, Refresh-Token, user_id, x-requested-with")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// Extract token and inject user_id header for RPC
			authHeader := r.Header.Get("Authorization")
			fmt.Printf("Auth header: %s\n", authHeader)
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					tokenStr := parts[1]
					claims, err := token.ParseJWT(tokenStr)
					if err == nil {
						// Inject user_id header
						r.Header.Set("user_id", fmt.Sprintf("%d", claims.UserID))
						r.Header.Set("Grpc-Metadata-User-Id", fmt.Sprintf("%d", claims.UserID))
						fmt.Printf("Injected user_id: %d\n", claims.UserID)
					}
				}
			}

			next(w, r)
		}
	})

	defer gw.Stop()
	fmt.Printf("Starting gateway at %s:%d...\n", c.Host, c.Port)
	gw.Start()
}
