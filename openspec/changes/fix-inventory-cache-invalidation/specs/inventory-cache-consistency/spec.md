## Purpose

库存读写路径的缓存一致性契约：任意库存写路径完成后，`GetInventory` 返回与 DB 一致的最新值；缓存同步失败降级不影响业务写正确性。

## ADDED Requirements

### Requirement: 库存写路径缓存一致性

所有修改库存（`sold`/`total`）的写路径（扣减、预扣、回补、预扣回补、库存更新）完成后 MUST 同步缓存，使随后 `GetInventory` 返回与 DB 一致的最新值。缓存同步的 delta 方向 MUST 与写路径对 `sold`/`total` 的实际变更一致，不得反向。

#### Scenario: 扣减后读库存一致
- **WHEN** 调用 `DecreaseInventory` 扣减某商品库存
- **THEN** 随后 `GetInventory` 返回的库存 MUST 等于 DB 中的最新值（缓存已同步，非陈旧）

#### Scenario: 预扣/回补/更新后读库存一致
- **WHEN** 调用 `DecreasePreInventory` / `ReturnPreInventory` / `UpdateInventory` 修改库存
- **THEN** 随后 `GetInventory` MUST 返回与 DB 一致的最新值

#### Scenario: 全量写路径无遗漏
- **WHEN** 审查所有改 `sold`/`total` 的 model 方法对应的 logic 写路径
- **THEN** 每个写路径 MUST 同步缓存，无遗漏

### Requirement: 缓存同步失败降级不影响业务写

缓存同步失败（如 redis 不可用）MUST NOT 阻断或回滚已完成的业务写（DB 已更新），仅记录日志；与既有 `ReturnInventory` 降级语义一致。

#### Scenario: 缓存失败不阻断写
- **WHEN** 库存写路径的 DB 更新成功而缓存同步失败
- **THEN** 业务写返回成功（不回滚），仅记录缓存同步失败日志

### Requirement: 并发读写一致性可验证

库存并发写场景下，写后读一致性 MUST 可被测试验证：`GetInventory` 在写完成后返回 DB 最新值，不因缓存滞后产生偏差。

#### Scenario: 并发扣减后最终一致
- **WHEN** 高并发扣减库存后读取库存
- **THEN** `GetInventory` 返回值 MUST 等于 DB 值（无缓存滞后偏差），并发用例稳定通过