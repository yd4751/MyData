package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/openworld-server/internal/cluster"
	"github.com/openworld-server/internal/connector"
	"github.com/openworld-server/internal/log"
	"github.com/openworld-server/internal/msg"
	"github.com/openworld-server/internal/network"
	"github.com/openworld-server/internal/worldmap"
)

type GateHandler struct {
	connector   *connector.Connector
	gridRouter  *worldmap.GridMapRouter
	playerGrids map[int64]int
	mu          sync.RWMutex
}

func NewGateHandler(cluster *cluster.Cluster) *GateHandler {
	return &GateHandler{
		connector:   connector.NewConnector(cluster),
		gridRouter:  worldmap.NewGridMapRouter(cluster, 9),
		playerGrids: make(map[int64]int),
	}
}

func (h *GateHandler) Handle(msgObj *network.Message) {
	if msgObj.ID != msg.MSG_PING {
		log.Info("Gate received message from ", msgObj.Session.RemoteAddr(), " - MsgID:", msgObj.ID, "(", msg.GetMsgName(msgObj.ID), ")")
	}

	targetNodeType := msg.GetMessageNodeType(msgObj.ID)
	h.routeMessage(msgObj, targetNodeType)
}

func (h *GateHandler) routeMessage(msgObj *network.Message, targetNodeType msg.NodeType) {
	switch targetNodeType {
	case msg.NodeTypeGate:
		return
	case msg.NodeTypeGridMap:
		h.handleGridMapMessage(msgObj)
	default:
		h.forwardToService(msgObj, targetNodeType)
	}
}

func (h *GateHandler) handleGridMapMessage(msgObj *network.Message) {
	switch msgObj.ID {
	case msg.MSG_MAP_PLAYER_ENTER:
		h.handleMapPlayerEnter(msgObj)
	case msg.MSG_MAP_PLAYER_MOVE:
		h.handleMapPlayerMove(msgObj)
	case msg.MSG_MAP_PLAYER_LEAVE:
		h.handleMapPlayerLeave(msgObj)
	default:
		h.forwardToRandomGridMap(msgObj)
	}
}

func (h *GateHandler) handleMapPlayerEnter(msgObj *network.Message) {
	var req msg.MapPlayerEnterRequest
	if err := json.Unmarshal(msgObj.Data, &req); err != nil {
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
	if err := json.Unmarshal(msgObj.Data, &req); err != nil {
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

func (h *GateHandler) handleMapPlayerLeave(msgObj *network.Message) {
	var req msg.MapPlayerLeaveRequest
	if err := json.Unmarshal(msgObj.Data, &req); err != nil {
		log.Error("Failed to unmarshal MapPlayerLeaveRequest: ", err)
		return
	}

	h.mu.RLock()
	gridID, ok := h.playerGrids[req.PlayerID]
	h.mu.RUnlock()

	if ok {
		h.forwardToGridMap(gridID, msgObj)

		h.mu.Lock()
		delete(h.playerGrids, req.PlayerID)
		h.mu.Unlock()
	}
}

func (h *GateHandler) forwardToRandomGridMap(msgObj *network.Message) {
	h.sendToService(msg.NodeTypeGridMap, msgObj)
}

func (h *GateHandler) forwardToService(msgObj *network.Message, nodeType msg.NodeType) {
	h.sendToService(nodeType, msgObj)
}

func (h *GateHandler) forwardToGridMap(gridID int, msgObj *network.Message) {
	nodeID := fmt.Sprintf("gridmap:gridmap-%d", gridID)
	err := h.connector.SendToNodeID(nodeID, msgObj.ID, msg.NodeTypeGridMap, msgObj.Data)
	if err != nil {
		log.Error("Failed to send message to gridmap ", gridID, ": ", err)
	}
}

func (h *GateHandler) sendToService(nodeType msg.NodeType, msgObj *network.Message) {
	log.Info("Forwarding message to ", nodeType.String(), " - MsgID:", msgObj.ID, "(", msg.GetMsgName(msgObj.ID), ")")

	respMsgID, respNodeType, respData, err := h.connector.RequestToNodeType(nodeType, msgObj.ID, msgObj.Data, 5*time.Second)
	if err != nil {
		log.Error("Failed to send message to service: ", err)
		return
	}

	log.Info("Received response from service - MsgID:", respMsgID, "(", msg.GetMsgName(respMsgID), "), NodeType:", respNodeType, "(", respNodeType.String(), ")")
	err = network.SendRawMessage(msgObj.Session, respMsgID, respNodeType, respData)
	if err != nil {
		log.Error("Failed to send response to client: ", err)
	}
}
