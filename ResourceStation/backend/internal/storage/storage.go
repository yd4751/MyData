package storage

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"resource-station/internal/utils"

	"github.com/google/uuid"
)

// FileStorage 文件存储接口
type FileStorage interface {
	SaveFile(file *multipart.FileHeader, resourceID string, chunkIndex int) (string, int64, error)
	MergeChunks(resourceID string, totalChunks int) error
	MoveSingleFile(resourceID string) error
	GetFilePath(resourceID string) (string, error)
	GetTempFilePath(resourceID string, chunkIndex int) (string, error)
	DeleteFile(resourceID string) error
	GetFileSize(resourceID string) (int64, error)
	CalculateFileHash(filePath string) (string, error)
}

// LocalFileStorage 本地文件存储实现
type LocalFileStorage struct {
	BasePath string
	TempPath string
}

// NewLocalFileStorage 创建本地文件存储实例
func NewLocalFileStorage() (*LocalFileStorage, error) {
	config := utils.GetConfig().Storage

	// 创建存储目录
	if err := os.MkdirAll(config.BasePath, 0755); err != nil {
		return nil, fmt.Errorf("创建存储目录失败: %v", err)
	}

	// 创建临时目录
	if err := os.MkdirAll(config.TempPath, 0755); err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %v", err)
	}

	return &LocalFileStorage{
		BasePath: config.BasePath,
		TempPath: config.TempPath,
	}, nil
}

// SaveFile 保存文件（支持分片上传）
func (s *LocalFileStorage) SaveFile(file *multipart.FileHeader, resourceID string, chunkIndex int) (string, int64, error) {
	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		return "", 0, fmt.Errorf("打开上传文件失败: %v", err)
	}
	defer src.Close()

	// 创建目标文件路径
	var destPath string
	if chunkIndex >= 0 {
		// 分片上传，保存到临时目录
		chunkDir := filepath.Join(s.TempPath, resourceID)
		// 创建分片目录
		if err := os.MkdirAll(chunkDir, 0755); err != nil {
			return "", 0, fmt.Errorf("创建分片目录失败: %v", err)
		}
		destPath = filepath.Join(chunkDir, fmt.Sprintf("chunk_%d", chunkIndex))
		log.Printf("保存分片文件到: %s", destPath)
	} else {
		// 普通上传，保存到最终目录
		destDir := filepath.Join(s.BasePath, resourceID[:2], resourceID[2:4])
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return "", 0, fmt.Errorf("创建目标目录失败: %v", err)
		}
		destPath = filepath.Join(destDir, resourceID)
	}

	// 创建目标文件
	dest, err := os.Create(destPath)
	if err != nil {
		return "", 0, fmt.Errorf("创建目标文件失败: %v", err)
	}
	defer dest.Close()

	// 复制文件内容
	written, err := io.Copy(dest, src)
	if err != nil {
		return "", 0, fmt.Errorf("复制文件内容失败: %v", err)
	}

	return destPath, written, nil
}

// MergeChunks 合并分片文件
func (s *LocalFileStorage) MergeChunks(resourceID string, totalChunks int) error {
	// 创建最终文件路径
	destDir := filepath.Join(s.BasePath, resourceID[:2], resourceID[2:4])
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}
	destPath := filepath.Join(destDir, resourceID)

	// 创建最终文件
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("创建最终文件失败: %v", err)
	}
	defer destFile.Close()

	// 合并所有分片
	for i := 0; i < totalChunks; i++ {
		chunkPath := filepath.Join(s.TempPath, resourceID, fmt.Sprintf("chunk_%d", i))

		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			return fmt.Errorf("打开分片文件失败: %v", err)
		}

		_, err = io.Copy(destFile, chunkFile)
		chunkFile.Close()
		if err != nil {
			return fmt.Errorf("合并分片文件失败: %v", err)
		}
	}

	// 删除临时分片目录
	tempDir := filepath.Join(s.TempPath, resourceID)
	if err := os.RemoveAll(tempDir); err != nil {
		log.Printf("警告: 删除临时目录失败: %v", err)
	}

	return nil
}

