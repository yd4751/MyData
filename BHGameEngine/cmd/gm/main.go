package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/openworld-server/internal/cluster"
	"github.com/openworld-server/internal/db"
	"github.com/openworld-server/internal/network"
	"github.com/openworld-server/internal/redis"
	"github.com/openworld-server/pkg/config"
	"github.com/openworld-server/pkg/logger"
)

var configPath = flag.String("config", "./config/config.toml", "config file path")

type GMHandler struct{}

func (h *GMHandler) Handle(msg *network.Message) {
	logger.Info("GM server received message from ", msg.Session.RemoteAddr())
}

func main() {
	flag.Parse()

	err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Fatal("Failed to load config:", err)
	}

	loggerConfig := config.GetLoggerConfig()
	logger.Init(loggerConfig.Path, loggerConfig.Level, "gm")
	logger.Info("GM server starting...")

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

	mysqlConfig := config.GetMySQLConfig()
	database, err := db.NewDatabase(db.DBConfig{
		Host:         mysqlConfig.Host,
		Port:         mysqlConfig.Port,
		User:         mysqlConfig.User,
		Password:     mysqlConfig.Password,
		DBName:       mysqlConfig.DBName,
		MaxOpenConns: mysqlConfig.MaxOpenConns,
		MaxIdleConns: mysqlConfig.MaxIdleConns,
	})
	if err != nil {
		logger.Fatal("Failed to connect to MySQL:", err)
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

	listenAddr := config.GetGMListenAddr()
	err = cluster.RegisterService("gm", listenAddr, map[string]string{
		"type": "gm",
	})
	if err != nil {
		logger.Fatal("Failed to register service:", err)
	}

	logger.Info("GM server starting network server on ", listenAddr)
	handler := &GMHandler{}
	server := network.NewServer(handler)

	go func() {
		if err := server.Start(listenAddr); err != nil {
			logger.Fatal("Server failed to start:", err)
		}
	}()
	logger.Info("GM server network server started")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down GM server...")
	server.Stop()
	redisClient.Close()
	database.Close()
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
