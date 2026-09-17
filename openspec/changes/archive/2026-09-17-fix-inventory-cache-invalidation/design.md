## Context

见 `proposal.md`（Why）与 `explore-brief.md`（含「重探索：失败时刻实证 + 缓存键语义裁定」节）。核心事实：

- 缓存键 `inventory:product:{pid}` 被**双重语义**共用：预扣/回补 Lua 当**可用库存**（`decreaselua.lua` DECRBY / `returnlua.lua` INCRBY），DB 镜像写入当**账面 total**（`svc/cache.go:43-49` 写 `inventoryRecord.Total`）。
- `GetInventory` 读该键返回（`getinventorylogic.go:47/72`）；缓存缺失回填分支用早期 DB 读值（`:69` ← `:35`）。
- 实证（8 次重跑）：3 次 FAIL（actual 390/990/730），DB 恒 `total=0, sold=1000`（扣满无丢扣），预扣/扣减 RPC 传输层 err 计数 0，缓存键测试后为空。
- 2B-0 调用方核证：6 处实质调用方**全部需要可用库存**，无消费者需要账面 total 镜像。

## Goals / Non-Goals

- **Goals**：缓存键口径统一为可用库存；读路径回填一致；预扣响应体状态可判定；`TestInventoryService_HighConcurrency` ≥10 次 0 失败。
- **Non-Goals**：不改 DB 扣减正确性；不修 ES 栈（另案）；不改 ruleset / 不设 Integration required；不引入持久化预留表。

## Decisions

### D1 缓存键方案：A 统一键（推荐）vs B 双键

- **A 统一键（选定）**：`inventory:product:{pid}` 唯一口径 = **可用库存**。预扣/回补 Lua 维持 DECRBY/INCRBY（已是可用语义）；DB 镜像类写入改为按可用库存计算后写入；`GetInventory` 直接返回该键。
  - **理由**：2B-0 证明**无任何调用方需要账面 total 镜像**——双键会造出无消费者的键；改动最小（读路径 + 回填口径 + 写路径核对）；与预扣 Lua 既有语义天然一致。
- **B 双键（被否）**：`inventory:product:{pid}`（可用）+ `inventory:product:{pid}:reserved`（预留），回填时 `available = total − reserved`。
  - **被否代价**：多一个需同步维护的键（且预留仅存在于缓存，双键仍需定义预留的持久化，否则回填仍不精确）；无消费者；复杂度上升。
- **结论**：选 A；B 仅在"缓存丢失仍需精确还原在途预留"时才必要，本变更范围不需要。

### D1b 预热（`PreheatInventoryCache`）纳入范围：只回填缺失键

- 现状 `PreheatInventoryCache`（`servicecontext.go:70-86`）服务启动时遍历全量 DB 库存，`Rdb.Set(productKey, inv.Total)` **无条件覆盖**缓存键。统一可用库存语义后，若 Redis 存活而 inventory Pod 重启（滚动更新/迁移/OOM），预热会用 `total` 覆盖含在途预扣的可用库存 → **可用库存虚高（超卖面）**。
- **决定**：预热改为**只回填缺失键**（键不存在时才写），不覆盖已存在值；与 `EnsureInventoryCacheCtx` 的"缺失才回填"语义一致。此为既有隐患，统一语义后必须一并修复（否则本变更放大事故面）。

### D2 可用库存公式与预留限制

- 领域权威定义：`AvailableStock() = TotalStock − LockedStock`（`domain/aggregate/inventory.go:36-38`）。
- DB 仅有 `inventory.total` / `inventory.sold` 两列（`\d inventory` 实测），**无预留列**；预扣量仅存在于缓存（`decreaselua.lua` 只 DECRBY 库存键 + SET 幂等锁键，不写 DB）。
- 对齐公式（用于验收断言）：`可用库存期望值 = (DB total + DB sold) − DB sold − 在途预留 = DB total − 在途预留`。全部确认后 `在途预留 = 0` → 期望值 = `DB total`（测试卖光场景 = `0`）。
- **限制（显式声明）**：预留为**缓存态**，缓存冷启动（`LoadInventoryFromDBToCache`）时无在途预留可还原 → 回填可用库存 = `DB total`（即 `预留 = 0` 的冷启动假设）。此为既有设计约束，本变更不引入持久化预留表。
  - **可执行性说明**：`LoadInventoryFromDBToCache` 的改动 = 在 `SetInventoryCacheCtx(total)` 基础上**加语义注释**（注明"写入值 = 可用库存，冷启动预留=0 假设；DB 无预留列，无法在回填时还原在途预扣"），**不要求写出无法从 DB 计算的"公式"**。验收标准 = "写 total + 语义声明 + 注释"而非"真公式"。

### D3 写路径 × 可用库存增量（逐条核对，避免二次扣减/反向）

