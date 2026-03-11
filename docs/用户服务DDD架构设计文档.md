# 用户服务DDD架构设计文档

## 1. 领域建模
### 1.1 限界上下文
**用户身份与访问上下文（IAM Context）**
- 职责：负责用户身份管理、认证授权、用户信息管理、收货地址管理等核心用户领域
- 边界：包含用户聚合、地址实体、认证相关的所有业务逻辑
- 对外接口：用户注册、登录、登出、用户信息CRUD、地址CRUD等

### 1.2 领域模型映射表
| 类型 | 名称 | 说明 | 业务规则 |
|------|------|------|----------|
| 聚合根 | User | 用户核心聚合 | 1. 邮箱唯一<br>2. 密码必须加密存储<br>3. 最多10个收货地址<br>4. 必须有一个默认地址 |
| 实体 | Address | 收货地址实体 | 1. 属于唯一用户<br>2. 地址信息完整有效 |
| 值对象 | Email | 邮箱 | 1. 格式符合RFC标准<br>2. 系统内唯一 |
| 值对象 | PasswordHash | 密码哈希 | 1. 不可逆加密<br>2. 支持密码验证 |
| 值对象 | AccessToken | 访问令牌 | 1. 有效期2小时<br>2. 签名防篡改 |
| 值对象 | RefreshToken | 刷新令牌 | 1. 有效期7天<br>2. 可刷新访问令牌 |
| 值对象 | AddressInfo | 地址详情 | 1. 省市区+详细地址完整<br>2. 邮编可选 |

### 1.3 聚合关联关系
```
User (聚合根)
├── ID (int64)
├── Email (值对象)
├── PasswordHash (值对象)
├── Username (string)
├── Avatar (string)
├── Status (int)
├── Addresses (Address实体列表，最多10个)
├── CreateTime (time)
├── UpdateTime (time)
├── LastLoginTime (time)
└── LastLoginIP (string)

Address (实体)
├── ID (int64)
├── UserID (int64)
├── Receiver (string)
├── Phone (string)
├── Address (AddressInfo值对象)
├── IsDefault (bool)
├── CreateTime (time)
└── UpdateTime (time)
```

### 1.4 领域事件
| 事件名称 | 触发时机 | 包含字段 | 订阅方示例 |
|----------|----------|----------|------------|
| UserRegisteredEvent | 用户注册成功 | 用户ID、邮箱、IP | 积分服务（赠送新用户积分）、营销服务（发送欢迎邮件） |
| UserLoggedInEvent | 用户登录成功 | 用户ID、邮箱、IP | 风控服务（异地登录检测）、统计服务（登录统计） |
| UserLoggedOutEvent | 用户登出 | 用户ID、IP | 会话服务（销毁会话） |
| UserInfoUpdatedEvent | 用户信息修改 | 用户ID、修改字段、新旧值 | 搜索服务（更新用户索引） |
| Address*Event | 地址增删改 | 用户ID、地址ID | 订单服务（地址更新同步） |

---

## 2. 分层架构设计
### 2.1 四层架构图
```
┌─────────────────────────────────────────────────────────┐
│ Interface Layer (接口层)                                │
│  - gRPC API 适配 go-zero Logic 层                       │
│  - 对外提供统一的服务接口                                │
├─────────────────────────────────────────────────────────┤
│ Application Layer (应用层)                              │
│  - 应用服务：AuthAppService、UserAppService            │
│  - DTO转换、用例编排、事务协调、领域事件发布            │
│  - 不包含业务规则，只协调领域对象完成业务流程            │
├─────────────────────────────────────────────────────────┤
│ Domain Layer (领域层) 【核心】                          │
│  - 聚合根：User                                         │
│  - 实体：Address                                        │
│  - 值对象：Email、PasswordHash、AddressInfo等           │
│  - 领域事件定义                                         │
│  - 仓储接口：UserRepository                             │
│  - 核心业务规则全部在此层实现，不依赖任何外部框架        │
├─────────────────────────────────────────────────────────┤
│ Infrastructure Layer (基础设施层)                       │
│  - 仓储实现：UserRepositoryImpl (MySQL操作)             │
│  - 消息发布：RabbitMQEventPublisher                      │
│  - 第三方服务集成：短信、邮件等                         │
│  - 实现领域层定义的接口，为上层提供技术实现              │
└─────────────────────────────────────────────────────────┘
```

