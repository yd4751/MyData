package main

import (
	"encoding/json"

	"github.com/openworld-server/internal/battlecore"
	"github.com/openworld-server/internal/entity"
	"github.com/openworld-server/internal/log"
	"github.com/openworld-server/internal/msg"
	"github.com/openworld-server/internal/network"
	"github.com/openworld-server/internal/player"
	"github.com/openworld-server/internal/skill"
	"github.com/openworld-server/internal/worldmap"
)

type BattleHandler struct {
	battleManager *battlecore.BattleManager
}

func NewBattleHandler() *BattleHandler {
	return &BattleHandler{
		battleManager: battlecore.GetBattleManager(),
	}
}

func (h *BattleHandler) Handle(msgObj *network.Message) {
	log.Info("Battle server received message - MsgID:", msgObj.ID, "(", msg.GetMsgName(msgObj.ID), ")")

	switch msgObj.ID {
	case msg.MSG_ATTACK_REQ:
		h.handleAttack(msgObj)
	case msg.MSG_SKILL_REQ:
		h.handleSkill(msgObj)
	case msg.MSG_BATTLE_END_REQ:
		h.handleBattleEnd(msgObj)
	default:
		log.Warn("Unknown message ID:", msgObj.ID, "(", msg.GetMsgName(msgObj.ID), ")")
	}
}

func (h *BattleHandler) handleAttack(msgObj *network.Message) {
	var req struct {
		BattleID int64 `json:"battle_id"`
		PlayerID int64 `json:"player_id"`
		TargetID int64 `json:"target_id"`
	}

	err := json.Unmarshal(msgObj.Data, &req)
	if err != nil {
		log.Error("Failed to unmarshal attack request: ", err)
		return
	}

	battle, ok := h.battleManager.GetBattle(req.BattleID)
	if !ok {
		log.Error("Battle not found: ", req.BattleID)
		return
	}

	p, ok := battle.GetPlayer(req.PlayerID)
	if !ok {
		log.Error("Player not found: ", req.PlayerID)
		return
	}

	monster, monsterOk := battle.GetMonster(req.TargetID)
	if !monsterOk {
		log.Error("Monster not found: ", req.TargetID)
		return
	}

	combatLog := battlecore.ProcessMonsterAttack(monster, p)
	if combatLog != nil {
		battle.AddCombatLog(combatLog)
	}

	response := map[string]interface{}{
		"result":      0,
		"message":     "attack success",
		"attacker_id": req.PlayerID,
		"target_id":   req.TargetID,
	}

	h.sendResponse(msgObj.Session, msg.MSG_ATTACK_RES, response)
}

func (h *BattleHandler) handleSkill(msgObj *network.Message) {
	var req struct {
		BattleID int64 `json:"battle_id"`
		PlayerID int64 `json:"player_id"`
		SkillID  int64 `json:"skill_id"`
		TargetID int64 `json:"target_id"`
	}

	err := json.Unmarshal(msgObj.Data, &req)
	if err != nil {
		log.Error("Failed to unmarshal skill request: ", err)
		return
	}

	battle, ok := h.battleManager.GetBattle(req.BattleID)
	if !ok {
		log.Error("Battle not found: ", req.BattleID)
		h.sendError(msgObj.Session, msg.MSG_SKILL_RES, "battle not found")
		return
	}

	if !battle.IsActive() {
		log.Error("Battle is not active: ", req.BattleID)
		h.sendError(msgObj.Session, msg.MSG_SKILL_RES, "battle not active")
		return
	}

	caster, ok := battle.GetPlayer(req.PlayerID)
	if !ok {
		log.Error("Caster not found in battle: ", req.PlayerID)
		h.sendError(msgObj.Session, msg.MSG_SKILL_RES, "caster not found")
		return
	}

	skillLog, err := battle.CastSkill(caster, req.SkillID, req.TargetID)
	if err != nil {
		log.Error("Failed to cast skill: ", err)
		h.sendError(msgObj.Session, msg.MSG_SKILL_RES, err.Error())
		return
	}

	log.Info("Skill cast successful: ", req.SkillID, " by ", req.PlayerID, " on ", req.TargetID)

	response := map[string]interface{}{
		"result":    0,
		"message":   "skill cast success",
		"battle_id": req.BattleID,
		"caster_id": req.PlayerID,
		"skill_id":  req.SkillID,
		"target_id": req.TargetID,
		"damage":    skillLog.Damage,
		"heal":      skillLog.Heal,
		"is_combo":  skillLog.IsCombo,
	}

	h.sendResponse(msgObj.Session, msg.MSG_SKILL_RES, response)

	if battle.CheckBattleEnd() {
		rewards := battle.GetRewards()
		levelUpInfo := h.processBattleRewards(battle, rewards)
		h.battleManager.EndBattle(req.BattleID)

		battleEndResponse := map[string]interface{}{
			"result":      0,
			"message":     "battle ended",
			"battle_id":   req.BattleID,
			"rewards":     rewards,
			"level_up":    levelUpInfo.leveledUp,
			"level_count": levelUpInfo.levelCount,
			"new_level":   levelUpInfo.newLevel,
		}
		h.sendResponse(msgObj.Session, msg.MSG_BATTLE_END_RES, battleEndResponse)
	}
}

type LevelUpInfo struct {
	leveledUp  bool
	levelCount int32
	newLevel   int32
}

func (h *BattleHandler) processBattleRewards(battle *battlecore.Battle, rewards map[int64]int64) LevelUpInfo {
	info := LevelUpInfo{
		leveledUp:  false,
		levelCount: 0,
		newLevel:   0,
	}

	for playerID, exp := range rewards {
		player, ok := battle.GetPlayer(playerID)
		if !ok {
			continue
		}

		leveledUp, levelCount := player.AddExp(exp)
		if leveledUp {
			info.leveledUp = true
			info.levelCount += levelCount
			info.newLevel = player.GetLevel()
			log.Info("Player ", playerID, " leveled up! New level: ", info.newLevel, ", Total levels gained: ", info.levelCount)
		}
	}

	return info
}

func (h *BattleHandler) handleBattleEnd(msgObj *network.Message) {
	var req struct {
		BattleID int64 `json:"battle_id"`
	}

	err := json.Unmarshal(msgObj.Data, &req)
	if err != nil {
		log.Error("Failed to unmarshal battle end request: ", err)
		return
	}

	h.battleManager.EndBattle(req.BattleID)
	log.Info("Ended battle: ", req.BattleID)

	response := map[string]interface{}{
		"result":    0,
		"message":   "battle ended",
		"battle_id": req.BattleID,
	}

	h.sendResponse(msgObj.Session, msg.MSG_BATTLE_END_RES, response)
}

func (h *BattleHandler) sendResponse(conn interface{ Write([]byte) (int, error) }, msgID uint32, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Error("Failed to marshal response: ", err)
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
