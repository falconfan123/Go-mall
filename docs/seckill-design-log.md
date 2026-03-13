# 秒杀系统设计与实现工作日志

> 设计日期：2026-03-14
> 设计师：Claude
> 项目：Go-mall 秒杀系统

---

## 一、系统分析与设计思路

### 1.1 需求分析

根据设计文档，秒杀系统的核心目标是：
1. **时钟同步**：解决前端本地时间不准确导致的倒计时误差
2. **原子性控制**：防止超卖、确保一人一单
3. **路径隐藏**：动态生成下单路径，防止恶意刷接口

### 1.2 架构设计

```
┌─────────────────────────────────────────────────────────────────┐
│                         前端 (Front-End)                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────────┐  │
│  │ 时钟同步    │  │ 按钮状态    │  │ 接口调用顺序控制         │  │
│  │ /time API  │  │ 等待/临界/  │  │ token → order 串行化    │  │
│  │ 30s 校准   │  │ 触发期     │  │ 防抖 + 随机延迟 0-300ms │  │
│  └─────────────┘  └─────────────┘  └─────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                       API Gateway (APISIX)                       │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ 限流：IP 5req/s, UserID 1req/活动期                         ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      后端服务 (Back-End)                         │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐ │
│  │ System 服务  │  │ Activity 服务│  │ Order 服务          │ │
│  │ /system/time │  │ /activity/   │  │ /order/{path_key}   │ │
│  │ 返回Unix时间 │  │   token      │  │ 秒杀下单入口         │ │
│  └──────────────┘  └──────────────┘  └──────────────────────┘ │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────────┐│
│  │                    Redis + Lua 原子性控制                     ││
│  │  act_start_limit, prod_stock, bought_user_set, path_key    ││
│  └──────────────────────────────────────────────────────────────┘│
│                                                                 │
│  ┌──────────────────────────────────────────────────────────────┐│
│  │                    RabbitMQ 异步下单                          ││
│  │  Redis 成功后 → 消息队列 → 异步写 MySQL 订单表               ││
│  └──────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
```

---

## 二、详细设计

### 2.1 核心接口定义

| 接口 | 方法 | 功能 | 关键参数 |
|------|------|------|----------|
| `/api/v1/system/time` | GET | 返回服务器 Unix 毫秒时间戳 | `{ "now": 1740000000000 }` |
| `/api/v1/activity/token` | GET | 获取下单动态路径 (path_key) | Header: Authorization |
| `/api/v1/order/{path_key}` | POST | 核心秒杀下单接口 | Header: Authorization, Body: `{ "prod_id": 1 }` |

### 2.2 Redis Key 设计

| Key | 类型 | 说明 |
|-----|------|------|
| `act_start_limit` | String | 活动开始时间戳 (毫秒) |
| `act_{activity_id}_stock` | String | 商品库存数量 |
| `act_{activity_id}_bought` | Set | 已购买用户集合 |
| `act_{activity_id}_path_{user_id}` | String | 用户 path_key (有效期 1 分钟) |

### 2.3 Lua 脚本逻辑

```lua
-- 秒杀核心原子性脚本
-- KEYS[1]: user_id
-- KEYS[2]: activity_id
-- ARGV[1]: 当前服务器时间 (毫秒)
-- ARGV[2]: path_key

local userId = KEYS[1]
local activityId = KEYS[2]
local nowTime = tonumber(ARGV[1])
local pathKey = ARGV[2]

-- 1. 校验活动时间
local startTime = tonumber(redis.call('GET', 'act_start_limit'))
if nowTime < startTime then
    return {code = -1, msg = "活动未开始"}
end

-- 2. 校验 path_key
local validPath = redis.call('GET', 'act_' .. activityId .. '_path_' .. userId)
if validPath ~= pathKey then
    return {code = -3, msg = "无效的路径"}
end

-- 3. 校验重复购买
if redis.call('SISMEMBER', 'act_' .. activityId .. '_bought', userId) == 1 then
    return {code = -2, msg = "您已购买"}
end

-- 4. 校验并扣减库存
local stock = tonumber(redis.call('GET', 'act_' .. activityId .. '_stock'))
if stock <= 0 then
    return {code = 0, msg = "已售罄"}
end

redis.call('DECR', 'act_' .. activityId .. '_stock')
redis.call('SADD', 'act_' .. activityId .. '_bought', userId)
redis.call('DEL', 'act_' .. activityId .. '_path_' .. userId)

return {code = 1, msg = "success"}
```

