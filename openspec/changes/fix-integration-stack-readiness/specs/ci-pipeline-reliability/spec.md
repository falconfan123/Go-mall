## Purpose

CI 流水线依赖栈就绪契约增量：ES 数据卷跨平台可写、ES 就绪判据为健康检查（wait_for_status=yellow）、栈启动后业务服务监听且集成测试真正执行。

## MODIFIED Requirements

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