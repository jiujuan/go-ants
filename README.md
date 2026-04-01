
## go-ants

go-ants 是一个现代化的 Go Web 应用开发框架，提供一个单体应用解决脚手架。

## 特性

- **分层架构**: Transport -> Service -> Business -> Data 层
- **模块化设计**: 高度模块化，可按需引入
- **函数选项模式**: 灵活的组件配置方式
- **依赖注入**: 支持 Google Wire
- **多种 HTTP 框架**: 支持 Gin 和 Fiber
- **完整的技术栈**:
  - 数据库: GORM (MySQL/PostgreSQL)
  - 缓存: Redis + 内存缓存
  - 消息队列: Kafka / RabbitMQ
  - 搜索引擎: Elasticsearch
  - 认证: JWT
  - 日志: Zap
  - 验证: Validator
  - 监控: Prometheus + pprof
- **CLI 工具**: 自动生成项目骨架


## 安装

```bash
go install github.com/jiujuan/go-ants/cmd/ants@latest
```

## 快速开始

### 创建新项目

```bash
# 创建新项目
ants new myproject

# 进入项目目录
cd myproject

# 运行项目
ants run
```

### 基本使用

```go
package main

import (
	"context"
	"os"

	"github.com/jiujuan/go-ants/pkg/app"
	"github.com/jiujuan/go-ants/pkg/conf"
	"github.com/jiujuan/go-ants/pkg/log"
	"github.com/jiujuan/go-ants/pkg/transport"
	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化日志
	log.InitZap()

	// 加载配置
	c := conf.New()
	c.Load("configs/config.yaml")

	// 创建 HTTP 服务器
	engine := transport.NewGinServer("server",
		transport.WithAddr(":8080"),
	)

	// 注册路由
	engine.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to go-ants!",
		})
	})

	// 创建应用
	application, cleanup := app.New(
		app.WithName("myapp"),
		app.WithLogger(log.DefaultLogger()),
		app.WithComponents(engine),
	)
	defer cleanup()

	// 运行应用
	if err := application.Run(); err != nil {
		log.Error("application failed", log.Error(err))
		os.Exit(1)
	}
}
```

## 项目结构

```
myproject/
├── cmd/                    # 应用入口
│   └── myproject/
│       └── main.go
├── configs/               # 配置文件
│   └── config.yaml
├── internal/              # 内部代码
│   ├── domain/              # 业务逻辑层
│   ├── data/             # 数据访问层
│   ├── server/           # HTTP/gRPC 服务器
│   └── service/          # 应用服务层
├── pkg/                   # 公共库（框架核心）
│   ├── app/
│   ├── auth/
│   ├── cache/
│   ├── conf/
│   ├── database/
│   ├── es/
│   ├── log/
│   ├── metric/
│   ├── mq/
│   ├── redis/
│   ├── transport/
│   ├── validator/
│   ├── worker/
│   └── wire/
├── go.mod
└── README.md
```

## 框架组件

### pkg/app

应用生命周期管理

```go
app, cleanup := app.New(
    app.WithName("myapp"),
    app.WithLogger(logger),
    app.WithComponents(httpServer, workerPool),
)
defer cleanup()
app.Run()
```

### pkg/database

数据库操作（支持 MySQL 和 PostgreSQL）

```go
db, err := database.New(ctx,
    database.WithDSN(dsn),
    database.WithDriver("mysql"),
    database.WithMaxOpenConns(100),
)
```

### pkg/redis

Redis 客户端

```go
client, err := redis.New(ctx,
    redis.WithAddr("localhost:6379"),
    redis.WithPassword(""),
    redis.WithDB(0),
)
```

### pkg/auth

JWT 认证

```go
auth := auth.New(
    auth.WithHMACSigningKey("secret"),
    auth.WithIssuer("myapp"),
    auth.WithExpiration(time.Hour*24),
)

token, err := auth.GenerateToken(ctx, user, time.Hour*24)
```

### pkg/transport

HTTP 服务器（支持 Gin 和 Fiber）

```go
// Gin
server := transport.NewGinServer("server",
    transport.WithAddr(":8080"),
)

// Fiber
server := transport.NewFiberServer("server",
    transport.WithAddr(":8080"),
)
```

### pkg/worker

工作池

```go
pool := worker.New(10,
    worker.WithQueueSize(1000),
)
pool.Start()
defer pool.Stop()

pool.Submit(task)
```

### pkg/mq

消息队列（Kafka 和 RabbitMQ）

```go
// Kafka Producer
producer, _ := mq.NewKafkaProducer(
    mq.WithKafkaBrokers("localhost:9092"),
    mq.WithKafkaTopic("my-topic"),
)
producer.Produce(ctx, key, value)

// RabbitMQ Producer
producer, _ := mq.NewRabbitMQProducer(
    mq.WithRabbitMQURL("amqp://guest:guest@localhost:5672/"),
)
producer.Publish(ctx, "routing-key", data)
```

### pkg/es

Elasticsearch 客户端

```go
client, err := es.New(
    es.WithESAddresses("http://localhost:9200"),
    es.WithESIndexPrefix("myapp"),
)
```

### pkg/wire

依赖注入

```go
// wire.go
func InitializeApp() (*app.App, error) {
    wire.Build(
        NewDB,
        NewRedis,
        NewAuth,
        NewServer,
        app.New,
    )
    return nil, nil
}
```

## CLI 命令

```bash
# 创建新项目
ants new myproject

# 运行项目
ants run


# todo
# 生成 CRUD 代码
# ants gen crud user

# 生成 Proto 文件
# ants gen proto
```

## 配置示例

```yaml
server:
  addr: :8080
  mode: debug

database:
  driver: mysql
  dsn: "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
  max_idle_conns: 10
  max_open_conns: 100

redis:
  addr: "localhost:6379"
  password: ""
  db: 0

log:
  level: debug
  format: json

auth:
  jwt_secret: "your-secret-key"
  jwt_expiration: 3600
```

## 贡献

欢迎提交 Issue 和 Pull Request！

