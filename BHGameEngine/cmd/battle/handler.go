package main

import (
	"encoding/json"

	"github.com/openworld-server/internal/battlecore"
	"github.com/openworld-server/internal/entity"
	"github.com/openworld-server/internal/network"
	"github.com/openworld-server/internal/player"
	"github.com/openworld-server/internal/skill"
	"github.com/openworld-server/internal/worldmap"
	"github.com/openworld-server/pkg/logger"
)

type BattleHandler struct {
	battleManager *battlecore.BattleManager
}

func NewBattleHandler() *BattleHandler {
	return &BattleHandler{
		battleManager: battlecore.GetBattleManager(),
	}
}

func (h *BattleHandler) Handle(msg *network.Message) {
	switch msg.ID {
	case 4001:
		h.handleAttack(msg)
	case 4003:
		h.handleSkill(msg)
	case 4005:
		h.handleBattleEnd(msg)
	default:
		logger.Info("Battle server received unknown message ID: ", msg.ID)
	}
}

func (h *BattleHandler) handleAttack(msg *network.Message) {
	var req struct {
		BattleID int64 `json:"battle_id"`
		PlayerID int64 `json:"player_id"`
		TargetID int64 `json:"target_id"`
	}

	err := json.Unmarshal(msg.Data, &req)
	if err != nil {
		logger.Error("Failed to unmarshal attack request: ", err)
		return
	}

	battle, ok := h.battleManager.GetBattle(req.BattleID)
	if !ok {
		logger.Error("Battle not found: ", req.BattleID)
		return
	}

	player, ok := battle.GetPlayer(req.PlayerID)
	if !ok {
		logger.Error("Player not found: ", req.PlayerID)
		return
	}

	monster, monsterOk := battle.GetMonster(req.TargetID)
	if !monsterOk {
		logger.Error("Monster not found: ", req.TargetID)
		return
	}

	combatLog := battlecore.ProcessMonsterAttack(monster, player)
	if combatLog != nil {
		battle.AddCombatLog(combatLog)
	}

	response := map[string]interface{}{
		"result":      0,
		"message":     "attack success",
		"attacker_id": req.PlayerID,
		"target_id":   req.TargetID,
	}

	h.sendResponse(msg.Session, 4002, response)
}

func (h *BattleHandler) handleSkill(msg *network.Message) {
	var req struct {
		BattleID int64 `json:"battle_id"`
		PlayerID int64 `json:"player_id"`
		SkillID  int64 `json:"skill_id"`
		TargetID int64 `json:"target_id"`
	}

	err := json.Unmarshal(msg.Data, &req)
	if err != nil {
		logger.Error("Failed to unmarshal skill request: ", err)
		return
	}

	battle, ok := h.battleManager.GetBattle(req.BattleID)
	if !ok {
		logger.Error("Battle not found: ", req.BattleID)
		h.sendError(msg.Session, 4004, "battle not found")
		return
	}

	if !battle.IsActive() {
		logger.Error("Battle is not active: ", req.BattleID)
		h.sendError(msg.Session, 4004, "battle not active")
		return
	}

	caster, ok := battle.GetPlayer(req.PlayerID)
	if !ok {
		logger.Error("Caster not found in battle: ", req.PlayerID)
		h.sendError(msg.Session, 4004, "caster not found")
		return
	}

	log, err := battle.CastSkill(caster, req.SkillID, req.TargetID)
	if err != nil {
		logger.Error("Failed to cast skill: ", err)
		h.sendError(msg.Session, 4004, err.Error())
		return
	}

	logger.Info("Skill cast successful: ", req.SkillID, " by ", req.PlayerID, " on ", req.TargetID)

	response := map[string]interface{}{
		"result":    0,
		"message":   "skill cast success",
		"battle_id": req.BattleID,
		"caster_id": req.PlayerID,
		"skill_id":  req.SkillID,
		"target_id": req.TargetID,
		"damage":    log.Damage,
		"heal":      log.Heal,
		"is_combo":  log.IsCombo,
	}

	h.sendResponse(msg.Session, 4004, response)

	if battle.CheckBattleEnd() {
		rewards := battle.GetRewards()
		h.battleManager.EndBattle(req.BattleID)

		battleEndResponse := map[string]interface{}{
			"result":    0,
			"message":   "battle ended",
			"battle_id": req.BattleID,
			"rewards":   rewards,
		}
		h.sendResponse(msg.Session, 4006, battleEndResponse)
	}
}

func (h *BattleHandler) handleBattleEnd(msg *network.Message) {
	var req struct {
		BattleID int64 `json:"battle_id"`
	}

	err := json.Unmarshal(msg.Data, &req)
	if err != nil {
		logger.Error("Failed to unmarshal battle end request: ", err)
		return
	}

	h.battleManager.EndBattle(req.BattleID)
	logger.Info("Ended battle: ", req.BattleID)

	response := map[string]interface{}{
		"result":    0,
		"message":   "battle ended",
		"battle_id": req.BattleID,
	}

	h.sendResponse(msg.Session, 4006, response)
}

func (h *BattleHandler) sendResponse(conn interface{ Write([]byte) (int, error) }, msgID uint32, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Error("Failed to marshal response: ", err)
		return
	}

	packet := make([]byte, 12+len(jsonData))
	copy(packet[12:], jsonData)

	conn.Write(packet)
}

func (h *BattleHandler) sendError(conn interface{ Write([]byte) (int, error) }, msgID uint32, message string) {
	response := map[string]interface{}{
		"result":  1,
		"message": message,
	}
	h.sendResponse(conn, msgID, response)
}

func init() {
	skill.LoadDefaultSkills()
}

func CreateBattle(battleID int64) *battlecore.Battle {
	return battlecore.GetBattleManager().CreateBattle(battleID)
}

func AddPlayerToBattle(battleID int64, p *player.Player) error {
	battle, ok := battlecore.GetBattleManager().GetBattle(battleID)
	if !ok {
		return skill.ErrInvalidTarget
	}
	battle.AddPlayer(p)
	return nil
}

func AddMonsterToBattle(battleID int64, m *entity.Monster) error {
	battle, ok := battlecore.GetBattleManager().GetBattle(battleID)
	if !ok {
		return skill.ErrInvalidTarget
	}
	battle.AddMonster(m)
	return nil
}

func CreatePlayer(playerID int64, name string, accountID int64) *player.Player {
	pm := player.NewPlayerManager()
	return pm.CreatePlayer(playerID, name, accountID)
}

func CreateMonster(monsterID int64, name string, pos worldmap.Vec3, level int32) *entity.Monster {
	em := entity.NewEntityManager()
	return em.CreateMonster(monsterID, name, pos, level)
}
