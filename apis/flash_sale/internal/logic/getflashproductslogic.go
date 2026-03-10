package logic

import (
	"context"
	"time"

	"github.com/falconfan123/Go-mall/apis/flash_sale/internal/svc"
	"github.com/falconfan123/Go-mall/apis/flash_sale/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFlashProductsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFlashProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFlashProductsLogic {
	return &GetFlashProductsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFlashProductsLogic) GetFlashProducts(req *types.GetFlashProductsReq) (resp *types.GetFlashProductsResp, err error) {
	// 模拟秒杀商品数据
	products := []*types.FlashProduct{
		{
			ID:          101,
			Name:        "iPhone 15 Pro",
			Description: "最新款苹果手机，搭载 A17 芯片",
			Picture:     "📱",
			Price:       7999,
			FlashPrice:  5999,
			Stock:       50,
			Sold:        0,
			CreatedAt:   time.Now().Format(time.RFC3339),
			UpdatedAt:   time.Now().Format(time.RFC3339),
		},
		{
			ID:          102,
			Name:        "MacBook Pro 14\"",
			Description: "高性能笔记本电脑，M3 Pro 芯片",
			Picture:     "💻",
			Price:       14999,
			FlashPrice:  11999,
			Stock:       30,
			Sold:        0,
			CreatedAt:   time.Now().Format(time.RFC3339),
			UpdatedAt:   time.Now().Format(time.RFC3339),
		},
		{
			ID:          103,
			Name:        "Sony PS5",
			Description: "次世代游戏主机",
			Picture:     "🎮",
			Price:       3899,
			FlashPrice:  2999,
			Stock:       20,
			Sold:        0,
			CreatedAt:   time.Now().Format(time.RFC3339),
			UpdatedAt:   time.Now().Format(time.RFC3339),
		},
		{
			ID:          104,
			Name:        "AirPods Pro 2",
			Description: "最新款降噪耳机",
			Picture:     "🎧",
			Price:       1899,
			FlashPrice:  1499,
			Stock:       100,
			Sold:        0,
			CreatedAt:   time.Now().Format(time.RFC3339),
			UpdatedAt:   time.Now().Format(time.RFC3339),
		},
	}

	return &types.GetFlashProductsResp{
		Products: products,
		Total:    int64(len(products)),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}
