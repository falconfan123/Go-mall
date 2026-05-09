# Service Discovery

Go-mall 已切换为 etcd-only 服务发现。

服务端注册配置：

```yaml
Etcd:
  Hosts:
    - etcd.go-mall.svc.cluster.local:2379
  Key: order.rpc
```

客户端发现配置：

```yaml
OrderRpc:
  Etcd:
    Hosts:
      - etcd.go-mall.svc.cluster.local:2379
    Key: order.rpc
  NonBlock: true
  Timeout: 5000
```

本地开发时，可将 `Hosts` 改为 `localhost:2379`。

当前约定：

- 每个 RPC 服务使用 `<service>.rpc` 作为注册 key。
- `product` 服务沿用 `products.rpc`。
- Gateway 通过 etcd 发现后端，不再直连 Service DNS，也不再依赖 Consul。
