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
	"strings"
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

func getGateAddrFromEtcd() (string, error) {
	etcdAddr := config.GetEtcdAddr()
	c, err := cluster.NewCluster(etcdAddr)
	if err != nil {
		return "", err
	}
	service, err := c.GetRandomService("gate")
	if err != nil {
		return "", err
	}
	return service.Addr, nil
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade to WebSocket:", err)
		return
	}
	defer ws.Close()

	wsClosed := make(chan struct{})
	wsErr := make(chan error, 1)

	var tcpConn net.Conn
	connect := func() error {
		gateAddr, err := getGateAddrFromEtcd()
		if err != nil {
			return err
		}
		if strings.HasPrefix(gateAddr, ":") {
			gateAddr = "localhost" + gateAddr
		}
		tcpConn, err = net.Dial("tcp", gateAddr)
		if err != nil {
			return err
		}
		log.Println("Connected to gate server:", gateAddr)
		return nil
	}

	err = connect()
	if err != nil {
		log.Println("Failed to connect to gate server:", err)
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				err = connect()
				if err == nil {
					log.Println("Successfully connected to gate server after retry")
					goto connected
				}
				log.Println("Retry failed, waiting 10s:", err)
			case <-wsClosed:
				log.Println("WebSocket closed, aborting connection attempts")
				return
			}
		}
	}
connected:

	go func() {
		defer close(wsClosed)
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				wsErr <- err
				return
			}
			if tcpConn != nil {
				_, writeErr := tcpConn.Write(message)
				if writeErr != nil {
					log.Println("TCP write error:", writeErr)
				}
			}
		}
	}()

	buf := make([]byte, 4096)
	for {
		select {
		case err := <-wsErr:
			log.Println("WebSocket error:", err)
			if tcpConn != nil {
				tcpConn.Close()
			}
			return
		default:
		}

		if tcpConn == nil {
			ticker := time.NewTicker(10 * time.Second)
			for {
				select {
				case <-ticker.C:
					err := connect()
					if err == nil {
						log.Println("Reconnected to gate server")
						ticker.Stop()
						goto reconnected
					}
					log.Println("Reconnection failed, waiting 10s:", err)
				case err := <-wsErr:
					log.Println("WebSocket closed during reconnection:", err)
					ticker.Stop()
					return
				}
			}
		reconnected:
		}

		reader := bufio.NewReader(tcpConn)
		n, err := reader.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Println("TCP read error, reconnecting:", err)
			}
			tcpConn.Close()
			tcpConn = nil
			continue
		}
		ws.WriteMessage(websocket.BinaryMessage, buf[:n])
	}
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
