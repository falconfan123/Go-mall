package response

import (
	"context"
	"github.com/falconfan123/Go-mall/common/consts/code"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

// Response a represents standard API response.
type Response struct {
	StatusCode int    `json:"code"`
	StatusMsg  string `json:"msg"`
}

// NewResponse creates a new Response.
func NewResponse(statusCode int, statusMsg string) *Response {
	return &Response{
		StatusCode: statusCode,
		StatusMsg:  statusMsg,
	}
}

// Fail writes a failure response to the http.ResponseWriter.
func Fail(w http.ResponseWriter, statusCode int) {
	var msg string
	switch statusCode {
	case code.RateLimitExceeded:
		msg = code.RateLimitExceededMsg
	case code.Fail:
		msg = code.FailMsg
	case code.ServerError:
		msg = code.ServerErrorMsg
	default:
		msg = code.FailMsg
	}
	httpx.OkJson(w, NewResponse(statusCode, msg))
}

// NewParamError writes a parameter error response to the http.ResponseWriter.
func NewParamError(ctx context.Context, w http.ResponseWriter, err error) {
	logx.Infow("params invalid", logx.Field("err", err))
	httpx.OkJsonCtx(ctx, w, NewResponse(code.Fail, code.FailMsg))
}
