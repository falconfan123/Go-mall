# inventory-cache-consistency Specification

## Purpose

库存缓存一致性契约：缓存键 `inventory:product:{pid}` 的口径统一为**可用库存（可售，available）**；任意库存写路径完成后，`GetInventory` 返回正确的可用库存；读路径回填使用一致来源；缓存同步失败降级不影响业务写正确性。

## Requirements

### Requirement: 缓存键口径统一为可用库存

库存缓存键 `inventory:product:{pid}`（`svc/cache.go:12-13`）MUST 表示**可用库存**（可售数量 = 账面总量 − 已预留未确认量）。所有写入该键的路径（预扣/回补 Lua、`SetInventoryCacheCtx`、`LoadInventoryFromDBToCache`、`AdjustInventoryCacheCtx`）MUST 以可用库存口径写入，不得写入账面 `total` 镜像。`GetInventory` MUST 返回该可用库存值。

#### Scenario: 预扣后读库存为可用库存期望值
- **WHEN** 某商品可用库存为 `A`，调用 `DecreasePreInventory` 预扣 `q`（`q <= A`）成功
- **THEN** 随后 `GetInventory` 返回值 MUST 等于 `A - q`（可用库存口径；**不得**等于账面 `total`）

#### Scenario: 回补预扣后读库存为可用库存期望值
- **WHEN** 已预扣 `q` 的商品调用 `ReturnPreInventory` 回补 `q` 成功
- **THEN** 随后 `GetInventory` 返回值 MUST 等于回补前的可用库存 `+ q`

#### Scenario: 缓存缺失回填使用一致来源
- **WHEN** 缓存键缺失、`GetInventory` 触发回填
- **THEN** 返回的库存值 MUST 来自**本次回填写入的缓存值**（一致来源），**不得**来自函数开头的早期 DB 读值

### Requirement: 写路径对可用库存的增量一致

各写路径对可用库存的增量 MUST 与其业务语义一致：确认扣减（`DecreaseInventory`，账面 `total -= q`、`sold += q`、预扣已占用）MUST NOT 再次减少可用库存（避免二次扣减）；回补（`ReturnInventory`）MUST 增加可用库存 `+ q`；库存更新（`UpdateInventory`）MUST 将可用库存设为新值（无在途预扣时）。

#### Scenario: 确认扣减不二次扣减可用库存
- **WHEN** 对已预扣 `q` 的商品调用 `DecreaseInventory` 确认扣减
- **THEN** 可用库存 MUST 保持不变（预扣阶段已扣减，确认阶段不再变化）
- **已知限制（不修复）**：若预扣后、确认前缓存键丢失（Redis 闪断/误删），确认路径不加缓存写，后续回填会用 DB `total`（含预扣量）→ 可用库存虚高。此为"预留缓存态"既有限制，本变更不处理（需预留持久化，另案）；验收不覆盖该场景

#### Scenario: 回补增加可用库存
- **WHEN** 调用 `ReturnInventory` 回补 `q`
- **THEN** 可用库存 MUST 增加 `q`，且随后 `GetInventory` 反映该增量

#### Scenario: 更新库存设为可用库存
- **WHEN** 无在途预扣时调用 `UpdateInventory` 将库存设为 `q`
- **THEN** 可用库存 MUST 变为 `q`，且随后 `GetInventory` 返回 `q`
- **已知限制**：若存在在途预扣（未确认预扣），`SetInventoryCacheCtx(q)` 会以 `q` 覆盖缓存（忽略在途预扣）；本变更不引入预扣持久化，管理端调整库存前 MUST 确认无活跃预扣（见 design D2 限制声明）

#### Scenario: 服务启动预热不覆盖在途预扣
- **WHEN** 服务启动执行 `PreheatInventoryCache`（`servicecontext.go:70-86`），且缓存键已存在（含在途预扣后的可用库存值）
- **THEN** 预热 MUST 只回填**缺失**的缓存键，MUST NOT 用 DB `total` 覆盖已存在的可用库存值

### Requirement: 缓存同步失败降级不影响业务写

缓存同步失败（如 redis 不可用）MUST NOT 阻断或回滚已完成的业务写（DB 已更新），仅记录日志；与既有 `ReturnInventory` 降级语义一致。

#### Scenario: 缓存失败不阻断写
- **WHEN** 库存写路径的 DB 更新成功而缓存同步失败
- **THEN** 业务写返回成功（不回滚），仅记录缓存同步失败日志

### Requirement: 预扣响应体状态可判定

预扣路径的"库存不足"是**响应体状态**（`decreasepreinventorylogic.go` Lua `return 2` → `StatusCode = InventoryNotEnough`，RPC 传输层 `err == nil`）。调用方/测试 MUST 校验响应体 `StatusCode`，不得仅以传输层 `err` 判定成功；预扣未成功时 MUST NOT 继续后续真实扣减。

#### Scenario: 预扣库存不足即判定失败且不继续扣减
- **WHEN** 预扣响应体 `StatusCode == InventoryNotEnough`
- **THEN** 测试 MUST 判定该用户预扣失败（`t.Fail`），且 MUST NOT 对该用户继续调用 `DecreaseInventory`

### Requirement: 并发读写一致性可验证

高并发"预扣 + 确认扣减"场景下，全部完成后 `GetInventory` MUST 等于可用库存口径期望值（由 DB `total` 与在途预留对齐：全部确认后 = `DB total`）。

#### Scenario: 并发预扣确认后最终一致
- **WHEN** `N` 个用户并发对可用库存 `N*q` 的商品各"预扣 `q` + 确认扣减 `q`"，且全部预扣成功
- **THEN** 全部完成后 `GetInventory` MUST 等于 `0`（可用库存卖光口径），且 `TestInventoryService_HighConcurrency` 连续 ≥10 次重跑 MUST 0 失败