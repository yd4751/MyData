package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/openworld-server/internal/cluster"
	"github.com/openworld-server/internal/db"
	"github.com/openworld-server/internal/log"
	"github.com/openworld-server/internal/network"
	"github.com/openworld-server/internal/redis"
	"github.com/openworld-server/pkg/config"
)

var configPath = flag.String("config", "./config/config.toml", "config file path")

func main() {
	flag.Parse()

	log.Info("Logic server initializing...")

	err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	logger := log.NewLogger("logic")
	loggerConfig := config.GetLoggerConfig()
	logger.Init(loggerConfig.Path, loggerConfig.Level)
	log.Info("Logic server using log config: Path=", loggerConfig.Path, ", Level=", loggerConfig.Level)

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

	mysqlConfig := config.GetMySQLConfig()
	log.Info("Connecting to MySQL at ", mysqlConfig.Host, ":", mysqlConfig.Port)
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
		log.Fatal("Failed to connect to MySQL:", err)
	}
	log.Info("Connected to MySQL successfully")

	etcdAddr := config.GetEtcdAddr()
	log.Info("Connecting to etcd at ", etcdAddr)
	cluster, err := cluster.NewCluster(etcdAddr)
	if err != nil {
		log.Fatal("Failed to connect to etcd:", err)
	}
	log.Info("Connected to etcd successfully")

	logger.SetCluster(cluster)
	go logger.StartLogServerDiscovery()

	listenAddr := config.GetLogicListenAddr()
	log.Info("Registering logic service at ", listenAddr)
	err = cluster.RegisterService("logic", listenAddr, map[string]string{
		"type": "logic",
	})
	if err != nil {
		log.Fatal("Failed to register service:", err)
	}

	log.Info("Creating LogicHandler and starting network server")
	handler := NewLogicHandler(database, redisClient)
	server := network.NewServer(handler)

	go func() {
		if err := server.Start(listenAddr); err != nil {
			log.Fatal("Server failed to start:", err)
		}
	}()
	log.Info("Logic server started successfully on ", listenAddr)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("Shutting down logic server...")
	server.Stop()
	redisClient.Close()
	database.Close()
	log.Info("Logic server shutdown complete")
}
