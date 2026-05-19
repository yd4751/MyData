package battlecore

import (
	"math"
	"sync"
	"time"

	"github.com/openworld-server/internal/entity"
	"github.com/openworld-server/internal/player"
	"github.com/openworld-server/internal/skill"
	"github.com/openworld-server/internal/worldmap"
	"github.com/openworld-server/pkg/snowflake"
)

type BattleState int32 // 战斗状态

const (
	BattleStateActive BattleState = 1 // 进行中
	BattleStateEnded  BattleState = 2 // 已结束
)

// Battle 战斗实例
type Battle struct {
	ID          int64                     // 战斗ID
	Players     map[int64]*player.Player  // 参战玩家
	Monsters    map[int64]*entity.Monster // 参战怪物
	StartTime   int64                     // 开始时间
	EndTime     int64                     // 结束时间
	State       BattleState               // 战斗状态
	CombatLogs  []*CombatLog              // 战斗日志
	ComboTracks map[int64]*ComboTracker   // 连招追踪
	mu          sync.RWMutex              // 并发锁
}

// ComboTracker 连招追踪器
type ComboTracker struct {
	SkillSequence []int64 // 技能序列
	CurrentIndex  int     // 当前索引
	LastCastTime  int64   // 上次释放时间
	MaxDelay      int     // 最大间隔(毫秒)
}

// SkillCombatLog 技能战斗日志
type SkillCombatLog struct {
	Timestamp int64  // 时间戳
	Attacker  int64  // 攻击者ID
	Target    int64  // 目标ID
	SkillID   int64  // 技能ID
	Damage    int32  // 伤害值
	Heal      int32  // 治疗值
	IsCombo   bool   // 是否连招
	Effect    string // 效果描述
}

// CombatLog 战斗日志
type CombatLog struct {
	Timestamp int64  // 时间戳
	Attacker  int64  // 攻击者ID
	Target    int64  // 目标ID
	SkillID   int64  // 技能ID
	Damage    int32  // 伤害值
	Heal      int32  // 治疗值
	Effect    string // 效果描述
}

// BattleManager 战斗管理器
type BattleManager struct {
	battles map[int64]*Battle // 战斗列表
	mu      sync.RWMutex      // 并发锁
}

var battleManagerInstance *BattleManager
var battleOnce sync.Once

func GetBattleManager() *BattleManager {
	battleOnce.Do(func() {
		battleManagerInstance = &BattleManager{
			battles: make(map[int64]*Battle),
		}
	})
	return battleManagerInstance
}

func (m *BattleManager) CreateBattle(battleID int64) *Battle {
	battle := &Battle{
		ID:          battleID,
		Players:     make(map[int64]*player.Player),
		Monsters:    make(map[int64]*entity.Monster),
		StartTime:   time.Now().Unix(),
		State:       BattleStateActive,
		CombatLogs:  make([]*CombatLog, 0),
		ComboTracks: make(map[int64]*ComboTracker),
	}

	m.mu.Lock()
	m.battles[battleID] = battle
	m.mu.Unlock()

	return battle
}

func (m *BattleManager) GetBattle(battleID int64) (*Battle, bool) {
	m.mu.RLock()
	battle, ok := m.battles[battleID]
	m.mu.RUnlock()
	return battle, ok
}

func (m *BattleManager) EndBattle(battleID int64) {
	m.mu.Lock()
	battle, ok := m.battles[battleID]
	if ok {
		battle.mu.Lock()
		battle.State = BattleStateEnded
		battle.EndTime = time.Now().Unix()
		battle.mu.Unlock()
		delete(m.battles, battleID)
	}
	m.mu.Unlock()
}

func (b *Battle) AddPlayer(p *player.Player) {
	b.mu.Lock()
	b.Players[p.ID] = p
	b.ComboTracks[p.ID] = &ComboTracker{
		SkillSequence: make([]int64, 0),
		CurrentIndex:  0,
		MaxDelay:      3000,
	}
	b.mu.Unlock()
}

