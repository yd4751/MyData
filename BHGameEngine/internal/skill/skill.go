package skill

import (
	"sync"
	"time"

	"github.com/openworld-server/internal/timer"
	"github.com/openworld-server/pkg/snowflake"
)

type SkillType int32

const (
	SkillTypeActive  SkillType = 1
	SkillTypePassive SkillType = 2
	SkillTypeCombo   SkillType = 3
)

type SkillClass int32

const (
	SkillClassPhysical SkillClass = 1
	SkillClassMagic    SkillClass = 2
	SkillClassHealing  SkillClass = 3
	SkillClassBuff     SkillClass = 4
	SkillClassDebuff   SkillClass = 5
)

type TargetType int32

const (
	TargetTypeSingle TargetType = 1
	TargetTypeMulti  TargetType = 2
	TargetTypeArea   TargetType = 3
	TargetTypeSelf   TargetType = 4
	TargetTypeParty  TargetType = 5
	TargetTypeEnemy  TargetType = 6
)

type EffectType int32

const (
	EffectTypeDamage       EffectType = 1
	EffectTypeHeal         EffectType = 2
	EffectTypeBuff         EffectType = 3
	EffectTypeDebuff       EffectType = 4
	EffectTypeDot          EffectType = 5
	EffectTypeHot          EffectType = 6
	EffectTypeStun         EffectType = 7
	EffectTypeSilence      EffectType = 8
	EffectTypeKnockback    EffectType = 9
	EffectTypeSpeedBuff    EffectType = 10
	EffectTypeDamageShield EffectType = 11
)

type SkillConfig struct {
	SkillID     int64
	SkillName   string
	SkillType   SkillType
	SkillClass  SkillClass
	MaxLevel    int
	Cooldown    int
	ManaCost    int
	Range       float32
	TargetType  TargetType
	Description string
	Effects     []*EffectConfig
}

type EffectConfig struct {
	EffectID     int64
	EffectType   EffectType
	Value        float32
	Duration     int
	TickInterval int
	StackCount   int
}

type PlayerSkill struct {
	SkillID     int64
	Level       int
	LastCast    int64
	CooldownEnd int64
	IsUnlocked  bool
}

type ComboSequence struct {
	SkillIDs    []int64
	NextIndex   int
	LastTime    int64
	MaxDelay    int
	BonusEffect *EffectConfig
}

type SkillManager struct {
	skillConfigs   map[int64]*SkillConfig
	playerSkills   map[int64]map[int64]*PlayerSkill
	comboSequences map[int64]*ComboSequence
	timerManager   *timer.TimerManager
	mu             sync.RWMutex
}

func (m *SkillManager) TimerManager() *timer.TimerManager {
	return m.timerManager
}

var skillManagerInstance *SkillManager
var once sync.Once

func GetSkillManager() *SkillManager {
	once.Do(func() {
		skillManagerInstance = &SkillManager{
			skillConfigs:   make(map[int64]*SkillConfig),
			playerSkills:   make(map[int64]map[int64]*PlayerSkill),
			comboSequences: make(map[int64]*ComboSequence),
			timerManager:   timer.NewTimerManager(),
		}
		skillManagerInstance.timerManager.Start()
		LoadDefaultSkills()
	})
	return skillManagerInstance
}

