# Go-Mall 项目测试文档

## 项目概述

Go-Mall 是一个基于 Go-Zero 微服务架构的现代电商平台，包含以下主要功能模块：

- **用户管理**：注册、登录、个人信息管理、地址管理
- **商品管理**：商品列表、商品详情
- **购物车**：添加商品、删除商品、修改数量
- **订单管理**：创建订单、查看订单、订单状态管理
- **优惠券**：优惠券领取、使用
- **支付**：支付功能
- **结算**：订单结算

## 测试环境准备

### 1. 启动后端服务

确保所有微服务和API服务正常运行。可以使用以下命令启动：

```bash
# 启动所有服务（包含依赖）
docker-compose up -d

# 或者启动最小化服务
./start-minimal.sh

# 或者单独启动各个服务
./start.sh
```

### 2. 启动前端

```bash
# 进入前端目录
cd frontend

# 启动前端服务器（需要安装 http-server 或使用其他方式）
# 如果没有安装 http-server，可以使用 Python 或 Node.js 启动
python3 -m http.server 3000

# 或使用 Node.js
npx http-server -p 3000
```

### 3. 数据库初始化

确保数据库已经初始化，表结构已经创建：

```bash
# 执行初始化脚本
mysql -u root -p < init_all_tables.sql

# 插入测试数据
mysql -u root -p < insert_test_data.sql
mysql -u root -p < insert_test_products.sql
```

## 前端功能测试

### 1. 首页测试

**功能说明**：展示平台介绍和主要功能入口

**测试步骤**：
1. 访问 `http://localhost:3000`
2. 确认页面显示"欢迎来到 Go-Mall"标题
3. 确认三个功能卡片显示（高性能、安全可靠、分布式事务）
4. 点击"浏览商品"按钮，确认跳转到商品列表页

### 2. 用户注册测试

**功能说明**：用户可以通过邮箱和密码注册新账户

**测试步骤**：
1. 在首页点击"注册"按钮
2. 输入邮箱（如：test@example.com）
3. 输入密码（至少6位）
4. 确认密码
5. 点击"注册"按钮
6. 确认显示"注册成功！"提示
7. 确认自动跳转回首页
8. 确认导航栏显示用户名

### 3. 用户登录测试

**功能说明**：用户可以通过邮箱和密码登录

**测试步骤**：
1. 在首页点击"登录"按钮
2. 输入邮箱（test@example.com）和密码
3. 点击"登录"按钮
4. 确认显示"登录成功！"提示
5. 确认自动跳转回首页
6. 确认导航栏显示用户名和"退出"按钮

**演示账号**：
- 邮箱：admin
- 密码：admin

### 4. 商品列表测试

**功能说明**：展示所有商品列表，支持搜索功能

**测试步骤**：
1. 点击导航栏"商品"按钮
2. 确认显示商品列表（包含演示数据）
3. 在搜索框输入关键词（如：iPhone）
4. 确认商品列表根据关键词筛选
5. 点击商品卡片，确认跳转到商品详情页

### 5. 商品详情测试

**功能说明**：展示单个商品的详细信息

**测试步骤**：
1. 在商品列表页点击任意商品卡片
2. 确认显示商品详细信息（名称、价格、描述、库存）
3. 点击"加入购物车"按钮，确认商品加入购物车
4. 点击"返回列表"按钮，确认返回商品列表页

### 6. 购物车测试

**功能说明**：管理购物车商品，支持添加、删除、修改数量

**测试步骤**：
1. 点击导航栏"购物车"按钮
2. 确认购物车显示已添加的商品
3. 点击"+"按钮增加商品数量
4. 点击"-"按钮减少商品数量
5. 点击"删除"按钮删除商品
6. 确认购物车总计金额正确计算
7. 点击"结算"按钮，确认跳转到订单页

### 7. 订单测试

**功能说明**：查看和管理用户订单

**测试步骤**：
1. 确保用户已登录
2. 点击导航栏"我的订单"按钮
3. 确认显示订单列表
4. 确认订单显示订单号、状态、商品信息、总价等

## 后端API测试

### 测试工具

推荐使用以下工具进行API测试：

1. **Postman**：图形化API测试工具
2. **curl**：命令行工具
3. **Go 测试**：后端单元测试

### 用户API测试

**基础路径**：`http://localhost:8001/douyin/user`

#### 1. 注册接口

```http
POST /register
Content-Type: application/json

{
  "email": "test@example.com",
  "password": "123456",
  "confirmPassword": "123456"
}
```

**预期响应**：
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "access_token": "xxx",
    "user": {
      "id": 1,
      "email": "test@example.com",
      "user_name": "test"
    }
  }
}
```

#### 2. 登录接口

```http
POST /login
Content-Type: application/json

{
  "email": "test@example.com",
  "password": "123456"
}
```

**预期响应**：
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "access_token": "xxx",
    "user": {
      "id": 1,
      "email": "test@example.com",
      "user_name": "test"
    }
  }
}
```

#### 3. 获取用户信息

```http
GET /info
Authorization: Bearer {token}
```

**预期响应**：
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "id": 1,
    "email": "test@example.com",
    "user_name": "test",
    "avatar": ""
  }
}
```

### 商品API测试

**基础路径**：`http://localhost:8002/douyin/product`

#### 1. 获取商品列表

