package main

import (
	"EasyShare/handlers"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type Config struct {
	Port      int    `json:"port"`
	SharedDir string `json:"sharedDir"`
}

func loadConfig() Config {
	defaultConfig := Config{
		Port:      8080,
		SharedDir: "./shared",
	}

	file, err := os.ReadFile("config.json")
	if err != nil {
		log.Printf("Config file not found, using defaults: %v", err)
		return defaultConfig
	}

	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Printf("Error parsing config.json, using defaults: %v", err)
		return defaultConfig
	}

	if config.Port == 0 {
		config.Port = 8080
	}
	if config.SharedDir == "" {
		config.SharedDir = "./shared"
	}

	return config
}

func main() {
	config := loadConfig()

	handlers.SetSharedDir(config.SharedDir)

	http.HandleFunc("/api/files", handlers.GetFiles)
	http.HandleFunc("/api/upload", handlers.UploadFile)
	http.HandleFunc("/api/download", handlers.DownloadFile)
	http.HandleFunc("/api/delete", handlers.DeleteFile)
	http.HandleFunc("/api/mkdir", handlers.MakeDir)
	http.HandleFunc("/api/create-file", handlers.CreateFile)
	http.HandleFunc("/api/events", handlers.SSEHandler)

	http.Handle("/", http.FileServer(http.Dir("../frontend")))

	addr := fmt.Sprintf(":%d", config.Port)
	log.Printf("Server started on http://localhost:%d", config.Port)
	log.Printf("Shared directory: %s", config.SharedDir)
	log.Fatal(http.ListenAndServe(addr, nil))
}