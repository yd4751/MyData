package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/openworld-server/internal/cluster"
	"github.com/openworld-server/internal/network"
	"github.com/openworld-server/internal/redis"
	"github.com/openworld-server/pkg/config"
	"github.com/openworld-server/pkg/logger"
)

var configPath = flag.String("config", "./config/config.toml", "config file path")

type CrossHandler struct{}

func (h *CrossHandler) Handle(msg *network.Message) {
	logger.Info("Cross server received message from ", msg.Session.RemoteAddr())
}

func main() {
	flag.Parse()

	err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Fatal("Failed to load config:", err)
	}

	loggerConfig := config.GetLoggerConfig()
	logger.Init(loggerConfig.Path, loggerConfig.Level, "cross")
	logger.Info("Cross server starting...")

	redisConfig := config.GetRedisConfig()
	redisClient := redis.NewRedisClient(redis.RedisConfig{
		Addr:     redisConfig.Addr,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
		PoolSize: redisConfig.PoolSize,
	})
	if err := redisClient.Ping(); err != nil {
		logger.Fatal("Failed to connect to Redis:", err)
	}

	etcdAddr := config.GetEtcdAddr()
	cluster, err := cluster.NewCluster(etcdAddr)
	if err != nil {
		logger.Fatal("Failed to connect to etcd:", err)
	}

	go func() {
		time.Sleep(2 * time.Second)
		startLogServerDiscovery(cluster)
	}()

	listenAddr := config.GetCrossListenAddr()
	err = cluster.RegisterService("cross", listenAddr, map[string]string{
		"type": "cross",
	})
	if err != nil {
		logger.Fatal("Failed to register service:", err)
	}

	logger.Info("Cross server starting network server on ", listenAddr)
	handler := &CrossHandler{}
	server := network.NewServer(handler)

	go func() {
		if err := server.Start(listenAddr); err != nil {
			logger.Fatal("Server failed to start:", err)
		}
	}()
	logger.Info("Cross server network server started")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down cross server...")
	server.Stop()
	redisClient.Close()
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