### 2.2 目录结构
```
services/users/
├── domain/                     # 领域层（核心）
│   ├── aggregate/              # 聚合根定义
│   │   └── user.go             # User聚合根
│   ├── entity/                 # 实体定义
│   │   └── address.go          # Address实体
│   ├── valueobject/            # 值对象定义
│   │   ├── email.go
│   │   ├── password_hash.go
│   │   ├── address_info.go
│   │   └── token.go
│   ├── event/                  # 领域事件定义
│   │   └── events.go
│   ├── repository/             # 仓储接口定义
│   │   └── user_repository.go
│   └── service/                # 领域服务接口（可选）
├── application/                # 应用层
│   ├── dto/                    # 数据传输对象
│   │   ├── auth_dto.go
│   │   └── user_dto.go
│   ├── assembler/              # DTO与领域模型转换器
│   ├── service/                # 应用服务实现
│   │   ├── auth_service.go     # 认证应用服务
│   │   └── user_service.go     # 用户管理应用服务
│   └── event/                  # 事件发布器接口
│       └── event_publisher.go
├── infrastructure/             # 基础设施层
│   ├── persistence/            # 持久化实现
│   │   └── user_repo_impl.go   # User仓储MySQL实现
│   └── messaging/              # 消息队列实现
│       └── rabbitmq_event_publisher.go
├── interface/                  # 接口层
│   └── grpc/                   # gRPC接口适配
│       └── *.go                # 适配go-zero Logic层
└── 原有go-zero结构保留：
    ├── internal/
    │   ├── config/
    │   ├── logic/              # 原有Logic层，现在只做参数转发到应用服务
    │   ├── server/
    │   └── svc/
    ├── userspb/
    └── users.go
```

---

## 3. 核心业务重构说明
### 3.1 原有架构问题
- 业务逻辑混杂在Logic层，与go-zero框架强耦合
- 业务规则分散，难以单元测试
- 缺乏明确的边界，代码扩展性差
- 没有领域概念，需求变更时修改成本高

### 3.2 重构后优势
✅ **业务与技术解耦**：核心业务逻辑全部在领域层，不依赖go-zero、数据库等技术实现，可以独立演化
✅ **高可测试性**：领域层不需要依赖任何外部服务即可单元测试，测试成本降低80%
✅ **扩展性强**：新增需求（如第三方登录、积分体系）只需要在领域层扩展，不影响现有接口
✅ **事件驱动**：通过领域事件实现业务解耦，方便后续接入其他微服务
✅ **兼容现有框架**：完全保留go-zero作为通信层，对外接口不变，不需要修改上游服务调用方式
✅ **符合DDD规范**：明确的分层和职责划分，代码可维护性大幅提升

### 3.3 核心流程示例（登录）
**重构前**：
1. LoginLogic直接调用UsersModel查询数据库
2. 密码校验逻辑写在Logic层
3. 令牌生成逻辑写在Logic层
4. 直接返回响应

**重构后**：
1. LoginLogic接收参数 → 构造DTO → 调用AuthAppService.Login
2. AuthAppService：
   - 创建Email值对象（格式校验）
   - 调用UserRepository.FindByEmail查询用户
   - 调用User聚合根的VerifyPassword方法校验密码
   - 调用User.RecordLogin记录登录信息
   - 调用UserRepository.Update更新用户
   - 生成令牌
   - 发布UserLoggedInEvent领域事件
3. 返回响应给Logic层

---

## 4. 技术栈说明
- **框架**：go-zero（保留作为通信层，业务逻辑剥离至领域层）
- **数据库**：MySQL + go-zero sqlx（基础设施层实现）
- **消息队列**：RabbitMQ（领域事件发布）
- **缓存**：Redis（可在基础设施层扩展，不影响领域层）
- **事务**：应用层协调事务，保证最终一致性

---

## 5. 后续演进方向
1. 完善地址相关的领域逻辑和应用服务
2. 实现RabbitMQ事件发布器，替换当前的空实现
3. 添加单元测试，覆盖领域层核心业务规则
4. 引入领域服务处理跨聚合的业务逻辑
5. 引入Saga模式处理分布式事务场景
