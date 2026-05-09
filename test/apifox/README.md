# Apifox 接入

这套资产的目标是把测试入口固定到仓库里，让前端和后端都围绕测试层工作，而不是依赖手动点浏览器。

## 生成接口规范

```bash
node scripts/generate-apifox-openapi.mjs
```

生成产物:

- `test/apifox/go-mall-gateway.openapi.json`

这个文件同时包含两类接口:

- 网关里已经实现的真实后端路由
- 前端当前仍在调用、但后端未完全实现的 Mock 合同

这样前端可以先只对 Apifox Mock 测试，后端可以只对 Apifox 场景测试。

## Apifox App 导入

1. 打开 Apifox App
2. 导入 `test/apifox/go-mall-gateway.openapi.json`
3. 创建环境:

环境 `gateway-local`

- `baseUrl = http://127.0.0.1:18888`
- `activity_id = 1`
- `product_id = 1`
- `user_id = 1001`
- `address_id = 1`
- `device_id = device-demo-001`

环境 `mock`

- `baseUrl = https://mock.apifox.cn/m1/<your-project>`

## 前端切到测试层

前端不再写死真实网关地址，改为通过环境变量切换。

使用真实网关测试:

```bash
cp frontend/.env.gateway.example frontend/.env.local
```

使用 Apifox Mock/测试环境:

```bash
cp frontend/.env.apifox.example frontend/.env.local
```

然后把 `REPLACE_WITH_APIFOX_PROJECT` 改成实际的 Apifox Mock 地址。

## 后端通过 Apifox 跑链路

在 Apifox App 里基于导入的接口创建测试场景，例如:

1. `GET /api/v1/system/time`
2. `GET /api/v1/activity/token`
3. `POST /api/v1/order/seckill`
4. `GET /api/v1/order/detail`

创建好后，将测试场景导出为 `Apifox CLI` 格式，放到:

- `test/apifox/scenarios/go-mall-smoke.apifox-cli.json`

随后可直接命令行运行:

```bash
bash scripts/run-apifox.sh
```

输出报告目录:

- `test/apifox/reports`

## 网关本地暴露

如果你的网关跑在 Kubernetes 里，推荐把它转发到本地固定端口:

```bash
kubectl port-forward -n go-mall svc/gateway 18888:8888
```

这样前端和 Apifox 都统一打 `http://127.0.0.1:18888`。
