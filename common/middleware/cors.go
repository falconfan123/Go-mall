package middleware

import (
	"net/http"
)

func CorsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 设置 CORS 响应头
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Token, X-Real-IP, X-Forward-For")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// 如果是预检请求，直接返回 200 OK
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 继续处理其他请求
		next(w, r)
	}
}
