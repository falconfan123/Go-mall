package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/falconfan123/Go-mall/services/order/internal/mq/seckill"
	"github.com/falconfan123/Go-mall/services/order/internal/svc"
	order "github.com/falconfan123/Go-mall/services/order/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

const seckillLuaScript = `
local userId = KEYS[1]
local activityId = KEYS[2]
local nowTime = tonumber(ARGV[1])
local pathKey = ARGV[2]

local startTime = tonumber(redis.call('GET', 'act_start_limit'))
if startTime == nil then
    return {-4, "活动未配置"}
end
if nowTime < startTime then
    return {-1, "活动未开始"}
end

local validPath = redis.call('GET', 'act_' .. activityId .. '_path_' .. userId)
if validPath ~= pathKey then
    return {-3, "无效的路径"}
end

if redis.call('SISMEMBER', 'act_' .. activityId .. '_bought', userId) == 1 then
    return {-2, "您已购买"}
end

local stockKey = 'act_' .. activityId .. '_stock'
local stock = tonumber(redis.call('GET', stockKey))
if stock == nil or stock <= 0 then
    return {0, "已售罄"}
end

redis.call('DECR', stockKey)
redis.call('SADD', 'act_' .. activityId .. '_bought', userId)
redis.call('DEL', 'act_' .. activityId .. '_path_' .. userId)

return {1, "success"}
`

type SeckillLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSeckillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SeckillLogic {
	return &SeckillLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Seckill 秒杀下单
func (l *SeckillLogic) Seckill(in *order.SeckillRequest) (*order.SeckillResponse, error) {
	userId := int64(in.UserId)
	productId := int64(in.ProductId)
	pathKey := in.PathKey
	activityId := productId
	nowTime := time.Now().UnixMilli()

	logx.Infof("Seckill request: userId=%d, productId=%d, activityId=%d, pathKey=%s", userId, productId, activityId, pathKey)

	result, err := l.svcCtx.RedisClient.EvalCtx(l.ctx, seckillLuaScript,
		[]string{fmt.Sprintf("%d", userId), fmt.Sprintf("%d", activityId)},
		fmt.Sprintf("%d", nowTime), pathKey)
	if err != nil {
		logx.Errorf("seckill lua script error: %v", err)
		return &order.SeckillResponse{
			StatusCode: 1,
			StatusMsg:  "系统错误",
			Message:    "秒杀请求失败",
		}, nil
	}

	var luaResult []interface{}
	if resultBytes, ok := result.([]interface{}); ok {
		luaResult = resultBytes
	} else {
		resultStr, _ := json.Marshal(result)
		var parsed []interface{}
		json.Unmarshal(resultStr, &parsed)
		luaResult = parsed
	}

	if len(luaResult) < 2 {
		return &order.SeckillResponse{
			StatusCode: 1,
			StatusMsg:  "系统错误",
			Message:    "秒杀请求失败",
		}, nil
	}

	var retCode int
	var msg string

	switch v := luaResult[0].(type) {
	case float64:
		retCode = int(v)
	case int64:
		retCode = int(v)
	default:
		retCode = -4
	}

	switch v := luaResult[1].(type) {
	case string:
		msg = v
	default:
		msg = "未知错误"
	}

	switch retCode {
	case 1:
		orderID := l.generateOrderID(userId, productId)
		// 发送消息到 RabbitMQ 异步创建订单
		l.sendSeckillMessage(userId, productId, activityId, orderID)
		logx.Infof("Seckill success: order_id=%s, user_id=%d, product_id=%d", orderID, userId, productId)
		return &order.SeckillResponse{
			StatusCode: 0,
			StatusMsg:  "success",
			OrderId:    orderID,
			Message:    "秒杀成功，请尽快完成支付",
		}, nil
	case 0:
		return &order.SeckillResponse{StatusCode: 1, StatusMsg: msg, Message: "商品已售罄"}, nil
	case -1:
		return &order.SeckillResponse{StatusCode: 1, StatusMsg: msg, Message: "活动尚未开始"}, nil
	case -2:
		return &order.SeckillResponse{StatusCode: 1, StatusMsg: msg, Message: "您已购买过此商品"}, nil
	case -3:
		return &order.SeckillResponse{StatusCode: 1, StatusMsg: msg, Message: "秒杀路径无效，请重新获取"}, nil
	default:
		return &order.SeckillResponse{StatusCode: 1, StatusMsg: msg, Message: "秒杀失败"}, nil
	}
}

// ClearPurchasedRecord 清除用户的购买记录（用于测试或重试）
func (l *SeckillLogic) ClearPurchasedRecord(userId, activityId int64) error {
	key := fmt.Sprintf("act_%d_bought", activityId)
	_, err := l.svcCtx.RedisClient.SremCtx(l.ctx, key, fmt.Sprintf("%d", userId))
	return err
}

func (l *SeckillLogic) generateOrderID(userId, productId int64) string {
	return fmt.Sprintf("SK%d%d%d", time.Now().UnixMilli(), userId, productId)
}

func (l *SeckillLogic) sendSeckillMessage(userId, productId, activityId int64, orderID string) {
	// 如果 SeckillMQ 为 nil（RabbitMQ 连接失败），直接返回
	if l.svcCtx.SeckillMQ == nil {
		logx.Infof("SeckillMQ is nil, skipping message publish")
		return
	}

	msg := seckill.SeckillOrder{
		OrderID:    orderID,
		UserID:     userId,
		ProductID:  productId,
		ActivityID: activityId,
		Timestamp:  time.Now().UnixMilli(),
	}

	err := l.svcCtx.SeckillMQ.Publish(msg)
	if err != nil {
		logx.Errorf("failed to publish seckill message: %v", err)
	}
}
