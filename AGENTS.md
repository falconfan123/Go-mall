# Go-mall Harness Engineering — AI 行动准则

本文件与 CLAUDE.md 同步，供 Codex 读取。以下规则是你在 Go-mall 项目中的工程准则。

## Rule 1：修改后三件事（最高优先级）
每次代码修改后依次完成：编译通过 → 单元测试通过 → Lint 通过。三步不过任务不算完成。

## Rule 2：API 优先
新增功能前先定义 API 规范。测试必须基于 API 文档编写，禁止随机测试。

## Rule 3：go-zero 代码生成
handler 必须用模板生成，禁止手写。proto 修改后重新生成 pb 代码。

## Rule 4：DDD 红线（admin 服务）
pb 结构体禁止泄漏到 logic/service 层。进入逻辑层前必须转换。

## Rule 5：提交纪律
提交前运行 `make lint`。信息遵循 conventional-commits。每个提交一个逻辑单元。

## Rule 6：Mock 先行
外部依赖的测试必须使用 mock。修改业务代码后运行 `make mock`。

## Rule 7：分布式事务规范
跨服务事务使用 DTM Saga，禁止其他方案。每个 Saga 必须有补偿逻辑。

## 本地启动
```bash
./scripts/start-unified.sh       # 启动
./scripts/start-unified.sh stop  # 停止
```

## 可用 Makefile 命令
- `make lint` — 代码质量检查
- `make test-unit` — 单元测试
- `make mock` — 生成 mock
- `make build` — 编译所有服务
- `make coverage` — 覆盖率
- `make gatekeeper` — 主守门人检查
- `make fmt` — 自动格式化
- `make submit-ci MSG="..."` — 提 PR + CI + 自动合并

## 项目文档
- `docs/dev-map.md` — 开发导航图
- `docs/specs/template.md` — SPEC 编写模板
