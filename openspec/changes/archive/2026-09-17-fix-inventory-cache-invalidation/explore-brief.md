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

- [x] C1 所有库存写路径（decrease/pre/return/pre-return/update）写后缓存与 DB 一致；全量扫描无遗漏写路径。
- [x] C2 `GetInventory` 返回最新值（写后读一致性）；`AdjustInventoryCacheCtx` delta 方向正确（sold/total 不反向）。
- [x] C3 缓存同步失败不阻断业务写（与 ReturnInventory 既有降级一致），仅记日志。
- [x] C4 `TestInventoryService_HighConcurrency` 稳定通过（多次重跑 0 失败）+ 缓存/DB 一致性断言。
- [x] C5 不改扣减正确性/幂等锁语义；不删测试/放宽断言；Rule 1 门禁（编译/单测/lint）通过。
- [x] C6 新增/修改单测覆盖各写路径的缓存同步（mock 或真实 redis）。
---

## 重探索：失败时刻实证 + 缓存键语义裁定（2026-09-16，round 2 前置）

### 2A 失败时刻实证（同运行采集，8 次重跑）
| run | 结论 | actual(GetInventory) | 同刻缓存值 | 同刻 DB(total,sold) | 预扣失败 | 扣减失败 |
|---|---|---|---|---|---|---|
| 1 | ok | — | — | — | 0 | 0 |
| 2 | ok | — | — | — | 0 | 0 |
| 3 | **FAIL** | **390** | 未采到* | **0, 1000** | 0 | 0 |
| 4 | ok | — | — | — | 0 | 0 |
| 5 | **FAIL** | **990** | 未采到* | **0, 1000** | 0 | 0 |
| 6 | ok | — | — | — | 0 | 0 |
| 7 | ok | — | — | — | 0 | 0 |
| 8 | **FAIL** | **730** | 未采到* | **0, 1000** | 0 | 0 |

\* 同刻缓存值**未采到**（测试并发窗口 <1.5s，外部 redis 轮询（1s 间隔 + 后台 60s）无法命中；测试后缓存键 `inventory:product:9999` 实测为空）。采集命令：`/tmp/f2-probe.sh`（go test + redis-cli GET + psql）。
- **关键**：DB 恒 `0,1000`（并发扣满，无丢扣）；预扣/扣减 RPC 均无 error（失败计数 0）；但 GetInventory 返回 390/990/730 ≠ DB total=0 → **GetInventory 读的不是 DB total**。

### 2B 缓存键语义裁定（必须二选一）
- 缓存键 `inventory:product:{pid}`（svc/cache.go:12-13，`InventoryProductKey`）被**两方共用、双重语义**：
  1. **预扣路径（可用库存语义）**：`decreaselua.lua` 对库存键 `DECRBY`（扣减缓存）；`decreasepreinventorylogic.go:33` 注释明确"**并非真实扣除库存，而是在缓存进行--操作**"。
  2. **DB total 镜像语义**：`SetInventoryCacheCtx`（:16-17）/`LoadInventoryFromDBToCache`（:43-49，从 DB total 回填）/`AdjustInventoryCacheCtx`（:73）写 total；`GetInventory`（getinventorylogic.go:47/72）读并返回。
- **裁定：缓存键 = 双重语义混合（设计冲突）**——预扣当"可用库存"（只动缓存），GetInventory/回补当"DB total 镜像"。
- **对 D1-A 可行性的判定**：若采用"DecreaseInventory 补 Adjust（D1-A）"，标准生命周期（预扣 Lua -10 → DecreaseInventory DB-10 + Adjust -10 → 补偿 ReturnInventory DB+10/缓存+10）**缓存净漂移 -10（与 DB 反向）**；高并发下缓存终值大幅偏离 DB（如 -1000 vs 0）→ 测试从"偶发失败"变"**必失败**"。**D1-A 不可行**。

### 2C 写路径清单修正（4🔴 第 1 条）
- 事实修正：`updateinventorylogic.go:46` **已调 `SetInventoryCacheCtx`**（原 proposal"未同步"是错的）；`decreasepreinventorylogic`/`returnpreinventorylogic` **不写 DB**，走 Lua 原子改同一缓存键（"补 Adjust"会二次扣减，错误处方）。
- **真正未同步缓存（DB 写、缓存未跟）的写路径**：`DecreaseInventory`（DB 扣减，缓存未 Adjust）。扫描结论：`Batchdecrease`（被 DecreaseInventory 用）/`BatchReturn`（被 ReturnInventory 用，Return 已同步）/`DecreaseInventoryAtom`/`ReturnInventory`（单数，model 层，全仓无 logic 调用方）/`UpdateOrCreate`（被 UpdateInventory 用，已同步）——**唯一缺口 = DecreaseInventory**。

### 2D 根因候选面判定
| 候选 | 判定 | 依据 |
|---|---|---|
| (a) 缓存键缺失/淘汰 | **部分支持**，但非唯一根因 | 测试后缓存键空；但失败时 GetInventory 返回非 DB 值，更像读到可用库存 |
| (b) 测试忽略预扣 body 状态（inventory_test.go:310-323 仅传输层 error 时 return） | **在修复面外（测试写法）** | 预扣 Lua `return 2`（库存不足）是 body 层，RPC 返回 nil err → 测试认为预扣成功并继续扣减；预扣失败计数 0（t.Logf 仅传输层） |
| (c) getinventorylogic.go:69 回填用 :35 旧 DB 值 | **在修复面内（读路径）** | `cachedTotal = inventoryResp.Total`（:35 早读），非刚回填值 |

### 结论（触发 2E 判定）
- **根因不在当前制品假设的修复面（"DecreaseInventory 补 Adjust / 写路径统一同步"）内**——核心是**缓存键双重语义冲突（预扣可用库存 vs DB 镜像）+ 测试未校验预扣 body 状态 + GetInventory 回填旧值**三因素交织；D1-A 会加剧冲突（必失败）。
- **建议（供用户决策，不硬改制品）**：
  1. **换方向**：修复缓存键语义冲突（预扣与 GetInventory 解耦，或 GetInventory 在预扣存在时读可用库存口径）；
  2. **测试修正**：inventory_test 校验预扣 body 状态（失败即 t.Fail，不继续扣减）——但属测试写法，需确认是否为本变更范围；
  3. **读路径最小修**：getinventorylogic.go:69 回填用刚加载值（修 (c)）。
