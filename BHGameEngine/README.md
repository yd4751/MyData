# BHGameEngine

基于Go语言开发的高性能游戏服务器引擎。

## 项目结构

```
BHGameEngine/
├── bin/             # 编译后的可执行文件
├── cmd/             # 各服务入口
│   ├── battle/      # 战斗服务
│   ├── center/      # 中心服务
│   ├── cross/       # 跨服服务
│   ├── dataservice/ # 数据服务
│   ├── gate/        # 网关服务
│   ├── gm/          # GM管理服务
│   ├── gridmap/     # 网格地图服务
│   ├── logic/       # 逻辑服务
│   ├── login/       # 登录服务
│   └── webserver/   # Web服务
├── config/          # 配置文件
├── docker/          # Docker配置
├── internal/        # 内部模块
│   ├── ai/          # AI系统
│   ├── battlecore/  # 战斗核心
│   ├── cluster/     # 集群管理
│   ├── db/          # 数据库操作
│   ├── entity/      # 实体定义
│   ├── network/     # 网络模块
│   ├── player/      # 玩家管理
│   ├── redis/       # Redis操作
│   ├── timer/       # 定时器
│   └── worldmap/    # 世界地图
├── pkg/             # 公共包
│   ├── logger/      # 日志模块
│   ├── pool/        # 对象池
│   ├── proto/       # 协议定义
│   ├── resloader/   # 资源加载
│   ├── snowflake/   # 雪花算法
│   └── utils/       # 工具函数
├── scripts/         # 脚本文件
├── logs/            # 日志目录
├── pid/             # PID文件目录
└── web/             # Web静态资源
```

## 服务说明

| 服务 | 说明 |
|------|------|
| battle | 战斗服务，处理战斗逻辑 |
| center | 中心服务，协调各服务 |
| cross | 跨服服务，处理跨服务器交互 |
| dataservice | 数据服务，管理游戏数据 |
| gate | 网关服务，处理客户端连接 |
| gm | GM管理服务，提供管理接口 |
| gridmap | 网格地图服务 |
| logic | 逻辑服务，处理游戏逻辑 |
| login | 登录服务，处理用户登录 |
| webserver | Web服务，提供HTTP接口 |

## 环境要求

- Go 1.21+
- Redis 7.0+
- MySQL 8.0+

## 快速开始

### 编译

```bash
# 编译所有服务
go build -o bin/battle cmd/battle/main.go
go build -o bin/center cmd/center/main.go
go build -o bin/cross cmd/cross/main.go
go build -o bin/dataservice cmd/dataservice/main.go
go build -o bin/gate cmd/gate/main.go
go build -o bin/gm cmd/gm/main.go
go build -o bin/gridmap cmd/gridmap/main.go
go build -o bin/logic cmd/logic/main.go
go build -o bin/login cmd/login/main.go
go build -o bin/webserver cmd/webserver/main.go
```

### 运行

```bash
# 使用脚本启动所有服务
./scripts/start.sh

# 或单独启动某个服务
./bin/login
./bin/gate
./bin/logic
```

### 停止

```bash
./scripts/stop.sh
```

## Docker部署

```bash
# 使用docker-compose启动
cd docker
docker-compose up -d
```

## 配置

配置文件位于 `config/config.toml`，包含数据库连接、Redis配置、服务端口等参数。

## 日志

日志文件输出到 `logs/` 目录，每个服务有独立的日志文件。

## License

MIT License