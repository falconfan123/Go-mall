## Why

库存缓存键 `inventory:product:{pid}`（`svc/cache.go:12-13`）存在**双重语义冲突**：

- **预扣路径**把它当**可用库存（可售）**：`decreaselua.lua` 对库存键 `DECRBY`；`decreasepreinventorylogic.go:33` 注释明确"并非真实扣除库存，而是在缓存进行 -- 操作"；`returnlua.lua` 对称 `INCRBY`。
- **DB 镜像写入**把它当**账面 total**：`LoadInventoryFromDBToCache`（`svc/cache.go:43-49`）把 DB `total` 写入该键；`SetInventoryCacheCtx`（:16-17）由 `updateinventorylogic.go:46` 传入 `total`。

而 `GetInventory` 直接读该键返回（`getinventorylogic.go:47/72`），且缓存缺失回填分支用**函数开头的早期 DB 读值**（`getinventorylogic.go:69` ← `:35`）。后果：写后 `GetInventory` 返回的既非账面 total、也非正确可用库存，`TestInventoryService_HighConcurrency` 偶发失败（expected 0，实测 390/990/730，而 DB 恒 `total=0, sold=1000` 正确）。

**2B-0 调用方核证结论**：GetInventory 全部 6 处实质调用方（秒杀预热 `order/seckilllogic.go:167`、管理端展示 `admin/getinventorylogic.go:26`、管理端调整 `admin/adjustinventorylogic.go:27`、商品详情/列表/收藏 `product/getproductlogic.go:163` / `getallproductlogic.go:91` / `collection.go:64`）**均需要"可用库存（能卖多少）"**，无任何调用方依赖"账面 total 镜像"。故方向裁定 = **统一为「可用库存」语义**。

## What Changes

- **① 解除缓存键双重语义（统一为可用库存）**：声明 `inventory:product:{pid}` 口径为**可用库存**；DB 镜像类写入（`SetInventoryCacheCtx` / `LoadInventoryFromDBToCache` / `AdjustInventoryCacheCtx`，`svc/cache.go:16-73`）在**无在途预扣时 `DB total` 即等于可用库存**，写入时**声明可用库存口径 + 注释限制**（DB 无预留列，无法在回填时还原在途预扣；可执行性说明见 design D2）。**预热逻辑 `PreheatInventoryCache`（`servicecontext.go:70-86`）启动时用 `inv.Total` 无条件覆盖缓存键 → 改为只回填缺失键（不得覆盖在途预扣），否则统一语义后会放大超卖面**。方案对比（A 统一键 / B 双键）与推荐见 `design.md` D1。
- **② 读路径旧值修正**：`getinventorylogic.go:69` 回填分支使用 `:35` 的早期 DB 读值 → 改为使用**刚回填的缓存值/一致来源**。
- **③ 测试可判定性**：`test/rpc/inventory/inventory_test.go:310-323` 现仅对**传输层 error** 判失败，预扣 Lua `return 2`（库存不足，**body 层** `StatusCode=InventoryNotEnough`，RPC err=nil）被当作成功继续扣减 → 改为**校验预扣响应体状态**（不满足即 `t.Fail`、不继续扣减）。此条属测试写法，但缺了它本变更无法验收，**显式纳入范围**。

## Capabilities

### New Capabilities
- `inventory-cache-consistency`: 库存缓存一致性契约——缓存键 `inventory:product:{pid}` 口径统一为**可用库存**；任意写路径完成后 `GetInventory` 返回正确可用库存；读路径回填使用一致来源；缓存同步失败降级不影响业务写正确性。

### Modified Capabilities
<!-- 无：本变更新增库存缓存一致性契约；不改 async-settlement-reliability 等既有能力。 -->

## Impact

- **服务代码**：`services/inventory/internal/logic/getinventorylogic.go`（读路径旧值修正）、`services/inventory/internal/svc/cache.go`（可用库存口径）、`services/inventory/internal/svc/servicecontext.go`（预热只回填缺失键）、`services/inventory/internal/logic/{decreaseinventorylogic,decreasepreinventorylogic,returninventorylogic,returnpreinventorylogic,updateinventorylogic}.go`（写路径缓存口径核对）。
- **测试**：`test/rpc/inventory/inventory_test.go`（预扣响应体状态校验）。
- **门禁**：Integration 测试套 inventory 包（ES 栈修好前仍红，归因 = ES 栈启动，与本变更无关）。
- **风险**：可用库存口径若与预扣 Lua 现有 DECRBY/INCRBY 不一致会导致缓存与可用值反向（需按写路径逐一核对 + 一致性断言）；预留（预扣未确认量）仅存于缓存，回填时无法从 DB 还原（见 design D2 限制说明）；**服务重启预热若覆盖在途预扣会放大超卖面（已纳入 ① 修复，改为只回填缺失键）**；**确认扣减阶段若缓存键丢失，回填将用 DB total（含预扣量）导致可用库存虚高（已知限制，不修复，见 design）**。

## Non-Goals

- 不改 DB 扣减正确性（DB 恒正确：`total=0, sold=1000`）。
- 不修 Integration 的 ES 栈启动失败（另案 `fix-integration-stack-readiness`）。
- 不改 ruleset / 不把 Integration 设为 required（阶段 2 才做）。
- 不引入新的持久化预留表（预留仍为缓存态，见 design D2）。
