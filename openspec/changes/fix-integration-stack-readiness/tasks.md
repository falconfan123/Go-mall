> CI 编排变更（compose + ci-rpc-stack.sh）；白名单 = construct/depend/docker-compose.yaml + scripts/ci-rpc-stack.sh（越界即停）。
> 验收口径：栈启动 + 测试真正执行（非"workflow 绿"）。

## 1. 修复（D1/D2）

- [ ] 1.1 `docker-compose.yaml` ES 数据卷改为命名卷 `es-data:/usr/share/elasticsearch/data`，并在 `volumes:` 段声明 `es-data`；验证：`docker compose -f construct/depend/docker-compose.yaml config` 解析通过 + diff
- [ ] 1.2 `docker-compose.yaml` ES healthcheck 改为 `curl -sS "http://localhost:9200/_cluster/health?wait_for_status=yellow&timeout=5s" | grep -qE '"status"[[:space:]]*:[[:space:]]*"(yellow|green)"'`（解析响应体，非 `curl -f`；`CMD-SHELL` 参数用单引号包裹），参数 `interval: 10s, timeout: 5s, retries: 30, start_period: 40s`；验证：diff + `docker inspect --format='{{.State.Health.Status}}' go-mall-elasticsearch`（start_period 内健康状态为 `starting`，属正常；循环重试至 `healthy`/`unhealthy` 或超时，勿据此误判配置失败）
- [ ] 1.3 `ci-rpc-stack.sh` 移除 ES `wait_for_port 9200` 特判，ES 改走 `wait_for_container_health "$container" 180 2`（**360s**，不得低于原 240s 窗口，须覆盖 healthcheck 就绪检测窗口）（与其它依赖一致）；**全局 `DEPENDENCY_PORTS` 端口扫描（含 9200）保留为通用安全网，移除的只是 ES 单独特判——ES 主就绪判据 = 容器 healthcheck**；验证：diff + 本地运行输出

## 2. 本地验证

- [ ] 2.1 本地跑 `bash scripts/ci-rpc-stack.sh`（或依赖栈启动段）：ES 容器 `health=healthy`、无 `AccessDenied`、`audit` 服务 `listening`，且 `postgres/redis/rabbitmq/etcd/gorse` 容器均正常存活（compose 改动无副作用）；验证：原始输出（docker ps + ES 日志 + audit 日志）
- [ ] 2.2 栈启动后确认集成测试能真正执行（存在测试执行输出，非栈未起即失败）；验证：输出片段
- [ ] 2.3 本地 `start-unified.sh` 依赖栈启动验证（与 CI 共用同一 compose；命名卷变更不得破坏本地 dev 启动）；验证：`bash scripts/start-unified.sh` 输出（依赖容器 healthy）

## 3. CI 真验收

- [ ] 3.1 提 PR → 触发 Integration → 三件事分开取证：① 栈启动（ES 就绪、audit listen）② 测试是否真正执行 ③ 测试阶段真实失败清单（若有，逐条列用例名）；验证：`gh run view <run_id> --log | grep -E 'dependency port not ready|dependency not healthy|dependency failed to start|service failed to listen|AccessDenied|running tests|ok[[:space:]]|--- FAIL|PASS'`（含切换后新失败签名 `dependency not healthy` / `dependency failed to start`；`ok` 后为制表符）
- [ ] 3.2 合并后观察 2–3 次 Integration run，记录"栈启动成功率"；验证：run 列表 + 签名
- [ ] 3.3 阶段 2 口径重写落地（design.md「阶段 2 前置口径重写」节 + review-log.md 轮次记录）：F2 合并 + ES 栈修复合并 + A3 观察（14 天 20 run 绿率 ≥90% + 同 commit 3 次全绿）；注明"F1/F2 阻塞 Integration 表述已被 3A 推翻"；验证：文本位置（design 节名 + review-log）

## 4. 交付

- [ ] 4.1 红线：`git diff --name-only` 白名单 = construct/depend/docker-compose.yaml + scripts/ci-rpc-stack.sh（+ 本变更 openspec）；无测试断言改动、无 ruleset 改动；验证：命令输出
- [ ] 4.2 提 PR → Build + Quality 绿（Integration 按归因纪律；修复后应体现栈启动成功）→ 合并；验证：PR 号 + merge commit hash