func LoadDefaultSkills() {
	RegisterSkillConfig(&SkillConfig{
		SkillID:     1001,
		SkillName:   "猛击",
		SkillType:   SkillTypeActive,
		SkillClass:  SkillClassPhysical,
		MaxLevel:    10,
		Cooldown:    3000,
		ManaCost:    10,
		Range:       3,
		TargetType:  TargetTypeSingle,
		Description: "向目标发起猛击，造成物理伤害",
		Effects: []*EffectConfig{
			{
				EffectID:     1,
				EffectType:   EffectTypeDamage,
				Value:        50,
				Duration:     0,
				TickInterval: 0,
				StackCount:   0,
			},
		},
	})

	RegisterSkillConfig(&SkillConfig{
		SkillID:     1002,
		SkillName:   "火球术",
		SkillType:   SkillTypeActive,
		SkillClass:  SkillClassMagic,
		MaxLevel:    10,
		Cooldown:    2000,
		ManaCost:    20,
		Range:       15,
		TargetType:  TargetTypeSingle,
		Description: "发射火球攻击目标，造成魔法伤害",
		Effects: []*EffectConfig{
			{
				EffectID:     2,
				EffectType:   EffectTypeDamage,
				Value:        60,
				Duration:     3000,
				TickInterval: 1000,
				StackCount:   3,
			},
		},
	})

	RegisterSkillConfig(&SkillConfig{
		SkillID:     1003,
		SkillName:   "治愈术",
		SkillType:   SkillTypeActive,
		SkillClass:  SkillClassHealing,
		MaxLevel:    10,
		Cooldown:    4000,
		ManaCost:    25,
		Range:       10,
		TargetType:  TargetTypeSingle,
		Description: "为目标恢复生命值",
		Effects: []*EffectConfig{
			{
				EffectID:     3,
				EffectType:   EffectTypeHeal,
				Value:        80,
				Duration:     0,
				TickInterval: 0,
				StackCount:   0,
			},
		},
	})

	RegisterSkillConfig(&SkillConfig{
		SkillID:     2001,
		SkillName:   "力量提升",
		SkillType:   SkillTypePassive,
		SkillClass:  SkillClassBuff,
		MaxLevel:    5,
		Cooldown:    0,
		ManaCost:    0,
		Range:       0,
		TargetType:  TargetTypeSelf,
		Description: "被动提升角色力量属性",
		Effects: []*EffectConfig{
			{
				EffectID:     4,
				EffectType:   EffectTypeBuff,
				Value:        5,
				Duration:     0,
				TickInterval: 0,
				StackCount:   0,
			},
		},
	})

	RegisterSkillConfig(&SkillConfig{
		SkillID:     2002,
		SkillName:   "生命恢复",
		SkillType:   SkillTypePassive,
		SkillClass:  SkillClassHealing,
		MaxLevel:    5,
		Cooldown:    0,
		ManaCost:    0,
		Range:       0,
		TargetType:  TargetTypeSelf,
		Description: "每秒恢复少量生命值",
		Effects: []*EffectConfig{
			{
				EffectID:     5,
				EffectType:   EffectTypeHot,
				Value:        5,
				Duration:     0,
				TickInterval: 1000,
				StackCount:   0,
			},
		},
	})

	RegisterSkillConfig(&SkillConfig{
		SkillID:     3001,
		SkillName:   "烈焰风暴",
		SkillType:   SkillTypeCombo,
		SkillClass:  SkillClassMagic,
		MaxLevel:    5,
		Cooldown:    10000,
		ManaCost:    50,
		Range:       8,
		TargetType:  TargetTypeArea,
		Description: "召唤烈焰风暴，对范围内敌人造成持续伤害",
		Effects: []*EffectConfig{
			{
				EffectID:     6,
				EffectType:   EffectTypeDot,
				Value:        30,
				Duration:     5000,
				TickInterval: 1000,
				StackCount:   0,
			},
		},
	})
}

func RegisterSkillConfig(config *SkillConfig) {
	manager := GetSkillManager()
	manager.mu.Lock()
	manager.skillConfigs[config.SkillID] = config
	manager.mu.Unlock()
}

func GetSkillConfig(skillID int64) (*SkillConfig, bool) {
	manager := GetSkillManager()
	manager.mu.RLock()
	config, ok := manager.skillConfigs[skillID]
	manager.mu.RUnlock()
	return config, ok
}

func (m *SkillManager) InitializePlayerSkills(playerID int64) {
	m.mu.Lock()
	if _, ok := m.playerSkills[playerID]; !ok {
		m.playerSkills[playerID] = make(map[int64]*PlayerSkill)
	}
	m.mu.Unlock()
}