func (b *Battle) RemovePlayer(playerID int64) {
	b.mu.Lock()
	delete(b.Players, playerID)
	delete(b.ComboTracks, playerID)
	if len(b.Players) == 0 && len(b.Monsters) == 0 {
		b.State = BattleStateEnded
	}
	b.mu.Unlock()
}

func (b *Battle) AddMonster(m *entity.Monster) {
	b.mu.Lock()
	b.Monsters[m.ID] = m
	b.mu.Unlock()
}

func (b *Battle) RemoveMonster(monsterID int64) {
	b.mu.Lock()
	delete(b.Monsters, monsterID)
	if len(b.Players) == 0 && len(b.Monsters) == 0 {
		b.State = BattleStateEnded
	}
	b.mu.Unlock()
}

func (b *Battle) IsActive() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.State == BattleStateActive
}

func (b *Battle) GetPlayer(playerID int64) (*player.Player, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	p, ok := b.Players[playerID]
	return p, ok
}

func (b *Battle) GetMonster(monsterID int64) (*entity.Monster, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	m, ok := b.Monsters[monsterID]
	return m, ok
}

func (b *Battle) AddCombatLog(log *CombatLog) {
	b.mu.Lock()
	b.CombatLogs = append(b.CombatLogs, log)
	if len(b.CombatLogs) > 100 {
		b.CombatLogs = b.CombatLogs[len(b.CombatLogs)-100:]
	}
	b.mu.Unlock()
}

func CalculateDamage(attacker *player.Player, target *entity.Monster, skillConfig *skill.SkillConfig, skillLevel int) int32 {
	baseDamage := float32(0)
	for _, effect := range skillConfig.Effects {
		if effect.EffectType == skill.EffectTypeDamage {
			baseDamage += effect.Value
		}
	}

	levelBonus := float32(attacker.GetLevel()) * 2
	baseDamage += levelBonus

	var attributeBonus float32
	if skillConfig.SkillClass == skill.SkillClassPhysical {
		attributeBonus = float32(attacker.Strength) * 0.5
	} else if skillConfig.SkillClass == skill.SkillClassMagic {
		attributeBonus = float32(attacker.Intelligence) * 0.5
	}
	baseDamage += attributeBonus

	levelMultiplier := float32(1.0 + float32(skillLevel-1)*0.1)
	totalDamage := baseDamage * levelMultiplier

	armor := float32(target.Level * 5)
	armorReduction := armor / (armor + 100)
	finalDamage := totalDamage * (1 - armorReduction)

	if finalDamage < 1 {
		finalDamage = 1
	}

	return int32(finalDamage)
}

func CalculatePlayerDamage(attacker *player.Player, target *player.Player, skillConfig *skill.SkillConfig, skillLevel int) int32 {
	baseDamage := float32(0)
	for _, effect := range skillConfig.Effects {
		if effect.EffectType == skill.EffectTypeDamage {
			baseDamage += effect.Value
		}
	}

	levelBonus := float32(attacker.GetLevel()) * 2
	baseDamage += levelBonus

	var attributeBonus float32
	if skillConfig.SkillClass == skill.SkillClassPhysical {
		attributeBonus = float32(attacker.Strength) * 0.5
	} else if skillConfig.SkillClass == skill.SkillClassMagic {
		attributeBonus = float32(attacker.Intelligence) * 0.5
	}
	baseDamage += attributeBonus

	levelMultiplier := float32(1.0 + float32(skillLevel-1)*0.1)
	totalDamage := baseDamage * levelMultiplier

	armor := float32(target.Defense)
	armorReduction := armor / (armor + 100)
	finalDamage := totalDamage * (1 - armorReduction)

	if finalDamage < 1 {
		finalDamage = 1
	}

	return int32(finalDamage)
}

