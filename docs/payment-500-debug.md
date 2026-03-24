# 支付 500 错误调试报告

## 问题描述
用户点击"立即支付"按钮后，弹出 HTTP 500 错误。

## 错误排查过程

### 1. 初步分析
通过 Chrome DevTools Network 面板发现错误响应：
```
POST /douyin/payment/create HTTP/1.1 500
Response Body: rpc error: code = Unknown desc = sql: unknown driver "postgres" (forgotten import?)
```

### 2. 根因定位
Payment 服务使用 PostgreSQL 数据库 (`PostgresConfig`)，但 `go.mod` 中缺少 PostgreSQL 驱动 (`github.com/lib/pq`)。

### 3. 修复方案

#### 步骤 1: 添加 PostgreSQL 驱动依赖
在 `services/payment/go.mod` 中添加：
```go
github.com/lib/pq v1.12.0
```

#### 步骤 2: 导入驱动
在 `services/payment/internal/svc/servicecontext.go` 中添加空白导入：
```go
import (
    _ "github.com/lib/pq"
    // ... other imports
)
```

或者创建 `services/payment/internal/db/pq_dep.go`:
```go
package db
import _ "github.com/lib/pq"
```

### 4. 修复结果
重新编译并重启 payment 服务后，错误变为：
```
pq: syntax error at or near "," at column 20 (42601)
```

这说明 PostgreSQL 驱动已成功加载，但 SQL 查询语法不兼容。

### 5. 当前问题
SQL 查询使用 MySQL 语法（反引号），但 PostgreSQL 需要双引号：

**当前 (MySQL 语法)**:
```sql
SELECT `payment_id`,`pre_order_id`,... FROM `payments` WHERE `order_id` = 'xxx'
```

**应为 (PostgreSQL 语法)**:
```sql
SELECT "payment_id","pre_order_id",... FROM "payments" WHERE "order_id" = 'xxx'
```

## 涉及文件
- `services/payment/go.mod` - 添加 lib/pq 依赖
- `services/payment/internal/svc/servicecontext.go` - 添加 _ "github.com/lib/pq" 导入

## 待处理
- 需要修改 dal/model/payment 中的 SQL 查询，使用 PostgreSQL 兼容语法
- 或者统一使用 go-zero 的 sqlx 抽象，让其自动处理不同数据库的语法差异
