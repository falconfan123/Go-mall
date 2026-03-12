package logic

import (
	"github.com/falconfan123/Go-mall/apis/carts/internal/types"
	"github.com/falconfan123/Go-mall/services/carts/pb"
)

func ConvertCartInfoResponse(res []*carts.CartInfoResponse) []*types.CartInfoResponse {
	var result []*types.CartInfoResponse
	for _, item := range res {
		result = append(result, &types.CartInfoResponse{
			Id:        item.Id,
			UserId:    item.UserId,
			ProductId: item.ProductId,
			Quantity:  item.Quantity,
		})
	}
	return result
}
