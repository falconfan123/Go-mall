# 网关适配方案

## 当前问题
- 静态配置 12 个服务的 IP (如 `127.0.0.1:10001`)
- 不支持服务动态发现
- 重启网关需手动更新配置

## 解决方案

### 方案 1: 使用 K8s Service DNS (推荐)

修改 `services/gateway/etc/gateway.yaml`：

```yaml
Upstreams:
  - Grpc:
      # 使用 K8s DNS 格式
      Target: users-rpc.go-mall.svc.cluster.local:10001
    Mappings:
      - Method: post
        Path: /douyin/user/login
        RpcPath: users.Users/Login
```

### 方案 2: 使用 ConfigMap 动态注入

1. 创建 ConfigMap 存储网关配置
2. 通过 K8s volume mount 挂载到网关容器
3. 支持配置热更新

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: gateway-config
  namespace: go-mall
data:
  gateway.yaml: |
    Name: gateway
    Host: 0.0.0.0
    Port: 8888
    Upstreams:
      - Grpc:
          Target: users-rpc.go-mall.svc.cluster.local:10001
        Mappings: [...]
```

### 方案 3: 使用 Nginx Ingress (外部暴露)

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: go-mall-gateway
  namespace: go-mall
  annotations:
    kubernetes.io/ingress.class: nginx
    nginx.ingress.kubernetes.io/ssl-redirect: "false"
    # gRPC 代理配置
    nginx.ingress.kubernetes.io/backend-protocol: "GRPC"
spec:
  rules:
  - host: api.go-mall.local
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: gateway
            port:
              number: 8888
  tls:
  - hosts:
    - api.go-mall.local
    secretName: go-mall-tls
```

## 网关服务 K8s Manifest

```yaml
apiVersion: v1
kind: Service
metadata:
  name: gateway
  namespace: go-mall
spec:
  type: LoadBalancer  # 或 NodePort
  ports:
  - port: 80
    targetPort: 8888
    protocol: TCP
    name: http
  - port: 443
    targetPort: 8888
    protocol: TCP
    name: https
  selector:
    app: gateway

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gateway
  namespace: go-mall
spec:
  replicas: 2
  selector:
    matchLabels:
      app: gateway
  template:
    metadata:
      labels:
        app: gateway
    spec:
      containers:
      - name: gateway
        image: go-mall/gateway:latest
        ports:
        - containerPort: 8888
        volumeMounts:
        - name: config
          mountPath: /app/etc
          readOnly: true
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
        livenessProbe:
          httpGet:
            path: /health
            port: 8888
        readinessProbe:
          httpGet:
            path: /health
            port: 8888
      volumes:
      - name: config
        configMap:
          name: gateway-config
```

## 服务映射表

| 服务 | K8s Service 名称 | 端口 | 网关 Path 前缀 |
|-----|----------------|------|---------------|
| users | users-rpc | 10001 | /douyin/user |
| product | product-rpc | 10002 | /douyin/product |
| carts | carts-rpc | 10003 | /douyin/cart |
| order | order-rpc | 10004 | /douyin/order, /api/v1/order |
| checkout | checkout-rpc | 10005 | /api/v1/checkout |
| payment | payment-rpc | 10006 | /douyin/payment |
| inventory | inventory-rpc | 10007 | /api/v1/inventory |
| audit | audit-rpc | 10008 | (内部) |
| coupons | coupons-rpc | 10009 | (内部) |
| system | system-rpc | 10010 | /api/v1/system |
| activity | activity-rpc | 10011 | /api/v1/activity |
| admin | admin-rpc | 10012 | /api/v1/admin |