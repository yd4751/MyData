package main

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/openworld-server/internal/dataclient"
	"github.com/openworld-server/internal/log"
	"github.com/openworld-server/internal/msg"
	"github.com/openworld-server/internal/network"
	"github.com/openworld-server/internal/redis"
)

type LogicHandler struct {
	dataClient  *dataclient.DataClient
	redisClient *redis.RedisClient
}

func NewLogicHandler(dataClient *dataclient.DataClient, redisClient *redis.RedisClient) *LogicHandler {
	return &LogicHandler{
		dataClient:  dataClient,
		redisClient: redisClient,
	}
}

func (h *LogicHandler) Handle(msgObj *network.Message) {
	log.Info("Logic server received message from ", msgObj.Session.RemoteAddr(), " - MsgID:", msgObj.ID, "(", msg.GetMsgName(msgObj.ID), "), NodeType:", msgObj.NodeType, "(", msgObj.NodeType.String(), ")")

	switch msgObj.ID {
	case msg.MSG_PLAYER_INFO_REQ:
		h.handlePlayerInfo(msgObj)
	case msg.MSG_PLAYER_MOVE_REQ:
		h.handlePlayerMove(msgObj)
	case msg.MSG_INVENTORY_REQ:
		h.handleInventoryRequest(msgObj)
	case msg.MSG_ITEM_USE_REQ:
		h.handleItemUseRequest(msgObj)
	default:
		log.Warn("Unknown message ID:", msgObj.ID, "(", msg.GetMsgName(msgObj.ID), ")")
	}
}

func (h *LogicHandler) handlePlayerInfo(msgObj *network.Message) {
	var req msg.PlayerInfoRequest
	err := json.Unmarshal(msgObj.Data, &req)
	if err != nil {
		log.Error("Failed to parse player info request:", err)
		h.sendError(msgObj.Session, msg.MSG_PLAYER_INFO_RES, "Invalid request")
		return
	}

	log.Info("Player info request: PlayerID=", req.PlayerID)

	player, err := h.dataClient.GetPlayerByID(req.PlayerID)
	if err != nil {
		log.Error("Failed to get player:", err)
		h.sendError(msgObj.Session, msg.MSG_PLAYER_INFO_RES, "Player not found")
		return
	}

	if player == nil {
		log.Warn("Player not found: PlayerID=", req.PlayerID)
		h.sendError(msgObj.Session, msg.MSG_PLAYER_INFO_RES, "Player not found")
		return
	}

	resp := msg.PlayerInfoResponse{
		Result:     0,
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

	log.Info("Player info response sent: PlayerID=", req.PlayerID)
	h.sendResponse(msgObj.Session, msg.MSG_PLAYER_INFO_RES, resp)
}

func (h *LogicHandler) handlePlayerMove(msgObj *network.Message) {
	log.Debug("Raw player move request data: ", string(msgObj.Data))

	var req msg.PlayerMoveRequest
	err := json.Unmarshal(msgObj.Data, &req)
	if err != nil {
		log.Error("Failed to parse player move request:", err)
		h.sendMoveError(msgObj.Session, "Invalid request")
		return
	}

	log.Info("Player move request: PlayerID=", req.PlayerID, ", TargetX=", req.TargetX, ", TargetY=", req.TargetY)

	if req.PlayerID == 0 {
		log.Warn("PlayerID is 0, possibly JSON field name mismatch")
	}

	err = h.dataClient.UpdatePlayerPosition(req.PlayerID, req.TargetX, req.TargetY)
	if err != nil {
		log.Error("Failed to update player position:", err)
		h.sendMoveError(msgObj.Session, "Internal error")
		return
	}

	resp := msg.PlayerMoveResponse{
		Result: 0,
		PosX:   req.TargetX,
		PosY:   req.TargetY,
	}

	log.Info("Player move response sent: PlayerID=", req.PlayerID)
	h.sendMoveResponse(msgObj.Session, resp)
}

func (h *LogicHandler) sendError(conn net.Conn, msgID uint32, message string) {
	resp := msg.PlayerInfoResponse{
		Result:  1,
		Message: message,
	}
	h.sendResponse(conn, msgID, resp)
}

func (h *LogicHandler) sendResponse(conn net.Conn, msgID uint32, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Error("Failed to marshal response:", err)
		return
	}

	log.Debug("Sending response - MsgID:", msgID, "(", msg.GetMsgName(msgID), "), Length:", len(jsonData))
	err = network.SendRawMessage(conn, msgID, msg.NodeTypeLogic, jsonData)
	if err != nil {
		log.Error("Failed to send response:", err)
	}
}

