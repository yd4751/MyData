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

## 更新日志

### v1.1.0 - 无缝大地图系统

**新增功能:**

1. **GridMap分区管理**
   - 实现地图网格化分区，支持多GridMap节点分布式部署
   - 地图被划分为9个区域（3x3网格），每个区域由独立的GridMap服务器管理
   - 根据玩家坐标自动计算归属的GridMap节点

2. **动态消息路由**
   - Gate服务器根据玩家位置动态路由消息到对应GridMap
   - 维护玩家当前所在GridMap的映射关系
   - 支持玩家在不同GridMap之间无缝移动

3. **跨区传输机制**
   - 玩家跨越GridMap边界时自动传输到目标GridMap
   - 保持玩家状态和位置的连续性
   - 实现无缝的跨区体验

4. **地图区块管理**
   - 动态加载/卸载地图区块
   - 支持区块数据持久化存储
   - 实现视距内区块预加载

5. **玩家同步**
   - 实时同步附近玩家位置信息
   - 支持AOI（Area of Interest）区域管理

**启动方式:**

```bash
# 启动多个GridMap节点（每个节点负责不同区域）
./bin/gridmap -grid 1
./bin/gridmap -grid 2
./bin/gridmap -grid 3
./bin/gridmap -grid 4
./bin/gridmap -grid 5
./bin/gridmap -grid 6
./bin/gridmap -grid 7
./bin/gridmap -grid 8
./bin/gridmap -grid 9

# 启动Gate服务器（自动路由消息到对应GridMap）
./bin/gate
```

**架构设计:**

```
┌─────────────────────────────────────────────────────────────────┐
│                        客户端 (Client)                          │
└──────────────────────────┬──────────────────────────────────────┘
                           │ 网络通信
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Gate Server (网关层)                        │
│  - 客户端连接管理                                               │
│  - 消息路由转发                                               │
│  - 根据玩家位置路由到对应 gridmap                               │
└──────────────────────────┬──────────────────────────────────────┘
                           │ 内部通信
          ┌─────────────────┼─────────────────┐
          ▼                 ▼                 ▼
┌──────────────────┐ ┌──────────────────┐ ┌──────────────────┐
│   GridMap-001    │ │   GridMap-002    │ │   GridMap-003    │
│  (区块 A1-A4)    │ │  (区块 B1-B4)    │ │  (区块 C1-C4)    │
│  - 区块加载/卸载  │ │  - 区块加载/卸载  │ │  - 区块加载/卸载  │
│  - 实体管理      │ │  - 实体管理      │ │  - 实体管理      │
│  - 玩家位置同步  │ │  - 玩家位置同步  │ │  - 玩家位置同步  │
│  - AI 行为处理   │ │  - AI 行为处理   │ │  - AI 行为处理   │
└──────────────────┘ └──────────────────┘ └──────────────────┘
          │                 │                 │
          └─────────────────┼─────────────────┘
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Logic Server (逻辑层)                         │
│  - 玩家数据管理                                                 │
│  - 背包/物品系统                                                │
│  - 任务系统                                                    │
│  - 社交系统                                                    │
└──────────────────────────┬──────────────────────────────────────┘
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                   DataService (数据层)                          │
│  - MySQL: 玩家持久化数据                                        │
│  - Redis: 缓存/会话数据                                         │
│  - 地图数据加载                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**新增文件:**

| 文件 | 说明 |
|------|------|
| `internal/worldmap/gridmap_router.go` | GridMap分区路由器 |
| `internal/worldmap/map_loader.go` | 地图数据加载器 |
| `cmd/gridmap/handler.go` | GridMap处理逻辑 |
| `cmd/gate/handler.go` | Gate路由处理 |

## License

MIT License