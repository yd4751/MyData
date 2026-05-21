package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

type ServerStatus struct {
	Timestamp       string  `json:"timestamp"`
	Hostname        string  `json:"hostname"`
	OS              string  `json:"os"`
	Platform        string  `json:"platform"`
	PlatformVersion string  `json:"platform_version"`
	Uptime          uint64  `json:"uptime"`
	CPUUsage        float64 `json:"cpu_usage"`
	MemoryTotal     uint64  `json:"memory_total"`
	MemoryUsed      uint64  `json:"memory_used"`
	MemoryFree      uint64  `json:"memory_free"`
	MemoryUsage     float64 `json:"memory_usage"`
	DiskTotal       uint64  `json:"disk_total"`
	DiskUsed        uint64  `json:"disk_used"`
	DiskFree        uint64  `json:"disk_free"`
	DiskUsage       float64 `json:"disk_usage"`
	NetworkIn       uint64  `json:"network_in"`
	NetworkOut      uint64  `json:"network_out"`
	GoVersion       string  `json:"go_version"`
	NumCPU          int     `json:"num_cpu"`
}

type HardwareInfo struct {
	CPUInfo     []CPUInfo       `json:"cpu_info"`
	MemoryInfo  MemoryDetail    `json:"memory_info"`
	DiskInfo    []DiskDetail    `json:"disk_info"`
	NetworkInfo []NetworkDetail `json:"network_info"`
	HostInfo    HostDetail      `json:"host_info"`
}

type CPUInfo struct {
	CPU       int32    `json:"cpu"`
	VendorID  string   `json:"vendor_id"`
	Family    string   `json:"family"`
	Model     string   `json:"model"`
	ModelName string   `json:"model_name"`
	Stepping  int32    `json:"stepping"`
	Mhz       float64  `json:"mhz"`
	CacheSize int32    `json:"cache_size"`
	Flags     []string `json:"flags"`
	Cores     int      `json:"cores"`
}

type MemoryDetail struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

type DiskDetail struct {
	Device      string  `json:"device"`
	Mountpoint  string  `json:"mountpoint"`
	Fstype      string  `json:"fstype"`
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

type NetworkDetail struct {
	Name         string   `json:"name"`
	HardwareAddr string   `json:"hardware_addr"`
	Addrs        []string `json:"addrs"`
}

type HostDetail struct {
	Hostname        string `json:"hostname"`
	OS              string `json:"os"`
	Platform        string `json:"platform"`
	PlatformFamily  string `json:"platform_family"`
	PlatformVersion string `json:"platform_version"`
	Uptime          uint64 `json:"uptime"`
	Procs           uint64 `json:"procs"`
	KernelVersion   string `json:"kernel_version"`
	KernelArch      string `json:"kernel_arch"`
}

type ProcessInfo struct {
	PID           int32    `json:"pid"`
	Name          string   `json:"name"`
	Exe           string   `json:"exe"`
	Cmdline       string   `json:"cmdline"`
	CPUPercent    float64  `json:"cpu_percent"`
	MemoryPercent float32  `json:"memory_percent"`
	MemoryRSS     uint64   `json:"memory_rss"`
	CreateTime    int64    `json:"create_time"`
	Status        []string `json:"status"`
	Username      string   `json:"username"`
}

type HistoryRecord struct {
	gorm.Model
	Timestamp   time.Time `json:"timestamp"`
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	DiskUsage   float64   `json:"disk_usage"`
	NetworkIn   uint64    `json:"network_in"`
	NetworkOut  uint64    `json:"network_out"`
	MemoryUsed  uint64    `json:"memory_used"`
	DiskUsed    uint64    `json:"disk_used"`
}

var prevNetStats map[string]net.IOCountersStat
var db *gorm.DB
var config Config

type Config struct {
	Server struct {
		Port string `json:"port"`
	} `json:"server"`
	Database struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Username string `json:"username"`
		Password string `json:"password"`
		Database string `json:"database"`
	} `json:"database"`
}

func loadConfig() error {
	file, err := os.ReadFile("config.json")
	if err != nil {
		return err
	}
	return json.Unmarshal(file, &config)
}

func init() {
	prevNetStats = make(map[string]net.IOCountersStat)
	if err := loadConfig(); err != nil {
		log.Println("Failed to load config:", err)
	}
	initDatabase()
	go saveHistoryLoop()
}

func initDatabase() {
	var err error
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Database.Username,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.Database,
	)
	db, err = gorm.Open("mysql", dsn)
	if err != nil {
		log.Println("Failed to connect to MySQL, history feature disabled:", err)
		return
	}
	db.AutoMigrate(&HistoryRecord{})
	log.Println("Connected to MySQL successfully")
}

func saveHistoryLoop() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		if db != nil {
			saveHistoryRecord()
		}
	}
}

