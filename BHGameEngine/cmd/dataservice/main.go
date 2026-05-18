package main

import (
	"encoding/json"
	"flag"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/openworld-server/internal/db"
	"github.com/openworld-server/internal/log"
	"github.com/openworld-server/internal/msg"
	"github.com/openworld-server/internal/network"
	"github.com/openworld-server/pkg/config"
)

var configPath = flag.String("config", "./config/config.toml", "config file path")

type DataServiceHandler struct {
	database *db.Database
}

func (h *DataServiceHandler) Handle(msgObj *network.Message) {
	log.Info("Data service received message from ", msgObj.Session.RemoteAddr(), " - MsgID:", msgObj.ID, "(", msg.GetMsgName(msgObj.ID), ")")

	switch msgObj.ID {
	case msg.MSG_DB_ACCOUNT_GET:
		h.handleAccountGet(msgObj)
	case msg.MSG_DB_ACCOUNT_CREATE:
		h.handleAccountCreate(msgObj)
	case msg.MSG_DB_ACCOUNT_EXISTS:
		h.handleAccountExists(msgObj)
	case msg.MSG_DB_PLAYER_GET:
		h.handlePlayerGet(msgObj)
	case msg.MSG_DB_PLAYER_CREATE:
		h.handlePlayerCreate(msgObj)
	case msg.MSG_DB_PLAYER_UPDATE:
		h.handlePlayerUpdate(msgObj)
	case msg.MSG_DB_INVENTORY_GET:
		h.handleInventoryGet(msgObj)
	case msg.MSG_DB_ITEM_GET:
		h.handleItemGet(msgObj)
	default:
		log.Warn("Unknown message ID:", msgObj.ID)
	}
}

func (h *DataServiceHandler) sendResponse(conn net.Conn, msgID uint32, result int, message string, data interface{}) {
	resp := msg.DBResponse{
		Result:  result,
		Message: message,
		Data:    data,
	}
	jsonData, err := json.Marshal(resp)
	if err != nil {
		log.Error("Failed to marshal response:", err)
		return
	}
	err = network.SendRawMessage(conn, msgID, msg.NodeTypeData, jsonData)
	if err != nil {
		log.Error("Failed to send response:", err)
	}
}

func (h *DataServiceHandler) handleAccountGet(msgObj *network.Message) {
	var req map[string]string
	err := json.Unmarshal(msgObj.Data, &req)
	if err != nil {
		log.Error("Failed to parse request:", err)
		h.sendResponse(msgObj.Session, msg.MSG_DB_ACCOUNT_GET, -1, "Invalid request", nil)
		return
	}

	account, ok := req["account"]
	if !ok {
		h.sendResponse(msgObj.Session, msg.MSG_DB_ACCOUNT_GET, -1, "Missing account", nil)
		return
	}

	data, err := h.database.GetAccountByAccount(account)
	if err != nil {
		log.Error("Failed to get account:", err)
		h.sendResponse(msgObj.Session, msg.MSG_DB_ACCOUNT_GET, -1, "Internal error", nil)
		return
	}

	if data == nil {
		h.sendResponse(msgObj.Session, msg.MSG_DB_ACCOUNT_GET, -1, "Account not found", nil)
		return
	}

	accountData := &msg.AccountData{
		ID:       data.ID,
		Account:  data.Account,
		Password: data.Password,
		Salt:     data.Salt,
	}

	h.sendResponse(msgObj.Session, msg.MSG_DB_ACCOUNT_GET, 0, "success", accountData)
}

func (h *DataServiceHandler) handleAccountCreate(msgObj *network.Message) {
	var req map[string]string
	err := json.Unmarshal(msgObj.Data, &req)
	if err != nil {
		log.Error("Failed to parse request:", err)
		h.sendResponse(msgObj.Session, msg.MSG_DB_ACCOUNT_CREATE, -1, "Invalid request", nil)
		return
	}

	account, ok := req["account"]
	if !ok {
		h.sendResponse(msgObj.Session, msg.MSG_DB_ACCOUNT_CREATE, -1, "Missing account", nil)
		return
	}

	password, ok := req["password"]
	if !ok {
		h.sendResponse(msgObj.Session, msg.MSG_DB_ACCOUNT_CREATE, -1, "Missing password", nil)
		return
	}

	err = h.database.CreateAccount(account, password)
	if err != nil {
		log.Error("Failed to create account:", err)
		h.sendResponse(msgObj.Session, msg.MSG_DB_ACCOUNT_CREATE, -1, "Failed to create account", nil)
		return
	}

	h.sendResponse(msgObj.Session, msg.MSG_DB_ACCOUNT_CREATE, 0, "success", nil)
}

