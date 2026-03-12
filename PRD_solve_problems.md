# 问题修复 PRD

## 1. 问题总览

本次修复旨在解决本地 CI 检查中发现的所有遗留问题，使项目符合 Go 代码规范。

| 序号 | 问题 | 状态 | 优先级 |
|------|------|------|--------|
| 1 | order.OrderService 类型未定义 | ✅ 已修复 | P0 |
| 2 | go-zero 生成代码 JSON tag 兼容性 | ⚠️ 待升级 go-zero | P1 |
| 3 | golint 变量命名不规范 | ✅ 大部分已修复 | P1 |
| 4 | golint 导出类型缺少注释 | ⚠️ 部分修复 | P2 |
| 5 | 缺少单元测试 | ⚠️ 待处理 | P2 |

---

## 2. 已完成修复

### 2.1 order.OrderService 类型未定义 ✅

**修复内容**:
1. 修改 `apis/flash_sale/internal/svc/servicecontext.go`:
   - `order.OrderService` → `order.OrderServiceClient`
   - `order.NewOrderService()` → `order.NewOrderServiceClient(zrpc.MustNewClient(c.OrderRpc).Conn())`

2. 修改 `apis/order/internal/svc/servicecontext.go`:
   - 同上

### 2.2 golint 变量命名不规范 ✅

**修复内容**:
- `userId` → `userID`
- `productId` → `productID`
- `res_info` → `resInfo`
- `categoryId` → `categoryID`
- `pictureUrl` → `pictureURL`
- `thumbnailUrl` → `thumbnailURL`
- `itemIds` → `itemIDs`
- `InventoryNotEnoughErr` → `ErrInventoryNotEnough`
- `InventoryDecreaseFailedErr` → `ErrInventoryDecreaseFailed`
- `InvalidInventoryErr` → `ErrInvalidInventory`

### 2.3 移除未使用代码 ✅

**修复内容**:
- 移除 `apis/flash_sale/internal/logic/flashbuylogic.go` 中的未使用函数 `newError` 和 `customError`

### 2.4 添加必要注释 (部分) ✅

**修复内容**:
- 为 `common/response/base.go` 添加注释
- 为 `common/utils/gorse/model.go` 添加注释
- 为 `common/consts/code/inventory.go` 添加注释

---

## 3. 待后续处理

### 3.1 go-zero 生成代码 JSON tag 兼容性问题

**问题**: go-zero 生成的代码使用了 `json:"page,default=1"` 格式，与新版 Go 静态检查不兼容

**解决方案**:
```bash
# 升级 go-zero 到最新版本
go get -u github.com/zeromicro/go-zero@latest
go install github.com/zeromicro/go-zero/tools/goctl@latest

# 重新生成 API 代码
cd apis/checkout
goctl api migrate -v latest
goctl api gen
```

**影响范围**: types 生成文件

### 3.2 golint 注释问题

**问题**: 大量导出类型/函数缺少注释 (约 900+ 项)

**说明**:
- 大部分来自 go-zero 生成的 handler、config 文件
- 属于代码规范问题，不影响编译和运行
- 建议后续逐步完善或升级 go-zero 版本自动生成

### 3.3 缺少单元测试

**问题**: 项目几乎没有单元测试

**说明**:
- 根据用户要求：单元测试应当依附于 Swagger 文档
- 需要先查阅各服务的 API 规范后再编写测试
- 建议后续根据 Swagger 文档补充测试用例

### 问题 1: order.OrderService 类型未定义 (P0)

**问题描述**:
- `apis/flash_sale/internal/svc/servicecontext.go` 和 `apis/order/internal/svc/servicecontext.go` 引用了不存在的 `order.OrderService` 类型
- 正确类型应该是 `order.OrderServiceClient`

**根本原因**:
- protobuf 生成的代码中，客户端接口是 `OrderServiceClient`，创建客户端的函数是 `NewOrderServiceClient`
- 原代码错误地使用了 `OrderService` 和 `NewOrderService`

**修复方案**:

