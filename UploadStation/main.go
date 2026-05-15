package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// 配置结构体
type Config struct {
	HTTPPort  string `json:"http_port"`
	UploadDir string `json:"upload_dir"`
	StaticDir string `json:"static_dir"`
}

// 文件信息结构体
type FileInfo struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
	IsDir   bool      `json:"is_dir"`
	Path    string    `json:"path"`
	URL     string    `json:"url"`
}

// 全局变量
var (
	config    Config
	templates *template.Template
	uploadDir string
	staticDir string
)

// sendJSONError 发送JSON格式的错误响应
func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}

func main() {
	// 加载配置
	loadConfig()

	// 创建必要的目录
	createDirectories()

	// 加载模板
	loadTemplates()

	// 设置路由
	mux := http.NewServeMux()

	// 静态文件服务
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// API路由
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/upload", handleUpload)
	mux.HandleFunc("/upload-dir", handleUploadDir)
	mux.HandleFunc("/api/files", handleFileList)
	mux.HandleFunc("/api/download/", handleDownload)
	mux.HandleFunc("/api/delete/", handleDelete)
	mux.HandleFunc("/api/create-dir", handleCreateDir)
	mux.HandleFunc("/api/speedtest", handleSpeedTest)

		// 启动HTTP服务器
		startHTTPServer(mux)
}

func loadConfig() {
	// 默认配置
	config = Config{
		HTTPPort:  "7051",
		UploadDir: "uploads",
		StaticDir: "static",
	}

	// 尝试从配置文件加载
	configFile := "../config.json"
	if _, err := os.Stat(configFile); err == nil {
		data, err := os.ReadFile(configFile)
		if err == nil {
			var fullConfig struct {
				Backend struct {
					Port      string `json:"port"`
					UploadDir string `json:"upload_dir"`
					StaticDir string `json:"static_dir"`
				}
			}
			json.Unmarshal(data, &fullConfig)
			config.HTTPPort = fullConfig.Backend.Port
			uploadDir = fullConfig.Backend.UploadDir
			staticDir = fullConfig.Backend.StaticDir
		}
	} else {
		uploadDir = config.UploadDir
		staticDir = config.StaticDir
	}
}

func createDirectories() {
	os.MkdirAll(uploadDir, 0755)
	os.MkdirAll(staticDir, 0755)
	os.MkdirAll(filepath.Join(staticDir, "css"), 0755)
	os.MkdirAll(filepath.Join(staticDir, "js"), 0755)
	os.MkdirAll(filepath.Join(staticDir, "images"), 0755)
}