| 写路径 | DB 变更 | 对可用库存增量 | 缓存动作 |
|---|---|---|---|
| `DecreasePreInventory`（预扣） | 无 | `−q` | Lua `DECRBY`（已存在，保持） |
| `ReturnPreInventory`（回补预扣） | 无 | `+q` | Lua `INCRBY`（已存在，保持） |
| `DecreaseInventory`（确认扣减） | `total −= q, sold += q` | **`0`**（预扣已占用，确认不改变可用） | **不加缓存写**（修正旧"补 Adjust"错误处方，避免二次扣减） |
| `ReturnInventory`（回补） | `total += q, sold −= q` | `+q` | `AdjustInventoryCacheCtx(+q)`（已存在，保持；**barrier 与直连两条路径均调用**，见 `returninventorylogic.go:56/73/114`） |
| `UpdateInventory`（更新库存） | `UpdateOrCreate(total=q)` | 设为 `q`（全量 SET；无在途预扣时 = 可用库存） | `SetInventoryCacheCtx(q)`（已存在；口径注明为可用库存；**全量 SET 非增量**） |
| `PreheatInventoryCache`（服务启动预热） | 无 | 不改变（只回填缺失键） | 键缺失时 `Set(可用库存=total)`；**键已存在则跳过（不覆盖在途预扣）** |

> **已知限制（确认阶段缓存丢失）**：预扣 DECRBY 后、确认（DecreaseInventory）前若缓存键丢失（Redis 闪断/运维误删），因确认路径"不加缓存写"，下次 `GetInventory` 回填 DB `total`（该值**含**预扣量，预扣未写 DB）→ 可用库存虚高 → 可能超卖。此为既有"预留缓存态"限制的具体风险面；本变更**不处理**（需预留持久化，另案），spec 显式声明该场景不修复。

### D4 读路径旧值修正

- `getinventorylogic.go` 回填分支（`:56-69`）现用 `cachedTotal = inventoryResp.Total`（`:35` 早期 DB 读值）。
- **改为**使用**刚回填的缓存值**：让 `LoadInventoryFromDBToCache` **返回其写入的可用库存值**（避免再读一次缓存的多余 Redis 往返），回填分支以该返回值为准，保证返回值与缓存一致。

### D5 测试可判定性修正

- `inventory_test.go` 预扣调用后仅 `if preErr != nil` 判定（`:310-323`），而"库存不足"是响应体 `StatusCode = InventoryNotEnough`（`decreasepreinventorylogic.go` Lua `return 2`），传输层 err=nil。
- **改为**：预扣返回后校验 `resp.GetStatusCode()`；非成功（尤其 `InventoryNotEnough`）→ `t.Fail`（或 `t.Errorf`）并 `return`（不继续 `DecreaseInventory`）。

### D6 验证

- 单测/集成：`cd test/rpc && GOWORK=off GO_MALL_TEST_LOCAL=1 go test -count=10 ./inventory/...` → 0 失败（贴原始输出）。
- 一致性断言：写后 `GetInventory` == 可用库存期望值（`DB total − 在途预留`；测试卖光场景 = 0）。
- 门禁：`make build && make lint && make test-unit` 全绿。

## Risks / Trade-offs

- **[口径不一致导致反向漂移]** → D3 逐路径表 + D6 一致性断言；确认路径明确"不加缓存写"。
- **[服务重启预热覆盖在途预扣 → 超卖]** → D1b 改为只回填缺失键；spec 增加预热 Scenario。
- **[UpdateInventory 覆盖在途预扣]** → 已知限制（spec 声明）：管理端调整前确认无活跃预扣；本变更不引入预留持久化。
- **[预留缓存态限制]** → D2 显式声明；本变更不引入持久化预留。
- **[测试修正可能暴露更多预扣失败]** → 正是 D5 目的（可判定）；若修正后预扣确有失败，按 D1 口径修复缓存初始化后应消除。
- **[并发下缓存写竞争]** → `AdjustInventoryCacheCtx` 为原子 `IncrbyCtx`（`svc/cache.go:78`）；Lua DECRBY/INCRBY 原子。

## Migration Plan

1. 统一口径（D1/D3）：核对并修正 `svc/cache.go` 可用库存口径（`LoadInventoryFromDBToCache` 等）。
2. 读路径修正（D4）。
3. 测试可判定性（D5）。
4. 验证（D6）：`-count=10` + 一致性断言 + 门禁。

**回滚**：改动为 `services/inventory` + `test/rpc/inventory`，可 git revert。

## Open Questions

- 无在途预留时回填 = `DB total` 是否为可接受口径？——**本变更按 A 统一键采纳"冷启动预留=0"假设**（D2 限制已声明）；如需精确还原在途预留，另案引入预留持久化（超本变更范围）。
