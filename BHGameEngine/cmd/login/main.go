package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/openworld-server/internal/cluster"
	"github.com/openworld-server/internal/dataclient"
	"github.com/openworld-server/internal/log"
	"github.com/openworld-server/internal/network"
	"github.com/openworld-server/internal/redis"
	"github.com/openworld-server/pkg/config"
	"github.com/openworld-server/pkg/snowflake"
)

var configPath = flag.String("config", "./config/config.toml", "config file path")

func main() {
	flag.Parse()

	log.Info("Login server initializing...")

	err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	logger := log.NewLogger("login")
	loggerConfig := config.GetLoggerConfig()
	logger.Init(loggerConfig.Path, loggerConfig.Level)
	log.Info("Login server using log config: Path=", loggerConfig.Path, ", Level=", loggerConfig.Level)

	workerID := config.GetSnowflakeWorkerID()
	log.Info("Initializing snowflake with workerID: ", workerID)
	err = snowflake.Init(int64(workerID))
	if err != nil {
		log.Fatal("Failed to init snowflake:", err)
	}

	redisConfig := config.GetRedisConfig()
	log.Info("Connecting to Redis at ", redisConfig.Addr)
	redisClient := redis.NewRedisClient(redis.RedisConfig{
		Addr:     redisConfig.Addr,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
		PoolSize: redisConfig.PoolSize,
	})
	if err := redisClient.Ping(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	log.Info("Connected to Redis successfully")

	etcdAddr := config.GetEtcdAddr()
	log.Info("Connecting to etcd at ", etcdAddr)
	cluster, err := cluster.NewCluster(etcdAddr)
	if err != nil {
		log.Fatal("Failed to connect to etcd:", err)
	}
	log.Info("Connected to etcd successfully")

	logger.SetCluster(cluster)
	go logger.StartLogServerDiscovery()

	dataClient := dataclient.NewDataClient(cluster)

	listenAddr := config.GetLoginListenAddr()
	log.Info("Registering login service at ", listenAddr)
	err = cluster.RegisterService("login", listenAddr, map[string]string{
		"type": "login",
	})
	if err != nil {
		log.Fatal("Failed to register service:", err)
	}

	log.Info("Creating LoginHandler and starting network server")
	handler := NewLoginHandler(dataClient, redisClient)
	server := network.NewServer(handler)

	go func() {
		if err := server.Start(listenAddr); err != nil {
			log.Fatal("Server failed to start:", err)
		}
	}()
	log.Info("Login server started successfully on ", listenAddr)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("Shutting down login server...")
	server.Stop()
	redisClient.Close()
	log.Info("Login server shutdown complete")
}
