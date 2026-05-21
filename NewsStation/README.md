# GeoNews - 基于地图的新闻可视化工具

## 项目概述

GeoNews 是一个将新闻数据与地理位置绑定的 Web 应用，通过交互式地图展示全球各地的新闻动态。

## 技术栈

- **后端**: Go (Gin 框架)
- **前端**: Vue 3 / React + TypeScript
- **地图库**: Mapbox GL JS / Leaflet / Cesium
- **数据库**: MySQL 8.0
- **部署**: Docker + Docker Compose

## 快速开始

### 环境要求

- Docker >= 20.10
- Docker Compose >= 2.0

### 运行项目

```bash
# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f api
```

### 服务地址

- API: http://localhost:8080
- MySQL: localhost:3306

## API 接口

### POST /api/v1/news

接收新闻数据

**请求体:**
```json
{
  "title": "地震预警",
  "summary": "XX地区发生5.0级地震",
  "source": "新华社",
  "publish_time": "2026-05-21T10:00:00Z",
  "geo_level": "country",
  "latitude": 39.9042,
  "longitude": 116.4074,
  "is_breaking": true,
  "priority": 1
}
```

### GET /api/v1/news/map

获取地图点位

**参数:**
- `level`: 地理级别 (world, continent, country, city)
- `code`: 地区代码 (如 CN)

### GET /api/v1/news/breaking

获取突发新闻

### GET /api/v1/news/history

查询历史新闻

**参数:**
- `start`: 开始日期 (2026-01-01)
- `end`: 结束日期 (2026-01-31)
- `page`: 页码 (默认 1)
- `limit`: 每页数量 (默认 20)

## 项目结构

```
.
├── cmd/
│   └── server/
│       └── main.go          # 应用入口
├── internal/
│   ├── config/              # 配置管理
│   ├── controller/          # 控制器层
│   ├── service/             # 业务逻辑层
│   ├── repository/          # 数据访问层
│   ├── model/               # 数据模型
│   ├── router/              # 路由配置
│   └── database/            # 数据库连接
├── sql/
│   └── schema.sql           # 数据库建表脚本
├── config/
│   └── config.yaml          # 配置文件
├── Dockerfile
├── docker-compose.yml
└── go.mod
```

## 数据库设计

### news 表

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | BIGINT | 主键 |
| title | VARCHAR(255) | 新闻标题 |
| summary | TEXT | 摘要 |
| source | VARCHAR(50) | 来源 |
| publish_time | DATETIME | 发布时间 |
| lat | DECIMAL(10,8) | 纬度 |
| lng | DECIMAL(11,8) | 经度 |
| geo_level | ENUM | 地理级别 |
| is_breaking | TINYINT(1) | 突发标志 |
| priority | INT | 优先级 |
| created_at | TIMESTAMP | 入库时间 |

## 开发

```bash
# 安装依赖
go mod download

# 运行开发服务器
go run cmd/server/main.go
```