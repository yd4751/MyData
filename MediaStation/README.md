# MediaStation - 音视频媒体平台

一个功能完整的音视频媒体项目，支持视频在线播放、音频在线播放、图片/小说在线浏览。

## 功能特性

### 视频播放
- 在线视频流播放（支持HLS/DASH自适应码率）
- 支持播放/暂停/倍速（0.5x-2x）
- 进度条拖拽
- 剧集列表切换
- 截图功能
- 横屏/竖屏模式适配（针对短剧优化）

### 音频播放
- 在线音频流播放
- 播放控制功能

### 图片浏览
- 图片画廊展示
- 上下张切换浏览

### 小说阅读
- 在线小说阅读
- 章节切换

### 用户系统
- 用户注册/登录
- 播放历史记录
- 播放进度保存

### 其他特性
- 弱网环境下流畅访问（HTTP Range请求支持）
- 响应式设计（适配手机/平板/PC）
- 搜索功能

## 技术栈

### 后端
- **语言**: Go 1.22+
- **框架**: 原生net/http
- **数据库**: MySQL
- **ORM**: GORM

### 前端
- **HTML5**
- **CSS3**
- **JavaScript (ES6+)**
- **HTML5 Video/Audio API**

## 项目结构

```
MediaStation/
├── backend/                  # 后端代码
│   ├── cmd/                  # 命令入口
│   │   └── main.go           # 主入口
│   ├── config/               # 配置文件
│   │   └── config.go         # 配置加载
│   ├── internal/             # 内部模块
│   │   ├── handler/          # HTTP处理器
│   │   ├── service/          # 业务逻辑层
│   │   ├── repository/       # 数据访问层
│   │   └── model/            # 数据模型
│   ├── pkg/                  # 公共包
│   │   └── database/         # 数据库连接
│   └── go.mod                # Go模块配置
├── frontend/                 # 前端代码
│   ├── index.html            # 主页面
│   ├── css/                  # 样式文件
│   │   └── style.css         # 全局样式
│   └── js/                   # JavaScript文件
│       └── app.js            # 主应用逻辑
├── database/                 # 数据库相关
│   └── init.sql              # 初始化SQL脚本
├── media/                    # 媒体文件存储目录
└── README.md                 # 项目说明
```

## 快速开始

### 1. 环境要求

- Go 1.22+
- MySQL 5.7+
- Node.js (可选，用于开发)

### 2. 安装依赖

```bash
cd backend
go mod download
```

### 3. 数据库配置

创建数据库并导入初始化脚本：

```bash
mysql -u root -p < database/init.sql
```

### 4. 配置环境变量

创建 `.env` 文件或设置环境变量：

```bash
export PORT=8080
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=root
export DB_PASSWORD=password
export DB_NAME=mediastation
export MEDIA_DIR=./media
export STATIC_DIR=./frontend
```

### 5. 运行项目

```bash
cd backend/cmd
go run main.go
```

服务启动后访问: http://localhost:8080

## API 接口

### 媒体相关

| 接口 | 方法 | 描述 |
|------|------|------|
| `/api/media` | GET | 获取单个媒体信息 |
| `/api/media/list` | GET | 获取媒体列表 |
| `/api/media/search` | GET | 搜索媒体 |
| `/api/media/series` | GET | 获取剧集列表 |
| `/stream` | GET | 媒体流播放 |

### 用户相关

| 接口 | 方法 | 描述 |
|------|------|------|
| `/api/user/register` | POST | 用户注册 |
| `/api/user/login` | POST | 用户登录 |
| `/api/user` | GET | 获取用户信息 |

### 播放历史

| 接口 | 方法 | 描述 |
|------|------|------|
| `/api/history` | GET | 获取播放历史 |
| `/api/history/progress` | GET | 获取播放进度 |
| `/api/history/save` | POST | 保存播放进度 |
| `/api/history/remove` | DELETE | 删除历史记录 |
| `/api/history/clear` | DELETE | 清空历史记录 |

## 视频流技术说明

本项目支持**无需预先分片**的视频流播放：

1. **HTTP Range 请求**: 支持断点续传和分片加载
2. **动态流适配**: 根据网络状况自动调整请求大小
3. **弱网优化**: 支持慢速网络下的流畅播放

## 移动端适配

- 响应式布局，适配各种屏幕尺寸
- 触摸友好的控制按钮
- 竖屏模式优化（针对短剧）

## 许可证

MIT License