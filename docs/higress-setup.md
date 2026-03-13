# Higress 网关配置说明

## 概述

本项目使用 Higress 替换原有的 go-zero grpc-gateway。Higress 是阿里巴巴开源的基于 Nginx 的 API 网关，支持 gRPC 透传和 HTTP-gRPC 转换。

## 端口规划

| 服务 | 端口 | 说明 |
|------|------|------|
| Higress 网关 | 8888 | HTTP 网关入口（替换原 gateway） |
| Higress UI | 8889 | 管理界面 |
| Higress HTTPS | 8443 | HTTPS 网关入口 |

## 快速启动

### 1. 启动 Higress

```bash
cd /path/to/go-mall
./scripts/start-higress.sh
```

### 2. 访问 Higress UI

打开浏览器访问: **http://localhost:8889**

### 3. 配置路由

在 Higress UI 中进行以下配置：

#### 3.1 添加 McpBridge（后端服务）

创建以下 McpBridge 资源配置：

**users 服务**:
```yaml
apiVersion: networking.higress.io/v1
kind: McpBridge
metadata:
  name: users-service
  namespace: higress-system
spec:
  services:
    - name: users
      port: 10001
      protocol: grpc
```

**order 服务**:
```yaml
apiVersion: networking.higress.io/v1
kind: McpBridge
metadata:
  name: order-service
  namespace: higress-system
spec:
  services:
    - name: order
      port: 10004
      protocol: grpc
```

**product 服务**:
```yaml
apiVersion: networking.higress.io/v1
kind: McpBridge
metadata:
  name: product-service
  namespace: higress-system
spec:
  services:
    - name: product
      port: 10002
      protocol: grpc
```

**carts 服务**:
```yaml
apiVersion: networking.higress.io/v1
kind: McpBridge
metadata:
  name: carts-service
  namespace: higress-system
spec:
  services:
    - name: carts
      port: 10003
      protocol: grpc
```

**checkout 服务**:
```yaml
apiVersion: networking.higress.io/v1
kind: McpBridge
metadata:
  name: checkout-service
  namespace: higress-system
spec:
  services:
    - name: checkout
      port: 10005
      protocol: grpc
```

**payment 服务**:
```yaml
apiVersion: networking.higress.io/v1
kind: McpBridge
metadata:
  name: payment-service
  namespace: higress-system
spec:
  services:
    - name: payment
      port: 10006
      protocol: grpc
```

#### 3.2 添加 Http2Rpc（HTTP 到 gRPC 路由）

根据原有的 `gateway.yaml` 配置，创建以下 Http2Rpc 资源：

| HTTP 路径 | gRPC 服务 | gRPC 方法 |
|-----------|-----------|-----------|
| POST /douyin/user/login | users | Login |
| POST /douyin/user/register | users | Register |
| GET /douyin/user | users | GetUser |
| POST /douyin/user/address/add | users | AddAddress |
| GET /douyin/user/address/list | users | ListAddresses |
| GET /douyin/order/list | order | ListOrders |
| POST /douyin/order/create | order | CreateOrder |
| GET /douyin/order/detail | order | GetOrder |
| GET /douyin/product/list | product | GetAllProduct |
| GET /douyin/product | product | GetProduct |
| POST /douyin/product/upload | product | GetUploadURL |
| GET /douyin/product/list/cursor | product | ListProducts |
| POST /douyin/cart/add | carts | CreateCartItem |
| GET /douyin/cart/list | carts | CartItemList |
| POST /douyin/cart/delete | carts | DeleteCartItem |
| POST /douyin/cart/sub | carts | SubCartItem |
| POST /douyin/checkout/prepare | checkout | PrepareCheckout |
| POST /douyin/payment/create | payment | CreatePayment |

## 测试

配置完成后，可以使用以下命令测试：

```bash
# 测试用户登录
curl -X POST http://localhost:8888/douyin/user/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"test"}'
```

## 停止 Higress

```bash
docker-compose -f configs/docker-compose.higress.yml down
```

## 与原有 gateway 的区别

| 特性 | 原有 go-zero gateway | Higress |
|------|---------------------|---------|
| 配置方式 | YAML 文件 | UI + CRD |
| 服务发现 | 静态配置 | 动态发现(Consul/Nacos) |
| 性能 | 一般 | 基于 Nginx，性能更高 |
| UI | 无 | 完整的 Web UI |
| gRPC 支持 | 支持 | 支持透传和转换 |
