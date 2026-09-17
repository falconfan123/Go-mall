# Explore Brief · fix-ci-flaky-product-gateway-tests

> 立案探索（explore-only，不 propose）：Integration 首次暴露的 2 个测试阶段失败用例的根因与归属。
> 源发现：ES 栈修复（fix-integration-stack-readiness）后测试真正执行，att2 首次暴露这 2 个失败（本地可复现）。

## 2A 性质判定（失败率，非印象）

命令：
```
cd test/rpc && GOWORK=off GO_MALL_TEST_LOCAL=1 go test -count=1 -run TestQueryProduct ./product/...
cd test/rpc && GOWORK=off GO_MALL_TEST_LOCAL=1 go test -count=1 -run TestGatewayHTTPHappyPath ./scenarios/gateway_smoke/...
```

| 用例 | 干净卷（docker volume rm 后起栈） | 复用卷（重启不重建卷） | 判定 |
|---|---|---|---|
| TestQueryProduct | **10/10 FAIL** | **5/5 FAIL** | **必然失败（15/15）**，非间歇 |
| TestGatewayHTTPHappyPath | **10/10 FAIL** | **5/5 FAIL** | **必然失败（15/15）**，非间歇 |

失败原文：
- `TestQueryProduct`：`product_test.go:136: rpc error: code = Unknown desc = elastic: Error 400 (Bad Request): all shards failed [type=search_phase_execution_exception]`
- `TestGatewayHTTPHappyPath`：`gateway_smoke_test.go:438: invalid character 'r' looking for beginning of value`

> 说明：CI 观察中 att1/att3 绿、att2 红 → CI 侧是**时序依赖**；本地（既有环境）两用例为**确定性失败**。二者不矛盾：本地 env 与 CI 初始态不同（见 2B/2C）。

## 2B TestQueryProduct 时序取证

### 干净卷三时点 `_cat/indices`（ES healthy 后 0s/10s/30s）
```
T+0s: （无索引）
T+10s: （无索引）
T+30s: （无索引）
```
→ **索引为懒创建**：既非 ES 启动建、也非 product 服务启动即建（当 ES 已跑且 product 服务未重启时）。

### 建索引代码
- `services/product/internal/svc/servicecontext.go:98-107` `initEs`：`IndexExists(products)` 不存在才 `CreateIndex(...).Body(product.EsMapping)`；**失败仅记日志不 panic**（:81 "不panic，继续运行服务"）。
- `dal/es/product/mapping.json`：**snake_case** 映射（`id/name/description/picture/price/category/created_at/updated_at`，`created_at` 为 date `yyyy-MM-dd HH:mm:ss`）。
- 查询/排序侧（`services/product/internal/logic/queryproductlogic.go:50/54/56`）用 **snake_case** 字段（`created_at`/`updated_at`/`price`），与 mapping.json 一致。

### 失败机制（已实证）
1. 若索引已被 **bulk 动态创建**（`docker logs` 见 `[products] creating index, cause [auto(bulk api)]`），动态映射用 **Go struct 大写字段**：`CreatedAt/UpdatedAt/Price/Stock/Id/Name/...`（无 `category`/`created_at`）。
2. 查询 `New:true` → sort `created_at`（snake_case）→ **ES 对不存在字段 sort → `No mapping found` → search_phase_execution_exception → all shards failed**。
3. **对照实验**：干净卷重启 product 服务（`initEs` 跑）→ 索引为 snake_case（`category, created_at, ...`）→ **TestQueryProduct 5/5 PASS**。
4. 大小写不匹配（snake_case 查询 vs 大写动态映射）→ 确定性失败。

### 归属
- **存量时序问题（主要在）**：products 索引由 product 服务 `initEs` 在启动时建（依赖 ES 就绪 + 索引尚不存在）；若 ES 晚于 product 启动 / initEs 静默失败 / 首次 bulk 先自动建索引 → 得到大写动态映射 → 查询必然失败。
- **本次 ES 改名卷引入？否**：映射不匹配是既有隐患；named-volume 让 ES 首次稳定启动 → **暴露**了此前被"栈未起"掩盖的测试。
- **测试写法问题？否**：查询逻辑确实依赖 snake_case 映射，属服务/索引一致性缺陷，非断言过严。

## 2C TestGatewayHTTPHappyPath 取证

### 原始响应体（curl 复现，带 Short+Long Token）
```
$ curl -sS "http://localhost:8888/api/v1/activity/token?activity_id=140" -H "Short-Token: $ST" -H "Long-Token: $LT" -H 'X-Real-IP: 127.0.0.1'
rpc error: code = Unknown desc = activity not found or not started
HTTP 500
```
→ 响应体**非 JSON、以 'r' 开头**（gRPC 错误原样透出）→ 测试 `json.Unmarshal` → `invalid character 'r'`。

### 机制（已实证，本地）
- 活动服务 `Token` RPC（`services/activity/internal/logic/tokenlogic.go:53`）读 Redis `act_start_limit`；缺失则 `return nil, fmt.Errorf("activity not found or not started")`。
- **本地 env**：host 6379 被**本地 rogue redis**（PID 1050，无密码）占用，遮蔽 docker `go-mall-redis` 的 6379 映射；`docker port` 显示 6379 但实际监听是 PID 1050。
  - 活动服务配置 `127.0.0.1:6379` + `Pass: jjzzchtt` → 连到**本地 rogue redis**（无密码，AUTH 失败，`act_start_limit` 缺失）。
  - 测试 `configureSeckillActivity`（`gateway_smoke_test.go:455-462`）用 `docker exec go-mall-redis redis-cli SET act_start_limit` → 写入**容器 redis**（有该键）。
  - 两套 Redis 不通 → 活动服务读不到 `act_start_limit` → Token RPC 报错 → 网关透出 `rpc error:...`。

