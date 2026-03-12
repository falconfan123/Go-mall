# 抖音商城API文档

## 基础信息
- 基础URL：`http://localhost:8888/douyin`
- 认证方式：在请求头中携带 `Access-Token` 和 `Refresh-Token`

## 接口列表

### 1. 用户模块

#### 1.1 用户注册
- **接口地址**：`POST /user/register`
- **请求参数**：
  | 参数名 | 类型 | 必填 | 描述 |
  |--------|------|------|------|
  | email | string | 是 | 邮箱 |
  | password | string | 是 | 密码 |
  | confirmPassword | string | 是 | 确认密码 |
- **响应参数**：
  | 参数名 | 类型 | 描述 |
  |--------|------|------|
  | accessToken | string | 访问令牌 |
  | refreshToken | string | 刷新令牌 |
  | userId | integer | 用户ID |

#### 1.2 用户登录
- **接口地址**：`POST /user/login`
- **请求参数**：
  | 参数名 | 类型 | 必填 | 描述 |
  |--------|------|------|------|
  | email | string | 是 | 邮箱 |
  | password | string | 是 | 密码 |
- **响应参数**：
  | 参数名 | 类型 | 描述 |
  |--------|------|------|
  | accessToken | string | 访问令牌 |
  | refreshToken | string | 刷新令牌 |
  | userId | integer | 用户ID |

### 2. 商品模块

#### 2.1 获取商品列表
- **接口地址**：`GET /product/list`
- **查询参数**：
  | 参数名 | 类型 | 必填 | 默认值 | 描述 |
  |--------|------|------|--------|------|
  | page | integer | 否 | 1 | 页码 |
  | pageSize | integer | 否 | 100 | 每页数量 |
- **响应参数**：
  | 参数名 | 类型 | 描述 |
  |--------|------|------|
  | products | array | 商品列表 |
  | - id | integer | 商品ID |
  | - name | string | 商品名称 |
  | - description | string | 商品描述 |
  | - price | number | 商品价格 |
  | - stock | integer | 库存 |
  | - thumbnailUrl | string | 缩略图URL |

#### 2.2 获取文件上传预签名URL
- **接口地址**：`POST /product/upload`
- **请求参数**：
  | 参数名 | 类型 | 必填 | 描述 |
  |--------|------|------|------|
  | filename | string | 是 | 文件名 |
  | contentType | string | 是 | 文件类型 |
- **响应参数**：
  | 参数名 | 类型 | 描述 |
  |--------|------|------|
  | uploadUrl | string | 上传URL |
  | formData | object | 表单数据 |
  | key | string | 文件存储路径 |

### 3. 购物车模块

#### 3.1 添加商品到购物车
- **接口地址**：`POST /cart/add`
- **请求头**：
  | 头名 | 类型 | 必填 | 描述 |
  |------|------|------|------|
  | Access-Token | string | 是 | 访问令牌 |
- **请求参数**：
  | 参数名 | 类型 | 必填 | 描述 |
  |--------|------|------|------|
  | product_id | integer | 是 | 商品ID |
  | quantity | integer | 是 | 数量 |

#### 3.2 获取购物车列表
- **接口地址**：`GET /cart/list`
- **请求头**：
  | 头名 | 类型 | 必填 | 描述 |
  |------|------|------|------|
  | Access-Token | string | 是 | 访问令牌 |
- **响应参数**：
  | 参数名 | 类型 | 描述 |
  |--------|------|------|
  | data | array | 购物车列表 |
  | - id | integer | 购物车项ID |
  | - product_id | integer | 商品ID |
  | - quantity | integer | 数量 |

#### 3.3 减少购物车商品数量
- **接口地址**：`POST /cart/sub`
- **请求头**：
  | 头名 | 类型 | 必填 | 描述 |
  |------|------|------|------|
  | Access-Token | string | 是 | 访问令牌 |
- **请求参数**：
  | 参数名 | 类型 | 必填 | 描述 |
  |--------|------|------|------|
  | product_id | integer | 是 | 商品ID |
  | quantity | integer | 是 | 数量 |