func (h *LogicHandler) sendMoveError(conn net.Conn, message string) {
	resp := msg.PlayerMoveResponse{
		Result:  1,
		Message: message,
	}
	h.sendMoveResponse(conn, resp)
}

func (h *LogicHandler) sendMoveResponse(conn net.Conn, data msg.PlayerMoveResponse) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Error("Failed to marshal response:", err)
		return
	}

	log.Debug("Sending move response - Length:", len(jsonData))
	err = network.SendRawMessage(conn, msg.MSG_PLAYER_MOVE_RES, msg.NodeTypeLogic, jsonData)
	if err != nil {
		log.Error("Failed to send response:", err)
	}
}

func (h *LogicHandler) handleInventoryRequest(msgObj *network.Message) {
	var req msg.InventoryRequest
	err := json.Unmarshal(msgObj.Data, &req)
	if err != nil {
		log.Error("Failed to parse inventory request:", err)
		h.sendInventoryError(msgObj.Session, "Invalid request")
		return
	}

	log.Info("Inventory request received: PlayerID=", req.PlayerID, ", SessionID=", req.SessionID)
	if req.PlayerID == 0 {
		log.Warn("Inventory request with PlayerID=0, this will return empty inventory")
	}

	inventory, err := h.dataClient.GetPlayerInventory(req.PlayerID)
	if err != nil {
		log.Error("Failed to get player inventory:", err)
		h.sendInventoryError(msgObj.Session, "Internal error")
		return
	}

	items := make([]msg.ItemInfo, 0)
	itemConfigs := make([]msg.ItemConfig, 0)
	configMap := make(map[int64]bool)

	for _, item := range inventory {
		itemConfig, err := h.dataClient.GetItemConfig(item.ItemID)
		if err != nil {
			log.Warn("Item config not found for item ID:", item.ItemID)
			items = append(items, msg.ItemInfo{
				ItemID:      uint32(item.ItemID),
				Name:        "未知道具",
				Icon:        "unknown",
				Count:       item.Count,
				Slot:        item.Slot,
				Level:       0,
				UID:         fmt.Sprintf("%d", item.ID),
				Description: "未知道具",
			})
			continue
		}

		if !configMap[item.ItemID] {
			itemConfigs = append(itemConfigs, msg.ItemConfig{
				ID:          itemConfig.ID,
				Name:        itemConfig.Name,
				Type:        itemConfig.Type,
				EffectType:  itemConfig.EffectType,
				EffectValue: itemConfig.EffectValue,
				Icon:        itemConfig.Icon,
				Description: itemConfig.Description,
			})
			configMap[item.ItemID] = true
		}

		items = append(items, msg.ItemInfo{
			ItemID:      uint32(item.ItemID),
			Name:        itemConfig.Name,
			Icon:        itemConfig.Icon,
			Count:       item.Count,
			Slot:        item.Slot,
			Level:       0,
			UID:         fmt.Sprintf("%d", item.ID),
			Description: itemConfig.Description,
		})
	}

	resp := msg.InventoryResponse{
		Result:      0,
		Items:       items,
		Gold:        1000,
		Equipments:  []msg.EquipmentInfo{},
		Capacity:    50,
		ItemConfigs: itemConfigs,
	}

	log.Info("Inventory response sent: PlayerID=", req.PlayerID, ", ItemCount=", len(items))
	h.sendInventoryResponse(msgObj.Session, resp)
}

