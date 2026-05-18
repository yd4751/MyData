package main

import (
	"flag"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/openworld-server/internal/cluster"
	"github.com/openworld-server/internal/network"
	"github.com/openworld-server/pkg/config"
	"github.com/openworld-server/pkg/logger"
)

var configPath = flag.String("config", "./config/config.toml", "config file path")
var gridID = flag.Int("grid", 1, "gridmap ID")

func main() {
	flag.Parse()

	err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Fatal("Failed to load config:", err)
	}

	loggerConfig := config.GetLoggerConfig()
	logger.Init(loggerConfig.Path, loggerConfig.Level, "gridmap")
	logger.Info("GridMap server starting...")

	etcdAddr := config.GetEtcdAddr()
	cluster, err := cluster.NewCluster(etcdAddr)
	if err != nil {
		logger.Fatal("Failed to connect to etcd:", err)
	}

	go func() {
		time.Sleep(2 * time.Second)
		startLogServerDiscovery(cluster)
	}()

	listenAddr := config.GetGridMapListenAddr()
	serviceName := "gridmap-" + strconv.FormatInt(int64(*gridID), 10)

	err = cluster.RegisterService(serviceName, listenAddr, map[string]string{
		"type":    "gridmap",
		"grid_id": strconv.FormatInt(int64(*gridID), 10),
	})
	if err != nil {
		logger.Fatal("Failed to register service:", err)
	}

	logger.Info("GridMap server ", *gridID, " starting network server on ", listenAddr)

	handler := NewGridMapHandler(cluster, *gridID)
	server := network.NewServer(handler)

	go func() {
		if err := server.Start(listenAddr); err != nil {
			logger.Fatal("Server failed to start:", err)
		}
	}()
	logger.Info("GridMap server network server started")

	handler.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down gridmap server...")
	server.Stop()
	handler.Stop()
}

func startLogServerDiscovery(c *cluster.Cluster) {
	for {
		webService, err := c.GetRandomService("webserver")
		if err == nil && webService != nil {
			currentAddr := "http://" + webService.Addr
			logger.SetLogServer(currentAddr)
			logger.Info("Log server set to: ", webService.Addr)

			for {
				time.Sleep(5 * time.Second)
				webService, err = c.GetRandomService("webserver")
				if err != nil || webService == nil {
					logger.Warn("Log server disconnected, restarting discovery...")
					break
				}
				if "http://"+webService.Addr != currentAddr {
					currentAddr = "http://" + webService.Addr
					logger.SetLogServer(currentAddr)
					logger.Info("Log server changed to: ", webService.Addr)
				}
			}
		} else {
			logger.Warn("Failed to discover webserver for log forwarding, retrying in 5s...")
			time.Sleep(5 * time.Second)
		}
	}
}