### 2.4 前端实现要点

#### 时钟同步逻辑
```
1. 页面加载时调用 GET /api/v1/system/time
2. 计算偏移量: Offset = T_server - LocalTime_now
3. 展示时间 = LocalTime_current + Offset
4. 每 30 秒重新校准一次
```

#### 抢购按钮状态流转
```
等待期 → (倒计时 < 5s) → 临界期 → (倒计时 = 0) → 触发期
  │                                    │
  ▼                                    ▼
 按钮置灰                           点击后置灰2秒
 "距离开始还有 XX:XX"               "抢购中..."
```

---

## 三、实现计划

### Phase 1: System 服务 (时间同步服务)
- [x] 创建 services/system 目录结构
- [x] 实现 /api/v1/system/time 接口
- [x] 添加配置文件 etc/system.yaml
- [x] 集成 Consul 服务注册
- [x] 完成 protobuf 代码生成

### Phase 2: Activity 服务 (抢购 Token 服务)
- [x] 创建 services/activity 目录结构
- [x] 实现 /api/v1/activity/token 接口
- [x] path_key 生成逻辑 (MD5)
- [ ] 活动开始前 5s 生成 path_key 并存入 Redis (逻辑已实现)
- [ ] 添加限流逻辑

### Phase 3: Order 服务改造 (秒杀下单)
- [x] 新增 /api/v1/order/{path_key} 接口
- [x] 集成 Lua 脚本执行
- [ ] 添加 RabbitMQ 消息生产者 (TODO)
- [ ] 异步写入 MySQL 订单 (TODO)

### Phase 4: 网关配置
- [ ] APISIX 路由配置
- [ ] 限流策略配置
- [ ] 熔断降级配置

### Phase 5: 前端实现 (参考设计文档)
- [ ] 时钟同步组件
- [ ] 抢购按钮状态管理
- [ ] 接口调用顺序控制

---

## 四、关键代码片段

### 4.1 System 服务 - 时间接口

```go
// services/system/internal/logic/timelogic.go
func (l *TimeLogic) Time(req *types.TimeReq) (*types.TimeResp, error) {
    return &types.TimeResp{
        Now: time.Now().UnixMilli(),
    }, nil
}
```

### 4.2 Activity 服务 - Token 生成

```go
// services/activity/internal/logic/tokenlogic.go
func (l *TokenLogic) GenerateToken(userId int64, activityId int64) (string, error) {
    salt := "seckill_salt_2026"
    raw := fmt.Sprintf("%d_%d_%s", userId, activityId, salt)
    pathKey := md5.Sum([]byte(raw))
    return hex.EncodeToString(pathKey[:]), nil
}
```

### 4.3 Order 服务 - Lua 执行

```go
// services/order/internal/logic/seckilllogic.go
func (l *SeckillLogic) ExecuteSeckill(ctx context.Context, pathKey string, prodId int64) (*types.SeckillResp, error) {
    // 调用 Lua 脚本
    result, err := l.svcCtx.Redis.Eval(ctx, luaScript, []string{userId, activityId}, nowTime, pathKey).Result()
    // 处理结果
}
```

---

## 五、部署与测试

### 5.1 服务启动顺序
1. Redis (必须先启动)
2. RabbitMQ
3. Consul
4. System 服务
5. Activity 服务
6. Order 服务
7. Gateway (APISIX)

### 5.2 压测要点
- 单机 QPS 目标: 10000+
- 库存扣减一致性: 100%
- 抢购成功率: 库存售完为止

---

## 六、总结

本设计方案遵循了以下核心原则：

1. **前后端分离职责**：前端负责交互平滑，后端负责原子性校验
2. **Redis Lua 作为真理来源**：所有核心逻辑在 Redis 端原子执行
3. **路径动态化**：每次活动生成唯一 path_key，防止路径泄露
4. **异步削峰**：Redis 成功后立即返回，异步写入 MySQL

---

> 设计文档版本: v1.0
> 最后更新: 2026-03-14
