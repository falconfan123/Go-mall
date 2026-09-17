# Review Log · fix-inventory-cache-invalidation

## 审查轮次 1 — 2026-09-16（proposal + specs + design + tasks，方向级）

**审核方式**：@openspec-reviewer 逐路径核对证据代码（updateinventorylogic / decreasepreinventorylogic / returnpreinventorylogic / svc/cache.go / Lua 脚本 / inventorymodel.go）。

### 🔴 遗留（4 条，方向级，未冻结）
1. **"4 个写路径均未同步缓存"与代码不符（事实错误）**：
   - `updateinventorylogic.go:46` **已调** `SetInventoryCacheCtx`（写后 set 全量值）→ proposal 事实错误，task 2.4 为 no-op。
   - `decreasepreinventorylogic.go:78` / `returnpreinventorylogic.go:60` **不写 DB**，而是 Lua 原子改**同一缓存键**（`decreaselua.lua:19` DECRBY，键 = `biz.InventoryProductKey:pid`）→ 预扣的"同步"就是 Lua 本身；task 2.2/2.3"补 AdjustInventoryCacheCtx"会**二次扣减缓存**。
2. **核心修复与预扣 Lua 流双重语义冲突（方向级）**：同一缓存键兼两义——预扣路径当"可用库存"（`decreasepreinventorylogic.go:33` 注释），ReturnInventory 当"DB total 镜像"（delta=+q ↔ inventorymodel.go:156 total+q）。标准生命周期（预扣 Lua -10 → DecreaseInventory DB-10 → 补偿 ReturnInventory DB+10/缓存+10）：现状缓存==DB；**加修复后**缓存 = DB-10（补偿对不抵消）；高并发下缓存终值 -1000 ≠ DB 0，测试从"偶发失败"变"必失败"。制品未区分"镜像语义"与"可用库存语义"。
3. **spec Scenario"预扣/回补/更新后 GetInventory == DB"不可达成**：预扣只动缓存不动 DB，活跃预留时缓存(available) 必然 ≠ DB.total → C2 不可判定；强行验收会逼出放宽断言（触 C5 红线）。
4. **根因叙事未证明唯一**：(a) DB sold=1000 与失败断言 actual 270~790 若同一运行则自相矛盾；(b) 失败候选机制（缓存键缺失/淘汰 + 测试在预扣 body-status 失败后仍继续扣减，inventory_test.go:310-323 仅传输层 error 时 return）**不在本变更修复面**；(c) `getinventorylogic.go:69` 回填分支用 :35 较早的 DB 读值（本身是读路径陈旧来源）。需补"失败时刻"实证（同一运行内 缓存值/DB 值/预扣 Lua 返回码计数）+ 缓存键语义裁定。

### 🟡 建议（记录）
1. task 2.1 需区分 DecreaseInventory 的 barrier/Saga 路径（:51-70，生产结算走此）与直连路径（:74-98），ReturnInventory 既有做法是 3 处全补。
2. `AdjustInventoryCacheCtx` 原子性可关闭 Open Question：`svc/cache.go:78` 是 `Rdb.IncrbyCtx`（Redis INCRBY 原子）。
3. `getinventorylogic.go:69` 回填陈旧读纳入扫描或声明不在范围。
4. task 4.3 白名单过死（若需 dal/读路径修复会误报）。
5. spec "delta 方向"表述歧义：缓存只存 total，delta 仅对应 total 方向。

### 结论
**4 🔴 未清零 → 本批不冻结**。审核建议：先补"失败时刻实证 + 缓存键语义裁定"两项，再定 D1 方向。F2 需重新 explore（根因/缓存语义）。

## 审查轮次 2 — 2026-09-16（方向修订后整体重审，冻结）

**审核方式**：@openspec-reviewer 对完整变更包（proposal/design/tasks/specs/.openspec.yaml）重审（备选通道脚本）。方向已由用户裁定 = 统一「可用库存」语义。

### 修复循环（同一门禁内 3 次输出）
1. 首轮：🔴 1（`PreheatInventoryCache` 无条件覆盖缓存键 → 统一语义后超卖面）+ 5 🟡 → 修复：proposal/design/spec/tasks 纳入预热"只回填缺失键"（design D1b）、UpdateInventory 已知限制、spec 预热 Scenario、tasks 1.1b/5.1。
2. 次轮：🔴 3（tasks 缺写路径单测覆盖闭合 explore-brief C6；`LoadInventoryFromDBToCache` 可执行性模糊；确认阶段缓存丢失边界未声明）+ 3 🟡 → 修复：tasks 1.4 单测覆盖、design D2 可执行性说明、spec/design/proposal 已知限制、tasks 1.2b/2.1 连带。
3. 末轮：**🔴 无** → 冻结。

### 🟡 遗留（记录，不阻塞冻结）
1. `DecreaseInventory` 直连路径"不加缓存写"仅在"存在前置预扣"时正确；已核实**活调用方（order 结算/seckill）均先 `DecreasePreInventory` 再 `DecreaseInventory`**，且直连路径要求 `PreOrderId` 并走 `FindLockOrder` 锁检查（`decreaseinventorylogic.go:74`）→ 无无预扣直扣调用方。若未来出现"无预扣直扣"调用方，需在 D3 补 fallback（另案）。
2. spec 写入路径清单未显式列 `EnsureInventoryCacheCtx`（`cache.go:56`）；task 1.1 已要求核对，记录。
3. tasks 4.4 验收模板建议同步到 design D6（可选）。

### 结论
**🔴 清零 → 本批冻结**（方向已裁定，本轮允许 apply）。上一轮 4 🔴（事实错误/双重语义/spec 不可判定/根因）已全部由方向修订 + 重探索（2A-2D）闭合。

## Verify — 2026-09-17（归档前）

一致性四项 + openspec validate：
1. **proposal→tasks 可追溯**：proposal What Changes 三件事（①②③）分别落到 tasks 1.x/2.x/3.x（①口径统一→1.1/1.1b/1.2/1.3/1.4；②读路径→2.1/2.2；③测试可判定→3.1/3.2）。✓
2. **tasks 全勾且有证据**：17/17 `- [x]`；证据 = PR #79（9681da0）`-count=10` 10/10 PASS、make build/lint/test-unit 绿、svc 单测 2 例（`cache_test.go`：回填写可用值、预热只回填缺失键）。✓
3. **无越界改动**：`git show --stat 9681da0` = services/inventory（logic×2 + svc×3）+ test/rpc/inventory×1 + openspec 制品，白名单内。✓
4. **explore-brief C1–C6 逐条成立**（方向修订后口径）：
   - C1 写路径扫描：2C 清单修正后 DecreaseInventory 确认不加缓存写（避免二次扣减），UpdateInventory 已同步，pre/return-pre 走 Lua——无遗漏无错误处方。✓
   - C2 GetInventory 返回可用库存最新值：D4 回填用刚回填值；delta 方向 D3 表正确。✓
   - C3 缓存失败降级：ReturnInventory/UpdateInventory 失败记日志不阻断（1.2b 确认）。✓
   - C4 并发稳定 + 一致性断言：`-count=10` 10/10 PASS（0.40~0.54s）；断言 expected 0 == 可用库存。✓
   - C5 不改扣减正确性/幂等锁；无删测试；make build/lint/test-unit 绿。✓
   - C6 单测覆盖：svc 层新增回填/预热缓存同步单测；逻辑层行为由集成测试覆盖。✓

**openspec validate**：`Change 'fix-inventory-cache-invalidation' is valid`。
