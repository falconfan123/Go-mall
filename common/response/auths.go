package response

import (
	"github.com/falconfan123/Go-mall/common/consts/code"
)

type RefreshResponse struct {
	Response
	Data interface{} `json:"data"`
}

func NewRefreshResponse(data interface{}) RefreshResponse {
	return RefreshResponse{
		Response: Response{
			StatusCode: code.TokenRenewed,
			StatusMsg:  code.TokenRenewedMsg,
		},
		Data: data,
	}
}
