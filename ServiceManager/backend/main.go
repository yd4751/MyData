package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Config 定义配置结构体
type Config struct {
	Backend struct {
		Port    int    `json:"port"`
		Command string `json:"command"`
		PidFile string `json:"pid_file"`
	} `json:"backend"`
	Database struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Name     string `json:"name"`
	} `json:"database"`
	Frontend struct {
		URL string `json:"url"`
	} `json:"frontend"`
}

var config Config

// loadConfig 加载配置文件
func loadConfig() error {
	file, err := os.ReadFile("../config.json")
	if err != nil {
		return err
	}
	return json.Unmarshal(file, &config)
}

// Service 定义服务结构体
type Service struct {
	ID        int          `json:"id"`
	Name      string       `json:"name"`
	Status    string       `json:"status"`
	StartTime sql.NullTime `json:"startTime"`
	Uptime    *string      `json:"uptime"`
	URL       *string      `json:"url"`
	Address   *string      `json:"address"`
	Command   *string      `json:"command"`
	PID       *int64       `json:"pid"`
}

var db *sql.DB

func main() {
	// 加载配置
	if err := loadConfig(); err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 初始化数据库连接
	initDB()
	defer db.Close()

	// 设置静态文件服务
	fs := http.FileServer(http.Dir("../frontend"))
	http.Handle("/", fs)

	// 添加CORS中间件
	corsHandler := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == "OPTIONS" {
				log.Println("Handling OPTIONS preflight request")
				w.WriteHeader(http.StatusOK)
				return
			}

			next(w, r)
		}
	}

	// 设置API路由
	http.HandleFunc("/api/services", corsHandler(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			getServices(w, r)
		case "POST":
			addService(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	http.HandleFunc("/api/services/", corsHandler(deleteService))
	http.HandleFunc("/api/services/start", corsHandler(startService))
	http.HandleFunc("/api/services/stop", corsHandler(stopService))
	http.HandleFunc("/api/services/import", corsHandler(importServices))
	http.HandleFunc("/api/services/edit", corsHandler(updateService))

	log.Printf("Server starting on port %d...", config.Backend.Port)
	if err := http.ListenAndServe(":"+strconv.Itoa(config.Backend.Port), nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

// initDB 初始化数据库连接
func initDB() {
	var err error
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "root" // 默认用户名
	}
	dbPass := os.Getenv("DB_PASS")
	if dbPass == "" {
		dbPass = "12345678" // 默认密码
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "service_manager" // 默认数据库名
	}
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost:3306" // 默认主机
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true", dbUser, dbPass, dbHost, dbName)
	log.Printf("Connecting to database with DSN: %s", dsn)

	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Successfully connected to database")

	// 测试查询
	rows, err := db.Query("SELECT COUNT(*) FROM services")
	if err != nil {
		log.Println("Warning: services table may not exist:", err)
	} else {
		var count int
		rows.Next()
		rows.Scan(&count)
		log.Printf("Found %d services in database", count)
		rows.Close()
	}
}

// getServices 获取所有服务
func getServices(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM services")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var services []Service
	for rows.Next() {
		var s Service
		var name sql.NullString
		var status sql.NullString
		var url sql.NullString
		var address sql.NullString
		var command sql.NullString
		var pid sql.NullInt64

		err := rows.Scan(&s.ID, &name, &status, &s.StartTime, &s.Uptime, &url, &address, &command, &pid)
		// 处理null值并设置默认值
		s.Name = name.String
		if !name.Valid {
			s.Name = ""
		}
		s.Status = status.String
		if !status.Valid {
			s.Status = "stopped"
		}
		if url.Valid {
			s.URL = &url.String
		}
		if address.Valid {
			s.Address = &address.String
		}
		if command.Valid {
			s.Command = &command.String
		}
		if pid.Valid {
			s.PID = &pid.Int64
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		services = append(services, s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

// startService 启动服务
func startService(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var service Service
	err := json.NewDecoder(r.Body).Decode(&service)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if service.ID == 0 {
		http.Error(w, "Service ID is required", http.StatusBadRequest)
		return
	}

	// 从数据库获取服务命令
	var command string
	err = db.QueryRow("SELECT command FROM services WHERE id = ?", service.ID).Scan(&command)
	if err != nil {
		http.Error(w, "Failed to get service command: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if command == "" {
		http.Error(w, "Service command is empty", http.StatusBadRequest)
		return
	}

	// 执行命令
	cmd := exec.Command("sh", "-c", command)
	fmt.Printf("Starting service with command: %s", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // 创建进程组
	err = cmd.Start()
	if err != nil {
		http.Error(w, "Failed to start service: "+err.Error(), http.StatusInternalServerError)
		return
	}

	pid := int64(cmd.Process.Pid)
	service.PID = &pid
	service.Status = "running"
	service.StartTime = sql.NullTime{Time: time.Now(), Valid: true}

	// 更新数据库
	_, err = db.Exec("UPDATE services SET status = ?, start_time = ?, pid = ? WHERE id = ?",
		service.Status,
		service.StartTime,
		pid,
		service.ID)
	if err != nil {
		http.Error(w, "Failed to update service: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 写入PID文件
	pidFile := config.Backend.PidFile
	if pidFile != "" {
		err = os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0644)
		if err != nil {
			log.Printf("Warning: failed to write PID file: %v", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(service)
}

// stopService 停止服务
func stopService(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var service Service
	err := json.NewDecoder(r.Body).Decode(&service)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if service.ID == 0 {
		http.Error(w, "Service ID is required", http.StatusBadRequest)
		return
	}

	// 从数据库获取PID
	var pid int64
	err = db.QueryRow("SELECT pid FROM services WHERE id = ?", service.ID).Scan(&pid)
	if err != nil {
		http.Error(w, "Failed to get service PID: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if pid == 0 {
		http.Error(w, "Service is not running", http.StatusBadRequest)
		return
	}

	// 终止进程
	err = syscall.Kill(-int(pid), syscall.SIGTERM) // 使用负数PID终止整个进程组
	if err != nil {
		// 如果进程不存在，可能是已经停止
		if !errors.Is(err, syscall.ESRCH) {
			http.Error(w, "Failed to stop service: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// 更新数据库
	_, err = db.Exec("UPDATE services SET status = 'stopped', pid = NULL, uptime = TIMESTAMPDIFF(SECOND, start_time, NOW()) WHERE id = ?", service.ID)
	if err != nil {
		http.Error(w, "Failed to update service: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 删除PID文件
	pidFile := config.Backend.PidFile
	if pidFile != "" {
		os.Remove(pidFile)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

// updateService 更新服务信息
func updateService(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" && r.Method != "PATCH" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var service Service
	err := json.NewDecoder(r.Body).Decode(&service)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 验证必要字段
	if service.ID == 0 {
		http.Error(w, "Service ID is required", http.StatusBadRequest)
		return
	}
	if service.Name == "" {
		http.Error(w, "Service name is required", http.StatusBadRequest)
		return
	}

	// 检查服务是否存在
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM services WHERE id = ?", service.ID).Scan(&count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if count == 0 {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	// 更新服务
	_, err = db.Exec("UPDATE services SET name = ?, status = ?, url = ?, address = ?, command = ? WHERE id = ?",
		service.Name,
		service.Status,
		toNullString(service.URL),
		toNullString(service.Address),
		toNullString(service.Command),
		service.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(service)
}

// addService 添加新服务
func addService(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var service Service
	err := json.NewDecoder(r.Body).Decode(&service)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 验证必要字段
	if service.Name == "" {
		http.Error(w, "Service name is required", http.StatusBadRequest)
		return
	}

	// 设置默认状态
	if service.Status == "" {
		service.Status = "stopped"
	}

	result, err := db.Exec("INSERT INTO services (name, status, url, address, command) VALUES (?, ?, ?, ?, ?)",
		service.Name,
		service.Status,
		toNullString(service.URL),
		toNullString(service.Address),
		toNullString(service.Command))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	service.ID = int(id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(service)
}

// deleteService 删除服务
func deleteService(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 从URL路径中提取ID
	idStr := strings.TrimPrefix(r.URL.Path, "/api/services/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid service ID", http.StatusBadRequest)
		return
	}

	// 检查服务是否存在
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM services WHERE id = ?", id).Scan(&count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if count == 0 {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	// 删除服务
	_, err = db.Exec("DELETE FROM services WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// importServices 批量导入服务
func importServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var services []Service
	err := json.NewDecoder(r.Body).Decode(&services)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 批量插入服务
	for _, service := range services {
		// 验证必要字段
		if service.Name == "" {
			tx.Rollback()
			http.Error(w, "Service name is required", http.StatusBadRequest)
			return
		}

		// 设置默认状态
		if service.Status == "" {
			service.Status = "stopped"
		}

		_, err := tx.Exec("INSERT INTO services (name, status, url, address, command) VALUES (?, ?, ?, ?, ?)",
			service.Name,
			service.Status,
			toNullString(service.URL),
			toNullString(service.Address),
			toNullString(service.Command))
		if err != nil {
			tx.Rollback()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Successfully imported services"))
}

// toNullString 将*string转换为sql.NullString
func toNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}
