# CI-FIX 知识库

当处理 `ci-fix` 类型的 Issue 时，**必须先阅读本文件**，了解之前类似问题的修复模式后再动手。

---

## Build — `go: invalid GOWORK`

### 症状
- 所有或大部分服务构建失败
- 错误：`go: invalid GOWORK: not an absolute path`
- Go 1.25+

### 修复
在 CI 构建脚本中强制设置 `GOWORK` 为绝对路径：
```bash
export GOWORK="$ROOT_DIR/go.work"
```

### 验证
```bash
go build ./...
```

---

## Build — 单服务构建失败

### 症状
- 某个服务构建失败，其他正常
- 报错：undefined reference、import cycle、missing dependency

### 修复
1. 检查该服务的 `go.mod`，确认依赖是否正确
2. 运行 `go mod tidy` 修复依赖
3. 检查是否有未提交的 protobuf/mock 生成代码

### 验证
```bash
cd services/<name> && go build ./...
```

---

## Mock Consistency

### 症状
- `git diff --exit-code` 在 `make mock` 后返回非零
- diff 显示 `go.mod` 依赖变化或 mock 文件不同

### 根因
`make mock` 重新生成代码后，`go mod tidy` 自动清理了不再使用的 indirect 依赖，或 mock 文件本身有变更。

### 修复
```bash
make mock
git add -A
git commit -m "chore: sync generated code after mock regeneration"
```

### 预防
每次新增/修改接口后，先跑 `make mock` 确保生成代码已提交，避免 CI 中才发现不一致。

---

## Coverage Gate

### 症状
- 报错：`coverage below threshold (X% < 63.6%)`
- Quality workflow 中的 `Coverage Gate` job 失败

### 修复
1. 为新代码补充单元测试，提升覆盖率
2. 运行 `make test-unit` 确认覆盖率达标
3. 如果新代码是纯配置/框架代码，可以将测试放在已有测试文件中

### 验证
```bash
make test-unit
# 确认输出中有 coverage: XX.X% of statements
```

---

## Go Vet

### 症状
- `go vet ./...` 报错
- 常见：`unreachable code`, `composite literal uses unkeyed fields`, `mismatched args`

### 修复
```bash
go vet ./services/<name>/...
```
根据具体报错修复代码逻辑。

---

## Govulncheck

### 症状
- 依赖包存在已知安全漏洞
- 输出：`Vulnerability #1: GO-XXXX-XXXX`

### 修复
```bash
go get <package>@latest
go mod tidy
```

---

## Integration Test

### 症状
- Integration workflow 失败
- 通常需要外部服务（MySQL, Redis, Consul 等）

### 修复
1. 检查 CI 环境是否有必要的外部服务
2. 确认测试配置中的服务地址是否匹配 CI 环境
3. 如果需要 mock 外部服务，确认 mock 是否正确

---

## 案例目录

具体案例见 `CI-FIX-CASES/` 目录：

| Issue | PR | 问题 | 状态 |
|-------|----|------|------|
| #58 | #57 | Build (GOWORK), Mock Consistency, Coverage | 待修复 |
| — | — | **事故复盘**: CI 修复链膨胀 (2026-06) | `postmortem-2026-06-ci-chain-explosion.md` |

### 循环检测

如果 ci-fix-monitor 发现同一个 PR 产生了 ≥2 个 ci-fix Issue，或连续 3 个 Issue 都是 fix-of-fix：
- **暂停自动修复**，不要创建新 Issue
- 检查原始 PR 的状态，定位工具链或配置层面的根因
- 先修根因，再修具体代码
