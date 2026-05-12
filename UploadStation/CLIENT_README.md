# Go命令行客户端使用说明

## 概述

这是一个用于文件传输服务器的Go命令行客户端，支持文件上传、下载、目录操作等功能。

## 编译

```bash
go build -o client client.go
```

或者直接运行：

```bash
go run client.go [参数]
```

## 使用方法

### 基本语法

```bash
./client -server <服务器地址> -action <操作类型> -path <路径> [其他参数]
```

### 参数说明

| 参数 | 说明 | 默认值 | 必需 |
|------|------|--------|------|
| `-server` | 服务器地址 | `http://localhost:8080` | 否 |
| `-action` | 操作类型：`list`, `upload`, `upload-dir`, `download`, `delete`, `mkdir` | `list` | 否 |
| `-path` | 文件或目录路径 | 空 | 是（除list外） |
| `-target` | 目标路径（用于上传、下载等） | 空 | 否 |
| `-recursive` | 是否递归操作（用于删除目录） | `false` | 否 |

### 操作示例

#### 1. 列出文件
```bash
# 列出根目录
./client -action list

# 列出指定目录
./client -action list -path "uploads/subdir"
```

#### 2. 上传文件
```bash
# 上传单个文件到根目录
./client -action upload -path "/path/to/local/file.txt"

# 上传单个文件到指定目录
./client -action upload -path "/path/to/local/file.txt" -target "uploads/subdir"

# 上传多个文件
./client -action upload -path "/path/to/file1.txt,/path/to/file2.txt"

# 上传多个文件到指定目录
./client -action upload -path "/path/to/file1.txt,/path/to/file2.txt" -target "uploads/subdir"
```

#### 3. 上传目录（保留目录结构）
```bash
# 上传整个目录
./client -action upload-dir -path "/path/to/local/directory"
```

#### 4. 下载文件
```bash
# 下载文件到当前目录
./client -action download -path "uploads/file.txt"

# 下载文件到指定位置
./client -action download -path "uploads/file.txt" -target "/path/to/save/file.txt"
```

#### 5. 删除文件或目录
```bash
# 删除文件
./client -action delete -path "uploads/file.txt"

# 删除目录（需要确认）
./client -action delete -path "uploads/subdir"
```

#### 6. 创建目录
```bash
# 在根目录创建目录
./client -action mkdir -path "newdir"

# 在指定目录下创建目录
./client -action mkdir -path "newdir" -target "uploads"
```

## 功能特点

1. **保留目录结构**：上传目录时会保留原始目录结构
2. **交互式确认**：删除操作需要用户确认
3. **文件覆盖保护**：下载文件时如果目标文件已存在会询问是否覆盖
4. **进度显示**：上传目录时会显示文件添加进度
5. **错误处理**：详细的错误信息和友好的提示

## 服务器API兼容性

客户端与文件传输服务器的以下API端点兼容：

- `GET /api/files` - 获取文件列表
- `POST /upload` - 上传单个文件
- `POST /upload-dir` - 上传目录（保留结构）
- `GET /api/download/{path}` - 下载文件
- `DELETE /api/delete/{path}` - 删除文件或目录
- `POST /api/create-dir` - 创建目录

## 注意事项

1. 上传大文件时可能需要调整服务器和客户端的超时设置
2. 目录上传会递归遍历所有子目录
3. 删除操作不可逆，请谨慎使用
4. 确保服务器正在运行且可访问

## 故障排除

### 连接失败
- 检查服务器地址是否正确
- 确保服务器正在运行
- 检查防火墙设置

### 上传失败
- 检查文件路径是否正确
- 确保有读取权限
- 检查服务器存储空间

### 下载失败
- 检查文件路径是否存在
- 确保有写入权限
- 检查磁盘空间

## 源码结构

```
client.go
├── main() - 主函数，解析命令行参数
├── listFiles() - 列出文件
├── uploadFile() - 上传单个文件
├── uploadDirectory() - 上传目录（保留结构）
├── downloadFile() - 下载文件
├── deleteFile() - 删除文件或目录
├── createDirectory() - 创建目录
└── formatFileSize() - 格式化文件大小显示
```

./client.exe -action upload-dir -path "." -server http://localhost:5555