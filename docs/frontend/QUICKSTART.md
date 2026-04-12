# Go-Mall 快速启动指南

## 🚀 一键启动

### 方式一：使用启动脚本（推荐）

```bash
cd /Users/fan/go-mall
./start-all.sh
```

这个脚本会同时启动：
- ✅ 前端服务器 (http://localhost:3000)
- ✅ 后端微服务
- ✅ 显示所有访问地址

### 方式二：分别启动

#### 1. 启动前端

```bash
cd /Users/fan/go-mall/frontend
python3 -m http.server 3000
# 然后访问: http://localhost:3000
```

#### 2. 启动后端

在另一个终端窗口：

```bash
cd /Users/fan/go-mall
go run run.go -services=auths,users,product,inventory
```

## 📱 访问地址

| 服务 | 地址 | 说明 |
|------|------|------|
| 🌐 前端界面 | http://localhost:3000 | 电商平台首页 |
| 🔧 Consul UI | http://localhost:8500 | 服务注册中心 |
| 🔍 Elasticsearch | http://localhost:9200 | 搜索引擎 |
| 🐰 RabbitMQ | http://localhost:15672 | 消息队列 (admin/admin) |

## 🎯 前端功能

### 用户模块
- ✅ 用户注册
- ✅ 用户登录
- ✅ 用户登出
- ✅ Token 持久化

### 商品模块
- ✅ 商品列表展示
- ✅ 商品搜索过滤
- ✅ 商品详情查看
- ✅ 商品信息展示

### 购物车模块
- ✅ 添加商品到购物车
- ✅ 更新商品数量
- ✅ 删除商品
- ✅ 购物车统计
- ✅ 结算功能

### 订单模块
- ✅ 订单列表
- ✅ 订单状态展示
- ✅ 订单详情查看

## 💻 技术栈

### 后端
- Go 1.21+
- Go-Zero 微服务框架
- MySQL 8.0
- Redis 6.0
- Elasticsearch 8.x
- Consul 服务发现
- RabbitMQ 消息队列
- DTM 分布式事务

### 前端
- HTML5
- CSS3 (响应式设计)
- Vanilla JavaScript (ES6+)
- LocalStorage 本地存储
- Fetch API 网络请求

## 📂 项目结构

```
go-mall/
├── services/           # 后端 RPC 服务
│   ├── auths/         # 认证服务
│   ├── users/         # 用户服务
│   ├── product/       # 商品服务
│   ├── inventory/     # 库存服务
│   ├── carts/         # 购物车服务
│   ├── order/         # 订单服务
│   ├── payment/       # 支付服务
│   ├── coupons/       # 优惠券服务
│   ├── checkout/      # 结算服务
│   └── audit/         # 审计服务
├── apis/              # API 网关
│   ├── user/          # 用户 API
│   ├── product/       # 商品 API
│   ├── carts/         # 购物车 API
│   ├── order/         # 订单 API
│   ├── payment/       # 支付 API
│   ├── coupon/        # 优惠券 API
│   └── checkout/      # 结算 API
├── frontend/          # 前端界面
│   ├── index.html     # 主页面
│   ├── styles.css     # 样式文件
│   └── app.js         # 应用逻辑
├── construct/         # 基础设施
│   └── depend/        # Docker 依赖
├── dal/               # 数据访问层
├── common/            # 公共模块
├── run.go             # 后端启动文件
├── start-all.sh       # 一键启动脚本
└── QUICKSTART.md      # 本文档
```

## 🛠️ 开发指南

### 添加新的前端页面

1. 在 `frontend/index.html` 中添加新的页面模板
2. 在 `frontend/app.js` 中添加页面逻辑
3. 在 `frontend/styles.css` 中添加样式

### 连接真实 API

编辑 `frontend/app.js` 中的 `API_BASE` 配置：

```javascript
const API_BASE = {
    user: 'http://localhost:8001/douyin/user',
    product: 'http://localhost:8002/douyin/product',
    // ...
};
```

## 🔧 故障排查

### 端口被占用

```bash
# 查找占用端口的进程
lsof -ti:3000 | xargs kill -9

# 或清理所有相关端口
lsof -ti:3000,10000,10001,10002,8001,8002 | xargs kill -9
```

### 前端无法访问

确认前端服务器正在运行：
```bash
ps aux | grep "python3.*http.server"
```

### 后端服务无法启动

1. 确认基础设施服务正在运行：
   - MySQL: localhost:3306
   - Redis: localhost:6379
   - Consul: localhost:8500
   - Elasticsearch: localhost:9200

2. 查看日志：
   ```bash
   tail -f /tmp/auths.log
   tail -f /tmp/users.log
   ```

## 📝 许可证

MIT License

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！