func CalculateHeal(caster *player.Player, target *player.Player, skillConfig *skill.SkillConfig, skillLevel int) int32 {
	baseHeal := float32(0)
	for _, effect := range skillConfig.Effects {
		if effect.EffectType == skill.EffectTypeHeal {
			baseHeal += effect.Value
		}
	}

	intellectBonus := float32(caster.Intelligence) * 0.3
	baseHeal += intellectBonus

	levelMultiplier := float32(1.0 + float32(skillLevel-1)*0.1)
	totalHeal := baseHeal * levelMultiplier

	return int32(totalHeal)
}

func IsInRange(pos1, pos2 worldmap.Vec3, r float32) bool {
	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	dz := pos1.Z - pos2.Z
	dist := float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
	return dist <= r
}

func (b *Battle) CastSkill(caster *player.Player, skillID int64, targetID int64) (*SkillCombatLog, error) {
	skillConfig, ok := skill.GetSkillConfig(skillID)
	if !ok {
		return nil, skill.ErrSkillNotFound
	}

	playerSkill, ok := skill.GetSkillManager().GetPlayerSkill(caster.ID, skillID)
	if !ok {
		return nil, skill.ErrSkillNotLearned
	}

	if skillConfig.SkillType == skill.SkillTypePassive {
		return nil, skill.ErrCannotCastPassive
	}

	if !skill.GetSkillManager().CheckCooldown(caster.ID, skillID) {
		return nil, skill.ErrSkillOnCooldown
	}

	if !caster.ConsumeMana(int32(skillConfig.ManaCost)) {
		return nil, skill.ErrInsufficientMana
	}

	skill.GetSkillManager().SetCooldown(caster.ID, skillID, skillConfig.Cooldown)

	var target *entity.Monster
	var targetPlayer *player.Player
	var isPlayerTarget bool

	b.mu.RLock()
	target, monsterOk := b.Monsters[targetID]
	if !monsterOk {
		targetPlayer, isPlayerTarget = b.Players[targetID]
	}
	b.mu.RUnlock()

	if target == nil && targetPlayer == nil {
		return nil, skill.ErrInvalidTarget
	}

	var targetPos worldmap.Vec3
	if isPlayerTarget {
		targetPos = targetPlayer.GetPosition()
	} else {
		targetPos = target.GetPosition()
	}

	if skillConfig.Range > 0 && !IsInRange(caster.GetPosition(), targetPos, skillConfig.Range) {
		return nil, skill.ErrTargetOutOfRange
	}

	return b.ExecuteSkill(caster, skillConfig, playerSkill.Level, targetID, isPlayerTarget)
}

