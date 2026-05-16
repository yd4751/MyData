package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Port       string `json:"port"`
	DBHost     string `json:"db_host"`
	DBPort     string `json:"db_port"`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
	DBName     string `json:"db_name"`
	MediaDir   string `json:"media_dir"`
	StaticDir  string `json:"static_dir"`
}

func LoadConfig(configPath string) (*Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	setDefaults(&cfg)
	return &cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.DBHost == "" {
		cfg.DBHost = "localhost"
	}
	if cfg.DBPort == "" {
		cfg.DBPort = "3306"
	}
	if cfg.DBUser == "" {
		cfg.DBUser = "root"
	}
	if cfg.DBPassword == "" {
		cfg.DBPassword = "password"
	}
	if cfg.DBName == "" {
		cfg.DBName = "mediastation"
	}
	if cfg.MediaDir == "" {
		cfg.MediaDir = "./media"
	}
	if cfg.StaticDir == "" {
		cfg.StaticDir = "./frontend"
	}
}
