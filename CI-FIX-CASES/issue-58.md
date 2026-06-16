# Issue 58 — PR #57 CI 修复

## 基本信息

| 字段 | 值 |
|------|-----|
| Issue | #58 |
| PR | #57 |
| 标题 | feat(rag): add DeepSeek AI client support |
| 创建时间 | 2026-06-14 |
| 状态 | 未修复 |

## 错误

### 1. Build — 所有 17 个服务构建失败

**Workflow:** Build (run ID: 27489270219)

**错误输出：**
```
go: invalid GOWORK: not an absolute path
```

**根因：**
GitHub Actions CI 环境中 `GOWORK` 环境变量被设为了相对路径，而 Go 1.25 要求 `GOWORK` 必须是绝对路径。`scripts/go-ci-build.sh` 调用 `go build ./...` 时继承了这个错误的 `GOWORK`。

**修复：**
在 `go-ci-build.sh` 中显式设置 `GOWORK` 为绝对路径：
```bash
export GOWORK="$ROOT_DIR/go.work"
```

或者在 CI workflow 的 `env` 中强制指定。

---

### 2. Mock Consistency

**Workflow:** Quality (run ID: 27489270220)

**错误输出：**
```
git diff --exit-code → exit code 1
```
`make mock` 生成的代码与已提交的代码不一致。

**根因：**
`make mock` 重新生成 mock 代码后，`common/go.mod` 的依赖列表会变化（`go mod tidy` 清理了不再使用的 indirect 依赖，包括 `github.com/fxamacker/cbor/v2`, `github.com/x448/float16` 等）。这些变更没有被提交到 PR 中。

**修复：**
```bash
make mock
# 检查 diff，把 go.mod 的自动变更一并提交
git add common/go.mod
git commit -m "chore: sync go.mod after mock generation"
```

---

### 3. 其他失败

| 失败项 | 关联根因 |
|--------|----------|
| Unit Tests | 可能因 mock 不一致导致 |
| Coverage Gate | 新代码覆盖率不足 63.6% |
| Go Vet | 可能因 GOWORK 问题导致 vet 失败 |
| Govulncheck | 依赖扫描 |

## PR 链接

https://github.com/falconfan123/Go-mall/pull/57