func (b *Battle) ExecuteSkill(caster *player.Player, skillConfig *skill.SkillConfig, skillLevel int, targetID int64, isPlayerTarget bool) (*SkillCombatLog, error) {
	log := &SkillCombatLog{
		Timestamp: time.Now().Unix(),
		Attacker:  caster.ID,
		Target:    targetID,
		SkillID:   skillConfig.SkillID,
	}

	for _, effectConfig := range skillConfig.Effects {
		switch effectConfig.EffectType {
		case skill.EffectTypeDamage:
			if isPlayerTarget {
				b.mu.RLock()
				target, ok := b.Players[targetID]
				b.mu.RUnlock()
				if ok {
					damage := CalculatePlayerDamage(caster, target, skillConfig, skillLevel)
					target.TakeDamage(damage)
					log.Damage += damage
				}
			} else {
				b.mu.RLock()
				target, ok := b.Monsters[targetID]
				b.mu.RUnlock()
				if ok {
					damage := CalculateDamage(caster, target, skillConfig, skillLevel)
					target.TakeDamage(damage)
					log.Damage += damage
				}
			}

		case skill.EffectTypeHeal:
			b.mu.RLock()
			target, ok := b.Players[targetID]
			b.mu.RUnlock()
			if ok {
				heal := CalculateHeal(caster, target, skillConfig, skillLevel)
				target.Heal(heal)
				log.Heal += heal
			}

		case skill.EffectTypeBuff:
			b.mu.RLock()
			target, ok := b.Players[targetID]
			b.mu.RUnlock()
			if ok {
				buff := &player.Buff{
					ID:      int32(effectConfig.EffectID),
					Name:    skillConfig.SkillName,
					Stacks:  1,
					EndTime: time.Now().UnixNano()/1e6 + int64(effectConfig.Duration),
				}
				target.ApplyBuff(buff)
			}

		case skill.EffectTypeDebuff:
			if isPlayerTarget {
				b.mu.RLock()
				target, ok := b.Players[targetID]
				b.mu.RUnlock()
				if ok {
					buff := &player.Buff{
						ID:      int32(effectConfig.EffectID),
						Name:    skillConfig.SkillName,
						Stacks:  1,
						EndTime: time.Now().UnixNano()/1e6 + int64(effectConfig.Duration),
					}
					target.ApplyBuff(buff)
				}
			}

		case skill.EffectTypeDot:
			b.applyDotEffect(caster.ID, targetID, effectConfig, skillLevel, isPlayerTarget)

		case skill.EffectTypeHot:
			b.applyHotEffect(targetID, effectConfig, skillLevel)

		case skill.EffectTypeStun:
			log.Effect = "stun"

		case skill.EffectTypeSilence:
			log.Effect = "silence"

		case skill.EffectTypeKnockback:
			log.Effect = "knockback"

		case skill.EffectTypeDamageShield:
			b.mu.RLock()
			target, ok := b.Players[targetID]
			b.mu.RUnlock()
			if ok {
				buff := &player.Buff{
					ID:      int32(effectConfig.EffectID),
					Name:    "Damage Shield",
					Stacks:  int32(effectConfig.Value),
					EndTime: time.Now().UnixNano()/1e6 + int64(effectConfig.Duration),
				}
				target.ApplyBuff(buff)
			}
		}
	}

	log.IsCombo = b.CheckCombo(caster.ID, skillConfig.SkillID)

	combatLog := &CombatLog{
		Timestamp: log.Timestamp,
		Attacker:  log.Attacker,
		Target:    log.Target,
		SkillID:   log.SkillID,
		Damage:    log.Damage,
		Heal:      log.Heal,
	}
	b.AddCombatLog(combatLog)

	return log, nil
}

func (b *Battle) applyDotEffect(casterID int64, targetID int64, effectConfig *skill.EffectConfig, skillLevel int, isPlayerTarget bool) {
	effectMultiplier := float32(1.0 + float32(skillLevel-1)*0.1)
	tickValue := effectConfig.Value * effectMultiplier
	tickCount := effectConfig.Duration / effectConfig.TickInterval

	taskID := snowflake.GenerateID()
	tick := 0

	timer := skill.GetSkillManager().TimerManager()
	timer.AddTask(taskID, 0, time.Duration(effectConfig.TickInterval)*time.Millisecond, func() {
		tick++
		b.mu.RLock()
		var target *entity.Monster
		var targetPlayer *player.Player
		if isPlayerTarget {
			targetPlayer, _ = b.Players[targetID]
		} else {
			target, _ = b.Monsters[targetID]
		}
		b.mu.RUnlock()

		if isPlayerTarget && targetPlayer != nil {
			targetPlayer.TakeDamage(int32(tickValue))
			log := &CombatLog{
				Timestamp: time.Now().Unix(),
				Attacker:  casterID,
				Target:    targetID,
				SkillID:   effectConfig.EffectID,
				Damage:    int32(tickValue),
				Effect:    "DoT",
			}
			b.AddCombatLog(log)
		} else if !isPlayerTarget && target != nil {
			target.TakeDamage(int32(tickValue))
			log := &CombatLog{
				Timestamp: time.Now().Unix(),
				Attacker:  casterID,
				Target:    targetID,
				SkillID:   effectConfig.EffectID,
				Damage:    int32(tickValue),
				Effect:    "DoT",
			}
			b.AddCombatLog(log)
		}

		if tick >= tickCount {
			timer.RemoveTask(taskID)
		}
	})
}