#### 3.4 删除购物车商品
- **接口地址**：`POST /cart/delete`
- **请求头**：
  | 头名 | 类型 | 必填 | 描述 |
  |------|------|------|------|
  | Access-Token | string | 是 | 访问令牌 |
- **请求参数**：
  | 参数名 | 类型 | 必填 | 描述 |
  |--------|------|------|------|
  | product_id | integer | 是 | 商品ID |

### 4. 订单模块

#### 4.1 创建订单
- **接口地址**：`POST /order/create`
- **请求头**：
  | 头名 | 类型 | 必填 | 描述 |
  |------|------|------|------|
  | Access-Token | string | 是 | 访问令牌 |
- **请求参数**：
  | 参数名 | 类型 | 必填 | 描述 |
  |--------|------|------|------|
  | pre_order_id | string | 是 | 预订单ID |
  | coupon_id | string | 否 | 优惠券ID |
  | address_id | integer | 是 | 地址ID |
  | payment_method | integer | 是 | 支付方式：1-微信支付，2-支付宝 |
- **响应参数**：
  | 参数名 | 类型 | 描述 |
  |--------|------|------|
  | data.order_id | string | 订单ID |

#### 4.2 获取订单列表
- **接口地址**：`GET /order/list`
- **请求头**：
  | 头名 | 类型 | 必填 | 描述 |
  |------|------|------|------|
  | Access-Token | string | 是 | 访问令牌 |
- **响应参数**：
  | 参数名 | 类型 | 描述 |
  |--------|------|------|
  | data.orders | array | 订单列表 |
  | - order_id | string | 订单ID |
  | - items | array | 订单商品 |
  | - payable_amount | number | 应付金额 |
  | - order_status | integer | 订单状态：0-待支付，1-已支付，2-已发货，3-已完成，4-已取消，5-已退款，6-已关闭 |
  | - created_at | string | 创建时间 |

### 5. 结算模块

#### 5.1 结算准备
- **接口地址**：`POST /checkout/prepare`
- **请求头**：
  | 头名 | 类型 | 必填 | 描述 |
  |------|------|------|------|
  | Access-Token | string | 是 | 访问令牌 |
- **请求参数**：
  | 参数名 | 类型 | 必填 | 描述 |
  |--------|------|------|------|
  | coupon_id | string | 否 | 优惠券ID |
  | order_items | array | 是 | 订单商品列表 |
  | - product_id | integer | 是 | 商品ID |
  | - quantity | integer | 是 | 数量 |
  | address_id | integer | 是 | 地址ID |
- **响应参数**：
  | 参数名 | 类型 | 描述 |
  |--------|------|------|
  | data.pre_order_id | string | 预订单ID |

### 6. 支付模块

#### 6.1 创建支付
- **接口地址**：`POST /payment/create`
- **请求头**：
  | 头名 | 类型 | 必填 | 描述 |
  |------|------|------|------|
  | Access-Token | string | 是 | 访问令牌 |
- **请求参数**：
  | 参数名 | 类型 | 必填 | 描述 |
  |--------|------|------|------|
  | order_id | string | 是 | 订单ID |
  | payment_method | integer | 是 | 支付方式：1-微信支付，2-支付宝 |

### 7. 秒杀模块

#### 7.1 秒杀购买
- **接口地址**：`POST /flash/buy`
- **请求头**：
  | 头名 | 类型 | 必填 | 描述 |
  |------|------|------|------|
  | Access-Token | string | 是 | 访问令牌 |
- **请求参数**：
  | 参数名 | 类型 | 必填 | 描述 |
  |--------|------|------|------|
  | productId | integer | 是 | 商品ID |
  | quantity | integer | 是 | 数量 |
- **响应参数**：
  | 参数名 | 类型 | 描述 |
  |--------|------|------|
  | data.orderId | string | 订单ID |
  | data.orderNo | string | 订单号 |
  | data.total | integer | 总金额（分） |

## 错误码说明
| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 10001 | 认证失败 |
| 10003 | 令牌过期 |
| 10004 | 令牌续期成功 |
