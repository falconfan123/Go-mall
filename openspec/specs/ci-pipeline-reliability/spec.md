# ci-pipeline-reliability Specification

## Purpose

CI 流水线（Integration / Quality / Build）可靠性契约：Integration 稳定转绿、Quality 门禁全绿、失败时产出可诊断日志、required status checks 分阶段生效且不阻塞误伤。

## Requirements

### Requirement: Integration 流水线稳定转绿

Integration 流水线（RPC Integration 起依赖 + 启动全部业务服务 + 集成测试）MUST 稳定通过，不得因环境编排（服务启动/端口等待/依赖就绪）系统性失败。判定口径：从本 change 首个修复 commit 合并后计数，按 run 序（非按天），观察窗口 14 天内满 20 次 Integration run 绿率 ≥90%，且同一 commit 至少重复触发 3 次全绿（验证非 flaky）。依赖栈（尤其 Elasticsearch）的数据卷 MUST 跨平台可写，就绪判据 MUST 为集群健康检查而非仅端口就绪。

#### Scenario: 集成栈稳定启动
- **WHEN** 在 CI runner 上启动全部业务服务（起依赖容器 + 15 个服务）
- **THEN** 所有服务在超时内监听端口并保持存活，集成测试执行不因服务启动失败而中断

#### Scenario: 重复触发不 flaky
- **WHEN** 同一 commit 重复触发 Integration 流水线 3 次
- **THEN** 3 次均绿（无偶发服务启动超时/依赖未就绪）

#### Scenario: ES 数据卷跨平台可写
- **WHEN** 在 CI runner（ubuntu）或本地（macOS）执行 `docker compose -f construct/depend/docker-compose.yaml up -d` 启动依赖栈
- **THEN** ES 容器数据目录 `/usr/share/elasticsearch/data` MUST 可写（不出现 `AccessDeniedException` / `failed to obtain node locks`），ES MUST 完成 bootstrap 且容器存活

#### Scenario: ES 就绪判据为健康检查而非端口就绪
- **WHEN** `ci-rpc-stack.sh` 等待 ES 就绪
- **THEN** MUST 以 ES 容器健康检查为就绪判据；healthcheck test 为 `curl -sS "http://localhost:9200/_cluster/health?wait_for_status=yellow&timeout=5s" | grep -qE '"status"[[:space:]]*:[[:space:]]*"(yellow|green)"'`，参数 MUST 为 `interval: 10s, timeout: 5s, retries: 30, start_period: 40s`；**端口 9200 就绪本身 MUST NOT 视为就绪**（ES 绑定端口早于集群可用，audit 连入可能 `connection reset`）；`/_cluster/health` 恒返回 200，MUST 解析响应体 `status` 为 yellow/green 而非仅 `curl -f` 校验状态码；`ci-rpc-stack` 对 ES 的就绪等待超时（`wait_for_container_health "$container" 180 2`，360s）MUST ≥ healthcheck 就绪检测最坏窗口

#### Scenario: 栈启动后业务服务监听且测试真正执行
- **WHEN** Integration 流水线在依赖栈就绪后启动全部业务服务
- **THEN** `audit` 等服务 MUST 成功监听端口（不再出现 `service failed to listen: audit` / `audit mq init attempt` 循环），且集成测试 MUST 真正开始执行（存在测试执行输出，非栈未起即失败）

### Requirement: Quality 门禁全绿

Quality 流水线各 job（Go Vet / Govulncheck / Unit Tests / Coverage Gate）MUST 全部通过；模块隔离编译（GOWORK=off）不得因依赖清单缺失（如 `replace`）而失败。

#### Scenario: 模块隔离编译通过
- **WHEN** CI 以 GOWORK=off 独立编译各 Go 模块（含 dal 模块）
- **THEN** 编译成功，不出现 `undefined: biz.ErrReturnAlreadyLocked` 类依赖缺失错误

#### Scenario: 质量门禁全绿
- **WHEN** 对同一 commit 触发 Quality 流水线
- **THEN** Go Vet / Govulncheck / Unit Tests / Coverage Gate 全部成功

### Requirement: 失败可诊断

CI 流水线失败时 MUST 产出可诊断产物，供 PR 审阅者定位根因；不得静默失败。

#### Scenario: 失败上传诊断产物
- **WHEN** Integration 流水线失败
- **THEN** 上传诊断 artifact：各业务服务 stderr/stdout 日志、`docker ps -a` 全量、依赖容器健康状态、端口监听快照（`ss -ltnp`）、etcd/rabbitmq/minio 健康探针结果；在 PR 的 Actions run Artifacts 页面可访问，保留期明确（≥7 天）

#### Scenario: Quality 失败诊断适配环境
- **WHEN** Quality 流水线（无 docker 栈/服务日志的 runner）的某 job 失败
- **THEN** 诊断 artifact 聚焦该 job 可得的诊断输出（如 go vet/govulncheck 的 stderr 落盘 + runner 环境快照），不伪造 docker/服务日志；artifact 空时须显式记录"该 runner 无栈，输出为空"，不得静默产生空包

### Requirement: 门禁分阶段 required 且不阻塞误伤

main 分支 required status checks MUST 分阶段生效：阶段 1 仅 Build + Quality 为 required（Integration 仍非 required，不阻塞合并）；阶段 2 在 Integration 满足"稳定转绿"判定后加入 required。任一阶段 MUST 不以未修复的必红流水线阻塞合并（防自锁）。

#### Scenario: 阶段 1 不阻塞
- **WHEN** 阶段 1 生效而 Integration 尚未修复（可能红）
- **THEN** 仅 Build + Quality 为 required，Integration 红不阻塞 PR 合并

#### Scenario: 阶段 2 后 Integration 成为必过
- **WHEN** Integration 已满足"稳定转绿"判定并加入 required
- **THEN** Integration 红则阻塞合并（门禁真实生效）

### Requirement: 不削弱门禁强度

CI 修复 MUST NOT 通过删测试、skip、continue-on-error 或移除 required checks 来"改绿"。

#### Scenario: 修复不削弱强度
- **WHEN** 审查本 change 的 CI 改动 diff
- **THEN** 无 `continue-on-error` 掩盖、无测试删除/skip、required checks 未因修复而移除或弱化