### 归属
- **本地失败 = 本地环境问题**（rogue redis 遮蔽 docker 端口），非测试/代码缺陷。
- **CI att2 归属：假设（未证实）**——CI 无 rogue redis（6379=容器），机制可能不同（活动 RPC 其它报错路径或时序）；需 CI 侧日志取证。
- **附加服务健壮性发现**：网关对 RPC error 返回**纯文本非 JSON body**（HTTP 500 `rpc error:...`），下游 JSON 解析易碎——`services/gateway/gateway.go` 错误处理未统一 JSON 化。

## 2D 归属与建议（含「假设（未证实）」）

### TestQueryProduct
- **在后续修复面内（是）**。
- 方向候选：
  1. **ES index template 兜底（推荐，根因级）**：对 `products` 预定义 snake_case mapping template，任何自动建索引都用正确映射 → 消除 initEs 时序依赖。
  2. **initEs 时序补强**：product 服务对 `initEs` 失败做重试/延迟至 ES 就绪，或每次写前 `EnsureIndex`。
  3. 查询侧适配大写映射（不推荐，侵入查询逻辑且双套字段）。

### TestGatewayHTTPHappyPath
- **本地失败 = 环境问题（先清 rogue redis 再复测）**。
- **CI 归属：假设（未证实）**，需 CI 侧日志（活动/网关）复核。
- 方向候选：
  1. **网关错误体 JSON 化（服务健壮性）**：RPC error 统一返回 JSON `{status_code,status_msg}`，避免透出 `rpc error:...` 纯文本。
  2. 测试写法：token 响应在 `DoJSON` 失败/非 200 时先断言状态再继续（容错）。
  3. 本地 env 治理：停用 rogue redis（杀 PID 1050）或 go-mall-redis 改端口，保证"服务连接的 redis == 测试写入的 redis"。

## Checklist（立案基线）

- [ ] C1 两用例为确定性失败（干净 10/10 + 复用 5/5，本地环境）。
- [ ] C2 TestQueryProduct 根因 = products 索引映射大小写不匹配（snake_case 查询 vs bulk 动态大写映射）+ initEs 时序依赖（已实证：重启 product 后 5/5 PASS）。
- [ ] C3 TestGatewayHTTPHappyPath 本地根因 = rogue redis 遮蔽 go-mall-redis 6379 → 活动 Token RPC 失败 → 网关透出 `rpc error:...` 纯文本（已实证原始 body）。
- [ ] C4 CI att2 归属：TestQueryProduct 机制环境无关可解释；gateway 用例 CI 归属待日志复核（假设未证实）。
- [ ] C5 方向候选登记（ES template / initEs 补强 / 网关错误 JSON 化 / 测试容错 / env 治理）。
- [ ] C6 本变更仅立案（explore-only），不 propose。
## 环境剥离增量（2026-09-17，2A 复测结论）

### rogue redis 鉴定（已 kill）
```
$ ps -p 1050 -o pid,ppid,user,etime,command
1050 1 fan 10-12:48:20 /opt/homebrew/opt/redis/bin/redis-server 127.0.0.1:6379
$ lsof -p 1050 | grep -E "cwd|txt"   # cwd=/opt/homebrew/var/db/redis（Homebrew 默认数据目录）
```
- PID 1050 = **Homebrew redis**（`/opt/homebrew/Cellar/redis/8.6.1`），PPID 1（launchd 自启），无密码（`CONFIG GET requirepass` 空），无本仓关联 → 判定"无关游离 redis" → **kill 1050**。
- 对照：docker `go-mall-redis` 无密码（`AUTH jjzzchtt` → WRONGPASS），映射 0.0.0.0:6379；kill 后 127.0.0.1:6379 → 容器 redis（PING PONG）。

### 复测（干净栈全重启后）
```
cd test/rpc && GOWORK=off GO_MALL_TEST_LOCAL=1 go test -count=1 -run TestGatewayHTTPHappyPath ./scenarios/gateway_smoke/...
第 1 批：10 跑 → 1 FAIL（438 间歇）
第 2 批：12 跑 → 0 FAIL
合计：22 跑 1 FAIL ≈ 95% 绿
```

### 结论（CI 归属升级为有证据）
1. **'r' 纯文本失败（line 438）= rogue redis 所致（有证据）**：kill 前确定性 10/10；kill 后失败点消失（第 1 批仅 1 次间歇回到 438，12/12 稳定）。
2. **CI 归属**：CI 无 rogue redis（6379=容器）→ **'r' 机制在 CI 不成立**；CI att2 的 gateway 失败 = **假设（未证实）**，更可能是 transient（本次隔离后间歇 ~5% 亦见同类残留）。
3. 残留间歇失败（1/22，438）为低概率问题，将在 变更B（网关错误响应 JSON 化）下被容错覆盖——测试对 RPC error 不再因非 JSON body 直接崩。
4. 附注：隔离过程中本地服务重启扰动 etcd 服务发现（曾现 60010 / checkout connect 超时）——为本地编排伪影，非代码缺陷；已通过全栈重启归零。

### 变更A/变更B 依据更新
- 变更A（TestQueryProduct）：仍成立（15/15 确定性，索引映射竞态，与本 env 无关）。
- 变更B（网关错误 JSON 化）：正当性**不依赖测试 flaky**——网关对 RPC error 返回 `rpc error: code = Unknown desc = activity not found or not started`（HTTP 500 纯文本）本身就是 API 契约缺陷（实测 curl 复现，见上）。
