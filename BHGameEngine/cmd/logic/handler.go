package main

import (
	"encoding/json"

	"github.com/openworld-server/internal/msg"
	"github.com/openworld-server/internal/network"
	"github.com/openworld-server/internal/player"
	"github.com/openworld-server/pkg/logger"
)

type LogicHandler struct {
}

func NewLogicHandler() *LogicHandler {
	return &LogicHandler{}
}

func (h *LogicHandler) Handle(msgObj *network.Message) {
	logger.Info("Logic received message - MsgID:", msgObj.ID, "(", msg.GetMsgName(msgObj.ID), ")")

	switch msgObj.ID {
	case msg.MSG_PLAYER_MOVE_REQ:
		h.handlePlayerMoveRequest(msgObj)
	case msg.MSG_ATTACK_REQ:
		h.handleAttackRequest(msgObj)
	case msg.MSG_SKILL_REQ:
		h.handleSkillRequest(msgObj)
	case msg.MSG_ITEM_USE_REQ:
		h.handleItemUseRequest(msgObj)
	}
}

func (h *LogicHandler) handlePlayerMoveRequest(msgObj *network.Message) {
	logger.Info("Handling player move request")

	req := &msg.PlayerMoveRequest{}
	if err := json.Unmarshal(msgObj.Data, req); err != nil {
		logger.Error("Failed to unmarshal player move request:", err)
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
		logger.Error("Failed to marshal player move response:", err)
		return
	}

	msgObj.Session.Send(msg.MSG_PLAYER_MOVE_RES, msg.NodeTypeLogic, data)
	logger.Info("Player move response sent - PlayerID:", req.PlayerID, "Pos:", req.TargetX, ",", req.TargetY)
}

func (h *LogicHandler) handleAttackRequest(msgObj *network.Message) {
	logger.Info("Handling attack request")

	req := &msg.AttackRequest{}
	if err := json.Unmarshal(msgObj.Data, req); err != nil {
		logger.Error("Failed to unmarshal attack request:", err)
		return
	}

	res := &msg.AttackResponse{
		Result:     0,
		Message:    "success",
		AttackerID: req.PlayerID,
		TargetID:   req.TargetID,
		Damage:     20,
		TargetHP:   60,
	}

	data, err := json.Marshal(res)
	if err != nil {
		logger.Error("Failed to marshal attack response:", err)
		return
	}

	msgObj.Session.Send(msg.MSG_ATTACK_RES, msg.NodeTypeLogic, data)
	logger.Info("Attack response sent - Attacker:", req.PlayerID, "Target:", req.TargetID)
}

func (h *LogicHandler) handleSkillRequest(msgObj *network.Message) {
	logger.Info("Handling skill request")

	req := &msg.SkillRequest{}
	if err := json.Unmarshal(msgObj.Data, req); err != nil {
		logger.Error("Failed to unmarshal skill request:", err)
		return
	}

	res := &msg.SkillResponse{
		Result:   0,
		Message:  "success",
		SkillID:  req.SkillID,
		TargetID: req.TargetID,
		Damage:   35,
		TargetHP: 45,
	}

	data, err := json.Marshal(res)
	if err != nil {
		logger.Error("Failed to marshal skill response:", err)
		return
	}

	msgObj.Session.Send(msg.MSG_SKILL_RES, msg.NodeTypeLogic, data)
	logger.Info("Skill response sent - PlayerID:", req.PlayerID, "SkillID:", req.SkillID)
}

func (h *LogicHandler) handleItemUseRequest(msgObj *network.Message) {
	logger.Info("Handling item use request")

	req := &msg.ItemUseRequest{}
	if err := json.Unmarshal(msgObj.Data, req); err != nil {
		logger.Error("Failed to unmarshal item use request:", err)
		return
	}

	playerManager := player.NewPlayerManager()
	p, ok := playerManager.GetPlayer(req.PlayerID)
	if !ok {
		res := &msg.ItemUseResponse{
			Result:  -1,
			Message: "玩家不存在",
		}
		data, _ := json.Marshal(res)
		msgObj.Session.Send(msg.MSG_ITEM_USE_RES, msg.NodeTypeLogic, data)
		return
	}

	success, message := p.UseItem(req.Position)

	res := &msg.ItemUseResponse{
		Result:  0,
		Message: message,
	}
	if !success {
		res.Result = -1
	}

	data, err := json.Marshal(res)
	if err != nil {
		logger.Error("Failed to marshal item use response:", err)
		return
	}

	msgObj.Session.Send(msg.MSG_ITEM_USE_RES, msg.NodeTypeLogic, data)
	logger.Info("Item use response sent - PlayerID:", req.PlayerID, "ItemID:", req.ItemID, "Result:", success)
}
