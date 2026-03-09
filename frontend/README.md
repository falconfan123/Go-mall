# Go-Mall 前端界面

## 概述

这是为 Go-Mall 微服务电商平台创建的简洁、美观的前端界面，展示了项目的各项功能。

## 功能特性

### 1. 用户认证
- **用户注册** - 新用户可以注册账号
- **用户登录** - 现有用户可以登录
- **用户登出** - 用户可以安全退出
- Token 自动保存到 localStorage

### 2. 商品浏览
- **商品列表** - 网格布局展示所有商品
- **商品搜索** - 实时搜索过滤商品
- **商品详情** - 查看商品详细信息
- 包含商品图片、价格、库存、描述等信息

### 3. 购物车
- **添加商品** - 将商品加入购物车
- **更新数量** - 增减购物车商品数量
- **删除商品** - 从购物车移除商品
- **购物车统计** - 显示购物车商品数量和总价
- **结算功能** - 模拟订单创建

### 4. 订单管理
- **订单列表** - 查看所有订单
- **订单状态** - 显示待支付、已支付、已发货、已完成等状态
- **订单详情** - 查看订单商品和金额

### 5. 界面设计
- 响应式布局，支持移动端
- 现代化渐变配色
- 流畅的动画效果
- Toast 消息提示
- 清晰的导航栏

## 文件结构

```
frontend/
├── index.html      # 主页面 HTML
├── styles.css      # 样式文件
├── app.js          # 应用逻辑
└── README.md       # 说明文档
```

## 技术栈

- **HTML5** - 页面结构
- **CSS3** - 样式和动画
- **JavaScript (ES6+)** - 应用逻辑
- **LocalStorage** - 本地存储
- **Fetch API** - 网络请求

## 使用方法

### 直接打开

直接用浏览器打开 `index.html` 文件即可：

```bash
open frontend/index.html
```

### 使用 HTTP 服务器（推荐）

```bash
# 使用 Python
cd frontend
python3 -m http.server 8080

# 或使用 Node.js
npx serve -p 8080

# 然后访问 http://localhost:8080
```

## API 配置

在 `app.js` 文件顶部配置 API 端点：

```javascript
const API_BASE = {
    user: 'http://localhost:8001/douyin/user',
    product: 'http://localhost:8002/douyin/product',
    cart: 'http://localhost:8003/douyin/carts',
    order: 'http://localhost:8004/douyin/order',
    // ...
};
```

## 演示数据

当前使用演示商品数据进行展示：

- iPhone 15 - ¥7999
- MacBook Pro - ¥14999
- Nike Air Max - ¥899
- Sony WH-1000XM5 - ¥2699
- iPad Pro - ¥8999
- Apple Watch - ¥2999

## 状态管理

应用状态包括：

- `user` - 当前登录用户信息
- `token` - JWT 认证令牌
- `cart` - 购物车商品列表
- `products` - 商品列表
- `orders` - 订单列表

## 页面路由

通过 `showPage(pageName)` 函数切换页面：

- `home` - 首页
- `login` - 登录页
- `register` - 注册页
- `products` - 商品列表页
- `product-detail` - 商品详情页
- `cart` - 购物车页
- `orders` - 订单页

## 后续优化

1. **真实 API 集成** - 连接到后端实际 API
2. **更多服务** - 添加优惠券、支付、结算等功能
3. **用户地址管理** - 完善收货地址功能
4. **图片上传** - 支持商品和用户头像上传
5. **更多筛选** - 添加价格区间、分类等筛选功能
6. **购物车持久化** - 将购物车保存到后端
7. **订单状态更新** - 支持订单状态流转

## 浏览器兼容性

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

## 许可证

MIT License
