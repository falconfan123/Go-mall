# Explore Brief: 商品搜索热搜词

> 本文件是 explore 阶段需求讨论的落盘基线（checklist），供提案审核全程引用，防止会话上下文丢失。
> 对应文章优化点 4：`/opsx-explore` 讨论必须落盘。

## 背景

- SearchService 当前仅暴露 `ParseQuery` / `Search` 两个 RPC（services/search/search.proto）。
- 商城前端缺少"热门搜索词"引导位；运营侧缺少搜索热度数据。
- 搜索行为目前未落库，无法做任何热度聚合。
- 已有 `internal/stats` 包（Postgres store + Metrics），但只统计对话/RAG 事件，与搜索词热度无关，**不可复用其表结构**，可复用其"事件落库 + 定时聚合"的模式。

## 目标（用户视角）

1. 前端可调用 `HotKeywords` RPC 获取热门搜索词列表（默认 Top 10 / 近 7 天）。
2. 每次用户搜索行为可被记录，支撑热度聚合与未来运营分析。

## 非目标（明确排除）

- 不做个性化推荐（不按 user_id 区分热度）。
- 不做实时 TopK 推送（允许分钟级延迟，聚合任务定时跑即可）。
- 不改 Search 主链路性能（埋点必须异步/旁路，失败不影响搜索主流程）。
- 不新增前端页面（只提供 RPC 出口）。

## 待决策点

- D1: 热度聚合放哪一层？（候选：a) Postgres 定时聚合到 hot_keywords 表；b) 每次查询实时 count 聚合 → 倾向 a，避免大表实时聚合）
- D2: 搜索词事件的去重口径（同一 user 同一窗口内重复搜索同一词，算 1 次还是 N 次？→ 倾向按"用户去重"计热度，防刷）
- D3: 规范化：热度基于 `normalized_query`（去空白/大小写/同义实体）还是原始 query？（→ 倾向 normalized，避免"iPhone"与"iphone"分裂成两条）
- D4: 数据保留窗口（→ 倾向 30 天原始事件 + 7 天聚合表，定期清理）

## Checklist（审核时逐项核对）

- [ ] proposal 是否覆盖上述全部目标与非目标
- [ ] 埋点是否明确"异步旁路、失败不影响主流程"
- [ ] 热度口径（用户去重、normalized）是否落到 requirement
- [ ] HotKeywords 请求/响应字段是否完整定义（窗口、条数、排序）
- [ ] 数据清理/容量控制是否有 requirement 兜底
