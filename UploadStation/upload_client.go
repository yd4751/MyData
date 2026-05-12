package main

import (
	"bytes"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

const (
	defaultServer = "http://localhost:7051"
)

var (
	serverURL  string
	rateLimit  int // KB/s
	uploadPath string
)

func init() {
	flag.StringVar(&serverURL, "server", defaultServer, "服务器地址")
	flag.IntVar(&rateLimit, "rate-limit", 0, "上传速率限制(KB/s)，0表示自动测速80%")
	flag.Parse()
}

func main() {
	args := flag.Args()
	if len(args) == 0 {
		log.Fatal("请指定要上传的文件或目录路径")
	}
	uploadPath = args[0]

	// 检查路径是否存在
	info, err := os.Stat(uploadPath)
	if os.IsNotExist(err) {
		log.Fatalf("路径不存在: %s", uploadPath)
	}

	// 自动测速
	if rateLimit == 0 {
		rateLimit = measureBandwidth() * 8 / 10 // 80%带宽
		log.Printf("自动设置上传速率限制为: %d KB/s", rateLimit)
	}

	// 创建速率限制器
	limiter := NewRateLimiter(rateLimit)

	if info.IsDir() {
		// 上传目录
		err = uploadDirectory(uploadPath, limiter)
	} else {
		// 上传单个文件
		err = uploadFile(uploadPath, limiter)
	}

	if err != nil {
		log.Fatal(err)
	}
}

// 测速函数
func measureBandwidth() int {
	start := time.Now()
	
	// 创建测试数据 (1MB)
	testData := make([]byte, 1024*1024)
	_, err := rand.Read(testData)
	if err != nil {
		log.Printf("测速数据生成失败: %v", err)
		return 1024 // 默认1MB/s
	}

	// 上传测试数据
	url := fmt.Sprintf("%s/speedtest", serverURL)
	resp, err := http.Post(url, "application/octet-stream", bytes.NewReader(testData))
	if err != nil {
		log.Printf("测速请求失败: %v", err)
		return 1024 // 默认1MB/s
	}
	defer resp.Body.Close()

	// 计算带宽 (KB/s)
	duration := time.Since(start)
	speed := float64(len(testData)) / duration.Seconds() / 1024
	
	log.Printf("测速完成: %.2f KB/s", speed)
	return int(speed)
}

// 速率限制器
type RateLimiter struct {
	bucket chan struct{}
	stop   chan struct{}
}

func NewRateLimiter(rateKB int) *RateLimiter {
	if rateKB <= 0 {
		return nil
	}

	limiter := &RateLimiter{
		bucket: make(chan struct{}, rateKB),
		stop:   make(chan struct{}),
	}

	// 计算每个令牌的间隔时间(毫秒)
	interval := time.Second / time.Duration(rateKB)

	// 启动令牌填充goroutine
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				select {
				case limiter.bucket <- struct{}{}:
				default: // 桶已满，丢弃令牌
				}
			case <-limiter.stop:
				return
			}
		}
	}()

	return limiter
}

// 获取令牌，控制上传速率
func (r *RateLimiter) Take(n int) {
	if r == nil {
		return
	}
	for i := 0; i < n; i++ {
		<-r.bucket
	}
}

// 停止令牌桶
func (r *RateLimiter) Stop() {
	if r == nil {
		return
	}
	close(r.stop)
}

// 上传目录
func uploadDirectory(dirPath string, limiter *RateLimiter) error {
	const smallFileSize = 1024 * 1024 // 1MB
	var smallFiles []string
	
	// 先收集小文件
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		
		if info.Size() < smallFileSize {
			smallFiles = append(smallFiles, path)
		} else {
			// 大文件直接上传
			if err := uploadFile(path, limiter); err != nil {
				return err
			}
		}
		return nil
	})
	
	if err != nil {
		return err
	}
	
	// 批量上传小文件(每次最多10个或总大小5MB)
	batchSize := 0
	batch := make([]string, 0, 10)
	
	for _, file := range smallFiles {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		
		if batchSize+int(info.Size()) > 5*1024*1024 || len(batch) >= 10 {
			if err := uploadBatch(batch, dirPath, limiter); err != nil {
				return err
			}
			batch = batch[:0]
			batchSize = 0
		}
		
		batch = append(batch, file)
		batchSize += int(info.Size())
	}
	
	// 上传剩余文件
	if len(batch) > 0 {
		return uploadBatch(batch, dirPath, limiter)
	}
	
	return nil
}