func (h *DataServiceHandler) handleAccountExists(msgObj *network.Message) {
	var req map[string]string
	err := json.Unmarshal(msgObj.Data, &req)
	if err != nil {
		log.Error("Failed to parse request:", err)
		h.sendResponse(msgObj.Session, msg.MSG_DB_ACCOUNT_EXISTS, 0, "Invalid request", nil)
		return
	}

	account, ok := req["account"]
	if !ok {
		h.sendResponse(msgObj.Session, msg.MSG_DB_ACCOUNT_EXISTS, 0, "Missing account", nil)
		return
	}

	exists, err := h.database.AccountExists(account)
	if err != nil {
		log.Error("Failed to check account:", err)
		h.sendResponse(msgObj.Session, msg.MSG_DB_ACCOUNT_EXISTS, 0, "Internal error", nil)
		return
	}

	result := 0
	if exists {
		result = 1
	}
	h.sendResponse(msgObj.Session, msg.MSG_DB_ACCOUNT_EXISTS, result, "success", nil)
}

func (h *DataServiceHandler) handlePlayerGet(msgObj *network.Message) {
	var req map[string]interface{}
	err := json.Unmarshal(msgObj.Data, &req)
	if err != nil {
		log.Error("Failed to parse request:", err)
		h.sendResponse(msgObj.Session, msg.MSG_DB_PLAYER_GET, -1, "Invalid request", nil)
		return
	}

	var playerData *db.PlayerData

	if playerID, ok := req["player_id"]; ok {
		data, err := h.database.GetPlayerByID(int64(playerID.(float64)))
		if err != nil {
			log.Error("Failed to get player:", err)
			h.sendResponse(msgObj.Session, msg.MSG_DB_PLAYER_GET, -1, "Internal error", nil)
			return
		}
		playerData = data
	} else if accountID, ok := req["account_id"]; ok {
		data, err := h.database.GetPlayerByAccountID(int64(accountID.(float64)))
		if err != nil {
			log.Error("Failed to get player:", err)
			h.sendResponse(msgObj.Session, msg.MSG_DB_PLAYER_GET, -1, "Internal error", nil)
			return
		}
		playerData = data
	} else {
		h.sendResponse(msgObj.Session, msg.MSG_DB_PLAYER_GET, -1, "Missing player_id or account_id", nil)
		return
	}

	if playerData == nil {
		h.sendResponse(msgObj.Session, msg.MSG_DB_PLAYER_GET, -1, "Player not found", nil)
		return
	}

	result := &msg.PlayerData{
		ID:        playerData.ID,
		Name:      playerData.Name,
		AccountID: playerData.AccountID,
		Level:     playerData.Level,
		Exp:       playerData.Exp,
		PosX:      playerData.PosX,
		PosY:      playerData.PosY,
		Health:    playerData.Health,
		MaxHealth: playerData.MaxHealth,
		Mana:      playerData.Mana,
		MaxMana:   playerData.MaxMana,
	}

	h.sendResponse(msgObj.Session, msg.MSG_DB_PLAYER_GET, 0, "success", result)
}

func (h *DataServiceHandler) handlePlayerCreate(msgObj *network.Message) {
	var req map[string]interface{}
	err := json.Unmarshal(msgObj.Data, &req)
	if err != nil {
		log.Error("Failed to parse request:", err)
		h.sendResponse(msgObj.Session, msg.MSG_DB_PLAYER_CREATE, -1, "Invalid request", nil)
		return
	}

	playerID, ok := req["player_id"]
	if !ok {
		h.sendResponse(msgObj.Session, msg.MSG_DB_PLAYER_CREATE, -1, "Missing player_id", nil)
		return
	}

	accountID, ok := req["account_id"]
	if !ok {
		h.sendResponse(msgObj.Session, msg.MSG_DB_PLAYER_CREATE, -1, "Missing account_id", nil)
		return
	}

	name, ok := req["name"].(string)
	if !ok {
		h.sendResponse(msgObj.Session, msg.MSG_DB_PLAYER_CREATE, -1, "Missing name", nil)
		return
	}

	err = h.database.CreatePlayer(int64(playerID.(float64)), int64(accountID.(float64)), name)
	if err != nil {
		log.Error("Failed to create player:", err)
		h.sendResponse(msgObj.Session, msg.MSG_DB_PLAYER_CREATE, -1, "Failed to create player", nil)
		return
	}

	h.sendResponse(msgObj.Session, msg.MSG_DB_PLAYER_CREATE, 0, "success", nil)
}

