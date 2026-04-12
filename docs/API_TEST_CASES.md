# Go-Mall 接口测试用例

## 测试账号信息
```
测试用户1:
邮箱: test@example.com
密码: 12345678
确认密码: 12345678

测试用户2:
邮箱: test2@example.com
密码: 12345678
确认密码: 12345678
```

---

## 1. 用户服务接口测试用例

### 1.1 用户注册
**接口地址**: POST /douyin/user/register
**请求参数**:
```json
{
  "email": "test@example.com",
  "password": "12345678",
  "confirmPassword": "12345678"
}
```
**预期结果**:
- 返回200状态码
- 返回access_token和refresh_token
- 用户信息成功写入数据库

### 1.2 用户登录
**接口地址**: POST /douyin/user/login
**请求参数**:
```json
{
  "email": "test@example.com",
  "password": "12345678"
}
```
**预期结果**:
- 返回200状态码
- 返回有效的access_token和refresh_token
- 可以使用access_token访问需要认证的接口

### 1.3 获取用户信息
**接口地址**: GET /douyin/user/info
**请求头**:
```
Authorization: Bearer <access_token>
```
**预期结果**:
- 返回200状态码
- 返回用户的详细信息，包括user_id、email、user_name等

### 1.4 更新用户信息
**接口地址**: PUT /douyin/user/update
**请求头**:
```
Authorization: Bearer <access_token>
```
**请求参数**:
```json
{
  "user_name": "测试用户",
  "avatar": "https://example.com/avatar.jpg"
}
```
**预期结果**:
- 返回200状态码
- 返回更新后的用户信息
- 数据库中用户信息已更新

### 1.5 添加用户地址
**接口地址**: POST /douyin/user/address
**请求头**:
```
Authorization: Bearer <access_token>
```
**请求参数**:
```json
{
  "recipient_name": "张三",
  "phone_number": "13800138000",
  "province": "北京市",
  "city": "北京市",
  "detailed_address": "朝阳区某某街道123号",
  "is_default": true
}
```
**预期结果**:
- 返回200状态码
- 返回新增地址的详细信息
- 地址成功添加到用户地址列表

### 1.6 获取地址列表
**接口地址**: GET /douyin/user/address/list
**请求头**:
```
Authorization: Bearer <access_token>
```
**预期结果**:
- 返回200状态码
- 返回用户的所有地址列表
- 包含刚刚添加的地址信息

### 1.7 用户登出
**接口地址**: POST /douyin/user/logout
**请求头**:
```
Authorization: Bearer <access_token>
```
**预期结果**:
- 返回200状态码
- 返回logout_at时间戳
- token失效，无法再使用该token访问需要认证的接口

---

## 2. 商品服务接口测试用例

### 2.1 获取商品列表
**接口地址**: GET /douyin/product/list
**请求参数**:
```
page: 1
size: 10
```
**预期结果**:
- 返回200状态码
- 返回商品列表，包含至少1个商品
- 每个商品包含id、name、price、stock等字段

### 2.2 获取商品详情
**接口地址**: GET /douyin/product/
**请求参数**:
```
id: 1
```
**预期结果**:
- 返回200状态码
- 返回商品ID为1的详细信息
- 包含商品名称、描述、价格、库存等完整信息

---

## 3. 购物车服务接口测试用例

### 3.1 添加商品到购物车
**前置条件**: 用户已登录，获取到access_token，存在商品ID为1的商品
**接口地址**: POST /douyin/cart/add
**请求头**:
```
Authorization: Bearer <access_token>
```
**请求参数**:
```json
{
  "product_id": 1
}
```
**预期结果**:
- 返回200状态码
- 返回新增购物车项的id
- 购物车中成功添加该商品，数量为1

### 3.2 再次添加同一商品到购物车
**接口地址**: POST /douyin/cart/add
**请求头**:
```
Authorization: Bearer <access_token>
```
**请求参数**:
```json
{
  "product_id": 1
}
```
**预期结果**:
- 返回200状态码
- 返回购物车项的id
- 购物车中该商品数量增加为2

### 3.3 获取购物车列表
**接口地址**: GET /douyin/cart/list
**请求头**:
```
Authorization: Bearer <access_token>
```
**预期结果**:
- 返回200状态码
- 返回购物车列表，包含之前添加的商品
- 商品数量为2，总价计算正确

### 3.4 减少购物车商品数量
**接口地址**: POST /douyin/cart/sub
**请求头**:
```
Authorization: Bearer <access_token>
```
**请求参数**:
```json
{
  "product_id": 1,
  "quantity": 1
}
```
**预期结果**:
- 返回200状态码
- 返回购物车项的id
- 购物车中该商品数量减少为1

### 3.5 删除购物车商品
**接口地址**: POST /douyin/cart/delete
**请求头**:
```
Authorization: Bearer <access_token>
```
**请求参数**:
```json
{
  "product_id": 1
}
```
**预期结果**:
- 返回200状态码
- 返回success: true
- 购物车列表中该商品被移除

---

## 4. 优惠券服务接口测试用例

### 4.1 获取优惠券列表
**接口地址**: GET /douyin/coupon/list
**请求参数**:
```
page: 1
size: 10
```
**预期结果**:
- 返回200状态码
- 返回优惠券列表，包含可用的优惠券信息

### 4.2 领取优惠券
**前置条件**: 存在优惠券ID为"coupon_001"的优惠券
**接口地址**: POST /douyin/coupon/claim
**请求头**:
```
Authorization: Bearer <access_token>
```
**请求参数**:
```json
{
  "coupon_id": "coupon_001"
}
```
**预期结果**:
- 返回200状态码
- 优惠券成功领取，加入用户的优惠券列表

