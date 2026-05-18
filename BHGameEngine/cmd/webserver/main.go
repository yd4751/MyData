package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/openworld-server/internal/cluster"
	"github.com/openworld-server/pkg/config"
	"github.com/openworld-server/pkg/logger"
)

var configPath = flag.String("config", "./config/config.toml", "config file path")

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Service   string `json:"service"`
	Message   string `json:"message"`
}

var (
	logEntries     []LogEntry
	logEntriesMu   sync.Mutex
	logBroadcaster = make(chan LogEntry, 100)
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade to WebSocket:", err)
		return
	}
	defer ws.Close()

	gateAddr := config.GlobalConfig.Gate.ListenAddr
	if gateAddr == "" {
		gateAddr = ":7060"
	}

	tcpConn, err := net.Dial("tcp", gateAddr)
	if err != nil {
		log.Println("Failed to connect to gate server:", err)
		return
	}
	defer tcpConn.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				return
			}
			tcpConn.Write(message)
		}
	}()

	go func() {
		buf := make([]byte, 4096)
		reader := bufio.NewReader(tcpConn)
		for {
			n, err := reader.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Println("TCP read error:", err)
				}
				return
			}
			ws.WriteMessage(websocket.BinaryMessage, buf[:n])
		}
	}()

	<-done
}

func logHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var entry LogEntry
		if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		logEntriesMu.Lock()
		logEntries = append(logEntries, entry)
		if len(logEntries) > 1000 {
			logEntries = logEntries[len(logEntries)-1000:]
		}
		logEntriesMu.Unlock()

		select {
		case logBroadcaster <- entry:
		default:
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	serviceFilter := r.URL.Query().Get("service")
	levelFilter := r.URL.Query().Get("level")

	logEntriesMu.Lock()
	filtered := make([]LogEntry, 0)
	for _, entry := range logEntries {
		if serviceFilter != "" && entry.Service != serviceFilter {
			continue
		}
		if levelFilter != "" && entry.Level != levelFilter {
			continue
		}
		filtered = append(filtered, entry)
	}
	logEntriesMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filtered)
}

func logStreamHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	serviceFilter := r.URL.Query().Get("service")

	logEntriesMu.Lock()
	for _, entry := range logEntries {
		if serviceFilter == "" || entry.Service == serviceFilter {
			data, _ := json.Marshal(entry)
			w.Write([]byte("data: " + string(data) + "\n\n"))
		}
	}
	logEntriesMu.Unlock()
	flusher.Flush()

	for entry := range logBroadcaster {
		if serviceFilter != "" && entry.Service != serviceFilter {
			continue
		}
		data, _ := json.Marshal(entry)
		w.Write([]byte("data: " + string(data) + "\n\n"))
		flusher.Flush()
	}
}

func getServicesHandler(w http.ResponseWriter, r *http.Request) {
	services := map[string][]string{
		"gate":        {"gate"},
		"login":       {"login"},
		"logic":       {"logic"},
		"battle":      {"battle"},
		"gridmap":     {"gridmap"},
		"cross":       {"cross"},
		"dataservice": {"dataservice"},
		"gm":          {"gm"},
		"webserver":   {"webserver"},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(services)
}

func main() {
	flag.Parse()

	err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Fatal("Failed to load config:", err)
	}

	loggerConfig := config.GetLoggerConfig()
	logger.Init(loggerConfig.Path, loggerConfig.Level, "webserver")
	logger.Info("Web server starting...")

	listenAddr := config.GetWebServerListenAddr()

	http.Handle("/", http.FileServer(http.Dir("./web")))
	http.Handle("/admin/", http.StripPrefix("/admin/", http.FileServer(http.Dir("./web"))))
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/api/logs", logHandler)
	http.HandleFunc("/log", logHandler)
	http.HandleFunc("/api/logs/stream", logStreamHandler)
	http.HandleFunc("/api/services", getServicesHandler)

	go func() {
		if err := http.ListenAndServe(listenAddr, nil); err != nil {
			log.Fatalf("Failed to start web server: %v", err)
		}
	}()

	logger.Info("Web server running on http://localhost" + listenAddr)

	etcdAddr := config.GetEtcdAddr()
	go startClusterRegistration(etcdAddr, listenAddr)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down web server...")
}

func startClusterRegistration(etcdAddr, listenAddr string) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	var c *cluster.Cluster
	var registered bool

	for {
		if !registered {
			logger.Info("Connecting to etcd at ", etcdAddr)
			var err error
			c, err = cluster.NewCluster(etcdAddr)
			if err != nil {
				logger.Warn("Failed to connect to etcd, retrying in 10s:", err)
				<-ticker.C
				continue
			}
			logger.Info("Connected to etcd successfully")

			logger.Info("Registering webserver service at ", listenAddr)
			err = c.RegisterService("webserver", listenAddr, map[string]string{
				"type": "webserver",
			})
			if err != nil {
				logger.Warn("Failed to register service, retrying in 10s:", err)
				<-ticker.C
				continue
			}
			logger.Info("Webserver service registered successfully")
			registered = true
			logger.Info("Cluster registration successful, stopping reconnection attempts")
		}

		<-ticker.C
	}
}