func saveHistoryRecord() {
	status, err := getServerStatus()
	if err != nil {
		return
	}
	record := HistoryRecord{
		Timestamp:   time.Now(),
		CPUUsage:    status.CPUUsage,
		MemoryUsage: status.MemoryUsage,
		DiskUsage:   status.DiskUsage,
		NetworkIn:   status.NetworkIn,
		NetworkOut:  status.NetworkOut,
		MemoryUsed:  status.MemoryUsed,
		DiskUsed:    status.DiskUsed,
	}
	db.Create(&record)
}

func getServerStatus() (*ServerStatus, error) {
	now := time.Now().Format(time.RFC3339)
	hostname, _ := os.Hostname()
	hostInfo, _ := host.Info()
	cpuUsage, _ := cpu.Percent(0, false)
	memInfo, _ := mem.VirtualMemory()
	diskInfo, _ := disk.Usage("/")
	netStats, _ := net.IOCounters(true)

	var netIn, netOut uint64
	for _, stat := range netStats {
		if prevStat, ok := prevNetStats[stat.Name]; ok {
			netIn += stat.BytesRecv - prevStat.BytesRecv
			netOut += stat.BytesSent - prevStat.BytesSent
		}
		prevNetStats[stat.Name] = stat
	}

	return &ServerStatus{
		Timestamp:       now,
		Hostname:        hostname,
		OS:              hostInfo.OS,
		Platform:        hostInfo.Platform,
		PlatformVersion: hostInfo.PlatformVersion,
		Uptime:          hostInfo.Uptime,
		CPUUsage:        cpuUsage[0],
		MemoryTotal:     memInfo.Total,
		MemoryUsed:      memInfo.Used,
		MemoryFree:      memInfo.Free,
		MemoryUsage:     memInfo.UsedPercent,
		DiskTotal:       diskInfo.Total,
		DiskUsed:        diskInfo.Used,
		DiskFree:        diskInfo.Free,
		DiskUsage:       diskInfo.UsedPercent,
		NetworkIn:       netIn / 10,
		NetworkOut:      netOut / 10,
		GoVersion:       runtime.Version(),
		NumCPU:          runtime.NumCPU(),
	}, nil
}

func getHardwareInfo() (*HardwareInfo, error) {
	cpuInfo, _ := cpu.Info()
	memInfo, _ := mem.VirtualMemory()
	diskInfo, _ := disk.Partitions(true)
	netInfo, _ := net.Interfaces()
	hostInfo, _ := host.Info()

	var cpuList []CPUInfo
	for _, c := range cpuInfo {
		cpuList = append(cpuList, CPUInfo{
			CPU:       c.CPU,
			VendorID:  c.VendorID,
			Family:    c.Family,
			Model:     c.Model,
			ModelName: c.ModelName,
			Stepping:  c.Stepping,
			Mhz:       c.Mhz,
			CacheSize: c.CacheSize,
			Flags:     c.Flags,
			Cores:     runtime.NumCPU(),
		})
	}

	var diskList []DiskDetail
	for _, d := range diskInfo {
		usage, _ := disk.Usage(d.Mountpoint)
		diskList = append(diskList, DiskDetail{
			Device:      d.Device,
			Mountpoint:  d.Mountpoint,
			Fstype:      d.Fstype,
			Total:       usage.Total,
			Used:        usage.Used,
			Free:        usage.Free,
			UsedPercent: usage.UsedPercent,
		})
	}

	var netList []NetworkDetail
	for _, n := range netInfo {
		var addrs []string
		for _, addr := range n.Addrs {
			addrs = append(addrs, addr.Addr)
		}
		netList = append(netList, NetworkDetail{
			Name:         n.Name,
			HardwareAddr: n.HardwareAddr,
			Addrs:        addrs,
		})
	}

	return &HardwareInfo{
		CPUInfo: cpuList,
		MemoryInfo: MemoryDetail{
			Total:       memInfo.Total,
			Available:   memInfo.Available,
			Used:        memInfo.Used,
			Free:        memInfo.Free,
			UsedPercent: memInfo.UsedPercent,
		},
		DiskInfo:    diskList,
		NetworkInfo: netList,
		HostInfo: HostDetail{
			Hostname:        hostInfo.Hostname,
			OS:              hostInfo.OS,
			Platform:        hostInfo.Platform,
			PlatformFamily:  hostInfo.PlatformFamily,
			PlatformVersion: hostInfo.PlatformVersion,
			Uptime:          hostInfo.Uptime,
			Procs:           hostInfo.Procs,
			KernelVersion:   hostInfo.KernelVersion,
			KernelArch:      runtime.GOARCH,
		},
	}, nil
}

