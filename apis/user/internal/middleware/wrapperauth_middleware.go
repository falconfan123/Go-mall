package middleware

import "net/http"

type WrapperAuthMiddleware struct {
}

func NewWrapperAuthMiddleware() *WrapperAuthMiddleware {
	return &WrapperAuthMiddleware{}
}

// Handle does something.
func (m *WrapperAuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO generate middleware implement function, delete after code implementation

		// Passthrough to next handler if need
		next(w, r)
	}
}
