## Context

见 `proposal.md`（Why）与 explore-brief（2B 实证）。核心事实：
- 写路径 `BodyJson(productRes)` 用 DB 模型（无 json 标签）→ ES 文档 PascalCase（`_source` 实测 `Id/Name/Price/CreatedAt/...`）。
- 查询/映射 snake_case（`queryproductlogic.go:50/54/56`、`mapping.json`）→ sort 缺失字段 → `search_phase_execution_exception`。
- `initEs` 仅在索引缺失时创建、失败不重试。
- 对照：重启后正确映射 5/5 PASS；当前本地索引为混合映射（大写+snake 并存）。

## Goals / Non-Goals

- **Goals**：ES 文档字段名 = snake_case（写路径 DTO）；索引映射确定化（模板 + `dynamic:false`）；启动确保索引就绪/映射正确且失败可察觉。
- **Non-Goals**：不改查询逻辑；不掩盖；不改 search 读侧；不补 category 索引完整性（另案）。

## Decisions

### D1 方案对比

| 方案 | 做法 | 评价 |
|---|---|---|
| A 启动期建索引 + 等待 + 重试 | `ensureProductIndex` 就绪重试 + 缺失则建 | 解决时序；**不解决**文档字段 PascalCase（映射再对，文档字段名仍错） |
| B bulk 前确保索引存在 | 每次写前 EnsureIndex | 兜底写路径；仍未解决文档字段名 |
| C 索引模板 + `dynamic:false` | `products*` 模板（snake_case 映射 + dynamic:false） | 映射确定化，防动态漂移；**单独不解决**文档字段名（PascalCase 文档仍与 snake_case 映射不匹配 → dynamic:false 下静默丢字段） |
| **C′ 写路径 DTO + C 模板（推荐）** | ① ES 文档用 snake_case DTO（`dal/es/product/document.go`）② 模板确定化映射 | **根治**：文档字段名 = 映射 = 查询 = snake_case |
| D 查询端忽略缺失字段 | 容忍缺失/降级 | **只掩盖**：静默空结果/错序，掩盖生产故障；**否决** |

- **选定 C′（D1 根因）+ A（时序补强）+ 重建安全（D2）**；否决 D。

### D2 错误映射重建（安全、分页、失败阻断）

- **触发**：启动时索引存在但映射无 snake_case 关键字段（`created_at`/`price`/`category`）。
- **步骤**：① 删除旧索引 → ② 用 `EsMapping` 重建 → ③ **分页重索引**：`FindPage(offset, limit)`（批次如 500）从 DB 全量读 + `EsClient.Index()`（snake_case DTO）写入；循环至游标末 → ④ 完成。
- **安全**：任一步失败（DB 查询/ES 写入/网络）→ **panic/返回 error 阻断启动**（健康检查不过），MUST NOT 静默带空索引运行（防"删除后重建失败 → 空索引服务"事故）。
- **多实例**：以**单实例部署**为前提（product 服务副本数=1）；重建前再次检查映射（幂等），若已被其它实例修复则跳过。
- **不使用 ES `_reindex`**：旧文档字段名即错（PascalCase），reindex 不转换字段名；必须以 DB 为源用 DTO 重建。

### D3 写路径 DTO

- `dal/es/product/document.go`：`ProductDocument`（json 标签 snake_case：id/name/description/picture/price/stock/category/created_at/updated_at）+ `NewProductDocument(p *product2.Products) *ProductDocument`（category 当前缺省空串，索引完整性另案）。
- `createproductlogic.go:111` / `updateproductlogic.go:101`：`BodyJson(NewProductDocument(productRes))` / `Doc(NewProductDocument(...))`。

### D4 模板与映射

- `template.json`：`index_patterns:["products*"]`、`template.mappings` = snake_case（与 `mapping.json` 一致）+ `"dynamic":"false"`。
- `mapping.json` 同步加 `"dynamic":"false"`（显式 CreateIndex 也防漂移）。
- 模板名 `go-mall-products`（幂等 Put，同名覆盖即更新）。

### D5 验证

- 干净卷：`docker compose stop/rm/volume rm/up` + 全栈重启 → `_cat/indices` 就绪后 snake_case 映射；写路径 `_source` 键 = snake_case。
- `cd test/rpc && GOWORK=off GO_MALL_TEST_LOCAL=1 go test -count=10 -run TestQueryProduct ./product/...` → 0 失败。
- `make build && make lint && make test-unit` 全绿。

## Risks / Trade-offs

- **[删除重建搜索短暂空]** → 一次性、分页、失败阻断启动（可察觉）。
- **[单实例前提]** → 多副本需另加互斥（本服务当前单实例；design 注明）。
- **[category 缺省]** → 另案；本变更聚焦字段名/映射一致（crash 根因）。

## Migration Plan

1. 新增 `dal/es/product/document.go` + `template.json` + embed。
2. 写路径改 DTO（D3）。
3. 重写 `ensureProductIndex`（A + D2）。
4. 干净卷验收 + 门禁（D5）。
5. 提 PR → 合并。

**回滚**：product 服务 + dal/es，可 git revert；索引重建一次性，回滚不影响已正确映射。