func (h *LogicHandler) handleItemUseRequest(msgObj *network.Message) {
	var req msg.ItemUseRequest
	err := json.Unmarshal(msgObj.Data, &req)
	if err != nil {
		log.Error("Failed to parse item use request:", err)
		h.sendItemUseError(msgObj.Session, "Invalid request")
		return
	}

	log.Info("Item use request: PlayerID=", req.PlayerID, ", ItemID=", req.ItemID, ", Position=", req.Position)

	playerData, err := h.dataClient.GetPlayerByID(req.PlayerID)
	if err != nil || playerData == nil {
		log.Error("Player not found:", req.PlayerID)
		h.sendItemUseError(msgObj.Session, "Player not found")
		return
	}

	inventory, err := h.dataClient.GetPlayerInventory(req.PlayerID)
	if err != nil {
		log.Error("Failed to get inventory:", err)
		h.sendItemUseError(msgObj.Session, "Internal error")
		return
	}

	var targetItem *msg.InventoryItem
	for _, item := range inventory {
		if item.Slot == req.Position {
			targetItem = item
			break
		}
	}

	if targetItem == nil {
		log.Warn("Item not found in slot:", req.Position)
		h.sendItemUseError(msgObj.Session, "Item not found")
		return
	}

	if targetItem.Count <= 0 {
		log.Warn("Item count is zero:", req.Position)
		h.sendItemUseError(msgObj.Session, "Item count is zero")
		return
	}

	itemConfig, err := h.dataClient.GetItemConfig(targetItem.ItemID)
	if err != nil {
		log.Error("Item config not found:", targetItem.ItemID)
		h.sendItemUseError(msgObj.Session, "Item config not found")
		return
	}

	if itemConfig.Type != 1 {
		log.Warn("Item is not consumable:", targetItem.ItemID)
		h.sendItemUseError(msgObj.Session, "Item is not consumable")
		return
	}

	switch itemConfig.EffectType {
	case 1:
		playerData.Health += itemConfig.EffectValue
		if playerData.Health > playerData.MaxHealth {
			playerData.Health = playerData.MaxHealth
		}
	case 2:
		playerData.Mana += itemConfig.EffectValue
		if playerData.Mana > playerData.MaxMana {
			playerData.Mana = playerData.MaxMana
		}
	default:
		log.Warn("Unknown effect type:", itemConfig.EffectType)
		h.sendItemUseError(msgObj.Session, "Unknown effect type")
		return
	}

	err = h.dataClient.UpdatePlayer(playerData)
	if err != nil {
		log.Error("Failed to save player data:", err)
		h.sendItemUseError(msgObj.Session, "Internal error")
		return
	}

	resp := msg.ItemUseResponse{
		Result:  0,
		Message: "Use success",
	}

	log.Info("Item use response sent: PlayerID=", req.PlayerID, ", ItemID=", req.ItemID)
	h.sendItemUseResponse(msgObj.Session, resp)
}

func (h *LogicHandler) sendInventoryError(conn net.Conn, message string) {
	resp := msg.InventoryResponse{
		Result:  1,
		Message: message,
	}
	h.sendInventoryResponse(conn, resp)
}

func (h *LogicHandler) sendInventoryResponse(conn net.Conn, data msg.InventoryResponse) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Error("Failed to marshal inventory response:", err)
		return
	}

	log.Debug("Sending inventory response - Length:", len(jsonData))
	err = network.SendRawMessage(conn, msg.MSG_INVENTORY_RES, msg.NodeTypeLogic, jsonData)
	if err != nil {
		log.Error("Failed to send inventory response:", err)
	}
}

func (h *LogicHandler) sendItemUseError(conn net.Conn, message string) {
	resp := msg.ItemUseResponse{
		Result:  1,
		Message: message,
	}
	h.sendItemUseResponse(conn, resp)
}

func (h *LogicHandler) sendItemUseResponse(conn net.Conn, data msg.ItemUseResponse) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Error("Failed to marshal item use response:", err)
		return
	}

	log.Debug("Sending item use response - Length:", len(jsonData))
	err = network.SendRawMessage(conn, msg.MSG_ITEM_USE_RES, msg.NodeTypeLogic, jsonData)
	if err != nil {
		log.Error("Failed to send item use response:", err)
	}
}