func (b *Battle) applyHotEffect(targetID int64, effectConfig *skill.EffectConfig, skillLevel int) {
	effectMultiplier := float32(1.0 + float32(skillLevel-1)*0.1)
	tickValue := effectConfig.Value * effectMultiplier

	taskID := snowflake.GenerateID()
	timer := skill.GetSkillManager().TimerManager()

	timer.AddTask(taskID, 0, time.Duration(effectConfig.TickInterval)*time.Millisecond, func() {
		b.mu.RLock()
		target, ok := b.Players[targetID]
		b.mu.RUnlock()

		if ok {
			target.Heal(int32(tickValue))
			log := &CombatLog{
				Timestamp: time.Now().Unix(),
				Attacker:  0,
				Target:    targetID,
				SkillID:   effectConfig.EffectID,
				Heal:      int32(tickValue),
				Effect:    "HoT",
			}
			b.AddCombatLog(log)
		}
	})
}

func (b *Battle) CheckCombo(playerID int64, skillID int64) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	tracker, ok := b.ComboTracks[playerID]
	if !ok {
		return false
	}

	now := time.Now().UnixNano() / 1e6
	if now-tracker.LastCastTime > int64(tracker.MaxDelay) {
		tracker.CurrentIndex = 0
		tracker.SkillSequence = make([]int64, 0)
	}

	tracker.SkillSequence = append(tracker.SkillSequence, skillID)
	tracker.LastCastTime = now

	if len(tracker.SkillSequence) >= 3 {
		tracker.SkillSequence = tracker.SkillSequence[len(tracker.SkillSequence)-3:]
	}

	if len(tracker.SkillSequence) >= 3 {
		tracker.CurrentIndex = 0
		tracker.SkillSequence = make([]int64, 0)
		return true
	}

	return false
}

func (b *Battle) UpdateBuffs() {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, p := range b.Players {
		p.CleanExpiredBuffs()
	}
}

func (b *Battle) CheckBattleEnd() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	allMonstersDead := true
	for _, m := range b.Monsters {
		if m.GetState() != entity.EntityStateDead {
			allMonstersDead = false
			break
		}
	}

	allPlayersDead := true
	for _, p := range b.Players {
		if p.GetState() != player.StateDead {
			allPlayersDead = false
			break
		}
	}

	return allMonstersDead || allPlayersDead
}

func (b *Battle) GetRewards() map[int64]int64 {
	rewards := make(map[int64]int64)
	b.mu.RLock()
	for _, monster := range b.Monsters {
		isDead := monster.GetState() == entity.EntityStateDead
		expReward := monster.ExpReward

		if isDead {
			for playerID := range b.Players {
				rewards[playerID] += expReward
			}
		}
	}
	b.mu.RUnlock()
	return rewards
}

func ProcessMonsterAttack(monster *entity.Monster, player *player.Player) *CombatLog {
	now := time.Now().UnixNano() / 1e6
	if now-monster.LastAttackTime < 1000 {
		return nil
	}

	if !IsInRange(monster.GetPosition(), player.GetPosition(), 3) {
		return nil
	}

	damage := int32(monster.Level * 5)
	armorReduction := float32(player.Defense) / (float32(player.Defense) + 100)
	finalDamage := int32(float32(damage) * (1 - armorReduction))
	if finalDamage < 1 {
		finalDamage = 1
	}

	player.TakeDamage(finalDamage)
	monster.LastAttackTime = now

	return &CombatLog{
		Timestamp: now / 1e6,
		Attacker:  monster.ID,
		Target:    player.ID,
		SkillID:   0,
		Damage:    finalDamage,
	}
}
