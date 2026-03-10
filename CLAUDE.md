The role of this file is to describe common mistakes and confusion points that agents might encounter as they work in this project. If you ever encounter something in the project that surprises you, please alert the developer working with you and indicate that this is the case in the AgentMD file to help prevent future agents from having the same issue

## go-zero 开发规范
1. 必须使用 go-zero 的模板生成代码，禁止手写 handler 层代码
2. 编写 go-zero 相关代码时，必须调用 mcp-zero 工具进行代码生成和相关操作

## 基础服务地址（OrbStack 部署，禁止修改）
| 服务名称 | 地址 | 端口 | 说明 |
|---------|------|------|------|
| MySQL | 127.0.0.1 | 3306 | 关系型数据库 |
| Redis | 127.0.0.1 | 6379 | 缓存服务 |
| Consul | 127.0.0.1 | 8500 | 服务注册与发现 |
| Elasticsearch | 127.0.0.1 | 9200 | 搜索引擎 |
| RabbitMQ | 127.0.0.1 | 5672 | 消息队列（AMQP） |
| RabbitMQ管理界面 | 127.0.0.1 | 15672 | 消息队列管理后台 |
| DTM | 127.0.0.1 | 36789 | 分布式事务服务（HTTP） |
| DTM | 127.0.0.1 | 36790 | 分布式事务服务（gRPC） |
| MinIO | 127.0.0.1 | 9000 | 对象存储服务（API） |
| MinIO管理界面 | 127.0.0.1 | 9001 | 对象存储管理后台 |

## 分布式事务规范
项目采用 DTM Saga 模式实现分布式事务，禁止使用其他分布式事务方案

## 前端开发规范
修改前端服务代码后，必须调用 Chrome DevTools MCP 工具进行功能测试、兼容性测试和性能测试，确保修改不会破坏现有功能的正常运行

## 接口测试规范
当访问前后端交互边界时，必须使用 Apifox-mcp 进行接口测试（包括添加、删除接口等操作），禁止从前端直接越过网关（gateway）访问接口，也禁止从后端直接越过网关访问接口

### Apifox 配置信息
- API Key: `afxp_1bd4b5Je1OTIXl6b6AlwXgwQ7qxzzttqqTlk`
- 项目 ID: `7907732`