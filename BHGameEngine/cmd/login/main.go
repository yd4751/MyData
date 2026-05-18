package main

import (
	"encoding/json"
	"flag"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/openworld-server/internal/cluster"
	"github.com/openworld-server/internal/dataclient"
	"github.com/openworld-server/internal/log"
	"github.com/openworld-server/internal/msg"
	"github.com/openworld-server/internal/network"
	"github.com/openworld-server/internal/redis"
	"github.com/openworld-server/pkg/config"
	"github.com/openworld-server/pkg/snowflake"
)

var configPath = flag.String("config", "./config/config.toml", "config file path")

type LoginHandler struct {
	dataClient  *dataclient.DataClient
	redisClient *redis.RedisClient
}

func (h *LoginHandler) Handle(msgObj *network.Message) {
	log.Info("Login server received message from ", msgObj.Session.RemoteAddr(), " - MsgID:", msgObj.ID, "(", msg.GetMsgName(msgObj.ID), "), NodeType:", msgObj.NodeType, "(", msgObj.NodeType.String(), ")")

	switch msgObj.ID {
	case msg.MSG_LOGIN_REQ:
		h.handleLogin(msgObj)
	case msg.MSG_REGISTER_REQ:
		h.handleRegister(msgObj)
	case msg.MSG_LOGOUT_REQ:
		h.handleLogout(msgObj)
	case msg.MSG_PLAYER_INFO_REQ:
		h.handlePlayerInfoRequest(msgObj)
	default:
		log.Warn("Unknown message ID:", msgObj.ID, "(", msg.GetMsgName(msgObj.ID), ")")
	}
}

func (h *LoginHandler) handleLogin(msgObj *network.Message) {
	var req msg.LoginRequest
	err := json.Unmarshal(msgObj.Data, &req)
	if err != nil {
		log.Error("Failed to parse login request:", err)
		h.sendError(msgObj.Session, msg.MSG_LOGIN_RES, "Invalid request")
		return
	}

	log.Info("Login request: Account=", req.Account, ", DeviceID=", req.DeviceID)

	account, err := h.dataClient.GetAccount(req.Account)
	if err != nil {
		log.Error("Failed to get account:", err)
		h.sendError(msgObj.Session, msg.MSG_LOGIN_RES, "Account not found")
		return
	}

	if account == nil {
		log.Warn("Account not found: ", req.Account)
		h.sendError(msgObj.Session, msg.MSG_LOGIN_RES, "Account not found")
		return
	}

	if account.Password != req.Password+account.Salt {
		log.Warn("Invalid password for account: ", req.Account)
		h.sendError(msgObj.Session, msg.MSG_LOGIN_RES, "Invalid password")
		return
	}

	sessionID := generateSessionID()
	err = h.redisClient.SetSession(sessionID, account.ID, 24*time.Hour)
	if err != nil {
		log.Error("Failed to set session:", err)
		h.sendError(msgObj.Session, msg.MSG_LOGIN_RES, "Internal error")
		return
	}

	log.Info("Session created: SessionID=", sessionID, ", AccountID=", account.ID)

	player, err := h.dataClient.GetPlayerByAccountID(account.ID)
	if err != nil {
		log.Error("Failed to get player:", err)
		h.sendError(msgObj.Session, msg.MSG_LOGIN_RES, "Internal error")
		return
	}

	resp := msg.LoginResponse{
		Result:     0,
		SessionID:  sessionID,
		PlayerID:   player.ID,
		PlayerName: player.Name,
		Level:      player.Level,
		Health:     player.Health,
	}

	log.Info("Login successful: Account=", req.Account, ", PlayerID=", player.ID)
	h.sendResponse(msgObj.Session, msg.MSG_LOGIN_RES, resp)
}

func (h *LoginHandler) handleRegister(msgObj *network.Message) {
	var req msg.RegisterRequest
	err := json.Unmarshal(msgObj.Data, &req)
	if err != nil {
		log.Error("Failed to parse register request:", err)
		h.sendRegisterError(msgObj.Session, "Invalid request")
		return
	}

	log.Info("Register request: Account=", req.Account)

	exists, err := h.dataClient.AccountExists(req.Account)
	if err != nil {
		log.Error("Failed to check account:", err)
		h.sendRegisterError(msgObj.Session, "Internal error")
		return
	}

	if exists {
		log.Warn("Account already exists: ", req.Account)
		h.sendRegisterError(msgObj.Session, "Account already exists")
		return
	}

	err = h.dataClient.CreateAccount(req.Account, req.Password)
	if err != nil {
		log.Error("Failed to create account:", err)
		h.sendRegisterError(msgObj.Session, "Failed to create account")
		return
	}

	account, err := h.dataClient.GetAccount(req.Account)
	if err != nil {
		log.Error("Failed to get new account:", err)
		h.sendRegisterError(msgObj.Session, "Internal error")
		return
	}

	playerID := snowflake.GenerateID()
	err = h.dataClient.CreatePlayer(playerID, account.ID, req.Account)
	if err != nil {
		log.Error("Failed to create player:", err)
		h.sendRegisterError(msgObj.Session, "Failed to create player")
		return
	}

	log.Info("Register successful: Account=", req.Account, ", PlayerID=", playerID)

	resp := msg.RegisterResponse{
		Result:  0,
		Message: "Register success",
	}

	h.sendRegisterResponse(msgObj.Session, resp)
}