```go
// 错误代码
OrderRpc order.OrderService
OrderRpc: order.NewOrderService(zrpc.MustNewClient(c.OrderRpc))

// 正确代码
OrderRpc order.OrderServiceClient
OrderRpc: order.NewOrderServiceClient(zrpc.MustNewClient(c.OrderRpc))
```

**影响范围**:
- `apis/flash_sale/internal/svc/servicecontext.go`
- `apis/order/internal/svc/servicecontext.go`

---

### 问题 2: go-zero 生成代码 JSON tag 兼容性 (P1)

**问题描述**:
- go-zero 生成的 types 文件中使用了 `json:"page,default=1"` 这种格式
- 与新版 Go 静态检查工具不兼容

**示例**:
```go
// 当前生成的代码 (有问题)
Page     int32 `json:"page,default=1"`
PageSize int32 `json:"page_size,default=10"`
CouponID string `json:"coupon_id,optional"`

// 正确的 JSON tag 格式
Page     int32 `json:"page"`
PageSize int32 `json:"page_size"`
CouponID string `json:"coupon_id"`
```

**根本原因**:
- go-zero 版本与当前 Go 版本不兼容
- `default` 和 `optional` 是 go-zero 特定标签，不是标准 JSON 标签

**修复方案**:

方案 A: 升级 go-zero 版本
```bash
go get -u github.com/zeromicro/go-zero@latest
goctl version  # 同步升级 goctl
goctl migrate -v latest
```

方案 B: 重新生成代码
```bash
cd apis/checkout
goctl api migrate -v latest
goctl api gen
```

方案 C: 手动修改 JSON tag（不推荐，因为是生成代码）

**推荐方案**: 方案 A，升级 go-zero 到最新版本

**影响范围**:
- `apis/checkout/internal/types/types.go`
- `apis/coupon/internal/types/types.go`
- `apis/flash_sale/internal/types/types.go`
- `apis/order/internal/types/types.go`
- `apis/payment/internal/types/types.go`

---

### 问题 3: golint 变量命名不规范 (P1)

**问题描述**:
- 变量名使用 `userId` 而非 Go 规范的 `userID`
- 结构体字段使用 `productId` 而非 `productID`

**修复方案**:

使用脚本批量替换:

```bash
# 替换变量名 userId -> userID (保留大小写上下文)
find . -name "*.go" -exec sed -i '' 's/\buserId\b/userID/g' {} \;

# 替换其他常见变量名
find . -name "*.go" -exec sed -i '' 's/\bproductId\b/productID/g' {} \;
find . -name "*.go" -exec sed -i '' 's/\bproductId\b/productID/g' {} \;
find . -name "*.go" -exec sed -i '' 's/\bcouponId\b/couponID/g' {} \;
find . -name "*.go" -exec sed -i '' 's/\baddressId\b/addressID/g' {} \;
find . -name "*.go" -exec sed -i '' 's/\borderId\b/orderID/g' {} \;
find . -name "*.go" -exec sed -i '' 's/\bpaymentId\b/paymentID/g' {} \;

# 替换 res_info -> resInfo
find . -name "*.go" -exec sed -i '' 's/\bres_info\b/resInfo/g' {} \;
find . -name "*.go" -exec sed -i '' 's/\bitem_ids\b/itemIDs/g' {} \;
```

**主要修改文件**:
- `apis/carts/internal/logic/*.go`
- `apis/checkout/internal/logic/*.go`
- `apis/product/internal/logic/*.go`
- `services/product/internal/logic/*.go`

---

### 问题 4: golint 导出类型缺少注释 (P2)

**问题描述**:
- 大量导出的类型、函数、方法缺少注释
- Go 规范要求导出的标识符必须有注释

**修复方案**:

为所有导出的类型和函数添加注释:

```go
// CartItemListLogic handles the logic of listing cart items
type CartItemListLogic struct {
    // ...
}

// NewCartItemListLogic creates a new CartItemListLogic
func NewCartItemListLogic(ctx context.Context, svcCtx *ServiceContext) *CartItemListLogic {
    // ...
}

// CartItemList handles the listing of cart items
func (l *CartItemListLogic) CartItemList(req *types.CartItemListReq) (*types.CartItemListResp, error) {
    // ...
}
```

