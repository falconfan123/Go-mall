## Context

见 `proposal.md`（Why）与 `explore-brief.md`（3A/3B/3C 实证）。核心事实：

- ES bootstrap 失败 = `/usr/share/elasticsearch/data` 不可写（CI 附件 `elasticsearch.log`：`AccessDeniedException: .../node.lock`）。
- `docker-compose.yaml:95-96` ES 数据卷 = macOS 绝对路径 `/Volumes/Fan/docker-data/depend/es-data`（ubuntu runner 上不存在）。
- `ci-rpc-stack.sh:283-291` ES 就绪 = `wait_for_port 9200`（端口就绪），未用 `docker-compose.yaml:100-102` 的 ES healthcheck。
- 15/15 失败 run 签名一致；测试断言失败 = 0（栈未起、测试未执行）。

## Goals / Non-Goals

- **Goals**：ES 容器在 CI 与本地均可 bootstrap 成功；ES 就绪判据可靠；栈启动后业务服务监听且测试真正执行。
- **Non-Goals**：不改测试断言、不修 F2、不设 Integration required、不动 ruleset。

## Decisions

### D1 数据卷方案：B 命名卷（推荐）vs A 相对路径 vs C tmpfs

| 方案 | 做法 | 取舍 |
|---|---|---|
| **A 相对路径 bind mount** | 卷路径改相对/可覆盖（如 `${ES_DATA_DIR:-...}`），CI 设 runner 可写目录 | 保留宿主机持久化；但相对路径相对 compose 文件目录解析、需 runner 预建目录与权限，跨平台仍有隐患 |
| **B 命名卷（选定）** | `es-data:/usr/share/elasticsearch/data` + `volumes:` 声明 `es-data` | Docker 管理、**任何平台默认可写**、无宿主机路径依赖；CI 每次新卷、dev 数据存 docker 卷；历史 ES 索引不直接复用（可重建） |
| **C tmpfs** | `tmpfs: /usr/share/elasticsearch/data` | 内存态、最快；但容器/重启即丢数据，dev 下 ES 索引每次丢失，影响面大 |

- **选定 B**：唯一能同时满足"CI 可写 + 本地可写 + 无需宿主机路径"的方案，改动最小（compose 两处），无权限预置。A 的 env 覆盖可作为后续需要宿主机持久化时的演进（本变更不引入）；C 因数据易失被否。

### D2 就绪判据：容器健康检查 + wait_for_status=yellow（推荐）vs 仅端口就绪

- **端口就绪不足的原因**：ES 进程在 bootstrap 后即绑定 9200（监听端口），但**集群状态尚未就绪**（分片分配/集群状态评估进行中）。此时 `curl http://localhost:9200/_cluster/health` 可能返回非 200 或连接被 reset；`audit` 服务启动即连 9200，撞上未就绪窗口 → `audit mq init attempt ... connection reset` 循环 → `service failed to listen: audit`。
- **选定**：ES healthcheck 改为 `curl -sS "http://localhost:9200/_cluster/health?wait_for_status=yellow&timeout=5s" | grep -qE '"status"[[:space:]]*:[[:space:]]*"(yellow|green)"'`（YAML `CMD-SHELL` 参数用单引号包裹避免双引号转义），healthcheck 参数 `interval: 10s, timeout: 5s, retries: 30, start_period: 40s`。**关键**：`/_cluster/health` 无论集群状态均返回 HTTP 200，`curl -f` 只校验状态码、无法区分 yellow/red（超时时返回 200 + `status:red`）——因此必须**解析响应体**断言 `status` 为 yellow/green，否则容器会在集群未就绪时被误判 `healthy`（与 `wait_for_port` 缺陷行为等价）。
- **超时对齐（关键）**：`wait_for_container_health` 默认 `60 attempts × 2s = 120s`，而 healthcheck 标记 `healthy` 的检测粒度受 `interval` 影响。ES 改为调用 `wait_for_container_health "$container" 180 2`（**360s**，覆盖 ES 冷启动 + healthcheck 首个成功检测窗口；原 ES 特判为 120s×2=240s，不得低于原窗口）。**脚本超时 MUST ≥ healthcheck 就绪检测最坏窗口**（`start_period 40s + interval 10s` 级别），否则会在 ES 实际就绪前以 `dependency not healthy` 退出，修复目标不达成。
- `ci-rpc-stack.sh` 移除 ES 的 `wait_for_port 9200` 特判，改走 `wait_for_container_health`（与 postgres/redis/rabbitmq/etcd/gorse 一致）；**全局 `DEPENDENCY_PORTS` 端口扫描（含 9200）作为所有依赖的通用安全网保留**，移除的只是 ES 单独的 `wait_for_port` 特判——ES 的**主就绪判据 = 容器 healthcheck**。
- **理由**：`wait_for_status=yellow` 是 ES 官方的集群就绪判据，强于"端口监听"与"裸 `_cluster/health` 返回 200"（后者在 red 状态也可能返回 200）。

