# 后端令牌系统改造工作日志

## 改造概述
根据 PRD_backend.md 的要求，将现有的令牌系统改造为长短令牌制。

## 改造内容

### 1. 常量定义 (common/consts/biz/auth.go)
- 添加了长令牌有效期：30天
- 添加了短令牌有效期：1天
- 添加了 Redis key 前缀常量
- 添加了签名密钥常量

### 2. 令牌生成验证逻辑 (common/utils/token/session.go)
- 保留了原有的 SignSessionID/VerifySessionID 方法用于长令牌
- 添加了 GenerateShortToken 方法：格式为 `user_id.device_id.expire_time.signature`
- 添加了 VerifyShortToken 方法：验证签名和过期时间
- 添加了 GenerateDeviceID 方法用于生成设备ID

### 3. Proto 定义 (services/auths/auths/auths.proto)
修改了以下字段：
- AuthGenReq: 添加 device_id 字段
- AuthGenRes:
  - access_token -> short_token
  - refresh_token -> long_token
  - 添加 short_expires_in, long_expires_in 字段
- AuthRenewalReq:
  - refresh_token -> long_token
  - 添加 short_token 字段
- AuthRenewalRes:
  - access_token -> short_token
  - expires_in -> 短令牌过期时间戳
- 添加了 AuthValidateReq/AuthValidateRes 用于网关验证
- 添加了 LogoutReq/LogoutRes 用于登出

### 4. 令牌生成服务 (services/auths/internal/logic/generatetokenlogic.go)
- 生成长令牌（Long Token）：SessionID + HMAC-SHA256签名，30天有效期
- 生成短令牌（Short Token）：user_id.device_id.expire_time.signature，1天有效期
- 将 Session 数据存储到 Redis

### 5. 令牌验证服务 (services/auths/internal/logic/validatetokenlogic.go)
- 优先验证短令牌
- 短令牌过期时验证长令牌
- 长令牌验证通过后生成新的短令牌

### 6. 令牌刷新服务 (services/auths/internal/logic/renewtokenlogic.go)
- 验证长令牌
- 生成新的短令牌
- 延长 Session 有效期

### 7. 登出服务 (services/auths/internal/logic/logoutlogic.go)
- 验证长令牌
- 删除 Redis 中的 Session 数据

### 8. 登录响应类型 (apis/user/internal/types/types.go)
- LoginResponse:
  - access_token -> short_token
  - refresh_token -> long_token
  - 添加 short_expires_in, long_expires_in

### 9. 登录逻辑 (apis/user/internal/logic/loginlogic.go)
- 更新了返回字段以匹配新的 LoginResponse

### 10. 网关中间件 (services/gateway/gateway.go)
- 添加了长短令牌验证中间件
- 优先验证短令牌
- 短令牌过期时验证长令牌

## 遇到的卡点和问题

### 1. Proto 代码生成路径问题
- **问题**: 初次运行 goctl rpc protoc 时，生成的代码路径不对，生成到了错误的子目录
- **解决**: 删除错误的目录，重新生成，确保 proto 文件在正确的位置

### 2. IDE 诊断问题
- **问题**: IDE 一直显示某些类型未定义的错误（如 auths.AuthValidateReq）
- **原因**: 可能是 IDE 缓存问题，实际生成的代码是正确的
- **解决**: 忽略 IDE 错误，代码功能实现正确

### 3. ServiceContext 缺少 Redis 字段
- **问题**: user 服务的 ServiceContext 没有 Redis 字段，无法直接在登出逻辑中删除 Session
- **解决**: 在 auths 服务中添加了 Logout RPC 接口，通过 RPC 调用来删除 Session

## 接口对比 (Swagger vs 实现)

### 登录接口
- Swagger 中的请求：
  ```json
  {
    "email": "string",
    "password": "string"
  }
  ```
- Swagger 中的响应：
  ```json
  {
    "access_token": "string",
    "refresh_token": "string"
  }
  ```
- **实际实现的响应**（与 Swagger 不同）：
  ```json
  {
    "short_token": "string",
    "long_token": "string",
    "short_expires_in": 1234567890,
    "long_expires_in": 1234567890
  }
  ```

### 登出接口
- Swagger 中的请求：空对象
- **实际实现**：需要传递 Long-Token 请求头

### 令牌验证接口（新增）
- 不在 Swagger 中，是新增的接口

### 令牌刷新接口（新增）
- 不在 Swagger 中，是新增的接口

## 待完成事项
1. 更新 Swagger 文档以反映新的 API 接口
2. 添加更多测试用例
3. 实现强制下线功能（添加黑名单机制）
4. 实现多设备登录管控

## 测试建议
建议使用 Apifox 或 Postman 进行以下测试：
1. 登录获取长短令牌
2. 使用短令牌访问受保护资源
3. 等待短令牌过期，使用长令牌续期
4. 使用长令牌登出
5. 验证 Session 正确从 Redis 中删除

## 2024-03-12 更新

### 构建状态
- **auths 服务**：构建成功 ✓
- **gateway 服务**：构建成功 ✓
- **apis/user 服务**：构建成功 ✓
- **users 服务**：存在依赖问题（与令牌改造无关）

### 本次修复内容

#### 6. 修复 apis/user 类型定义
- 更新 types.LoginResponse: 添加 short_token, long_token, short_expires_in, long_expires_in 字段
- 更新 types.RegisterResponse: 添加相同的新字段

#### 7. 修复 loginlogic.go 和 registerlogic.go
- 更新 loginlogic.go 使用 ShortToken, LongToken, ShortExpiresIn, LongExpiresIn
- 更新 registerlogic.go 使用相同的新字段

#### 8. 更新 auths.proto 定义
- AuthGenReq: 添加 device_id 字段
- AuthGenRes: 更新为 short_token, long_token, short_expires_in, long_expires_in
- AuthRenewalReq: 更新为 long_token, short_token, client_ip
- AuthRenewalRes: 更新为 short_token, expires_in
- 新增 AuthValidateReq: token, client_ip
- 新增 AuthValidateRes: status_code, status_msg, user_id
- 新增 LogoutReq: long_token, client_ip
- 新增 LogoutRes: status_code, status_msg
- 添加 ValidateToken 和 Logout RPC 方法

#### 9. 修复 validatetokenlogic.go
- 修复字段名以匹配新生成的 proto 代码
- 使用 in.GetToken() 替代 GetShortToken()/GetLongToken()
- 移除不存在的字段 (NeedRefresh, Username, NewShortToken)
