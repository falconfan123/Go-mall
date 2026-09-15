> 业务代码变更（services/inventory）；Rule 1 门禁适用。验证命令在对应服务目录执行。

## 1. 影响面扫描（D2）

- [ ] 1.1 全量列出 model 层改 `sold`/`total` 的方法（`Batchdecrease`/`BatchReturn`/`DecreaseInventoryAtom`/`ReturnInventory`/`UpdateOrCreate` 等）→ 映射 logic 写路径，产出"写路径 × 对 sold/total 的变更 × 缓存同步点"清单；验证：清单覆盖 model 全部写方法，无遗漏
- [ ] 1.2 确认 `AdjustInventoryCacheCtx`（svc/cache.go:73）的 delta 语义与并发原子性（INCRBY/DECRBY 或读-改-写）；验证：给出实现结论（影响并发断言设计）

## 2. 写路径补缓存同步（D1）

- [ ] 2.1 `decreaseinventorylogic` 补 `AdjustInventoryCacheCtx`（delta 方向 = 扣减对 sold/total 的实际变更）；验证：写后 `GetInventory` == DB 值（单测）
- [ ] 2.2 `decreasepreinventorylogic` 补缓存同步；验证：同上
- [ ] 2.3 `returnpreinventorylogic` 补缓存同步；验证：同上
- [ ] 2.4 `updateinventorylogic` 补缓存同步；验证：同上
- [ ] 2.5 缓存同步失败降级一致（仅记日志，不阻断业务写，对齐 ReturnInventory）；验证：注入缓存失败，业务写仍成功

## 3. 验证

- [ ] 3.1 单测：各写路径"写后 GetInventory == DB 值"一致性断言；验证：`cd services/inventory && go test ./...` 通过
- [ ] 3.2 集成测试稳定：`cd test/rpc && GOWORK=off GO_MALL_TEST_LOCAL=1 go test -count=1 -run TestInventoryService_HighConcurrency ./inventory/` 连续 ≥5 次全绿（附输出）；验证：0 失败
- [ ] 3.3 缓存/DB 一致性断言：并发扣减后 `GetInventory` == DB（`SELECT total,sold`）值；验证：断言通过

## 4. 门禁与验收

- [ ] 4.1 Rule 1：`make lint` / `make test-unit` / `make build` 全绿；验证：命令输出
- [ ] 4.2 对照 spec 各 Scenario 与 explore-brief C1–C6 逐项核对，产出验收记录（含多次重跑证据）；验证：记录含证据
- [ ] 4.3 红线：`git diff` 无删测试/skip/放宽断言；改动白名单 = services/inventory + test/rpc/inventory；验证：`git diff --name-only` 符合