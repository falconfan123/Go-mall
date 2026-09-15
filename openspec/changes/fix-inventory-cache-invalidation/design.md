## Context

- 见 `proposal.md` — Why 与 `explore-brief.md`。证据：`AdjustInventoryCacheCtx` 调用计数（return=3、decrease/pre/return-pre/update=0）；`GetInventory` 读缓存（getinventorylogic.go:47/72）；DB sold=1000 正确、集成测试偶发失败（读缓存旧值）。

## Goals / Non-Goals

- **Goals**：所有库存写路径写后缓存与 DB 一致；缓存失败降级不阻断业务写；一致性可验证。
- **Non-Goals**：不改扣减正确性（DB 已正确）、不改幂等锁语义、不引入新缓存层。

## Decisions

### D1 修复方向：写路径补同步（A），否 GetInventory 改读 DB（B）/ 缓存失效（C）

- **选定 A**：为 `decreaseinventorylogic` / `decreasepreinventorylogic` / `returnpreinventorylogic` / `updateinventorylogic` 补齐 `AdjustInventoryCacheCtx`（对齐 `returninventorylogic` 既有做法）。
  - 理由：与既有模式一致（return 已是此做法）；保留读路径缓存收益；改动局限在写路径。
- **被否代价**：
  - B（GetInventory 改读 DB）：改动小但丢弃读缓存收益（读高频）；且与既有缓存设计矛盾。
  - C（写后删缓存键）：需处理并发失效竞态（删-回填窗口读到 DB 旧值/重复回填）。

### D2 影响面：delta 语义 + 全量写路径扫描

- `AdjustInventoryCacheCtx(productID, delta)` 的 delta 方向须与写路径对库存的实际变更一致：
  - 扣减（sold+、total-）→ delta 取对应方向；预扣/预扣回补同理；更新库存按实际增量。
  - 实施时逐一核对每个写路径的 `sold`/`total` 变更方向，避免反向。
- 全量扫描 model 层改 `sold`/`total` 的方法（`Batchdecrease`/`BatchReturn`/`DecreaseInventoryAtom`/`ReturnInventory`/`UpdateOrCreate`），映射到 logic 写路径，确认无遗漏。

### D3 验证：一致性断言 + 并发重跑

- 新增/修改单测：各写路径写后 `GetInventory` == DB 值（mock redis 或真实 redis）。
- 集成测试 `TestInventoryService_HighConcurrency` 多次重跑（≥5 次）稳定通过 + 写后一致性断言。
- 断言口径：读 `GetInventory` 与 DB 直接查询值相等（消除缓存滞后偏差）。

### D4 降级：缓存失败不阻断业务写

- 各写路径缓存同步失败时仅记日志（与 `ReturnInventory` 既有降级一致），不回滚 DB 写。

## Risks / Trade-offs

- **[delta 方向取错]** → D2 逐一核对 + C2 一致性断言（写后读 == DB）。
- **[全量扫描遗漏写路径]** → C1 显式列出所有写路径与同步点；审查时核对 model 方法全集。
- **[缓存失败降级不一致]** → D4 统一为"记日志不阻断"，与既有 return 一致。
- **[并发下缓存写竞争]** → `AdjustInventoryCacheCtx` 若为原子增减（INCRBY/DECRBY 语义）则并发安全；若非原子（读-改-写）需评估——实施时确认其实现（svc/cache.go:73）。

## Migration Plan

1. 全量扫描写路径与 delta 语义（D2），产出"写路径 × delta × 同步点"清单。
2. 为 4 个写路径补齐 `AdjustInventoryCacheCtx`（D1）。
3. 补单测（各写路径写后读一致）+ 跑集成测试多次（D3）。
4. 门禁（Rule 1：编译/单测/lint）。
5. 验收：C1–C6 checklist 核对。

**回滚**：改动为 inventory logic + 测试，可 git revert。

## Open Questions

- `AdjustInventoryCacheCtx` 的并发原子性（影响 D3 并发断言设计）——实施首步确认 svc/cache.go:73 实现。