func (h *LoginHandler) handleLogout(msgObj *network.Message) {
	var req msg.LogoutRequest
	err := json.Unmarshal(msgObj.Data, &req)
	if err != nil {
		log.Error("Failed to parse logout request:", err)
		return
	}

	log.Info("Logout request: SessionID=", req.SessionID)

	err = h.redisClient.DeleteSession(req.SessionID)
	if err != nil {
		log.Error("Failed to delete session:", err)
		return
	}

	log.Info("Logout successful: SessionID=", req.SessionID)
}

func (h *LoginHandler) handlePlayerInfoRequest(msgObj *network.Message) {
	log.Info("Handling player info request")

	var req msg.PlayerInfoRequest
	err := json.Unmarshal(msgObj.Data, &req)
	if err != nil {
		log.Error("Failed to unmarshal player info request:", err)
		h.sendPlayerInfoError(msgObj.Session, "Invalid request")
		return
	}

	playerIDStr, err := h.redisClient.GetSession(req.SessionID)
	if err != nil {
		log.Error("Failed to get player ID from session:", err)
		h.sendPlayerInfoError(msgObj.Session, "Invalid session")
		return
	}

	if playerIDStr == "" {
		log.Warn("Session not found: ", req.SessionID)
		h.sendPlayerInfoError(msgObj.Session, "Invalid session")
		return
	}

	playerID, err := strconv.ParseInt(playerIDStr, 10, 64)
	if err != nil {
		log.Error("Failed to parse player ID:", err)
		h.sendPlayerInfoError(msgObj.Session, "Internal error")
		return
	}

	player, err := h.dataClient.GetPlayerByID(playerID)
	if err != nil {
		log.Error("Failed to get player:", err)
		h.sendPlayerInfoError(msgObj.Session, "Internal error")
		return
	}

	if player == nil {
		log.Warn("Player not found: ", playerID)
		h.sendPlayerInfoError(msgObj.Session, "Player not found")
		return
	}

	resp := msg.PlayerInfoResponse{
		Result:     0,
		Message:    "success",
		PlayerID:   player.ID,
		PlayerName: player.Name,
		Level:      player.Level,
		Exp:        player.Exp,
		Health:     player.Health,
		MaxHealth:  player.MaxHealth,
		Mana:       player.Mana,
		MaxMana:    player.MaxMana,
		PositionX:  player.PosX,
		PositionY:  player.PosY,
	}

	log.Info("Player info request successful: PlayerID=", player.ID)
	h.sendPlayerInfoResponse(msgObj.Session, resp)
}

func (h *LoginHandler) sendPlayerInfoError(conn net.Conn, message string) {
	resp := msg.PlayerInfoResponse{
		Result:  1,
		Message: message,
	}
	h.sendResponse(conn, msg.MSG_PLAYER_INFO_RES, resp)
}

func (h *LoginHandler) sendPlayerInfoResponse(conn net.Conn, resp msg.PlayerInfoResponse) {
	data, err := json.Marshal(resp)
	if err != nil {
		log.Error("Failed to marshal player info response:", err)
		return
	}
	err = network.SendRawMessage(conn, msg.MSG_PLAYER_INFO_RES, msg.NodeTypeData, data)
	if err != nil {
		log.Error("Failed to send player info response:", err)
	}
}

func (h *LoginHandler) sendError(conn net.Conn, msgID uint32, message string) {
	type errorResponse struct {
		Result  int    `json:"result"`
		Message string `json:"message"`
	}
	resp := errorResponse{
		Result:  1,
		Message: message,
	}
	data, err := json.Marshal(resp)
	if err != nil {
		log.Error("Failed to marshal error response:", err)
		return
	}
	err = network.SendRawMessage(conn, msgID, msg.NodeTypeData, data)
	if err != nil {
		log.Error("Failed to send error response:", err)
	}
}

func (h *LoginHandler) sendResponse(conn net.Conn, msgID uint32, resp interface{}) {
	data, err := json.Marshal(resp)
	if err != nil {
		log.Error("Failed to marshal response:", err)
		return
	}
	err = network.SendRawMessage(conn, msgID, msg.NodeTypeData, data)
	if err != nil {
		log.Error("Failed to send response:", err)
	}
}

func (h *LoginHandler) sendRegisterError(conn net.Conn, message string) {
	resp := msg.RegisterResponse{
		Result:  1,
		Message: message,
	}
	h.sendRegisterResponse(conn, resp)
}

func (h *LoginHandler) sendRegisterResponse(conn net.Conn, resp msg.RegisterResponse) {
	data, err := json.Marshal(resp)
	if err != nil {
		log.Error("Failed to marshal register response:", err)
		return
	}
	err = network.SendRawMessage(conn, msg.MSG_REGISTER_RES, msg.NodeTypeData, data)
	if err != nil {
		log.Error("Failed to send register response:", err)
	}
}

func generateSessionID() string {
	return snowflake.GenerateIDString()
}

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
	handler := &LoginHandler{dataClient: dataClient, redisClient: redisClient}
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
