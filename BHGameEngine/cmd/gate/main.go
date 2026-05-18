package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/openworld-server/internal/cluster"
	"github.com/openworld-server/internal/log"
	"github.com/openworld-server/internal/network"
	"github.com/openworld-server/pkg/config"
)

var configPath = flag.String("config", "./config/config.toml", "config file path")

func main() {
	flag.Parse()

	log.Info("Gate server initializing...")

	err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	logger := log.NewLogger("gate")
	loggerConfig := config.GetLoggerConfig()
	logger.Init(loggerConfig.Path, loggerConfig.Level)
	log.Info("Gate server using log config: Path=", loggerConfig.Path, ", Level=", loggerConfig.Level)

	etcdAddr := config.GetEtcdAddr()
	log.Info("Connecting to etcd at ", etcdAddr)
	cluster, err := cluster.NewCluster(etcdAddr)
	if err != nil {
		log.Fatal("Failed to connect to etcd:", err)
	}

	logger.SetCluster(cluster)
	go logger.StartLogServerDiscovery()

	listenAddr := config.GetGateListenAddr()
	log.Info("Registering gate service at ", listenAddr)
	err = cluster.RegisterService("gate", listenAddr, map[string]string{
		"type": "gateway",
	})
	if err != nil {
		log.Fatal("Failed to register service:", err)
	}

	log.Info("Creating GateHandler and starting network server")
	handler := NewGateHandler(cluster)
	server := network.NewServer(handler)

	go func() {
		if err := server.Start(listenAddr); err != nil {
			log.Fatal("Server failed to start:", err)
		}
	}()

	log.Info("Gate server started successfully on ", listenAddr)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("Shutting down gate server...")
	server.Stop()
}
