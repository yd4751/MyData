# 文件传输服务器

一个使用Go语言编写的文件传输服务器，支持HTTP/HTTPS协议，带有Web管理界面，可以直接通过Web界面上传单个文件或整个目录。

## 功能特性

- ✅ **支持HTTP和HTTPS** - 同时支持HTTP和HTTPS协议
- ✅ **Web管理界面** - 现代化的Web界面，支持中文
- ✅ **文件上传** - 支持单个文件和多文件上传
- ✅ **目录上传** - 支持整个目录上传（保持目录结构）
- ✅ **文件管理** - 文件列表查看、下载、删除
- ✅ **目录管理** - 创建新目录、目录导航
- ✅ **拖放上传** - 支持拖放文件/目录到上传区域
- ✅ **响应式设计** - 适配各种屏幕尺寸
- ✅ **实时更新** - 文件列表自动刷新
- ✅ **进度提示** - 上传和操作状态提示

## 快速开始

### 1. 安装Go
确保已安装Go 1.16或更高版本。

### 2. 下载和运行
```bash
# 克隆或下载项目
git clone <项目地址>
cd file-transfer-server

# 运行服务器
go run main.go
```

### 3. 访问Web界面
打开浏览器访问：http://localhost:8080

## 配置说明

服务器使用 `config.json` 文件进行配置，默认配置如下：

```json
{
  "http_port": "8080",
  "https_port": "8443",
  "tls_cert": "cert.pem",
  "tls_key": "key.pem",
  "upload_dir": "uploads",
  "static_dir": "static",
  "max_upload_size_mb": 100,
  "enable_cors": true,
  "allow_file_types": ["*"],
  "require_auth": false,
  "auth_username": "admin",
  "auth_password": "password"
}
```

### 配置项说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `http_port` | HTTP服务端口 | 8080 |
| `https_port` | HTTPS服务端口 | 8443 |
| `tls_cert` | TLS证书文件路径 | cert.pem |
| `tls_key` | TLS密钥文件路径 | key.pem |
| `upload_dir` | 上传文件存储目录 | uploads |
| `static_dir` | 静态文件目录 | static |
| `max_upload_size_mb` | 最大上传文件大小(MB) | 100 |
| `enable_cors` | 是否启用CORS | true |
| `allow_file_types` | 允许的文件类型，["*"]表示所有 | ["*"] |
| `require_auth` | 是否启用基本认证 | false |
| `auth_username` | 认证用户名 | admin |
| `auth_password` | 认证密码 | password |

## 启用HTTPS

要启用HTTPS，需要提供TLS证书和密钥文件：

### 1. 生成自签名证书（仅用于测试）
```bash
# 生成私钥
openssl genrsa -out key.pem 2048

# 生成证书签名请求
openssl req -new -key key.pem -out csr.pem

# 生成自签名证书
openssl x509 -req -days 365 -in csr.pem -signkey key.pem -out cert.pem

# 删除CSR文件
rm csr.pem
```

### 2. 使用Let's Encrypt证书（生产环境）
```bash
# 使用certbot获取证书
certbot certonly --standalone -d yourdomain.com
```

### 3. 更新配置文件
将生成的证书和密钥文件路径配置到 `config.json` 中。

## 使用指南

### 1. 上传文件
1. 点击"选择文件"按钮或拖放文件到上传区域
2. 选择要上传的文件（支持多选）
3. 点击"上传文件"按钮
4. 等待上传完成，文件将出现在文件列表中

### 2. 上传目录
1. 点击"选择目录"按钮或拖放目录到上传区域
2. 选择要上传的目录
3. 点击"上传目录"按钮
4. 目录及其所有文件将保持原有结构上传

### 3. 创建目录
1. 在"创建目录"区域输入目录名称
2. 点击"创建目录"按钮
3. 新目录将出现在文件列表中

### 4. 文件操作
- **打开目录**：点击目录的"打开"按钮进入目录
- **下载文件**：点击文件的"下载"按钮下载文件
- **删除文件/目录**：点击"删除"按钮删除项目
- **导航**：使用路径导航栏在不同目录间跳转

### 5. 拖放功能
- 可以直接拖放文件或目录到上传区域
- 系统会自动识别是文件还是目录
- 自动开始上传过程

## API接口

### 文件列表
```
GET /api/files?path=<目录路径>
```
返回指定目录下的文件列表

### 文件上传
```
POST /upload
```
上传单个或多个文件

### 目录上传
```
POST /upload-dir
```
上传整个目录（保持目录结构）

### 文件下载
```
GET /api/download/<文件路径>
```
下载指定文件

### 删除文件/目录
```
DELETE /api/delete/<路径>
```
删除文件或目录

### 创建目录
```
POST /api/create-dir
```
创建新目录

## 项目结构

```
file-transfer-server/
├── main.go              # 主程序文件
├── go.mod               # Go模块文件
├── config.json          # 配置文件
├── README.md            # 说明文档
├── static/              # 静态文件目录
│   ├── css/
│   │   └── style.css   # 样式文件
│   ├── js/
│   │   └── main.js     # JavaScript文件
│   └── index.html      # 主页面模板
├── uploads/             # 上传文件存储目录
└── cert.pem & key.pem   # TLS证书文件（可选）
```

## 构建和部署

### 构建可执行文件
```bash
go build -o file-transfer-server main.go
```

### 运行可执行文件
```bash
./file-transfer-server
```

### 作为系统服务运行（Linux）
```bash
# 创建服务文件
sudo nano /etc/systemd/system/file-transfer.service

# 服务文件内容
[Unit]
Description=File Transfer Server
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/path/to/file-transfer-server
ExecStart=/path/to/file-transfer-server/file-transfer-server
Restart=on-failure

[Install]
WantedBy=multi-user.target

# 启用并启动服务
sudo systemctl enable file-transfer.service
sudo systemctl start file-transfer.service
```

## 注意事项

1. **安全性**：生产环境请启用HTTPS和认证
2. **文件大小**：默认最大上传100MB，可在配置中调整
3. **存储空间**：确保服务器有足够的存储空间
4. **权限**：确保服务器有写入上传目录的权限
5. **防火墙**：确保防火墙允许配置的端口访问

## 故障排除

### 常见问题

1. **端口被占用**
   ```
   修改config.json中的端口号
   ```

2. **无法上传大文件**
   ```
   检查磁盘空间和文件权限
   调整max_upload_size_mb配置
   ```

3. **HTTPS无法启动**
   ```
   检查证书文件路径和权限
   确保证书和密钥文件存在
   ```

4. **无法访问Web界面**
   ```
   检查防火墙设置
   确认服务器正在运行
   ```

### 日志查看
服务器运行日志会输出到控制台，包含错误信息和访问日志。

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request来改进这个项目。

## 更新日志

### v1.0.0
- 初始版本发布
- 支持HTTP/HTTPS
- Web管理界面
- 文件/目录上传
- 文件管理功能