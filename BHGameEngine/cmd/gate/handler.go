package main

import (
	"encoding/json"

	"github.com/openworld-server/internal/cluster"
	"github.com/openworld-server/internal/msg"
	"github.com/openworld-server/internal/network"
	"github.com/openworld-server/pkg/logger"
)

type GateHandler struct {
	cluster *cluster.Cluster
}

func NewGateHandler(c *cluster.Cluster) *GateHandler {
	return &GateHandler{cluster: c}
}

func (h *GateHandler) Handle(msgObj *network.Message) {
	logger.Info("Gate received message from ", msgObj.Session.RemoteAddr(), " - MsgID:", msgObj.ID, "(", msg.GetMsgName(msgObj.ID), "), NodeType:", msgObj.NodeType, "(", msgObj.NodeType.String(), ")")

	if msgObj.NodeType == msg.NodeTypeGate {
		h.handleGateMessage(msgObj)
	} else {
		h.forwardMessage(msgObj)
	}
}

func (h *GateHandler) handleGateMessage(msgObj *network.Message) {
	logger.Info("Handling gate message, MsgID:", msgObj.ID, "(", msg.GetMsgName(msgObj.ID), ")")

	switch msgObj.ID {
	case msg.MSG_PLAYER_MOVE_REQ:
		h.handleMoveRequest(msgObj)
	}
}

func (h *GateHandler) handleMoveRequest(msgObj *network.Message) {
	logger.Info("Handling move request from ", msgObj.Session.RemoteAddr())

	req := &msg.PlayerMoveRequest{}
	if err := json.Unmarshal(msgObj.Data, req); err != nil {
		logger.Error("Failed to unmarshal move request:", err)
		return
	}

	res := &msg.PlayerMoveResponse{
		Result:  0,
		Message: "success",
		PosX:    req.TargetX,
		PosY:    req.TargetY,
	}

	data, err := json.Marshal(res)
	if err != nil {
		logger.Error("Failed to marshal move response:", err)
		return
	}

	msgObj.Session.Send(msg.MSG_PLAYER_MOVE_RES, msg.NodeTypeGate, data)
	logger.Info("Move response sent - X:", res.PosX, ", Y:", res.PosY)
}

func (h *GateHandler) forwardMessage(msgObj *network.Message) {
	serviceName := h.getServiceNameByNodeType(msgObj.NodeType)
	if serviceName == "" {
		logger.Error("Unknown node type:", msgObj.NodeType, "(", msgObj.NodeType.String(), ")")
		return
	}

	logger.Info("Forwarding message to ", serviceName, " - MsgID:", msgObj.ID, "(", msg.GetMsgName(msgObj.ID), ")")
	service, err := h.cluster.GetRandomService(serviceName)
	if err != nil {
		logger.Error("Failed to get service ", serviceName, ": ", err)
		return
	}

	if err := h.cluster.Send(service.ID, msgObj.ID, msgObj.NodeType, msgObj.Data); err != nil {
		logger.Error("Failed to send message to ", serviceName, ": ", err)
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
