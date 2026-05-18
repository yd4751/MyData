package main

import (
	"encoding/json"
	"net"
	"sync"

	"github.com/openworld-server/internal/cluster"
	"github.com/openworld-server/internal/log"
	"github.com/openworld-server/internal/msg"
	"github.com/openworld-server/internal/network"
	"github.com/openworld-server/internal/worldmap"
)

type GateHandler struct {
	cluster     *cluster.Cluster
	gridRouter  *worldmap.GridMapRouter
	playerGrids map[int64]int
	mu          sync.RWMutex
}

func NewGateHandler(cluster *cluster.Cluster) *GateHandler {
	return &GateHandler{
		cluster:     cluster,
		gridRouter:  worldmap.NewGridMapRouter(cluster, 9),
		playerGrids: make(map[int64]int),
	}
}

func (h *GateHandler) Handle(msgObj *network.Message) {
	log.Info("Gate received message from ", msgObj.Session.RemoteAddr(), " - MsgID:", msgObj.ID, "(", msg.GetMsgName(msgObj.ID), "), NodeType:", msgObj.NodeType, "(", msgObj.NodeType.String(), ")")

	if msgObj.NodeType == msg.NodeTypeGate {
		h.handleGateMessage(msgObj)
	} else {
		h.forwardMessage(msgObj)
	}
}

func (h *GateHandler) handleGateMessage(msgObj *network.Message) {
	switch msgObj.ID {
	case msg.MSG_MAP_PLAYER_ENTER:
		h.handleMapPlayerEnter(msgObj)
	case msg.MSG_MAP_PLAYER_MOVE:
		h.handleMapPlayerMove(msgObj)
	default:
		log.Info("Handling gate message, MsgID:", msgObj.ID, "(", msg.GetMsgName(msgObj.ID), ")")
	}
}

func (h *GateHandler) handleMapPlayerEnter(msgObj *network.Message) {
	var req msg.MapPlayerEnterRequest
	err := json.Unmarshal(msgObj.Data, &req)
	if err != nil {
		log.Error("Failed to unmarshal MapPlayerEnterRequest: ", err)
		return
	}

	gridID := h.gridRouter.GetGridMapID(worldmap.Vec3{X: req.PosX, Y: req.PosY, Z: req.PosZ})

	h.mu.Lock()
	h.playerGrids[req.PlayerID] = gridID
	h.mu.Unlock()

	h.forwardToGridMap(gridID, msgObj)
}

func (h *GateHandler) handleMapPlayerMove(msgObj *network.Message) {
	var req msg.MapPlayerMoveRequest
	err := json.Unmarshal(msgObj.Data, &req)
	if err != nil {
		log.Error("Failed to unmarshal MapPlayerMoveRequest: ", err)
		return
	}

	h.mu.RLock()
	currentGridID, ok := h.playerGrids[req.PlayerID]
	h.mu.RUnlock()

	newGridID := h.gridRouter.GetGridMapID(worldmap.Vec3{X: req.PosX, Y: req.PosY, Z: req.PosZ})

	if !ok {
		currentGridID = newGridID
		h.mu.Lock()
		h.playerGrids[req.PlayerID] = currentGridID
		h.mu.Unlock()
	}

	if newGridID != currentGridID {
		log.Info("Player ", req.PlayerID, " crossing grid boundary: ", currentGridID, " -> ", newGridID)

		h.mu.Lock()
		h.playerGrids[req.PlayerID] = newGridID
		h.mu.Unlock()
	}

	h.forwardToGridMap(newGridID, msgObj)
}

func (h *GateHandler) forwardMessage(msgObj *network.Message) {
	if msgObj.NodeType == msg.NodeTypeGridMap {
		var playerID int64
		switch msgObj.ID {
		case msg.MSG_MAP_PLAYER_ENTER:
			var req msg.MapPlayerEnterRequest
			json.Unmarshal(msgObj.Data, &req)
			playerID = req.PlayerID
		case msg.MSG_MAP_PLAYER_MOVE:
			var req msg.MapPlayerMoveRequest
			json.Unmarshal(msgObj.Data, &req)
			playerID = req.PlayerID
		case msg.MSG_MAP_PLAYER_LEAVE:
			var req msg.MapPlayerLeaveRequest
			json.Unmarshal(msgObj.Data, &req)
			playerID = req.PlayerID
		}

		if playerID != 0 {
			h.mu.RLock()
			gridID, ok := h.playerGrids[playerID]
			h.mu.RUnlock()

			if ok {
				h.forwardToGridMap(gridID, msgObj)
				return
			}
		}
	}

	serviceName := h.getServiceNameByNodeType(msgObj.NodeType)
	if serviceName == "" {
		log.Error("Unknown node type:", msgObj.NodeType, "(", msgObj.NodeType.String(), ")")
		return
	}

	log.Info("Forwarding message to ", serviceName, " - MsgID:", msgObj.ID, "(", msg.GetMsgName(msgObj.ID), ")")
	service, err := h.cluster.GetRandomService(serviceName)
	if err != nil {
		log.Error("Failed to get service ", serviceName, ": ", err)
		return
	}

	h.sendToService(service.Addr, msgObj)
}

func (h *GateHandler) forwardToGridMap(gridID int, msgObj *network.Message) {
	serviceName := "gridmap-" + string(rune('0'+gridID))
	services, err := h.cluster.DiscoverServices(serviceName)
	if err != nil || len(services) == 0 {
		log.Error("Failed to find gridmap service: ", serviceName, err)
		return
	}

	h.sendToService(services[0].Addr, msgObj)
}

func (h *GateHandler) sendToService(addr string, msgObj *network.Message) {
	log.Info("Connecting to service at ", addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Error("Failed to connect to service at ", addr, ": ", err)
		return
	}
	defer conn.Close()

	err = network.SendRawMessage(conn, msgObj.ID, msgObj.NodeType, msgObj.Data)
	if err != nil {
		log.Error("Failed to send message to service: ", err)
		return
	}

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		log.Error("Failed to read response from service: ", err)
		return
	}

	respMsgID, respNodeType, respData, err := network.ParseMessage(buf[:n])
	if err != nil {
		log.Error("Failed to parse response: ", err)
		return
	}

	log.Info("Received response from service - MsgID:", respMsgID, "(", msg.GetMsgName(respMsgID), "), NodeType:", respNodeType, "(", respNodeType.String(), ")")
	err = network.SendRawMessage(msgObj.Session, respMsgID, respNodeType, respData)
	if err != nil {
		log.Error("Failed to send response to client: ", err)
	}
}

func (h *GateHandler) getServiceNameByNodeType(nodeType msg.NodeType) string {
	switch nodeType {
	case msg.NodeTypeLogin:
		return "login"
	case msg.NodeTypeLogic:
		return "logic"
	case msg.NodeTypeBattle:
		return "battle"
	case msg.NodeTypeGridMap:
		return "gridmap"
	case msg.NodeTypeCross:
		return "cross"
	case msg.NodeTypeData:
		return "dataservice"
	case msg.NodeTypeGM:
		return "gm"
	default:
		return ""
	}
}