// 批量上传小文件
func uploadBatch(files []string, baseDir string, limiter *RateLimiter) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	// 添加路径字段
	err := writer.WriteField("path", ".")
	if err != nil {
		return fmt.Errorf("写入路径字段失败: %v", err)
	}
	
	// 添加多个文件
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			continue
		}
		defer f.Close()
		
		relPath, err := filepath.Rel(baseDir, file)
		if err != nil {
			relPath = filepath.Base(file)
		}
		
		part, err := writer.CreateFormFile("files", relPath)
		if err != nil {
			return fmt.Errorf("创建文件字段失败: %v", err)
		}
		
		if _, err := io.Copy(part, f); err != nil {
			return fmt.Errorf("复制文件内容失败: %v", err)
		}
	}
	
	// 应用速率限制(按总大小)
	totalSize := body.Len()
	limiter.Take(totalSize / 1024)
	
	// 完成表单
	if err := writer.Close(); err != nil {
		return fmt.Errorf("关闭表单失败: %v", err)
	}
	
		// 发送请求
		uploadURL, err := url.JoinPath(serverURL, "upload-dir")
		if err != nil {
			return fmt.Errorf("构建上传URL失败: %v", err)
		}
		req, err := http.NewRequest("POST", uploadURL, body)
		if err != nil {
			return fmt.Errorf("创建请求失败: %v", err)
		}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("上传失败: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("上传失败: %s", resp.Status)
	}
	
	fmt.Printf("\r批量上传完成: %d个小文件\n", len(files))
	return nil
}

// 上传文件
func uploadFile(filePath string, limiter *RateLimiter) error {
	log.Printf("开始上传文件: %s", filePath)
	
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %v", err)
	}

	// 计算相对路径
	relPath := ""
	if uploadPath != "" {
		var err error
		relPath, err = filepath.Rel(uploadPath, filePath)
		if err != nil {
			log.Printf("计算相对路径失败, 使用文件名: %v", err)
			relPath = filepath.Base(filePath)
		}
	} else {
		relPath = filepath.Base(filePath)
	}
	
	log.Printf("上传路径映射: %s -> %s", filePath, relPath)

	// 准备multipart表单
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加路径字段
	targetPath := filepath.Dir(relPath)
	log.Printf("设置上传目标路径: %s", targetPath)
	err = writer.WriteField("path", targetPath)
	if err != nil {
		return fmt.Errorf("写入路径字段失败: %v", err)
	}

	// 创建文件字段
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("创建文件字段失败: %v", err)
	}

	// 分片上传 (1MB chunks)
	chunkSize := 1024 * 1024
	totalSize := fileInfo.Size()
	bytesUploaded := 0
	buf := make([]byte, chunkSize)

	for {
		// 应用速率限制
		limiter.Take(chunkSize / 1024) // 每次取n个KB的令牌

		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("读取文件失败: %v", err)
		}
		if n == 0 {
			break
		}

		_, err = part.Write(buf[:n])
		if err != nil {
			return fmt.Errorf("写入表单失败: %v", err)
		}

		bytesUploaded += n
		progress := float64(bytesUploaded) / float64(totalSize) * 100
		fmt.Printf("\r上传中: %s %.2f%%", filepath.Base(filePath), progress)
	}

	// 完成表单
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("关闭表单失败: %v", err)
	}

			// 发送请求
			uploadURL, err := url.JoinPath(serverURL, "upload")
			if err != nil {
				return fmt.Errorf("构建上传URL失败: %v", err)
			}
			req, err := http.NewRequest("POST", uploadURL, body)
			if err != nil {
				return fmt.Errorf("创建请求失败: %v", err)
			}
			req.Header.Set("Content-Type", writer.FormDataContentType())

			log.Printf("发送上传请求到: %s", uploadURL)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("上传失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取响应失败: %v", err)
	}

	log.Printf("服务端响应: %s, 内容: %s", resp.Status, string(respBody))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("上传失败: %s, 响应: %s", resp.Status, string(respBody))
	}

	fmt.Printf("\r上传完成: %s\n", filepath.Base(filePath))
	return nil
}