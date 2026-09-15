## Why

库存写路径的缓存同步不一致：`ReturnInventory` 有 `AdjustInventoryCacheCtx` 同步，但 `DecreaseInventory` / `DecreasePreInventory` / `ReturnPreInventory` / `UpdateInventory` **均未同步缓存**，而 `GetInventory` 读缓存并返回缓存值（`getinventorylogic.go:47/72`）。后果：库存写后 `GetInventory` 返回陈旧值（DB 正确、缓存滞后），导致集成测试 `TestInventoryService_HighConcurrency` 偶发失败（读缓存旧值，DB 实测 sold=1000 正确）。这是缓存一致性缺陷，影响所有"写库存后读库存"的调用方。

## What Changes

- **统一写路径缓存同步**：为 `decreaseinventorylogic` / `decreasepreinventorylogic` / `returnpreinventorylogic` / `updateinventorylogic` 补齐 `AdjustInventoryCacheCtx`（对齐 `returninventorylogic` 既有做法），按各写路径正确取 delta（sold/total 方向）。
- **全量扫描写路径**：核对 model 层所有改 `sold`/`total` 的方法（`Batchdecrease`/`BatchReturn`/`DecreaseInventoryAtom`/`ReturnInventory`/`UpdateOrCreate` 等）对应的 logic 写路径无遗漏。
- **缓存失败降级一致**：缓存同步失败不阻断业务写（仅记日志），与 ReturnInventory 既有降级语义一致。
- **一致性验证**：写后 `GetInventory` == DB 值的一致性断言；`TestInventoryService_HighConcurrency` 多次重跑稳定通过。

## Capabilities

### New Capabilities
- `inventory-cache-consistency`: 库存读写路径的缓存一致性契约——任意库存写路径完成后，`GetInventory` 返回与 DB 一致的最新值；缓存同步失败降级不影响业务写正确性。

### Modified Capabilities
<!-- 无：本变更新增库存缓存一致性契约；不改 async-settlement-reliability 等既有能力。 -->

## Impact

- **服务代码**：`services/inventory/internal/logic/{decreaseinventorylogic,decreasepreinventorylogic,returninventorylogic,returnpreinventorylogic,updateinventorylogic}.go`（补缓存同步）；可能涉及 `internal/svc/cache.go`（delta 语义确认）。
- **测试**：inventory 单测（各写路径缓存同步覆盖）+ 集成测试 `TestInventoryService_HighConcurrency`。
- **门禁**：Integration 测试套 inventory 包转绿（阶段 2 前提之一）。
- **风险**：delta 方向取错会导致缓存与 DB 反向（需按写路径逐一核对 + 一致性断言）；缓存失败降级需与既有 ReturnInventory 行为一致。