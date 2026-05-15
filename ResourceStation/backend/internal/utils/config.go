package utils

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Upload   UploadConfig   `mapstructure:"upload"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	Database        string `mapstructure:"database"`
	Charset         string `mapstructure:"charset"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	BasePath     string   `mapstructure:"base_path"`
	TempPath     string   `mapstructure:"temp_path"`
	MaxFileSize  int64    `mapstructure:"max_file_size"`
	ChunkSize    int64    `mapstructure:"chunk_size"`
	AllowedTypes []string `mapstructure:"allowed_types"`
}

// UploadConfig 上传配置
type UploadConfig struct {
	MaxConcurrentUploads int  `mapstructure:"max_concurrent_uploads"`
	CleanupTempFiles     bool `mapstructure:"cleanup_temp_files"`
	TempFileExpiry       int  `mapstructure:"temp_file_expiry"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level    string `mapstructure:"level"`
	Output   string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
}

var AppConfig *Config

// LoadConfig 加载配置文件
func LoadConfig(configPath string) error {
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	// 获取配置文件的绝对路径
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return fmt.Errorf("获取配置文件绝对路径失败: %v", err)
	}

	// 直接读取JSON文件
	data, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析JSON到结构体
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 转换配置结构以匹配原有格式
	config.Server.Host = "0.0.0.0"
	config.Server.Port = config.Backend.Port
	config.Database.Host = config.Database.Host
	config.Database.Port = config.Database.Port
	config.Database.Username = config.Database.User
	config.Database.Password = config.Database.Password
	config.Database.Database = config.Database.Name

	AppConfig = &config
	log.Printf("配置文件加载成功: %s", absPath)
	return nil
}

// GetConfig 获取应用配置
func GetConfig() *Config {
	if AppConfig == nil {
		log.Fatal("配置未初始化，请先调用LoadConfig")
	}
	return AppConfig
}
