# Explore Brief · fix-integration-stack-readiness

> 立案探索（explore-only，不 propose）：Integration 流水线"栈启动失败"归因与修复方向。

## 背景（实证）

- **失败签名分类（3A，9-14 起 15 个 Integration failure run）**：**100% 为栈启动失败**（`audit mq init attempt N/30 failed: 127.0.0.1:9200 connection reset/refused` + `service failed to listen: audit`）；**测试断言失败计数 = 0**（栈未起、测试未执行）。
  代表 run：35091030961 / 35089155705 / 35088587189 / 35088089665 / 35079947384 / 34955862416 / 34955463117 / 34953758435 / 34953393830 / 34946062980 / 34940211709 / 34939050248 / 34937623982 / 34833333856 / 34832652000。
- **根因（3B，CI 附件 dependency-logs/elasticsearch.log）**：
  ```
  ERROR: fatal exception while booting Elasticsearch
  java.lang.IllegalStateException: failed to obtain node locks, tried [/usr/share/elasticsearch/data]
  Caused by: java.nio.file.NoSuchFileException: /usr/share/elasticsearch/data/node.lock
  Caused by: java.nio.file.AccessDeniedException: /usr/share/elasticsearch/data/node.lock
  ERROR: Elasticsearch exited unexpectedly, with exit code 1
  ```
  → **ES bootstrap 失败（数据目录不可写）→ 容器退出 → 9200 不监听 → audit 连不上**。
- **代码事实**：
  - `scripts/ci-rpc-stack.sh:283-291`：ES 分支用 `wait_for_port 9200`（端口就绪即放行；注释 = fix-ci-integration-pipeline D1），**未用 docker-compose 的 healthcheck**；ES 启动失败时 wait_for_port 120s 超时前容器已退出。
  - `construct/depend/docker-compose.yaml:95-96`：ES 数据卷 = **`/Volumes/Fan/docker-data/depend/es-data:/usr/share/elasticsearch/data`**（macOS 本机绝对路径）→ **ubuntu runner 上不存在/不可写**。
  - `construct/depend/docker-compose.yaml:100-102`：ES healthcheck（`curl -fsS localhost:9200/_cluster/health`，interval 30s）**存在但未被 ci-rpc-stack 使用**。

## 目标 / 非目标

- **目标**：Integration 能稳定进入"测试真正执行"阶段（消除栈启动失败；不含修测试断言/业务）。
- **非目标**：修 F2（inventory 缓存）；改 ruleset required；阶段 2 动作；修测试断言（F1 类）。

## 假设清单（均标注「假设（未证实）」+ 取证方法）

1. **假设（未证实）**：`/Volumes/Fan/docker-data/depend/es-data` 在 ubuntu runner 上不可写/不存在 → ES bootstrap 失败。证据支持（ES 日志 AccessDenied on node.lock + 数据卷为 macOS 路径）；取证：docker compose 运行后 `ls -la` 该挂载路径 + `docker inspect` ES 挂载状态。
2. **假设（未证实）**：修复 ES 数据卷（CI 覆盖为可写路径）后 ES 能正常启动 → audit 能 listen → Integration 进入测试执行。取证：修复后跑一次 Integration 看 audit 是否 listen。
3. **假设（未证实）**：ES 失败形态 = bootstrap 一次性退出（非运行中崩溃/OOMKilled）。日志支持（exit code 1，bootstrap 阶段）；取证：诊断附件的 docker-ps / ES 日志无 OOMKilled 字样。
4. **假设（未证实）**：其它依赖（postgres/redis/rabbitmq/etcd/gorse）在 CI 就绪正常。证据支持（其日志无致命错误）；取证：诊断附件各 dep-*.health。

## 下一步取证方案（修复前必须做）

1. 在 CI 加一步 `docker inspect go-mall-elasticsearch` + 挂载路径 `ls -la`（确认数据卷为何不可写）——或直接改 docker-compose ES 数据卷为 CI 可写路径（如 `./.artifacts/es-data` 或移除 bind mount 用匿名卷）。
2. 复用 docker-compose 的 ES healthcheck（让 ci-rpc-stack 的 `wait_for_container_health` 生效），或把 ES 就绪判断从"端口"改为"健康检查"。
3. 修复后跑 Integration 验证 audit listen + 测试真正执行（0 栈失败）。

## Checklist（立案基线）

- [ ] C1 Integration failure run 中栈启动失败占比 100%（已实证，15/15）。
- [ ] C2 ES 根因 = 数据卷不可写（bootstrap 失败），非测试断言（已实证日志）。
- [ ] C3 修复方向 = ES 数据卷路径/健康检查（不涉及测试/业务/ruleset）。
- [ ] C4 修复后 Integration 无"service failed to listen"栈失败，进入测试执行阶段。