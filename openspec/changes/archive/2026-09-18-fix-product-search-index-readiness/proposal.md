## Why

商品搜索 ES 索引 `products` 的**文档字段名与映射/查询不一致**，导致商品查询在特定时序下必然失败：

- **现象**：`TestQueryProduct` 本地 15/15 确定性失败。
  ```
  $ cd test/rpc && GOWORK=off GO_MALL_TEST_LOCAL=1 go test -count=1 -run TestQueryProduct ./product/...
  product_test.go:136: rpc error: code = Unknown desc = elastic: Error 400 (Bad Request): all shards failed [type=search_phase_execution_exception]
  ```
- **根因（已实证，双因素）**：
  1. **写路径文档字段名 = PascalCase**：`createproductlogic.go:46/111` 用 `product2.Products`（DB 模型，仅 `db` 标签、无 `json` 标签）`BodyJson` 序列化 → Go 默认大写字段名（`CreatedAt`/`Price`/`Stock`/`Id`…，`Description`/`Picture` 为 `sql.NullString` 嵌套 `{String,Valid}`）；实测 `_source` 键 = `['Id','Name','Description','Picture','Price','Stock','CreatedAt','UpdatedAt']`。
  2. **查询/映射 = snake_case**：`queryproductlogic.go:50/54/56` 按 `created_at`/`updated_at`/`price` 排序；`dal/es/product/mapping.json` 定义 snake_case 显式映射。文档字段是 PascalCase → sort `created_at` 在文档中无值/不存在 → `search_phase_execution_exception` → **all shards failed**。
  3. **时序放大**：`initEs`（`servicecontext.go:98-107`）只在索引缺失时用 snake_case 映射创建、失败不重试（:81）；ES 晚就绪 / initEs 静默失败 / 首次 `EsClient.Index()`（`createproductlogic.go:108`）auto-create（`docker logs`: `[products] creating index, cause [auto(bulk api)]`）→ 动态映射叠加 PascalCase 字段 → 混合映射复发。
  4. 对照：重启 product（initEs 重建 + 正确映射）后 5/5 PASS（见 explore-brief）。
- **生产影响**：商品搜索/列表在索引未就绪或文档/映射不一致时直接 500；且动态映射导致字段漂移不可控。

## What Changes

- **① 写路径文档 snake_case（根因）**：新增 ES 文档 DTO `dal/es/product/document.go`（`json` 标签全部 snake_case：`id/name/description/picture/price/stock/category/created_at/updated_at`）+ 转换函数；`createproductlogic` / `updateproductlogic` 的 ES 写改用该 DTO（`BodyJson`/`Doc`），文档字段与映射、查询一致。
- **② 索引模板（映射确定化）**：新增 `dal/es/product/template.json`（embedded）——`index_patterns: ["products*"]` + 与 `mapping.json` 一致的 snake_case 映射 + **`"dynamic": "false"`**（忽略未知/大写字段，防混合映射复发）；模板对任何 auto-create 生效。
- **③ 启动 ensure（时序补强 + 安全）**：`initEs` → `ensureProductIndex`：a) ES 就绪有限重试（60s、间隔 2s）；b) 幂等 Put 模板；c) 索引缺失 → `CreateIndex`(EsMapping)；d) 索引存在但映射错误（无 snake_case 关键字段）→ **重建**：删旧 → 用正确映射新建 → **分页**（`FindPage` 批次）从 DB 全量重索引（snake_case DTO）；e) **重建失败 → 阻断启动**（panic/健康检查失败），不静默带空索引运行。
- **④ 读取/查询不动**：search 读 `products` 无需改；查询逻辑保持 snake_case。

## Capabilities

### New Capabilities
- `product-search`: 商品搜索 ES 索引就绪与映射一致性契约——ES 文档字段名 MUST 为 snake_case（写路径 DTO），索引映射 MUST 为 snake_case 显式定义（模板 + `dynamic:false`），服务启动 MUST 确保索引存在且映射正确（就绪重试 + 错误映射重建 + 失败阻断），商品搜索/查询 MUST 不因索引就绪/字段不一致失败。

### Modified Capabilities
<!-- 无。 -->

## Impact

- **服务代码**：`services/product/internal/logic/createproductlogic.go`、`updateproductlogic.go`（ES 写改 DTO）；`services/product/internal/svc/servicecontext.go`（`ensureProductIndex`）。
- **资源**：`dal/es/product/document.go`（新增）、`template.json`（新增）、`var.go`（embed）。
- **测试**：`test/rpc/product/`（验收）。
- **风险**：错误映射索引重建为一次性破坏性操作（删→建→重索引），重建期间搜索短暂空；失败阻断启动（不静默）。

## Non-Goals

- 不改查询逻辑（snake_case 保持）。
- 不实现"查询端忽略缺失字段"的掩盖（design D5 否决）。
- 不补 category 索引完整性（现有功能缺口，另案）。
- 不改 search 读侧 / ruleset / 门禁。