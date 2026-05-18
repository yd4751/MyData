package main

import (
	"encoding/json"

	"github.com/openworld-server/internal/msg"
	"github.com/openworld-server/internal/network"
	"github.com/openworld-server/pkg/logger"
)

type BattleHandler struct {
}

func NewBattleHandler() *BattleHandler {
	return &BattleHandler{}
}

func (h *BattleHandler) Handle(msgObj *network.Message) {
	logger.Info("Battle received message - MsgID:", msgObj.ID, "(", msg.GetMsgName(msgObj.ID), ")")

	switch msgObj.ID {
	case msg.MSG_ATTACK_REQ:
		h.handleAttackRequest(msgObj)
	case msg.MSG_SKILL_REQ:
		h.handleSkillRequest(msgObj)
	}
}

func (h *BattleHandler) handleAttackRequest(msgObj *network.Message) {
	logger.Info("Handling battle attack request")

	req := &msg.AttackRequest{}
	if err := json.Unmarshal(msgObj.Data, req); err != nil {
		logger.Error("Failed to unmarshal attack request:", err)
		return
	}

	res := &msg.AttackResponse{
		Result:     0,
		Message:    "attack success",
		AttackerID: req.PlayerID,
		TargetID:   req.TargetID,
		Damage:     25,
		TargetHP:   55,
	}

	data, err := json.Marshal(res)
	if err != nil {
		logger.Error("Failed to marshal attack response:", err)
		return
	}

	msgObj.Session.Send(msg.MSG_ATTACK_RES, msg.NodeTypeBattle, data)
	logger.Info("Battle attack response sent - Attacker:", req.PlayerID, "Target:", req.TargetID)
}

func (h *BattleHandler) handleSkillRequest(msgObj *network.Message) {
	logger.Info("Handling battle skill request")

	req := &msg.SkillRequest{}
	if err := json.Unmarshal(msgObj.Data, req); err != nil {
		logger.Error("Failed to unmarshal skill request:", err)
		return
	}

	res := &msg.SkillResponse{
		Result:   0,
		Message:  "skill success",
		SkillID:  req.SkillID,
		TargetID: req.TargetID,
		Damage:   40,
		TargetHP: 40,
	}

	data, err := json.Marshal(res)
	if err != nil {
		logger.Error("Failed to marshal skill response:", err)
		return
	}

	msgObj.Session.Send(msg.MSG_SKILL_RES, msg.NodeTypeBattle, data)
	logger.Info("Battle skill response sent - PlayerID:", req.PlayerID, "SkillID:", req.SkillID)
}
