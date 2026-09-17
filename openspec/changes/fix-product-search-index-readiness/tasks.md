> 变更A：商品搜索索引就绪与映射一致性。白名单 = services/search/ + services/product/ + dal/es/ + test/rpc/product/（越界即停）。
> 口径：ES 文档字段名 = 索引映射 = 查询字段 = snake_case；索引 `products` 映射 MUST 为 snake_case + dynamic:false。

## 1. 写路径 DTO（D1-C′ / D3）

- [ ] 1.1 新增 `dal/es/product/document.go`：`ProductDocument`（json 标签 snake_case：id/name/description/picture/price/stock/category/created_at/updated_at）+ `NewProductDocument(p *product2.Products) *ProductDocument`；`dal/es/product/var.go` embed `template.json` 导出 `EsTemplate`；验证：`go vet` + 编译
- [ ] 1.2 `createproductlogic.go:111` `BodyJson` 改用 `NewProductDocument(productRes)`；`updateproductlogic.go:101` `Doc` 同上；验证：代码 diff + 单测（序列化输出含 snake_case 键）

## 2. 模板 + 映射确定化（D1-C / D4）

- [ ] 2.1 新增 `dal/es/product/template.json`：`index_patterns:["products*"]` + snake_case 映射 + `"dynamic":"false"`；`mapping.json` 同步加 `"dynamic":"false"`；验证：`curl localhost:9200/_template/go-mall-products` 返回模板体
- [ ] 2.2 确认 create/update/delete product 的 ES 写（`createproductlogic.go:108/150`、`updateproductlogic.go:98`）在模板下 auto-create 得正确映射；验证：结论 + 行号

## 3. 启动 ensure（A + D2）

- [ ] 3.1 `servicecontext.go` 重写 `initEs` → `ensureProductIndex`：a) ES 就绪重试（60s/2s）；b) 幂等 Put 模板；c) 索引缺失 → `CreateIndex`(EsMapping)；d) 映射错误 → 删→建→**分页重索引**（`FindPage` 批次 500 + DTO 写入）；e) 失败 → panic 阻断启动；验证：代码 + 单测/本地 ES
- [ ] 3.2 多实例幂等：重建前复查映射，已被修复则跳过；验证：结论 + 代码

## 4. 验证（D5）

- [ ] 4.1 干净卷验收：`docker compose -f construct/depend/docker-compose.yaml stop elasticsearch && docker rm go-mall-elasticsearch && docker volume rm depend_es-data && docker compose -f construct/depend/docker-compose.yaml up -d elasticsearch` 后全栈重启；`_cat/indices` 就绪后 snake_case 映射 + 写路径 `_source` 键 snake_case（贴就绪前/后对照）；验证：命令 + 输出
- [ ] 4.2 集成验收：`cd test/rpc && GOWORK=off GO_MALL_TEST_LOCAL=1 go test -count=10 -run TestQueryProduct ./product/...` → 0 失败；验证：**贴原始输出**（10 次）
- [ ] 4.3 Rule 1 门禁：`make build && make lint && make test-unit` 全绿；验证：命令输出
- [ ] 4.4 对照 spec 各 Scenario 逐项核对；验证：验收记录

## 5. 交付

- [ ] 5.1 红线：`git diff --name-only` 白名单 = services/search/ + services/product/ + dal/es/ + test/rpc/product/；验证：命令输出
- [ ] 5.2 提 PR → Build + Quality 绿 → 合并；验证：PR 号 + merge commit hash