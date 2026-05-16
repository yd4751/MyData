package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	Storage  StorageConfig  `yaml:"storage"`
	Upload   UploadConfig   `yaml:"upload"`
	Logging  LoggingConfig  `yaml:"logging"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	Database        string `yaml:"database"`
	Charset         string `yaml:"charset"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret      string `yaml:"secret"`
	ExpireHours int    `yaml:"expire_hours"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	BasePath     string   `yaml:"base_path"`
	TempPath     string   `yaml:"temp_path"`
	MaxFileSize  int64    `yaml:"max_file_size"`
	ChunkSize    int64    `yaml:"chunk_size"`
	AllowedTypes []string `yaml:"allowed_types"`
}

// UploadConfig 上传配置
type UploadConfig struct {
	MaxConcurrentUploads int  `yaml:"max_concurrent_uploads"`
	CleanupTempFiles     bool `yaml:"cleanup_temp_files"`
	TempFileExpiry       int  `yaml:"temp_file_expiry"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level    string `yaml:"level"`
	Output   string `yaml:"output"`
	FilePath string `yaml:"file_path"`
}

var AppConfig *Config

// LoadConfig 加载配置文件
func LoadConfig(jsonConfigPath, yamlConfigPath string) error {
	if yamlConfigPath == "" {
		yamlConfigPath = "configs/config.yaml"
	}
	if jsonConfigPath == "" {
		jsonConfigPath = "../../config.json"
	}

	// 加载JSON配置(只读取server部分)
	var jsonConfig struct {
		Backend struct {
			Host string `json:"host"`
			Port int    `json:"port"`
		} `json:"backend"`
	}

	jsonAbsPath, err := filepath.Abs(jsonConfigPath)
	if err != nil {
		return fmt.Errorf("获取JSON配置文件绝对路径失败: %v", err)
	}

	jsonData, err := os.ReadFile(jsonAbsPath)
	if err != nil {
		return fmt.Errorf("读取JSON配置文件失败: %v", err)
	}

	if err := json.Unmarshal(jsonData, &jsonConfig); err != nil {
		return fmt.Errorf("解析JSON配置文件失败: %v", err)
	}

	// 加载YAML配置
	yamlAbsPath, err := filepath.Abs(yamlConfigPath)
	if err != nil {
		return fmt.Errorf("获取YAML配置文件绝对路径失败: %v", err)
	}

	yamlData, err := os.ReadFile(yamlAbsPath)
	if err != nil {
		return fmt.Errorf("读取YAML配置文件失败: %v", err)
	}

	var yamlConfig Config
	if err := yaml.Unmarshal(yamlData, &yamlConfig); err != nil {
		return fmt.Errorf("解析YAML配置文件失败: %v", err)
	}

	// 合并配置(JSON中的server配置优先)
	yamlConfig.Server.Host = jsonConfig.Backend.Host
	yamlConfig.Server.Port = jsonConfig.Backend.Port

	AppConfig = &yamlConfig
	log.Printf("配置文件加载成功 - JSON: %s, YAML: %s", jsonAbsPath, yamlAbsPath)
	return nil
}

// GetConfig 获取应用配置
func GetConfig() *Config {
	if AppConfig == nil {
		log.Fatal("配置未初始化，请先调用LoadConfig")
	}
	return AppConfig
}