// MoveSingleFile 移动单文件从临时目录到最终目录
func (s *LocalFileStorage) MoveSingleFile(resourceID string) error {
	// 检查临时目录中的文件
	tempDir := filepath.Join(s.TempPath, resourceID)

	// 查找临时文件（可能是chunk_0或以resourceID命名的文件）
	var tempFilePath string

	// 首先检查是否有chunk_0文件（分片上传的单文件）
	chunkPath := filepath.Join(tempDir, "chunk_0")
	if _, err := os.Stat(chunkPath); err == nil {
		tempFilePath = chunkPath
	} else {
		// 检查是否有直接以resourceID命名的文件
		directPath := filepath.Join(tempDir, resourceID)
		if _, err := os.Stat(directPath); err == nil {
			tempFilePath = directPath
		} else {
			// 检查temp目录根目录
			rootPath := filepath.Join(s.TempPath, resourceID)
			if _, err := os.Stat(rootPath); err == nil {
				tempFilePath = rootPath
			} else {
				return fmt.Errorf("在临时目录中找不到文件: %s", resourceID)
			}
		}
	}

	// 创建最终目录
	destDir := filepath.Join(s.BasePath, resourceID[:2], resourceID[2:4])
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}

	destPath := filepath.Join(destDir, resourceID)

	// 移动文件
	if err := os.Rename(tempFilePath, destPath); err != nil {
		// 如果跨设备移动失败，尝试复制后删除
		if err := copyFile(tempFilePath, destPath); err != nil {
			return fmt.Errorf("移动文件失败: %v", err)
		}
		// 删除原文件
		os.Remove(tempFilePath)
	}

	// 删除临时目录（如果为空）
	if tempDir != s.TempPath {
		os.RemoveAll(tempDir)
	}

	return nil
}

// copyFile 复制文件（用于跨设备移动）
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// GetTempFilePath 获取临时文件路径
func (s *LocalFileStorage) GetTempFilePath(resourceID string, chunkIndex int) (string, error) {
	var filePath string
	if chunkIndex >= 0 {
		filePath = filepath.Join(s.TempPath, resourceID, fmt.Sprintf("chunk_%d", chunkIndex))
	} else {
		filePath = filepath.Join(s.TempPath, resourceID, fmt.Sprintf("chunk_%d", 0))
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 检查目录是否存在
		dirPath := filepath.Dir(filePath)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			return "", fmt.Errorf("临时目录不存在: %s", dirPath)
		}

		// 列出目录内容
		files, err := os.ReadDir(dirPath)
		if err != nil {
			return "", fmt.Errorf("临时文件不存在: %s (无法读取目录: %v)", filePath, err)
		}

		var contents []string
		for _, file := range files {
			contents = append(contents, file.Name())
		}
		return "", fmt.Errorf("临时文件不存在: %s (目录内容: %v)", filePath, contents)
	}

	log.Printf("获取临时文件路径: %s", filePath)

	return filePath, nil
}

// GetFilePath 获取文件路径
func (s *LocalFileStorage) GetFilePath(resourceID string) (string, error) {
	filePath := filepath.Join(s.BasePath, resourceID[:2], resourceID[2:4], resourceID)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("文件不存在: %s", resourceID)
	}

	return filePath, nil
}

// DeleteFile 删除文件
func (s *LocalFileStorage) DeleteFile(resourceID string) error {
	filePath := filepath.Join(s.BasePath, resourceID[:2], resourceID[2:4], resourceID)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // 文件不存在，无需删除
	}

	// 删除文件
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("删除文件失败: %v", err)
	}

	// 尝试删除空目录
	dirPath := filepath.Join(s.BasePath, resourceID[:2], resourceID[2:4])
	if err := os.Remove(dirPath); err != nil && !os.IsExist(err) {
		// 目录非空或删除失败，忽略错误
	}

	return nil
}

// GetFileSize 获取文件大小
func (s *LocalFileStorage) GetFileSize(resourceID string) (int64, error) {
	filePath, err := s.GetFilePath(resourceID)
	if err != nil {
		return 0, err
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, fmt.Errorf("获取文件信息失败: %v", err)
	}

	return fileInfo.Size(), nil
}

// CalculateFileHash 计算文件哈希值（MD5）
func (s *LocalFileStorage) CalculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("计算文件哈希失败: %v", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// GenerateResourceID 生成资源ID（UUID v4）
func GenerateResourceID() string {
	// 使用真正的UUID v4
	id, err := uuid.NewRandom()
	if err != nil {
		// 如果UUID生成失败，回退到时间戳+随机数
		timestamp := time.Now().UnixNano()
		randomPart := fmt.Sprintf("%x", timestamp)
		if len(randomPart) > 32 {
			randomPart = randomPart[:32]
		}
		return randomPart
	}
	return id.String()
}

// ValidateFileType 验证文件类型
func ValidateFileType(fileType string) bool {
	config := utils.GetConfig().Storage

	for _, allowedType := range config.AllowedTypes {
		if strings.HasSuffix(allowedType, "/*") {
			// 通配符匹配，如 "image/*"
			prefix := strings.TrimSuffix(allowedType, "/*")
			if strings.HasPrefix(fileType, prefix) {
				return true
			}
		} else if fileType == allowedType {
			// 精确匹配
			return true
		}
	}

	return false
}

// GetMimeTypeFromExtension 根据文件扩展名获取MIME类型
func GetMimeTypeFromExtension(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	mimeTypes := map[string]string{
		".txt":  "text/plain",
		".pdf":  "application/pdf",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".mp4":  "video/mp4",
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".zip":  "application/zip",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt":  "application/vnd.ms-powerpoint",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	}

	if mimeType, ok := mimeTypes[ext]; ok {
		return mimeType
	}

	return "application/octet-stream"
}