**主要修改文件**:
- `apis/carts/internal/logic/*.go`
- `apis/checkout/internal/logic/*.go`
- `apis/product/internal/logic/*.go`
- `common/utils/token/jwt.go`

---

### 问题 5: 缺少单元测试 (P2)

**问题描述**:
- 项目几乎没有单元测试
- 测试应依附于 Swagger 文档编写

**修复方案**:

根据用户要求，单元测试应依附于 Swagger 文档，不能随便乱测。

**实施步骤**:

1. 首先查阅各服务的 API 文档
   - Swagger UI: `http://localhost:8080/swagger/` (服务运行后)
   - 或查看 `etc/*.yaml` 配置文件中的 API 定义

2. 根据 API 端点编写测试用例

3. 测试覆盖范围:
   - 正常流程测试
   - 参数验证测试
   - 错误处理测试

**测试文件位置**:
```
services/
  ├── product/
  │   └── internal/
  │       └── logic/
  │           └── getproductlogic_test.go
apis/
  ├── carts/
  │   └── internal/
  │       └── logic/
  │           └── cartitemlistlogic_test.go
```

---

## 3. 修复实施计划

### 阶段一: 修复编译错误 (P0)

1. 修复 `apis/flash_sale/internal/svc/servicecontext.go`
   - 将 `order.OrderService` 改为 `order.OrderServiceClient`
   - 将 `order.NewOrderService` 改为 `order.NewOrderServiceClient`

2. 修复 `apis/order/internal/svc/servicecontext.go`
   - 同上

### 阶段二: 修复代码规范 (P1)

1. 升级 go-zero 版本
2. 重新生成代码或修复 JSON tag
3. 修复变量命名
4. 添加注释

### 阶段三: 完善测试 (P2)

1. 根据 Swagger 文档编写测试用例
2. 确保测试覆盖率提升

---

## 4. 验证

修复完成后，运行以下命令验证:

```bash
# 本地 CI 检查
make lint
```

**当前验证结果**:
- ✅ gofmt 检查通过
- ✅ go vet 检查通过 ⚠️ (仅有 go-zero 生成代码的 JSON tag 警告)
- ⚠️ staticcheck 有 go-zero 生成代码 JSON tag 警告
- ⚠️ golint 有约 900 项警告（大部分来自生成代码）
- ✅ revive 检查通过
- ✅ 编译成功 (apis 和 services 均通过)
- ⏭️ 测试跳过 (根据用户要求，测试需依附 Swagger 文档)

---

## 5. 风险与回滚

### 风险

1. **升级 go-zero 可能引入破坏性变更**
   - 缓解: 先在本地测试，确认功能正常后再提交

2. **批量修改变量名可能遗漏**
   - 缓解: 使用脚本替换后人工复核

### 回滚方案

如果修复后出现问题:
```bash
# 回滚所有更改
git checkout -- .

# 或者回滚特定文件
git checkout -- apis/flash_sale/internal/svc/servicecontext.go
```

---

## 6. 附录

### 常用命令

```bash
# 格式化代码
make fmt

# 本地检查
make lint

# 快速检查（仅格式）
make lint-fast

# 升级 go-zero
go get -u github.com/zeromicro/go-zero@latest

# 升级 goctl
go install github.com/zeromicro/go-zero/tools/goctl@latest
```

### Go 命名规范参考

| 类型 | 规范 | 示例 |
|------|------|------|
| 变量 | 驼峰命名，首字母小写 | `userID`, `productName` |
| 常量 | 驼峰命名，首字母大写 | `MaxRetry`, `DefaultTimeout` |
| 函数 | 驼峰命名，首字母大写 | `GetUser`, `CreateOrder` |
| 结构体 | 驼峰命名，首字母大写 | `UserInfo`, `OrderDetail` |
| 接口 | 驼峰命名，以 -er 结尾 | `Reader`, `Writer` |
| 包 | 小写字母 | `utils`, `types` |