func (m *SkillManager) LearnSkill(playerID int64, skillID int64) error {
	config, ok := GetSkillConfig(skillID)
	if !ok {
		return ErrSkillNotFound
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.playerSkills[playerID]; !ok {
		m.playerSkills[playerID] = make(map[int64]*PlayerSkill)
	}

	if existing, ok := m.playerSkills[playerID][skillID]; ok {
		if existing.Level >= config.MaxLevel {
			return ErrSkillMaxLevel
		}
		existing.Level++
	} else {
		m.playerSkills[playerID][skillID] = &PlayerSkill{
			SkillID:     skillID,
			Level:       1,
			LastCast:    0,
			CooldownEnd: 0,
			IsUnlocked:  true,
		}
	}

	return nil
}

func (m *SkillManager) UpgradeSkill(playerID int64, skillID int64) (int, error) {
	config, ok := GetSkillConfig(skillID)
	if !ok {
		return 0, ErrSkillNotFound
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	playerSkills, ok := m.playerSkills[playerID]
	if !ok {
		return 0, ErrSkillNotLearned
	}

	playerSkill, ok := playerSkills[skillID]
	if !ok {
		return 0, ErrSkillNotLearned
	}

	if playerSkill.Level >= config.MaxLevel {
		return 0, ErrSkillMaxLevel
	}

	playerSkill.Level++
	return playerSkill.Level, nil
}

func (m *SkillManager) GetPlayerSkill(playerID int64, skillID int64) (*PlayerSkill, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	playerSkills, ok := m.playerSkills[playerID]
	if !ok {
		return nil, false
	}

	skill, ok := playerSkills[skillID]
	return skill, ok
}

func (m *SkillManager) GetPlayerSkillList(playerID int64) []*PlayerSkill {
	m.mu.RLock()
	defer m.mu.RUnlock()

	playerSkills, ok := m.playerSkills[playerID]
	if !ok {
		return nil
	}

	skills := make([]*PlayerSkill, 0, len(playerSkills))
	for _, skill := range playerSkills {
		skills = append(skills, skill)
	}
	return skills
}

func (m *SkillManager) CheckCooldown(playerID int64, skillID int64) bool {
	playerSkill, ok := m.GetPlayerSkill(playerID, skillID)
	if !ok {
		return false
	}

	now := time.Now().UnixNano() / 1e6
	return now >= playerSkill.CooldownEnd
}

func (m *SkillManager) SetCooldown(playerID int64, skillID int64, cooldownMs int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	playerSkills, ok := m.playerSkills[playerID]
	if !ok {
		return
	}

	playerSkill, ok := playerSkills[skillID]
	if !ok {
		return
	}

	now := time.Now().UnixNano() / 1e6
	playerSkill.CooldownEnd = now + int64(cooldownMs)
	playerSkill.LastCast = now
}

type SkillCaster interface {
	GetID() int64
	ConsumeMana(amount int32) bool
}

func (m *SkillManager) CastSkill(caster SkillCaster, skillID int64, targetID int64) (*SkillResult, error) {
	config, ok := GetSkillConfig(skillID)
	if !ok {
		return nil, ErrSkillNotFound
	}

	playerSkill, ok := m.GetPlayerSkill(caster.GetID(), skillID)
	if !ok {
		return nil, ErrSkillNotLearned
	}

	if config.SkillType == SkillTypePassive {
		return nil, ErrCannotCastPassive
	}

	if !m.CheckCooldown(caster.GetID(), skillID) {
		return nil, ErrSkillOnCooldown
	}

	if !caster.ConsumeMana(int32(config.ManaCost)) {
		return nil, ErrInsufficientMana
	}

	m.SetCooldown(caster.GetID(), skillID, config.Cooldown)

	return m.ExecuteSkill(caster.GetID(), config, playerSkill.Level, targetID)
}

func (m *SkillManager) ExecuteSkill(casterID int64, config *SkillConfig, level int, targetID int64) (*SkillResult, error) {
	result := &SkillResult{
		SkillID:   config.SkillID,
		TargetID:  targetID,
		Effects:   make([]*EffectResult, 0),
		IsCombo:   false,
		ComboName: "",
	}

	effectMultiplier := float32(1.0 + float32(level-1)*0.1)

	for _, effectConfig := range config.Effects {
		effectResult := m.ApplyEffect(effectConfig, effectMultiplier)
		result.Effects = append(result.Effects, effectResult)
		m.ApplyEffectToTarget(targetID, effectResult, effectConfig, effectMultiplier)
	}

	m.CheckCombo(casterID, config.SkillID, result)

	return result, nil
}

func (m *SkillManager) ApplyEffect(config *EffectConfig, multiplier float32) *EffectResult {
	return &EffectResult{
		EffectID:   config.EffectID,
		EffectType: config.EffectType,
		Value:      config.Value * multiplier,
		Duration:   config.Duration,
	}
}

func (m *SkillManager) ApplyEffectToTarget(targetID int64, effect *EffectResult, config *EffectConfig, multiplier float32) {
	switch effect.EffectType {
	case EffectTypeDamage:
		targetEntity := GetTargetEntity(targetID)
		if targetEntity != nil {
			targetEntity.TakeDamage(int32(effect.Value))
		}
	case EffectTypeHeal:
		targetEntity := GetTargetEntity(targetID)
		if targetEntity != nil {
			targetEntity.Heal(int32(effect.Value))
		}
	case EffectTypeDot:
		tickValue := config.Value * multiplier
		tickCount := config.Duration / config.TickInterval

		taskID := snowflake.GenerateID()
		tick := 0

		m.timerManager.AddTask(taskID, 0, time.Duration(config.TickInterval)*time.Millisecond, func() {
			tick++
			targetEntity := GetTargetEntity(targetID)
			if targetEntity != nil {
				targetEntity.TakeDamage(int32(tickValue))
			}
			if tick >= tickCount {
				m.timerManager.RemoveTask(taskID)
			}
		})
	case EffectTypeHot:
		tickValue := config.Value * multiplier

		taskID := snowflake.GenerateID()
		m.timerManager.AddTask(taskID, 0, time.Duration(config.TickInterval)*time.Millisecond, func() {
			targetEntity := GetTargetEntity(targetID)
			if targetEntity != nil {
				targetEntity.Heal(int32(tickValue))
			}
		})
	}
}

func (m *SkillManager) CheckCombo(playerID int64, skillID int64, result *SkillResult) {
	combo, ok := m.comboSequences[playerID]
	if !ok {
		return
	}

	now := time.Now().UnixNano() / 1e6
	if now-combo.LastTime > int64(combo.MaxDelay) {
		combo.NextIndex = 0
	}

	if combo.NextIndex < len(combo.SkillIDs) && combo.SkillIDs[combo.NextIndex] == skillID {
		combo.NextIndex++
		combo.LastTime = now

		if combo.NextIndex == len(combo.SkillIDs) {
			result.IsCombo = true
			result.ComboName = "Combo!"
			if combo.BonusEffect != nil {
				effectResult := &EffectResult{
					EffectID:   combo.BonusEffect.EffectID,
					EffectType: combo.BonusEffect.EffectType,
					Value:      combo.BonusEffect.Value,
					Duration:   combo.BonusEffect.Duration,
				}
				result.Effects = append(result.Effects, effectResult)
			}
			combo.NextIndex = 0
		}
	} else {
		combo.NextIndex = 0
	}
}

func (m *SkillManager) RegisterCombo(playerID int64, skillIDs []int64, maxDelay int, bonusEffect *EffectConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.comboSequences[playerID] = &ComboSequence{
		SkillIDs:    skillIDs,
		NextIndex:   0,
		LastTime:    0,
		MaxDelay:    maxDelay,
		BonusEffect: bonusEffect,
	}
}

var entityGetter func(targetID int64) EntityTarget

func SetEntityGetter(getter func(targetID int64) EntityTarget) {
	entityGetter = getter
}

func GetTargetEntity(targetID int64) EntityTarget {
	if entityGetter != nil {
		return entityGetter(targetID)
	}
	return nil
}

type EntityTarget interface {
	TakeDamage(damage int32)
	Heal(amount int32)
}