func getTopProcesses(limit int) ([]ProcessInfo, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var processList []ProcessInfo
	for _, p := range processes {
		name, _ := p.Name()
		exe, _ := p.Exe()
		cmdline, _ := p.Cmdline()
		cpuPercent, _ := p.CPUPercent()
		memPercent, _ := p.MemoryPercent()
		memInfo, _ := p.MemoryInfo()
		createTime, _ := p.CreateTime()
		status, _ := p.Status()
		username, _ := p.Username()

		processList = append(processList, ProcessInfo{
			PID:           p.Pid,
			Name:          name,
			Exe:           exe,
			Cmdline:       cmdline,
			CPUPercent:    cpuPercent,
			MemoryPercent: memPercent,
			MemoryRSS:     memInfo.RSS,
			CreateTime:    createTime,
			Status:        status,
			Username:      username,
		})
	}

	for i := 0; i < len(processList)-1; i++ {
		for j := 0; j < len(processList)-1-i; j++ {
			if processList[j].CPUPercent < processList[j+1].CPUPercent {
				processList[j], processList[j+1] = processList[j+1], processList[j]
			}
		}
	}

	if len(processList) > limit {
		processList = processList[:limit]
	}

	return processList, nil
}

func searchProcesses(query string) ([]ProcessInfo, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var result []ProcessInfo
	for _, p := range processes {
		name, _ := p.Name()
		pidStr := strconv.Itoa(int(p.Pid))
		cmdline, _ := p.Cmdline()

		if containsIgnoreCase(name, query) || pidStr == query || containsIgnoreCase(cmdline, query) {
			exe, _ := p.Exe()
			cpuPercent, _ := p.CPUPercent()
			memPercent, _ := p.MemoryPercent()
			memInfo, _ := p.MemoryInfo()
			createTime, _ := p.CreateTime()
			status, _ := p.Status()
			username, _ := p.Username()

			result = append(result, ProcessInfo{
				PID:           p.Pid,
				Name:          name,
				Exe:           exe,
				Cmdline:       cmdline,
				CPUPercent:    cpuPercent,
				MemoryPercent: memPercent,
				MemoryRSS:     memInfo.RSS,
				CreateTime:    createTime,
				Status:        status,
				Username:      username,
			})
		}
	}

	return result, nil
}

func containsIgnoreCase(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalsIgnoreCase(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

func equalsIgnoreCase(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if toLower(a[i]) != toLower(b[i]) {
			return false
		}
	}
	return true
}

func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + 32
	}
	return c
}

func getProcessByPID(pid int32) (*ProcessInfo, error) {
	p, err := process.NewProcess(pid)
	if err != nil {
		return nil, err
	}

	name, _ := p.Name()
	exe, _ := p.Exe()
	cmdline, _ := p.Cmdline()
	cpuPercent, _ := p.CPUPercent()
	memPercent, _ := p.MemoryPercent()
	memInfo, _ := p.MemoryInfo()
	createTime, _ := p.CreateTime()
	status, _ := p.Status()
	username, _ := p.Username()

	return &ProcessInfo{
		PID:           p.Pid,
		Name:          name,
		Exe:           exe,
		Cmdline:       cmdline,
		CPUPercent:    cpuPercent,
		MemoryPercent: memPercent,
		MemoryRSS:     memInfo.RSS,
		CreateTime:    createTime,
		Status:        status,
		Username:      username,
	}, nil
}

func getHistoryData(hours int) ([]HistoryRecord, error) {
	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}
	var records []HistoryRecord
	db.Where("created_at >= ?", time.Now().Add(-time.Duration(hours)*time.Hour)).Order("created_at asc").Find(&records)
	return records, nil
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	status, _ := getServerStatus()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(status)
}

func hardwareHandler(w http.ResponseWriter, r *http.Request) {
	hardware, _ := getHardwareInfo()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(hardware)
}

func processesHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query != "" {
		processes, _ := searchProcesses(query)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(processes)
		return
	}

	limit := 10
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		limit, _ = strconv.Atoi(limitStr)
	}
	processes, _ := getTopProcesses(limit)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(processes)
}

func processDetailHandler(w http.ResponseWriter, r *http.Request) {
	pidStr := r.URL.Query().Get("pid")
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		http.Error(w, "Invalid PID", http.StatusBadRequest)
		return
	}
	processInfo, err := getProcessByPID(int32(pid))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(processInfo)
}

func historyHandler(w http.ResponseWriter, r *http.Request) {
	hours := 24
	hoursStr := r.URL.Query().Get("hours")
	if hoursStr != "" {
		hours, _ = strconv.Atoi(hoursStr)
	}
	records, err := getHistoryData(hours)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(records)
}

func main() {
	http.HandleFunc("/api/status", statusHandler)
	http.HandleFunc("/api/hardware", hardwareHandler)
	http.HandleFunc("/api/processes", processesHandler)
	http.HandleFunc("/api/process", processDetailHandler)
	http.HandleFunc("/api/history", historyHandler)
	http.Handle("/", http.FileServer(http.Dir("web")))

	port := config.Server.Port
	if port == "" {
		port = ":7077"
	}
	fmt.Printf("Server started on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
