> 业务代码变更（services/inventory）+ 测试（test/rpc/inventory）；Rule 1 门禁适用。
> 白名单 = services/inventory/** + test/rpc/inventory/**（越界即停）。
> 口径：缓存键 `inventory:product:{pid}` = 可用库存（可售）。

## 1. 口径统一（D1/D2/D3）

- [ ] 1.1 核对 `svc/cache.go` 全部写入点口径：`SetInventoryCacheCtx` / `LoadInventoryFromDBToCache` / `AdjustInventoryCacheCtx` / `EnsureInventoryCacheCtx`；`LoadInventoryFromDBToCache` **验收标准 = 写 DB total + 语义声明为可用库存 + 注释"冷启动预留=0 假设、DB 无预留列无法还原在途预扣"**（不要求写出无法计算的"公式"，见 design D2 可执行性说明）；验证：给出各写入点口径结论 + 代码行
- [ ] 1.1b 预热口径（D1b）：`PreheatInventoryCache`（`servicecontext.go:70-86`）改为**只回填缺失键**（键存在则跳过，不覆盖在途预扣）；验证：代码 diff + 结论（含"不覆盖已存在键"）
- [ ] 1.2 逐条核对写路径 × 可用库存增量（design D3 表）：`DecreaseInventory` **barrier/Saga 路径（`decreaseinventorylogic.go:51-70`）与直连路径（`:74-98`）均确认不加缓存写**（避免二次扣减破坏 Saga 补偿语义）、`ReturnInventory`（barrier `:56/73` + 直连 `:114`）保持 `+q`、`UpdateInventory` 为**全量 SET**（非增量，口径注明可用库存）；验证：逐路径结论（含"确认不二次扣减"）
- [ ] 1.2b 缓存失败降级验证（spec Requirement"缓存同步失败降级不影响业务写"）：确认 `ReturnInventory` / `UpdateInventory` 缓存同步失败仅记日志不阻断业务写（对齐既有行为）；`DecreaseInventory`/预扣/回补路径保持不加缓存写；验证：结论 + 代码行
- [ ] 1.3 核对预扣/回补 Lua（`decreaselua.lua` DECRBY / `returnlua.lua` INCRBY）与可用库存口径一致，无需改动则记录结论；验证：结论 + 行号
- [ ] 1.4 写路径单测覆盖（mock redis 或真实 redis，explore-brief C6 闭合）：`DecreaseInventory` 确认后可用库存**不变**（不二次扣减）；`ReturnInventory` 后可用库存 `+q`；`UpdateInventory` 全量 SET；`PreheatInventoryCache` 只回填缺失键（已存在键不覆盖）；验证：`cd services/inventory && go test ./...` 通过 + 各用例输出

## 2. 读路径旧值修正（D4）

- [ ] 2.1 `getinventorylogic.go` 回填分支（`:56-69`）改用**刚回填的缓存值**（`LoadInventoryFromDBToCache` 返回其写入的可用库存值，避免多余 Redis 往返；不再用 `:35` 早期 DB 读值）；**若调整 `LoadInventoryFromDBToCache` 签名，必须同步调整同包 `EnsureInventoryCacheCtx`（`cache.go:56`）**；验证：回填后返回值 == 缓存值（单测）
- [ ] 2.2 确认 `GetInventory` 缓存命中分支（`:47/72`）返回值口径为可用库存；验证：代码行 + 结论

## 3. 测试可判定性（D5）

- [ ] 3.1 `test/rpc/inventory/inventory_test.go` 预扣调用后校验响应体 `StatusCode`：非成功（含 `InventoryNotEnough`）→ 判定失败且 `return`（不继续 `DecreaseInventory`）；验证：代码 diff + 行号
- [ ] 3.2 确认测试未删除/放宽既有断言（红线）；验证：`git diff` 无删测试/skip

## 4. 验证（D6）

- [ ] 4.1 集成测试稳定：`cd test/rpc && GOWORK=off GO_MALL_TEST_LOCAL=1 go test -count=10 ./inventory/...` → 0 失败；验证：**贴原始输出**（≥10 次）
- [ ] 4.2 一致性断言：写后 `GetInventory` == 可用库存期望值（`DB total − 在途预留`；卖光场景 = 0）；验证：断言 + 输出
- [ ] 4.3 Rule 1 门禁：`make build && make lint && make test-unit` 全绿；验证：命令输出
- [ ] 4.4 对照 spec 各 Scenario 逐项核对，产出验收记录（最小模板：`Scenario 名 | 验证方式 | 证据输出`）；验证：记录含证据

## 5. 交付

- [ ] 5.1 红线：`git diff --name-only` 白名单 = services/inventory（含 `svc/cache.go`、`svc/servicecontext.go`、`logic/*`）+ test/rpc/inventory；验证：命令输出
- [ ] 5.2 提 PR → 等 Build + Quality 绿（Integration 若红按归因纪律写"ES 栈启动，与本变更无关"）→ 合并；验证：PR 号 + merge commit hash
- [ ] 5.3 本轮不归档（下一轮 verify 后再归档）；验证：—