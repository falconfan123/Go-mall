package biz

import "time"

const (
	OrderRpcPort = 10004
)
const (
	OrderExpireTime = time.Minute * 30
)

const (
	// 秒杀活动相关 Redis Key
	SeckillStartTimeKey = "act_start_limit" // 秒杀活动开始时间
	SeckillStockKey     = "act_%d_stock"    // 秒杀商品库存 Key 模板
	SeckillStartKey     = "act_%d_start"    // 秒杀活动开始时间 Key 模板
	SeckillBoughtKey    = "act_%d_bought"   // 已购买用户集合 Key 模板
	SeckillPathKey      = "act_%d_path_%d"  // 用户下单路径 Key 模板

	// 秒杀缓存 TTL - 活动结束后多保留 1 天
	SeckillCacheTTL = 24 * time.Hour
)
