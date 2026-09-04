# Review Log — observability-wiring-fix（整体反思 Reflect 阶段）

> 基线：explore-brief.md（Checklist 10 项）
> 审核方式：@openspec-reviewer 只读子智能体，逐轮独立审核

---

## 审查轮次 1（2026-08-30）

**范围**：完整变更包（proposal / design / specs / tasks）对照 explore-brief.md 基线整体复查。

**Checklist 10 项**：10/10 ✅（D3 项带瑕疵注记）。

**结论：FAIL（3 🔴 + 5 🟡 + 3 💡）**

### 🔴 遗留
1. **PrometheusExt 影响范围事实错误**：proposal/design/tasks 声称"仅 order 使用"，实为 13 个服务；且 **search 是活代码**（`search.go:34` 启动 StatsHTTPServer，承载 `/metrics` + `/stats/rag/overview|trend|breakdown|events` 四个业务 API）——按 tasks 1.2 执行会导致验证永不通过或砍掉业务功能。
2. **design Context 自相矛盾**："所有 dev 配置均无 DevServer" 失真，auths（11000）/audit（11008）dev 已有 DevServer。
3. **prometheus.yml target host 缺陷**：全部 target 用 `localhost`，但 Prometheus 跑在 bridge 网络容器内，抓不到宿主机服务——spec Req2"全部 target UP"场景无法验收。

### 🟡
1. 文档路径修正无 spec requirement；2. 采样率 scenario 不可验收；3. design 承诺的端口预检无 task；4. D3 表 search 行占位猜测值；5. Req6 未覆盖 PrometheusExt 移除。

---

## 修复（主智能体，第 1 轮后）

- search 显式例外：12 服务死配置全清理，search 保留机制仅对齐端口 9113→9081（proposal What Changes/BREAKING/Impact、design D1/Context/Risks、tasks 1.2/2.3、spec Req6 同步）。
- design Context 修正为"仅 auths/audit 已有"；tasks 2.2 对应调整。
- host 方案（`host.docker.internal`）纳入 proposal/design D3/spec Req2 新增场景/tasks 3.1。
- 🟡1→proposal 注明文档修正不进 spec；🟡2→spec Req3 弱化 + tasks 5.4 补抽验；🟡3→新增 tasks 4.4 端口预检；🟡4→D3 表 search 行改实证值 8081/9081；🟡5→Req6 扩展 3 场景。

---

## 审查轮次 2（2026-08-30）

**结论：FAIL（2 🔴 + 2 🟡 + 3 💡）**——第 1 轮 8 项中 6 项完全修复。

### 🔴 遗留
1. **死配置清单成员错位**：carts 被列入清单但实际 config.go 与 etc yaml 均无 PrometheusExt（零命中）；admin 有死配置（config.go:15,18 + admin.yaml:53）却被遗漏；users 仅 yaml 块（config.go 无字段）——spec Req6 场景 2 以错误清单为验收对象，验收失效。
2. **不存在的 prod 配置**：system/activity/admin（及 search/gateway）根本没有 prod.yaml，"为其 prod 配置补 DevServer"指向不存在文件，主动作不可执行。

### 🟡
1. gateway 8889 违反 spec Req1 "指标端口 = 服务端口 + 1000" MUST 条款（应 9888）；2. tasks 2.2 计数 13 错误（应 14）。

---

## 修复（主智能体，第 2 轮后）

- 清单修正（四份制品统一）：yaml 块删除 = auths/users/product/order/checkout/payment/inventory/audit/coupons/system/activity/admin（12 个）；config.go 字段删除 = 除 users 外 11 个；carts 不在清理范围。
- prod 修正：system/activity/admin/search/gateway 无 prod 配置 → 本 change 不创建，prod 侧随线上部署独立 change（proposal/design Context/D3 脚注/Risks/Migration/tasks 2.3 同步）。
- gateway 端口 8889→9888，服从 +1000 规则消除例外（design D3/tasks 2.1/3.1/4.4 同步）。
- tasks 2.2 计数 13→14；💡：design pprof 措辞对齐（dev 也统一 false）、tasks 3.2 前提修正、tasks 1.1 改 git grep 全仓库口径。

---

## 审查轮次 3（2026-08-30）

**结论：PASS（0 🔴）**——第 2 轮 7 项中 6 项完全修复，1 项残留降级为 🟡。

**关键实证**（grep + go-zero v1.10.1 源码）：
- PrometheusExt 分布：13 yaml（12 死 + search 活）/ 12 config.go（11 死 + search）/ carts、users 零命中 ✅
- prod yaml 恰 10 个且全部含 DevServer ✅；gateway 9888 全仓库无占用 ✅
- `devserver.StartAgent` 仅判 `Enabled` 不依赖 Mode（server.go:80-82）；gateway DevServer 链路源码级验证生效（gateway/server.go:48 → rest/server.go:49 → serviceconf.go:77）✅
- 服务清单、端口值（110xx/9888/9081）、计数、Non-goals 在四份制品间交叉一致，无新引入矛盾 ✅

### 🟡（已顺手修复）
1. design Risks 条目漏列 gateway（"四个服务"→"五个服务"）。

### 💡（已顺手采纳 2 项）
1. proposal Impact 补全 carts/gateway dev yaml 变更说明；2. design Open Questions 补 gateway DevServer 源码预验证结论与 promtail 9081 端口重合提示。

---

## 最终结论

**✅ 通过（🔴 清零，3 轮收敛）**。explore-brief.md Checklist 10/10 ✅，四项追溯（proposal↔specs、design↔specs、tasks↔specs、Non-goals 越界检查）全部通过。

变更包已冻结，可进入 `/openspec-apply-change` 实施。