```http
GET /list
```

**预期响应**：
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "products": [
      {
        "id": 1,
        "name": "iPhone 15",
        "description": "最新款苹果手机",
        "price": 7999,
        "stock": 100,
        "picture": "📱"
      }
    ]
  }
}
```

#### 2. 获取商品详情

```http
GET /?id=1
```

**预期响应**：
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "id": 1,
    "name": "iPhone 15",
    "description": "最新款苹果手机",
    "price": 7999,
    "stock": 100,
    "picture": "📱"
  }
}
```

### 购物车API测试

**基础路径**：`http://localhost:8003/douyin/carts`

#### 1. 添加购物车商品

```http
POST /
Content-Type: application/json
Authorization: Bearer {token}

{
  "product_id": 1,
  "quantity": 1
}
```

**预期响应**：
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "id": 1,
    "product_id": 1,
    "quantity": 1,
    "user_id": 1
  }
}
```

#### 2. 获取购物车列表

```http
GET /
Authorization: Bearer {token}
```

**预期响应**：
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "product_id": 1,
        "quantity": 1,
        "product": {
          "id": 1,
          "name": "iPhone 15",
          "price": 7999
        }
      }
    ]
  }
}
```

#### 3. 删除购物车商品

```http
DELETE /?id=1
Authorization: Bearer {token}
```

**预期响应**：
```json
{
  "code": 0,
  "msg": "success"
}
```

### 订单API测试

**基础路径**：`http://localhost:8004/douyin/order`

#### 1. 创建订单

```http
POST /
Content-Type: application/json
Authorization: Bearer {token}

{
  "items": [
    {
      "product_id": 1,
      "quantity": 1
    }
  ],
  "address_id": 1
}
```

**预期响应**：
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "id": 1,
    "order_no": "ORD123456",
    "total": 7999,
    "status": "pending"
  }
}
```

#### 2. 获取订单列表

```http
GET /list
Authorization: Bearer {token}
```

**预期响应**：
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "orders": [
      {
        "id": 1,
        "order_no": "ORD123456",
        "total": 7999,
        "status": "pending",
        "created_at": "2024-01-01T00:00:00Z"
      }
    ]
  }
}
```

## 后端单测运行

### 运行所有测试

```bash
cd /Users/fan/go-mall

# 运行所有测试
go test -v ./test/...

# 运行特定模块的测试
go test -v ./test/rpc/users/...
go test -v ./test/rpc/product/...
go test -v ./test/rpc/carts/...
go test -v ./test/rpc/order/...
```

### 运行单个测试文件

```bash
# 运行用户登录测试
cd /Users/fan/go-mall/test/rpc/users/login
go test -v users_test.go

# 运行商品测试
cd /Users/fan/go-mall/test/rpc/product
go test -v product_test.go
```

### 测试覆盖报告

```bash
# 生成测试覆盖报告
go test -coverprofile=coverage.out ./test/...
go tool cover -html=coverage.out -o coverage.html

# 查看测试覆盖报告
open coverage.html
```

## 常见问题排查

### 1. 前端无法连接后端

**问题**：前端请求后端API失败

**解决方案**：
- 检查后端服务是否正常运行
- 检查API服务端口是否正确（8001-8007）
- 检查防火墙是否允许连接
- 确认CORS配置正确

### 2. 数据库连接失败

**问题**：后端服务无法连接数据库

**解决方案**：
- 检查MySQL服务是否正常运行
- 检查数据库连接配置（.env文件）
- 确认数据库用户密码正确
- 检查数据库是否存在

### 3. 前端页面无法加载

**问题**：浏览器无法访问前端页面

**解决方案**：
- 检查前端服务器是否正常运行
- 确认端口号是否正确（3000）
- 检查防火墙是否允许访问
- 清除浏览器缓存

## 测试数据准备

### 用户测试数据

```sql
-- 测试用户
INSERT INTO users (email, password, user_name, created_at) VALUES
('test@example.com', '123456', '测试用户', NOW()),
('admin@example.com', 'admin123', '管理员', NOW()),
('test9@test.com', '1234567', '测试用户9', NOW());
```

### 商品测试数据

```sql
-- 测试商品
INSERT INTO products (name, description, price, stock, picture) VALUES
('iPhone 15', '最新款苹果手机，搭载 A17 芯片', 7999, 100, '📱'),
('MacBook Pro', '高性能笔记本电脑，M3 Pro 芯片', 14999, 50, '💻'),
('Nike Air Max', '舒适的运动鞋子，气垫设计', 899, 200, '👟');
```

### 地址测试数据

```sql
-- 测试地址
INSERT INTO user_address (user_id, receiver_name, receiver_phone, receiver_address) VALUES
(1, '张三', '13800138000', '北京市朝阳区'),
(1, '李四', '13900139000', '上海市浦东新区');
```

## 测试总结

通过以上测试步骤，可以全面验证Go-Mall平台的各项功能是否正常运行。建议在以下场景下进行测试：

1. 开发阶段：每次代码修改后运行相关测试
2. 发布前：进行完整的回归测试
3. 生产环境：定期进行监控和巡检

## 备注

- 测试过程中遇到的问题请记录下来，方便后续修复
- 可以根据实际业务需求扩展测试用例
- 建议使用自动化测试工具提高测试效率
