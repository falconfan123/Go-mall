
## 审查轮次 1 — 2026-09-17（冻结）

**审核方式**：@openspec-reviewer 对完整变更包重审（备选通道脚本）。

### 修复循环（同一门禁内）
- 首轮：🔴 3（写路径 PascalCase 根因缺失 / 全量重建 OOM / 重建失败无安全回退）→ 重写制品：D1-C′ 写路径 DTO、D2 分页重建 + 失败阻断、spec/tasks 同步。
- 末轮：**🔴 无 → 冻结**。

### 🟡 记录（本轮吸收/后续）
- 已吸收（实现内落实，不属设计变更）：① `createproductlogic` 写前回写 `productRes.Id=productId`（防 doc.id=0）；② `mapping.json` 补 `stock` 字段 + `"dynamic":"false"`（与 DTO/spec 一致）；③ 错误映射判定扩展为"无 snake_case 关键字段 **或** 存在 PascalCase 字段（混合映射）"即重建。
- 后续（另案）：① proto `crated_at` 拼写错误（`product.proto:96`，需重生成 pb）——读路径 `created_at` 无法回填响应；② 大 offset 时 `FindPage` 性能，可评估 `FindListByCursor` keyset 分页；③ category 索引完整性。

### 结论
**🔴 清零 → 冻结**，可 apply（单门禁）。

## Verify — 2026-09-18（归档前）

一致性四项 + openspec validate：
1. **proposal→tasks 可追溯**：What Changes ①②③（写路径 DTO / 模板映射 / 启动 ensure）→ tasks 1.x / 2.x / 3.x；验收 4.x。✓
2. **tasks 全勾且有证据**：12/12 `- [x]`；证据 = PR #87（ee7aee7）+ 上轮 apply 输出（见下）。
3. **无越界改动**：`git show --stat ee7aee7` = dal/es/product（document.go/mapping.json/template.json/var.go）+ services/product（servicecontext/create/update）+ openspec；白名单内。✓
4. **基线成立**：本变更无独立 explore-brief；基线取自 `fix-ci-flaky-product-gateway-tests` 2B（索引映射竞态 + 方向候选），实现 C′（写路径 DTO）+ C（模板）+ A（ensure）与基线方向一致，D（掩盖）被否决。✓

**验收证据（已在 apply 产生，引用不重跑）**：
- 干净卷（`docker volume rm depend_es-data` + 起栈 + product 重启）→ 映射 = `['category','created_at','description','id','name','picture','price','stock','updated_at']`（snake_case）；模板 `go-mall-products`（index_patterns `products*`）已注册。
- `_source` 键 = snake_case（`id/name/description/picture/price/stock/category/created_at/updated_at`）。
- `cd test/rpc && GOWORK=off GO_MALL_TEST_LOCAL=1 go test -count=10 -run TestQueryProduct ./product/...` → **10/10 PASS**（ok 0.806s）。
- `make build && make lint && make test-unit` 全绿（154 测 0 失败）。
- 实施发现：olivere v7 `GetMapping` 带 ES7 `include_type_name=true` 被 ES8 拒 400 → 改用 `PerformRequest` 裸读映射。

**openspec validate**：`Change 'fix-product-search-index-readiness' is valid`。