### 4.3 获取我的优惠券列表
**接口地址**: GET /douyin/coupon/my/list
**请求头**:
```
Authorization: Bearer <access_token>
```
**预期结果**:
- 返回200状态码
- 返回用户已领取的优惠券列表
- 包含刚刚领取的优惠券

---

## 5. 结算服务接口测试用例

### 5.1 预结算
**前置条件**: 购物车中有商品，用户有可用地址
**接口地址**: POST /douyin/checkout/prepare
**请求头**:
```
Authorization: Bearer <access_token>
```
**请求参数**:
```json
{
  "coupon_id": "coupon_001",
  "order_items": [
    {
      "product_id": 1,
      "quantity": 2
    }
  ],
  "address_id": 1
}
```
**预期结果**:
- 返回200状态码
- 返回预订单ID（pre_order_id）
- 返回订单金额、优惠金额、实付金额等信息

### 5.2 获取结算详情
**前置条件**: 已生成预订单ID
**接口地址**: GET /douyin/checkout/detail
**请求头**:
```
Authorization: Bearer <access_token>
```
**请求参数**:
```
pre_order_id: <pre_order_id>
```
**预期结果**:
- 返回200状态码
- 返回预订单的详细信息
- 包含商品列表、金额信息、收货地址等

---

## 6. 订单服务接口测试用例

### 6.1 创建订单
**前置条件**: 已生成有效的预订单ID
**接口地址**: POST /douyin/order/create
**请求头**:
```
Authorization: Bearer <access_token>
```
**请求参数**:
```json
{
  "pre_order_id": "<pre_order_id>",
  "coupon_id": "coupon_001",
  "address_id": 1,
  "payment_method": 1
}
```
**预期结果**:
- 返回200状态码
- 返回订单ID（order_id）
- 订单成功创建，状态为待支付

### 6.2 获取订单列表
**接口地址**: GET /douyin/order/list
**请求头**:
```
Authorization: Bearer <access_token>
```
**请求参数**:
```
page: 1
page_size: 10
```
**预期结果**:
- 返回200状态码
- 返回用户的订单列表
- 包含刚刚创建的订单

### 6.3 获取订单详情
**前置条件**: 已创建订单
**接口地址**: GET /douyin/order/detail
**请求头**:
```
Authorization: Bearer <access_token>
```
**请求参数**:
```
order_id: <order_id>
```
**预期结果**:
- 返回200状态码
- 返回订单的详细信息
- 包含商品信息、金额信息、收货地址、订单状态等

### 6.4 取消订单
**前置条件**: 存在待支付的订单
**接口地址**: POST /douyin/order/cancel
**请求头**:
```
Authorization: Bearer <access_token>
```
**请求参数**:
```json
{
  "order_id": "<order_id>",
  "cancel_reason": "不想买了"
}
```
**预期结果**:
- 返回200状态码
- 返回订单ID
- 订单状态更新为已取消

---

## 7. 支付服务接口测试用例

### 7.1 创建支付订单
**前置条件**: 存在待支付的订单
**接口地址**: POST /douyin/payment/create
**请求头**:
```
Authorization: Bearer <access_token>
```
**请求参数**:
```json
{
  "order_id": "<order_id>",
  "payment_method": 1
}
```
**预期结果**:
- 返回200状态码
- 返回支付信息，包含支付链接或支付参数
- 支付订单成功创建

### 7.2 获取支付列表
**接口地址**: GET /douyin/payment/list
**请求头**:
```
Authorization: Bearer <access_token>
```
**请求参数**:
```
page: 1
page_size: 10
```
**预期结果**:
- 返回200状态码
- 返回用户的支付记录列表
- 包含刚刚创建的支付订单

---

## 8. 秒杀服务接口测试用例

### 8.1 获取秒杀商品列表
**接口地址**: GET /douyin/flash/products
**请求参数**:
```
page: 1
size: 10
```
**预期结果**:
- 返回200状态码
- 返回秒杀商品列表，包含秒杀价格、库存等信息

### 8.2 秒杀商品
**前置条件**: 存在可秒杀的商品ID为1，用户已登录
**接口地址**: POST /douyin/flash/buy
**请求头**:
```
Authorization: Bearer <access_token>
```
**请求参数**:
```json
{
  "product_id": 1,
  "quantity": 1
}
```
**预期结果**:
- 返回200状态码
- 秒杀成功时返回订单信息
- 秒杀失败时返回明确的错误信息（如库存不足、已抢完等）

---

## 测试流程建议

1. **用户模块**: 先测试注册 -> 登录 -> 获取用户信息 -> 更新用户信息 -> 添加地址 -> 获取地址列表 -> 登出
2. **商品模块**: 获取商品列表 -> 获取商品详情
3. **购物车模块**: 登录 -> 添加商品到购物车 -> 再次添加同一商品 -> 获取购物车列表 -> 减少商品数量 -> 删除商品
4. **优惠券模块**: 获取优惠券列表 -> 领取优惠券 -> 获取我的优惠券列表
5. **结算模块**: 添加商品到购物车 -> 预结算 -> 获取结算详情
6. **订单模块**: 创建订单 -> 获取订单列表 -> 获取订单详情 -> 取消订单
7. **支付模块**: 创建支付订单 -> 获取支付列表
8. **秒杀模块**: 获取秒杀商品列表 -> 秒杀商品

所有测试用例都使用真实的测试账号和真实的商品ID，确保测试场景符合实际使用情况。