func (h *DataServiceHandler) handlePlayerUpdate(msgObj *network.Message) {
	var req map[string]interface{}
	err := json.Unmarshal(msgObj.Data, &req)
	if err != nil {
		log.Error("Failed to parse request:", err)
		h.sendResponse(msgObj.Session, msg.MSG_DB_PLAYER_UPDATE, -1, "Invalid request", nil)
		return
	}

	playerID, ok := req["player_id"]
	if !ok {
		h.sendResponse(msgObj.Session, msg.MSG_DB_PLAYER_UPDATE, -1, "Missing player_id", nil)
		return
	}

	if posX, ok := req["pos_x"]; ok {
		posY := req["pos_y"].(float64)
		err = h.database.UpdatePlayerPosition(int64(playerID.(float64)), posX.(float64), posY)
		if err != nil {
			log.Error("Failed to update player position:", err)
			h.sendResponse(msgObj.Session, msg.MSG_DB_PLAYER_UPDATE, -1, "Failed to update position", nil)
			return
		}
	} else {
		data, err := h.database.GetPlayerByID(int64(playerID.(float64)))
		if err != nil || data == nil {
			log.Error("Failed to get player:", err)
			h.sendResponse(msgObj.Session, msg.MSG_DB_PLAYER_UPDATE, -1, "Player not found", nil)
			return
		}

		if name, ok := req["name"]; ok {
			data.Name = name.(string)
		}
		if level, ok := req["level"]; ok {
			data.Level = int32(level.(float64))
		}
		if exp, ok := req["exp"]; ok {
			data.Exp = int64(exp.(float64))
		}
		if health, ok := req["health"]; ok {
			data.Health = int32(health.(float64))
		}
		if mana, ok := req["mana"]; ok {
			data.Mana = int32(mana.(float64))
		}

		err = h.database.SavePlayerData(data)
		if err != nil {
			log.Error("Failed to save player:", err)
			h.sendResponse(msgObj.Session, msg.MSG_DB_PLAYER_UPDATE, -1, "Failed to save player", nil)
			return
		}
	}

	h.sendResponse(msgObj.Session, msg.MSG_DB_PLAYER_UPDATE, 0, "success", nil)
}

func (h *DataServiceHandler) handleInventoryGet(msgObj *network.Message) {
	var req map[string]int64
	err := json.Unmarshal(msgObj.Data, &req)
	if err != nil {
		log.Error("Failed to parse request:", err)
		h.sendResponse(msgObj.Session, msg.MSG_DB_INVENTORY_GET, -1, "Invalid request", nil)
		return
	}

	playerID, ok := req["player_id"]
	if !ok {
		h.sendResponse(msgObj.Session, msg.MSG_DB_INVENTORY_GET, -1, "Missing player_id", nil)
		return
	}

	items, err := h.database.GetPlayerInventory(playerID)
	if err != nil {
		log.Error("Failed to get inventory:", err)
		h.sendResponse(msgObj.Session, msg.MSG_DB_INVENTORY_GET, -1, "Internal error", nil)
		return
	}

	result := make([]*msg.InventoryItem, 0)
	for _, item := range items {
		result = append(result, &msg.InventoryItem{
			ID:       item.ID,
			PlayerID: item.PlayerID,
			ItemID:   item.ItemID,
			Slot:     item.Slot,
			Count:    item.Count,
		})
	}

	h.sendResponse(msgObj.Session, msg.MSG_DB_INVENTORY_GET, 0, "success", result)
}

func (h *DataServiceHandler) handleItemGet(msgObj *network.Message) {
	var req map[string]int64
	err := json.Unmarshal(msgObj.Data, &req)
	if err != nil {
		log.Error("Failed to parse request:", err)
		h.sendResponse(msgObj.Session, msg.MSG_DB_ITEM_GET, -1, "Invalid request", nil)
		return
	}

	itemID, ok := req["item_id"]
	if !ok {
		h.sendResponse(msgObj.Session, msg.MSG_DB_ITEM_GET, -1, "Missing item_id", nil)
		return
	}

	item, err := h.database.GetItemConfig(itemID)
	if err != nil {
		log.Error("Failed to get item:", err)
		h.sendResponse(msgObj.Session, msg.MSG_DB_ITEM_GET, -1, "Internal error", nil)
		return
	}

	if item == nil {
		h.sendResponse(msgObj.Session, msg.MSG_DB_ITEM_GET, -1, "Item not found", nil)
		return
	}

	result := &msg.ItemConfig{
		ID:          item.ID,
		Name:        item.Name,
		Type:        item.Type,
		EffectType:  item.EffectType,
		EffectValue: item.EffectValue,
		Icon:        item.Icon,
		Description: item.Description,
	}

	h.sendResponse(msgObj.Session, msg.MSG_DB_ITEM_GET, 0, "success", result)
}

func main() {
	flag.Parse()

	log.Info("Data service initializing...")

	err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	logger := log.NewLogger("dataservice")
	loggerConfig := config.GetLoggerConfig()
	logger.Init(loggerConfig.Path, loggerConfig.Level)

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

	err = database.InitTables()
	if err != nil {
		log.Fatal("Failed to init tables:", err)
	}

	listenAddr := config.GetDataServiceListenAddr()
	log.Info("Data service starting on ", listenAddr)

	handler := &DataServiceHandler{database: database}
	server := network.NewServer(handler)

	go func() {
		if err := server.Start(listenAddr); err != nil {
			log.Fatal("Server failed to start:", err)
		}
	}()
	log.Info("Data service started successfully")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("Shutting down data service...")
	server.Stop()
	database.ForceFlush()
	database.Close()
	log.Info("Data service shutdown complete")
}
