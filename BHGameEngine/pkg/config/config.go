package config

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Gate struct {
		ListenAddr     string `toml:"listen_addr"`
		MaxConnections int    `toml:"max_connections"`
		PingInterval   int    `toml:"ping_interval"`
		PingTimeout    int    `toml:"ping_timeout"`
		Encrypt        bool   `toml:"encrypt"`
	} `toml:"gate"`

	Login struct {
		ListenAddr    string `toml:"listen_addr"`
		SessionExpire int    `toml:"session_expire"`
	} `toml:"login"`

	Etcd struct {
		Addr string `toml:"addr"`
	} `toml:"etcd"`

	GridMap struct {
		ListenAddr         string `toml:"listen_addr"`
		ChunkSize          int    `toml:"chunk_size"`
		ViewDistance       int    `toml:"view_distance"`
		MaxPlayersPerChunk int    `toml:"max_players_per_chunk"`
		MapDataDir         string `toml:"map_data_dir"`
		GridCount          int    `toml:"grid_count"`
	} `toml:"gridmap"`

	Logic struct {
		ListenAddr string `toml:"listen_addr"`
		MaxPlayers int    `toml:"max_players"`
	} `toml:"logic"`

	Battle struct {
		ListenAddr string `toml:"listen_addr"`
		MaxBattles int    `toml:"max_battles"`
	} `toml:"battle"`

	Cross struct {
		ListenAddr string `toml:"listen_addr"`
	} `toml:"cross"`

	DataService struct {
		ListenAddr    string `toml:"listen_addr"`
		BatchSize     int    `toml:"batch_size"`
		FlushInterval int    `toml:"flush_interval"`
	} `toml:"dataservice"`

	GM struct {
		ListenAddr string `toml:"listen_addr"`
	} `toml:"gm"`

	MySQL struct {
		Host         string `toml:"host"`
		Port         int    `toml:"port"`
		User         string `toml:"user"`
		Password     string `toml:"password"`
		DBName       string `toml:"db_name"`
		MaxOpenConns int    `toml:"max_open_conns"`
		MaxIdleConns int    `toml:"max_idle_conns"`
	} `toml:"mysql"`

	Redis struct {
		Addr     string `toml:"addr"`
		Password string `toml:"password"`
		DB       int    `toml:"db"`
		PoolSize int    `toml:"pool_size"`
	} `toml:"redis"`

	Logger struct {
		Level string `toml:"level"`
		Path  string `toml:"path"`
	} `toml:"logger"`

	Snowflake struct {
		WorkerID int `toml:"worker_id"`
	} `toml:"snowflake"`

	WebServer struct {
		ListenAddr string `toml:"listen_addr"`
	} `toml:"webserver"`
}

var GlobalConfig *Config

func LoadConfig(path string) error {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return err
	}
	GlobalConfig = &cfg
	return nil
}

func GetGateListenAddr() string {
	if GlobalConfig == nil || GlobalConfig.Gate.ListenAddr == "" {
		return ":7060"
	}
	return GlobalConfig.Gate.ListenAddr
}

func GetLoginListenAddr() string {
	if GlobalConfig == nil || GlobalConfig.Login.ListenAddr == "" {
		return ":8081"
	}
	return GlobalConfig.Login.ListenAddr
}

func GetGridMapListenAddr() string {
	if GlobalConfig == nil || GlobalConfig.GridMap.ListenAddr == "" {
		return ":8083"
	}
	return GlobalConfig.GridMap.ListenAddr
}

func GetGridMapChunkSize() int {
	if GlobalConfig == nil || GlobalConfig.GridMap.ChunkSize == 0 {
		return 256
	}
	return GlobalConfig.GridMap.ChunkSize
}

func GetGridMapViewDistance() int {
	if GlobalConfig == nil || GlobalConfig.GridMap.ViewDistance == 0 {
		return 3
	}
	return GlobalConfig.GridMap.ViewDistance
}

func GetGridMapMapDataDir() string {
	if GlobalConfig == nil || GlobalConfig.GridMap.MapDataDir == "" {
		return "./data/maps"
	}
	return GlobalConfig.GridMap.MapDataDir
}

func GetGridMapGridCount() int {
	if GlobalConfig == nil || GlobalConfig.GridMap.GridCount == 0 {
		return 9
	}
	return GlobalConfig.GridMap.GridCount
}

func GetLogicListenAddr() string {
	if GlobalConfig == nil || GlobalConfig.Logic.ListenAddr == "" {
		return ":8084"
	}
	return GlobalConfig.Logic.ListenAddr
}

func GetBattleListenAddr() string {
	if GlobalConfig == nil || GlobalConfig.Battle.ListenAddr == "" {
		return ":8085"
	}
	return GlobalConfig.Battle.ListenAddr
}

func GetCrossListenAddr() string {
	if GlobalConfig == nil || GlobalConfig.Cross.ListenAddr == "" {
		return ":8086"
	}
	return GlobalConfig.Cross.ListenAddr
}

func GetDataServiceListenAddr() string {
	if GlobalConfig == nil || GlobalConfig.DataService.ListenAddr == "" {
		return ":8087"
	}
	return GlobalConfig.DataService.ListenAddr
}

func GetGMListenAddr() string {
	if GlobalConfig == nil || GlobalConfig.GM.ListenAddr == "" {
		return ":8088"
	}
	return GlobalConfig.GM.ListenAddr
}

func GetWebServerListenAddr() string {
	if GlobalConfig == nil || GlobalConfig.WebServer.ListenAddr == "" {
		return ":7062"
	}
	return GlobalConfig.WebServer.ListenAddr
}

func GetEtcdAddr() string {
	if GlobalConfig == nil || GlobalConfig.Etcd.Addr == "" {
		return "localhost:2379"
	}
	return GlobalConfig.Etcd.Addr
}

func GetMySQLConfig() MySQLConfig {
	if GlobalConfig == nil {
		return MySQLConfig{}
	}
	return MySQLConfig{
		Host:         GlobalConfig.MySQL.Host,
		Port:         GlobalConfig.MySQL.Port,
		User:         GlobalConfig.MySQL.User,
		Password:     GlobalConfig.MySQL.Password,
		DBName:       GlobalConfig.MySQL.DBName,
		MaxOpenConns: GlobalConfig.MySQL.MaxOpenConns,
		MaxIdleConns: GlobalConfig.MySQL.MaxIdleConns,
	}
}

func GetRedisConfig() RedisConfig {
	if GlobalConfig == nil {
		return RedisConfig{}
	}
	return RedisConfig{
		Addr:     GlobalConfig.Redis.Addr,
		Password: GlobalConfig.Redis.Password,
		DB:       GlobalConfig.Redis.DB,
		PoolSize: GlobalConfig.Redis.PoolSize,
	}
}

func GetSnowflakeWorkerID() int {
	if GlobalConfig == nil {
		return 1
	}
	return GlobalConfig.Snowflake.WorkerID
}

func GetLoggerConfig() LoggerConfig {
	if GlobalConfig == nil {
		return LoggerConfig{}
	}
	return LoggerConfig{
		Level: GlobalConfig.Logger.Level,
		Path:  GlobalConfig.Logger.Path,
	}
}

type MySQLConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	DBName       string
	MaxOpenConns int
	MaxIdleConns int
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
}

type LoggerConfig struct {
	Level string
	Path  string
}
