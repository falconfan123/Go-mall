package logic

import (
	"context"

	"github.com/falconfan123/Go-mall/services/admin/internal/db"
	"github.com/falconfan123/Go-mall/services/admin/internal/svc"
	"github.com/falconfan123/Go-mall/services/admin/pb"
)

type GetActivityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetActivityLogic {
	return &GetActivityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetActivityLogic) GetActivity(in *pb.GetActivityRequest) (*pb.GetActivityResponse, error) {
	activity, err := db.GetActivityByID(l.svcCtx.DB, in.Id)
	if err != nil {
		return &pb.GetActivityResponse{
			StatusCode: 404,
			StatusMsg:  "activity not found",
		}, nil
	}

	return &pb.GetActivityResponse{
		StatusCode: 200,
		StatusMsg:  "success",
		Activity:   convertActivity(activity),
	}, nil
}

func convertActivity(a *db.Activity) *pb.Activity {
	if a == nil {
		return nil
	}
	return &pb.Activity{
		Id:           a.ID,
		Name:         a.Name,
		ProductId:    a.ProductID,
		SeckillPrice: a.SeckillPrice,
		TotalStock:   a.TotalStock,
		LimitPerUser: a.LimitPerUser,
		StartTime:    a.StartTime.Format("2006-01-02 15:04:05"),
		EndTime:      a.EndTime.Format("2006-01-02 15:04:05"),
		Status:       a.Status,
		CreatedAt:    a.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    a.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