### D3 验证（真验收，非"workflow 绿"）

三件事分开取证：
1. **栈启动**：ES 是否 bootstrap 成功（无 AccessDenied）、容器 health=healthy、`audit` 是否 `listening`。
2. **测试执行**：日志出现测试执行行（`running tests` / `ok` / `--- FAIL` / `PASS`），非栈未起即失败。
3. **测试阶段真实失败清单**：首次暴露的失败用例逐条列出（若有）。

命令：`gh run view <run_id> --log | grep -E "dependency port not ready|service failed to listen|running tests|ok  |--- FAIL|PASS"`。合并后观察 2–3 次 Integration run，记录"栈启动成功率"。

## 阶段 2 前置口径重写（③）

> **此前表述（已被 3A 实证推翻）**：曾记为"F1/F2 阻塞 Integration"。
> **3A 实证**：15/15 失败 run 均为栈启动失败（ES 数据卷不可写），测试断言失败 = 0 —— F1/F2 与 Integration 红无因果关系。
>
> **阶段 2 前置条件（重写）**：
> 1. F2（fix-inventory-cache-invalidation）合并；
> 2. ES 栈就绪修复（本变更）合并；
> 3. 按 A3 口径观察：14 天内满 20 次 Integration run 绿率 ≥90%，且同一 commit 至少 3 次全绿。
>
> 满足后方可执行阶段 2（把 Integration 设为 required）。**不改 CI 配置或 ruleset。**

## Risks / Trade-offs

- **[命名卷导致 dev 历史 ES 数据不复用]** → audit 索引可重建；如需保留可后续引入 env 覆盖路径（A 方案演进）。
- **[wait_for_status=yellow 超时]** → 单节点 yellow 正常；healthcheck `interval: 10s, timeout: 5s, retries: 30, start_period: 40s`（就绪检测最坏窗口 ≈ 340s，该 340s 为 Docker 标记 `unhealthy` 的阈值；假设 CI runner 资源足以让 512m 单节点 ES 在该窗口内完成 bootstrap），`ci-rpc-stack` 对 ES 用 `wait_for_container_health "$container" 180 2`（360s）覆盖该窗口（任务 1.3 超时 MUST ≥ 检测窗口）。
- **[改动 CI 编排影响其它依赖]** → 仅改 ES 服务段与 ES 就绪分支，其它依赖不变。
- **[其它依赖同为 macOS 绝对路径 bind mount（postgres/etcd/gorse）]** → 低概率风险（explore-brief 假设其就绪正常；runner 环境若变化可能遇到与 ES 同类不可写问题）。本变更不改它们；**task 2.1 的本地验证（确认其它依赖容器均存活）同时作为该风险的取证点**；若未来复现同类失败，按 ES 同口径处理（命名卷/健康检查）。

## Migration Plan

1. 改 `docker-compose.yaml`（ES 命名卷 + healthcheck）。
2. 改 `ci-rpc-stack.sh`（ES 走 `wait_for_container_health`）。
3. 本地跑 `ci-rpc-stack.sh` 验证栈启动（ES healthy + audit listening）。
4. 提 PR → 触发 Integration → 三件事分开取证。
5. 合并后观察 2–3 次 run 记录栈启动成功率。

**回滚**：改动为 compose + 脚本，可 git revert。