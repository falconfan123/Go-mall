## Purpose

商品搜索 ES 索引（`products`）就绪与映射一致性契约：ES **文档字段名**与**索引映射** MUST 均为 snake_case（写路径 DTO + 模板保障），服务启动 MUST 确保索引存在且映射正确（就绪重试 + 错误映射分页重建 + 失败阻断），商品搜索/查询 MUST 不因索引就绪或字段不一致失败。

## ADDED Requirements

### Requirement: ES 文档字段名为 snake_case

product 服务写入 `products` 索引的 ES 文档 MUST 使用 snake_case 字段名（`id/name/description/picture/price/stock/category/created_at/updated_at`），通过显式 DTO（`json` 标签）序列化，MUST NOT 使用无 `json` 标签结构体的默认 PascalCase 序列化。

#### Scenario: 写路径文档字段为 snake_case
- **WHEN** 调用 `CreateProduct` / `UpdateProduct` 后查询 `products` 索引 `_source`
- **THEN** `_source` 字段名 MUST 为 snake_case（`id`/`name`/`price`/`created_at`/`updated_at`…），MUST NOT 出现 `Id`/`Price`/`CreatedAt` 等 PascalCase 字段

### Requirement: 商品搜索索引映射确定化

索引 `products` 的映射 MUST 为 snake_case 显式定义（`id/name/description/picture/price/category/created_at/updated_at`），且 MUST 设置 `"dynamic":"false"`（忽略未知字段，防动态漂移）。任何自动创建该索引的路径（bulk auto-create）MUST 应用该显式映射（索引模板）。

#### Scenario: 自动创建索引使用正确映射
- **WHEN** `products` 索引不存在，通过任意路径（bulk `Index()` / 显式 `CreateIndex`）触发创建
- **THEN** 创建后的索引映射 MUST 含 snake_case 字段且无 `CreatedAt`/`Price` 等大写字段（模板 + dynamic:false 生效）

#### Scenario: 干净卷下映射确定
- **WHEN** 干净 ES 卷（`docker volume rm` 后起栈）下服务全部就绪
- **THEN** `_cat/indices` 显示 `products` 索引为 snake_case 映射，随后按 `created_at` 排序的查询 MUST 不报 `search_phase_execution_exception`

### Requirement: 服务启动确保索引就绪与映射正确

product 服务启动时 MUST 确保 `products` 索引存在且映射正确：ES 未就绪 MUST 有限重试直至超时；索引缺失 MUST 用 snake_case 映射创建；索引存在但映射错误（不含 snake_case 关键字段）MUST 分页重建（删→建→DB 全量重索引）；**重建任一步失败 MUST 阻断启动**（不得静默带空索引运行）。

#### Scenario: ES 就绪后索引自动就绪
- **WHEN** product 服务在 ES 尚不可达时启动，随后 ES 就绪
- **THEN** product 服务 MUST 在重试窗口内完成索引确保（创建/修正）

#### Scenario: 错误映射被分页重建
- **WHEN** `products` 索引已存在但映射为大写/混合（无 `created_at`/`price`/`category`）
- **THEN** product 服务启动 MUST 删除重建该索引为 snake_case 映射，并分页从 DB 全量重建索引数据；若任一步失败 MUST 阻断启动（启动失败可察觉），MUST NOT 静默运行于空索引

### Requirement: 搜索查询不因索引就绪时序失败

商品搜索/列表查询 MUST 在索引就绪且映射、文档字段一致的前提下执行，MUST NOT 因索引未就绪或字段不一致返回 500（`search_phase_execution_exception`）。

#### Scenario: 干净卷下查询稳定
- **WHEN** 干净 ES 卷下服务全部就绪后调用 `QueryProduct`（`New:true`）
- **THEN** MUST 返回正常结果，`TestQueryProduct` 连续 ≥10 次重跑 MUST 0 失败