func loadTemplates() {
	// 创建默认模板
	tmplContent := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>文件传输服务器</title>
    <link rel="stylesheet" href="/static/css/style.css">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
</head>
<body>
    <div class="container">
        <header>
            <h1><i class="fas fa-cloud-upload-alt"></i> 文件传输服务器</h1>
            <p>支持HTTP/HTTPS，可直接上传文件或目录</p>
        </header>

        <div class="upload-section">
            <h2><i class="fas fa-upload"></i> 上传文件</h2>
            <form id="uploadForm" enctype="multipart/form-data">
                <div class="form-group">
                    <input type="file" id="fileInput" name="file" multiple>
                    <label for="fileInput" class="file-label">
                        <i class="fas fa-folder-open"></i> 选择文件
                    </label>
                </div>
                <div class="form-group">
                    <button type="submit" class="btn btn-primary">
                        <i class="fas fa-upload"></i> 上传文件
                    </button>
                </div>
            </form>

            <h2><i class="fas fa-folder"></i> 上传目录</h2>
            <form id="uploadDirForm" enctype="multipart/form-data">
                <div class="form-group">
                    <input type="file" id="dirInput" name="files" multiple webkitdirectory directory>
                    <label for="dirInput" class="file-label">
                        <i class="fas fa-folder"></i> 选择目录
                    </label>
                </div>
                <div class="form-group">
                    <button type="submit" class="btn btn-primary">
                        <i class="fas fa-upload"></i> 上传目录
                    </button>
                </div>
            </form>

            <h2><i class="fas fa-folder-plus"></i> 创建目录</h2>
            <form id="createDirForm">
                <div class="form-group">
                    <input type="text" id="dirName" placeholder="输入目录名称" required>
                </div>
                <div class="form-group">
                    <button type="submit" class="btn btn-secondary">
                        <i class="fas fa-plus"></i> 创建目录
                    </button>
                </div>
            </form>

            <h2><i class="fas fa-tachometer-alt"></i> 网络测速</h2>
            <div class="form-group">
                <button id="speedTestBtn" class="btn btn-info">
                    <i class="fas fa-bolt"></i> 开始测速
                </button>
                <div id="speedTestResult" style="margin-top: 10px; display: none;">
                    <p>测速结果: <span id="bandwidthValue">0</span> Mbps</p>
                    <p>带宽限制(80%): <span id="bandwidthLimitValue">0</span> Mbps</p>
                </div>
            </div>
        </div>

        <div class="file-list-section">
            <h2><i class="fas fa-list"></i> 文件列表</h2>
            <div class="path-nav">
                <span class="path-item" data-path="">根目录</span>
            </div>
            <div id="fileList" class="file-list">
                <!-- 文件列表将通过JavaScript动态加载 -->
            </div>
        </div>

        <footer>
            <p>服务器运行在: <span id="serverInfo">加载中...</span></p>
            <p>当前时间: <span id="currentTime"></span></p>
        </footer>
    </div>

    <script src="/static/js/main.js"></script>
</body>
</html>`

	// 创建模板文件
	tmplFile := filepath.Join(staticDir, "index.html")
	os.WriteFile(tmplFile, []byte(tmplContent), 0644)

	// 解析模板
	var err error
	templates, err = template.New("index").Parse(tmplContent)
	if err != nil {
		log.Printf("模板解析错误: %v", err)
	}
}

func startHTTPServer(mux http.Handler) {
	log.Printf("HTTP服务器启动在 :%s", config.HTTPPort)
	err := http.ListenAndServe(":"+config.HTTPPort, mux)
	if err != nil {
		log.Fatalf("HTTP服务器启动失败: %v", err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if templates != nil {
		templates.Execute(w, nil)
	} else {
		// 如果模板未加载，返回静态文件
		indexFile := filepath.Join(staticDir, "index.html")
		http.ServeFile(w, r, indexFile)
	}
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		sendJSONError(w, "只支持POST方法", http.StatusMethodNotAllowed)
		return
	}

	// 解析表单
	err := r.ParseMultipartForm(100 << 20) // 100MB
	if err != nil {
		sendJSONError(w, "表单解析错误: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 获取目标路径
	targetPath := r.FormValue("path")
	if targetPath == "" {
		targetPath = uploadDir
	} else {
		targetPath = filepath.Join(uploadDir, targetPath)
	}

	// 确保目录存在
	os.MkdirAll(targetPath, 0755)

	// 处理上传的文件
	files := r.MultipartForm.File["file"]
	results := make([]map[string]string, 0)

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			log.Printf("打开文件错误: %v", err)
			continue
		}
		defer file.Close()

		// 创建目标文件
		dstPath := filepath.Join(targetPath, fileHeader.Filename)
		dst, err := os.Create(dstPath)
		if err != nil {
			log.Printf("创建文件错误: %v", err)
			continue
		}
		defer dst.Close()

		// 复制文件内容
		_, err = io.Copy(dst, file)
		if err != nil {
			log.Printf("复制文件错误: %v", err)
			continue
		}

		results = append(results, map[string]string{
			"filename": fileHeader.Filename,
			"size":     fmt.Sprintf("%d", fileHeader.Size),
			"path":     dstPath,
			"status":   "success",
		})
	}

	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("成功上传 %d 个文件", len(results)),
		"files":   results,
	})
}

func handleUploadDir(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		sendJSONError(w, "只支持POST方法", http.StatusMethodNotAllowed)
		return
	}

	// 解析表单
	err := r.ParseMultipartForm(100 << 20) // 100MB
	if err != nil {
		sendJSONError(w, "表单解析错误: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 获取目标路径
	targetPath := r.FormValue("path")
	if targetPath == "" {
		targetPath = uploadDir
	} else {
		targetPath = filepath.Join(uploadDir, targetPath)
	}

	// 确保目标目录存在
	os.MkdirAll(targetPath, 0755)

	// 处理上传的文件
	files := r.MultipartForm.File["files"]
	results := make([]map[string]string, 0)
	uploadedCount := 0

	// 添加调试信息
	log.Printf("开始处理目录上传，共 %d 个文件", len(files))

	// 即使没有文件上传，也确保目录结构存在
	if len(files) == 0 {
		// 从表单获取目录名
		dirName := r.FormValue("dirName")
		if dirName != "" {
			// 创建完整路径
			fullPath := filepath.Join(targetPath, dirName)
			if err := os.MkdirAll(fullPath, 0755); err != nil {
				log.Printf("创建空目录失败: %v", err)
			} else {
				log.Printf("成功创建空目录: %s", fullPath)
			}
		}
	}

	for _, fileHeader := range files {
		// 获取文件名（可能包含相对路径）
		filename := fileHeader.Filename

		// 从Content-Disposition头获取完整路径
		if contentDisposition, ok := fileHeader.Header["Content-Disposition"]; ok {
			if len(contentDisposition) > 0 && strings.Contains(contentDisposition[0], "filename=") {
				// 提取带路径的文件名
				disposition := contentDisposition[0]
				start := strings.Index(disposition, "filename=") + len("filename=")
				end := strings.LastIndex(disposition, `"`)
				if start > 0 && end > start {
					filename = strings.TrimSpace(disposition[start:end])
					filename = strings.Trim(filename, `"`)
				}
			}
		}

		// 处理路径分隔符
		filename = strings.ReplaceAll(filename, "\\", "/")
		relativePath := filename

		// 获取目录路径并创建
		dirPath := filepath.Join(targetPath, filepath.Dir(relativePath))
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			log.Printf("创建目录失败: %v", err)
			continue
		}

		// 创建目标文件
		dstPath := filepath.Join(dirPath, filepath.Base(relativePath))
		file, err := fileHeader.Open()
		if err != nil {
			log.Printf("打开文件错误: %v", err)
			continue
		}
		defer file.Close()

		dst, err := os.Create(dstPath)
		if err != nil {
			log.Printf("创建文件错误: %v", err)
			continue
		}
		defer dst.Close()

		// 复制文件内容
		if _, err = io.Copy(dst, file); err != nil {
			log.Printf("复制文件错误: %v", err)
			continue
		}

		results = append(results, map[string]string{
			"filename": filename,
			"size":     fmt.Sprintf("%d", fileHeader.Size),
			"path":     dstPath,
			"status":   "success",
		})
		uploadedCount++
	}

	// 返回JSON响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("成功上传 %d/%d 个文件", uploadedCount, len(files)),
		"files":   results,
	})
}
func handleFileList(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		path = uploadDir
	} else {
		path = filepath.Join(uploadDir, path)
	}

	// 安全检查：确保路径在uploadDir内
	if !strings.HasPrefix(path, uploadDir) {
		sendJSONError(w, "无效的路径", http.StatusBadRequest)
		return
	}

	// 读取目录
	entries, err := os.ReadDir(path)
	if err != nil {
		sendJSONError(w, "无法读取目录: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("读取目录 %s 成功，找到 %d 个文件/目录", path, len(entries))

	fileInfos := make([]FileInfo, 0)
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		relPath, _ := filepath.Rel(uploadDir, filepath.Join(path, entry.Name()))
		urlPath := "/api/download/" + relPath

		fileInfos = append(fileInfos, FileInfo{
			Name:    entry.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   entry.IsDir(),
			Path:    relPath,
			URL:     urlPath,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fileInfos)
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/download/")
	if path == "" {
		sendJSONError(w, "未指定文件", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(uploadDir, path)

	// 安全检查
	if !strings.HasPrefix(filePath, uploadDir) {
		sendJSONError(w, "无效的文件路径", http.StatusBadRequest)
		return
	}

	// 检查文件是否存在
	info, err := os.Stat(filePath)
	if err != nil {
		sendJSONError(w, "文件不存在", http.StatusNotFound)
		return
	}

	if info.IsDir() {
		// 如果是目录，创建ZIP压缩包
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", filepath.Base(path)))
		// 这里可以添加ZIP压缩逻辑，简化版本直接返回错误
		sendJSONError(w, "目录下载功能暂未实现", http.StatusNotImplemented)
		return
	}

	// 如果是文件，直接提供下载
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(path)))
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, filePath)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		sendJSONError(w, "只支持DELETE方法", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/delete/")
	if path == "" {
		sendJSONError(w, "未指定文件", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(uploadDir, path)

	// 安全检查
	if !strings.HasPrefix(filePath, uploadDir) {
		sendJSONError(w, "无效的文件路径", http.StatusBadRequest)
		return
	}

	err := os.RemoveAll(filePath)
	if err != nil {
		sendJSONError(w, "删除失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "删除成功",
	})
}

func handleCreateDir(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		sendJSONError(w, "只支持POST方法", http.StatusMethodNotAllowed)
		return
	}

	dirName := r.FormValue("name")
	parentPath := r.FormValue("path")

	if dirName == "" {
		sendJSONError(w, "目录名不能为空", http.StatusBadRequest)
		return
	}

	targetPath := filepath.Join(uploadDir, parentPath, dirName)

	// 安全检查
	if !strings.HasPrefix(targetPath, uploadDir) {
		sendJSONError(w, "无效的路径", http.StatusBadRequest)
		return
	}

	err := os.MkdirAll(targetPath, 0755)
	if err != nil {
		sendJSONError(w, "创建目录失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "目录创建成功",
		"path":    targetPath,
	})
}

// handleSpeedTest 处理测速请求
func handleSpeedTest(w http.ResponseWriter, r *http.Request) {
	// 默认测试数据大小为1MB
	sizeStr := r.URL.Query().Get("size")
	size := int64(1 * 1024 * 1024) // 1MB默认值

	if sizeStr != "" {
		if s, err := strconv.ParseInt(sizeStr, 10, 64); err == nil && s > 0 {
			size = s
		}
	}

	// 限制最大大小为100MB，防止滥用
	if size > 100*1024*1024 {
		size = 100 * 1024 * 1024
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// 生成随机数据
	const chunkSize = 32 * 1024 // 32KB chunks
	buf := make([]byte, chunkSize)

	// 初始化随机数生成器
	rand.Seed(time.Now().UnixNano())

	// 写入数据
	written := int64(0)
	for written < size {
		// 生成随机数据
		rand.Read(buf)

		// 计算本次写入的大小
		toWrite := int64(len(buf))
		if written+toWrite > size {
			toWrite = size - written
		}

		// 写入响应
		n, err := w.Write(buf[:toWrite])
		if err != nil {
			log.Printf("测速响应写入错误: %v", err)
			return
		}

		written += int64(n)

		// 刷新缓冲区，确保数据立即发送
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}
