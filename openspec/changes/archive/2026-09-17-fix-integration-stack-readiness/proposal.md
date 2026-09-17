## Why

Integration 流水线**系统性失败于 ES 栈启动**（15/15 失败 run 签名一致：`audit mq init attempt N/30 failed: 127.0.0.1:9200` + `service failed to listen: audit`，测试断言失败 = 0）。根因（3B，CI 附件 `dependency-logs/elasticsearch.log`）：

```
ERROR: fatal exception while booting Elasticsearch
java.lang.IllegalStateException: failed to obtain node locks, tried [/usr/share/elasticsearch/data]
Caused by: java.nio.file.AccessDeniedException: /usr/share/elasticsearch/data/node.lock
ERROR: Elasticsearch exited unexpectedly, with exit code 1
```

- **数据卷不可写**：`construct/depend/docker-compose.yaml:95-96` 将 ES 数据卷挂载为 **macOS 本机绝对路径 `/Volumes/Fan/docker-data/depend/es-data`**；ubuntu runner 上该目录不存在/不可写 → ES bootstrap 失败 → 容器退出 → 9200 不监听 → `audit` 服务连不上 → 集成测试从未真正执行。
- **就绪判据不足**：`scripts/ci-rpc-stack.sh:283-291` 对 ES 用 `wait_for_port 9200`（端口就绪即放行），未用 docker-compose 已声明的 ES healthcheck（`docker-compose.yaml:100-102`，`curl _cluster/health`）。端口就绪不等于集群可用（ES 绑定 9200 早于集群就绪），audit 连入仍可能 `connection reset`。

## What Changes

- **① ES 数据卷 → 命名卷（named volume）**：`docker-compose.yaml` 中 ES 的 `volumes` 由主机 bind mount（macOS 绝对路径）改为 Docker 命名卷 `es-data:/usr/share/elasticsearch/data`，并在 `volumes:` 段声明 `es-data`。命名卷由 Docker 管理、任何平台默认可写，消除对宿主机路径的依赖。方案对比见 design D1。
- **② ES 就绪判据 → 容器健康检查（wait_for_status=yellow）**：`docker-compose.yaml` 的 ES healthcheck 改为 `curl -fsS "http://localhost:9200/_cluster/health?wait_for_status=yellow&timeout=5s"`；`ci-rpc-stack.sh` 移除 ES 的 `wait_for_port 9200` 特判，改走 `wait_for_container_health`（与其它依赖一致），保留端口兜底。为何端口就绪不足见 design D2。
- **③ 阶段 2 前置口径重写**：本变更 review-log/design 中"阶段 2 前置条件"改为 = F2 合并 + ES 栈就绪修复合并 + 按 A3 口径观察（14 天 20 run 绿率 ≥90% + 同 commit 3 次全绿），并注明"此前记为 F1/F2 阻塞 Integration 的表述已被 3A 实证推翻"。

## Capabilities

### New Capabilities
<!-- 无：不新增能力。 -->

### Modified Capabilities
- `ci-pipeline-reliability`: "Integration 流水线稳定转绿" 需求新增场景——ES 数据卷跨平台可写、ES 就绪判据为健康检查（wait_for_status=yellow）、栈启动后业务服务监听且测试真正执行。

## Impact

- **CI 编排**：`construct/depend/docker-compose.yaml`（ES 数据卷 + healthcheck）、`scripts/ci-rpc-stack.sh`（ES 就绪判据）。
- **门禁**：Integration 流水线；Build/Quality 不受影响。
- **本地开发**：dev 环境 ES 数据卷由 bind mount 改为命名卷，原宿主机路径 `/Volumes/Fan/docker-data/depend/es-data` 的历史 ES 数据对容器不可见（audit 索引可重建，影响可接受）；**如需保留历史数据，可手动将原目录内容复制进命名卷 `es-data`**（或 `docker volume rm es-data` 彻底重置）。

## Non-Goals

- 不改任何测试断言（不修 F1/F2 类测试问题）。
- 不修 F2 inventory 缓存缺陷（另案，已合并）。
- 不把 Integration 设为 required status check（阶段 2 才做）。
- 不动 ruleset / branch protection。
- 不修 ES 运行期资源（内存）等非栈启动类问题（未观测到，超出本变更范围）。