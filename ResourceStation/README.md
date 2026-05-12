# Resource Station - 资源管理存储站

一个基于Go语言开发的资源管理存储站，支持文本、图片、视频、音频等多种资源类型，提供Web端查询和管理，具备用户和权限管理功能，支持断点续传、批量上传，单一文件可达TB级别。

## 功能特性

- ✅ 多类型文件支持：文本、图片、视频、音频等
- ✅ 用户认证和权限管理（JWT）
- ✅ 统一格式上传接口
- ✅ 断点续传（分片上传）
- ✅ 批量上传
- ✅ 资源唯一标记（UUID + SHA256哈希）
- ✅ TB级别大文件支持
- ✅ Web端资源管理界面
- ✅ 响应式设计，支持移动端
- ✅ MySQL数据库存储

## 技术栈

### 后端
- Go 1.21+
- Gin Web框架
- GORM (MySQL驱动)
- JWT认证
- 本地文件存储

### 前端
- HTML5 + CSS3
- JavaScript (ES6+)
- Bootstrap 5
- Font Awesome图标

### 数据库
- MySQL 5.7+

## 快速开始

### 1. 环境要求

- Go 1.21+
- MySQL 5.7+
- Node.js (可选，用于前端开发)

### 2. 克隆项目

```bash
git clone <repository-url>
cd ResourceStation
```

### 3. 配置数据库

创建MySQL数据库：

```sql
CREATE DATABASE resource_station CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 4. 配置后端

编辑 `backend/configs/config.yaml` 文件，修改数据库连接信息：

```yaml
database:
  host: "localhost"
  port: 3306
  username: "root"
  password: "your_password"
  database: "resource_station"
  charset: "utf8mb4"
```

### 5. 安装后端依赖

```bash
cd backend
go mod download
```

### 6. 运行后端服务

```bash
cd backend
go run cmd/server/main.go
```

后端服务将在 `http://localhost:8080` 启动。

### 7. 访问前端

打开浏览器访问 `http://localhost:8080` 即可使用资源管理站。

## 项目结构

```
ResourceStation/
├── backend/                    # 后端代码
│   ├── cmd/
│   │   └── server/
│   │       └── main.go        # 主程序入口
│   ├── configs/
│   │   └── config.yaml        # 配置文件
│   ├── internal/
│   │   ├── api/              # API处理器
│   │   ├── auth/             # 认证模块
│   │   ├── database/         # 数据库连接
│   │   ├── models/           # 数据模型
│   │   ├── storage/          # 文件存储
│   │   └── utils/            # 工具函数
│   └── go.mod                # Go模块文件
├── frontend/                  # 前端代码
│   ├── index.html            # 主页面
│   └── js/
│       └── app.js            # 前端JavaScript
├── docker/                   # Docker配置
└── README.md                # 项目说明
```

## API接口

### 用户认证
- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录

### 资源管理
- `POST /api/v1/resources` - 创建资源（开始上传）
- `GET /api/v1/resources` - 获取资源列表
- `GET /api/v1/resources/:id` - 获取资源详情
- `PUT /api/v1/resources/:id` - 更新资源信息
- `DELETE /api/v1/resources/:id` - 删除资源
- `GET /api/v1/resources/:id/download` - 下载资源

### 分片上传
- `POST /api/v1/resources/:id/chunks/:chunkIndex` - 上传文件分片
- `POST /api/v1/resources/:id/complete` - 完成上传（合并分片）
- `GET /api/v1/resources/:id/progress` - 获取上传进度

### 批量上传
- `POST /api/v1/resources/batch` - 批量创建资源

## 配置说明

### 存储配置
```yaml
storage:
  base_path: "./storage"          # 文件存储路径
  temp_path: "./temp"            # 临时文件路径
  max_file_size: 1099511627776   # 最大文件大小（1TB）
  chunk_size: 10485760           # 分片大小（10MB）
  allowed_types:                 # 允许的文件类型
    - "text/*"
    - "image/*"
    - "video/*"
    - "audio/*"
    - "application/pdf"
    - "application/zip"
```

### 服务器配置
```yaml
server:
  host: "0.0.0.0"    # 监听地址
  port: 8080          # 监听端口
  mode: "debug"       # 运行模式：debug/release/test
```

### JWT配置
```yaml
jwt:
  secret: "your-secret-key-change-in-production"  # JWT密钥
  expire_hours: 24                                # Token过期时间（小时）
```

## 文件上传流程

1. **创建资源**：调用 `POST /api/v1/resources` 创建资源记录
2. **分片上传**：对于大文件，启用分片上传，调用 `POST /api/v1/resources/:id/chunks/:chunkIndex` 上传每个分片
3. **完成上传**：所有分片上传完成后，调用 `POST /api/v1/resources/:id/complete` 合并分片
4. **查询进度**：上传过程中可调用 `GET /api/v1/resources/:id/progress` 查询上传进度

## 批量上传

支持一次性创建多个资源记录，适用于需要上传多个文件的场景：

```json
POST /api/v1/resources/batch
[
  {
    "filename": "file1.jpg",
    "file_type": "image/jpeg",
    "file_size": 1024000,
    "chunk_count": 1,
    "chunk_size": 0,
    "description": "示例图片",
    "tags": "图片,示例",
    "is_public": false
  }
]
```

## 开发指南

### 添加新的文件类型支持

1. 在 `config.yaml` 的 `storage.allowed_types` 中添加MIME类型
2. 在 `storage.go` 的 `GetMimeTypeFromExtension` 函数中添加扩展名映射

### 修改分片大小

在 `config.yaml` 中修改 `storage.chunk_size` 值（单位：字节）

### 修改最大文件大小

在 `config.yaml` 中修改 `storage.max_file_size` 值（单位：字节）

## 部署说明

### 生产环境建议

1. 修改 `config.yaml` 中的JWT密钥
2. 设置 `server.mode` 为 `release`
3. 使用Nginx等反向代理
4. 配置HTTPS证书
5. 定期备份数据库和存储文件

### Docker部署

项目包含Docker配置，可以使用Docker Compose一键部署：

```bash
cd docker
docker-compose up -d
```

## 许可证

MIT License

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建Pull Request

## 问题反馈

如有问题或建议，请提交Issue或Pull Request。