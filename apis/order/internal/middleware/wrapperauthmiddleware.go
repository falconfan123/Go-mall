package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	// "github.com/zeromicro/go-zero/rest/httpx"
	"github.com/falconfan123/Go-mall/common/consts/biz"
	"github.com/falconfan123/Go-mall/common/utils/token"
)

type WrapperAuthMiddleware struct {
}

func NewWrapperAuthMiddleware() *WrapperAuthMiddleware {
	return &WrapperAuthMiddleware{}
}

func (m *WrapperAuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logx.Info("WrapperAuthMiddleware: Start handling request")
		// Directly return for debugging
		// httpx.OkJsonCtx(r.Context(), w, map[string]string{"status": "middleware_reached"})
		// return

		tokenString := r.Header.Get("Access-Token")
		if tokenString == "" {
			auth := r.Header.Get("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				tokenString = auth[7:]
			}
		}

		if tokenString != "" {
			logx.Infof("WrapperAuthMiddleware: Token found: %s", tokenString)
			claims, err := token.ParseJWT(tokenString)
			if err == nil && claims != nil {
				logx.Infof("WrapperAuthMiddleware: Token valid, UserID: %d", claims.UserID)
				ctx := r.Context()
				ctx = context.WithValue(ctx, biz.UserIDKey, claims.UserID)
				next(w, r.WithContext(ctx))
				return
			} else {
				logx.Errorf("WrapperAuthMiddleware: Token parse failed: %v", err)
			}
		} else {
			logx.Info("WrapperAuthMiddleware: No token found")
		}

		next(w, r)
	}
}
