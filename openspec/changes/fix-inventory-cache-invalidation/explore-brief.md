# Explore Brief · fix-inventory-cache-invalidation

> 探索结论落盘（来源：fix-ci-integration-pipeline 的发现项 F2），作为提案审核需求基线。

## 背景（实证）

- **库存写路径缓存同步不一致**（`AdjustInventoryCacheCtx` 调用计数，`services/inventory/internal/logic/`）：
  - `returninventorylogic`：**3 处**（有同步）
  - `decreaseinventorylogic`：**0**、`decreasepreinventorylogic`：**0**、`returnpreinventorylogic`：**0**、`updateinventorylogic`：**0**（**均不同步缓存**）
- **读路径**：`GetInventory` 读缓存并返回缓存值——`getinventorylogic.go:47`（`GetInventoryCacheCtx`）、`:72`（`res.Inventory = cachedTotal`）。
- **后果**：扣减/预扣/更新库存后，缓存未同步 → `GetInventory` 返回陈旧值（**DB 正确，缓存滞后**）。
- **证据**：
  - DB 实测：`SELECT total,sold FROM inventory WHERE product_id=9999` → `total=0, sold=1000`（并发扣满，**无丢扣**）。
  - 集成测试 `TestInventoryService_HighConcurrency` 偶发失败（expected 0 / actual 270~790，5 次重跑 4 过 1 败）——断言读 `GetInventory`（缓存值），非 DB 真值。
  - 服务端日志：预扣失败 0 + 扣减失败 0（服务返回成功），失败纯为缓存读陈旧。
- **性质**：缓存一致性缺陷（**非并发丢扣**）；影响面覆盖所有"写库存后读库存"的调用方。

## 目标 / 非目标

- **目标**：库存写路径与缓存保持一致——任意写路径（扣减/预扣/回补/更新）完成后，`GetInventory` 返回最新值（与 DB 一致）。
- **非目标**：不改库存扣减正确性（DB 已正确）；不引入新缓存层/中间件；不改扣减幂等锁语义。

## 待决策点

- **D1 修复方向**（倾向 A）：
  - **A（推荐）写路径补同步**：为 `decreaseinventorylogic` / `decreasepreinventorylogic` / `returnpreinventorylogic` / `updateinventorylogic` 补齐 `AdjustInventoryCacheCtx`（对齐 `returninventorylogic` 既有做法）。改动面 = 4 个 logic + 各自 delta 语义；与既有模式一致。
  - **B GetInventory 读 DB**：`GetInventory` 改读 DB（放弃缓存读）。改动小但丢失缓存收益（读路径高频）。
  - **C 缓存失效**：写后删除缓存键（下次读回填）。需处理并发失效竞态。
- **D2 影响面（读后写一致性）**：`AdjustInventoryCacheCtx` 的 delta 语义（sold/total 增减方向）需按各写路径正确取值；需全量扫描其它可能改 `sold`/`total` 的 model 方法（`Batchdecrease`/`BatchReturn`/`DecreaseInventoryAtom`/`ReturnInventory`/`UpdateOrCreate` 等）确认无遗漏。
- **D3 并发测试稳定性**：`TestInventoryService_HighConcurrency` 应在修复后稳定通过；需给出"缓存与 DB 一致性断言"（写后 GetInventory == DB 值）与多次重跑证据。
- **D4 缓存不可用降级**：`AdjustInventoryCacheCtx` 失败时的行为（当前 ReturnInventory 记日志不阻断）——各写路径应保持一致（不因缓存失败回滚业务写）。

## Checklist（反思审核基线）

- [ ] C1 所有库存写路径（decrease/pre/return/pre-return/update）写后缓存与 DB 一致；全量扫描无遗漏写路径。
- [ ] C2 `GetInventory` 返回最新值（写后读一致性）；`AdjustInventoryCacheCtx` delta 方向正确（sold/total 不反向）。
- [ ] C3 缓存同步失败不阻断业务写（与 ReturnInventory 既有降级一致），仅记日志。
- [ ] C4 `TestInventoryService_HighConcurrency` 稳定通过（多次重跑 0 失败）+ 缓存/DB 一致性断言。
- [ ] C5 不改扣减正确性/幂等锁语义；不删测试/放宽断言；Rule 1 门禁（编译/单测/lint）通过。
- [ ] C6 新增/修改单测覆盖各写路径的缓存同步（mock 或真实 redis）。