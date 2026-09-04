# Proposal: hot-search-keywords

## Purpose

为 SearchService 增加"商品搜索热搜词"能力：将用户搜索行为异步落库，定时聚合成热门搜索词，并通过新的 `HotKeywords` RPC 对外提供，供商城前端搜索页展示热门搜索引导。

现状：SearchService 仅有 `ParseQuery` / `Search` 两个 RPC（services/search/search.proto）；搜索行为不落库，无热度数据出口，运营与前端均无法获取热门搜索词。

## Requirements

### Requirement: 搜索词事件埋点

系统 SHALL 在每次 `Search` RPC 成功后，异步记录一条搜索词事件，包含 user_id、normalized_query、原始 query、搜索时间。

#### Scenario: 搜索成功后记录事件

- **When** 用户调用 `Search` 且返回成功
- **Then** 系统以异步旁路方式写入一条搜索词事件，且该写入失败不影响 Search 响应结果

#### Scenario: 埋点失败不阻塞主流程

- **When** 事件存储写入失败（如数据库不可用）
- **Then** Search 请求仍正常返回结果，仅在日志中记录埋点错误

### Requirement: 搜索词事件去重口径

系统 SHALL 在热度统计时，对同一用户在同一统计窗口内重复搜索的同一 normalized_query 仅计一次。

#### Scenario: 同一用户重复搜索同一词

- **When** 同一 user_id 在近 7 天内多次搜索 normalized_query 为 "iphone" 的词
- **Then** 该词对热度的贡献按 1 次计

#### Scenario: 不同用户搜索同一词

- **When** 两个不同 user_id 在窗口内分别搜索 "iphone"
- **Then** 该词热度计数为 2

### Requirement: 热搜词查询 RPC

系统 SHALL 新增 `HotKeywords` RPC，接收窗口天数（默认 7）与返回条数（默认 10），返回按热度降序的热门搜索词列表。

#### Scenario: 默认参数查询

- **When** 调用 `HotKeywords` 且不传任何参数
- **Then** 返回近 7 天热度最高的 10 个词，按热度降序

#### Scenario: 自定义窗口与条数

- **When** 调用 `HotKeywords` 且 window_days=1、limit=5
- **Then** 返回近 1 天热度最高的 5 个词

#### Scenario: 窗口无数据

- **When** 指定窗口内没有任何搜索事件
- **Then** 返回空列表，且不报错

### Requirement: 数据容量控制

系统 SHALL 定期清理超出保留期的搜索词原始事件，聚合表按窗口滚动重建，防止无界增长。

#### Scenario: 过期数据清理

- **When** 原始事件超过保留期（默认 30 天）
- **Then** 该事件被清理任务删除

## Non-Goals

- 不提供个性化热搜（热度不做 user 维度区分，仅全局 TopK）。
- 不提供实时推送/实时 TopK（允许聚合延迟，聚合任务按分钟级/小时级周期执行即可）。
- 不修改 Search 主链路（埋点为异步旁路，且不得引入同步依赖）。
- 不包含前端页面改造（本变更仅提供 RPC 出口与数据管道）。

## Assumptions

- 搜索行为的事件存储可复用现有 Postgres 基础设施（参考 internal/stats 的 postgres_store 模式），但表结构独立新建。
- 热度聚合结果可缓存，`HotKeywords` 读取聚合表即可，无需实时计算。
