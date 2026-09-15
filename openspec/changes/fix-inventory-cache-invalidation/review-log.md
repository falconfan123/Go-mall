# Review Log · fix-inventory-cache-invalidation

## 整体反思审核（proposal + specs + design + tasks）

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
