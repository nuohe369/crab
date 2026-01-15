# Crab 框架

一个模块化的 Go Web 框架，清晰的分层架构，同一套代码可以单体运行或按模块拆分为多个服务。

[English](README.md) | 简体中文

## 架构图

![架构图](docs/architecture.jpg)

## 快速开始

```bash
go run . init              # 生成配置文件
go run . serve             # 启动所有模块
go run . serve -s api      # 按服务名启动
go run . serve -m testapi  # 按模块名启动
go run . list              # 列出模块和服务
```

## 项目结构

```
├── main.go
├── boot/                   # 启动层
│   ├── boot.go            # 生命周期管理
│   ├── cmd.go             # 命令行
│   └── context.go         # 模块上下文
├── common/                 # 业务公共层
│   ├── config/            # 配置管理
│   ├── middleware/        # HTTP 中间件
│   ├── response/          # 响应结构
│   ├── service/           # 公共服务
│   └── util/              # 公共工具
├── pkg/                    # 基础设施层（无业务依赖）
│   ├── cache/             # 缓存抽象（Redis/本地）
│   ├── config/            # 配置加载（TOML + 热更新 + 加密）
│   ├── cron/              # 定时任务调度
│   ├── jwt/               # JWT 认证
│   ├── logger/            # 结构化日志（按模块分文件）
│   ├── metrics/           # Prometheus 指标
│   ├── mq/                # 消息队列（Redis/RabbitMQ）
│   ├── pgsql/             # PostgreSQL（xorm）
│   ├── redis/             # Redis 客户端
│   ├── server/            # HTTP 服务（Fiber）
│   ├── snowflake/         # 雪花 ID 生成器
│   ├── storage/           # 存储抽象（本地/S3/OSS）
│   ├── trace/             # OpenTelemetry 链路追踪
│   ├── util/              # 工具函数
│   └── ws/                # WebSocket Hub
└── module/                 # 业务模块
    ├── testapi/           # API 示例模块
    │   ├── module.go      # 模块入口
    │   └── internal/      # 私有实现
    └── ws/                # WebSocket 示例
        ├── module.go      # 模块入口
        ├── example_01_basic/
        ├── example_02_multiuser/
        ├── example_03_callback/
        ├── example_04_cluster/
        └── example_05_service/
```

## 架构规则

| 规则 | 说明 |
|------|------|
| pkg 无业务依赖 | 可独立使用，各包自带 Config 结构 |
| common 依赖 pkg | 业务公共层 |
| module 依赖 common + pkg | 业务模块 |
| 有 internal = 私有模块 | 内部实现，不导出 |

## 示例模块

### testapi - API 示例

演示基础 HTTP 处理器、消息队列集成、响应格式化。

### ws - WebSocket 示例

演示 `pkg/ws` 的各种用法，仅作为示例参考。详见 `module/ws/README.md`

## 配置

```toml
# config.toml
[database]
host = "localhost"
port = 5432
user = "postgres"
password = "ENC(xxxxx...)"  # 加密值
dbname = "crab"

[redis]
addr = "localhost:6379"
password = ""
db = 0

[[services]]
name = "api"
addr = ":3000"
modules = ["testapi", "ws"]
```

### 加密敏感值

```bash
# 加密
go run . encrypt -k your-secret-key -v "password123"
# 输出: ENC(xxxxx...)

# 启动时传入解密密钥
go run . serve -k your-secret-key
```

## 模块开发

```go
// module/xxx/module.go
package xxx

import "server/boot"

func init() {
    boot.Register(&Module{})
}

type Module struct{}

func (m *Module) Name() string { return "xxx" }
func (m *Module) Models() []any { return nil }

func (m *Module) Init(ctx *boot.ModuleContext) error {
    // 设置路由
    ctx.Router.Get("/hello", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{"msg": "hello"})
    })
    return nil
}

func (m *Module) Start() error { return nil }
func (m *Module) Stop() error { return nil }
```

**重要：在 main.go 中添加导入**

```go
// main.go
import (
    _ "server/module/xxx"  // 必须添加新模块
)
```

## 基础设施包

`pkg/` 中的所有包都是独立的，可在其他项目中使用：

- **cache** - 统一缓存接口（Redis/本地）
- **config** - TOML 配置 + 热更新 + 加密
- **cron** - 定时任务调度器
- **logger** - 结构化日志 + 按模块分文件
- **metrics** - Prometheus 指标中间件
- **mq** - 消息队列抽象（Redis/RabbitMQ）
- **pgsql** - PostgreSQL + xorm
- **redis** - Redis 客户端 + 连接池
- **storage** - 存储抽象（本地/S3/OSS）
- **ws** - WebSocket Hub + 发布订阅

## 多服务部署

```toml
# config.toml
[[services]]
name = "all"
addr = ":3000"
modules = ["testapi", "ws"]

[[services]]
name = "api"
addr = ":3001"
modules = ["testapi"]

[[services]]
name = "ws"
addr = ":3002"
modules = ["ws"]
```

```bash
go run . serve -s api  # 只启动 API 服务
go run . serve -s ws   # 只启动 WebSocket 服务
go run . serve -s all  # 启动所有模块
```

## 响应格式

```json
{"code": 0, "msg": "success", "data": {...}}
{"code": 4001, "msg": "错误信息"}
```

## 日志

```go
import "server/pkg/logger"

var log = logger.NewWithName[struct{}]("模块名")

log.Info("消息 %s", arg)
log.Error("错误 %v", err)
log.InfoCtx(ctx, "带 traceId 的消息")
```

日志写入 `logs/{模块名}/日期.log`

## 许可证

MIT

## 鸣谢

- [Kiro](https://kiro.dev) + [Claude](https://claude.ai) - 代码辅助与架构设计
