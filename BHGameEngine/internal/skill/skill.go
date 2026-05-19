package skill

import (
	"sync"
	"time"

	"github.com/openworld-server/internal/timer"
	"github.com/openworld-server/pkg/snowflake"
)

type SkillType int32 // 技能类型

const (
	SkillTypeActive  SkillType = 1 // 主动技能，需要手动释放
	SkillTypePassive SkillType = 2 // 被动技能，持续生效
	SkillTypeCombo   SkillType = 3 // 连招技能，需按特定顺序释放
)

type SkillClass int32 // 技能分类

const (
	SkillClassPhysical SkillClass = 1 // 物理类技能
	SkillClassMagic    SkillClass = 2 // 魔法类技能
	SkillClassHealing  SkillClass = 3 // 治疗类技能
	SkillClassBuff     SkillClass = 4 // 增益类技能
	SkillClassDebuff   SkillClass = 5 // 减益类技能
)

type TargetType int32 // 目标类型

const (
	TargetTypeSingle TargetType = 1 // 单体目标
	TargetTypeMulti  TargetType = 2 // 多个目标
	TargetTypeArea   TargetType = 3 // 区域范围
	TargetTypeSelf   TargetType = 4 // 自身
	TargetTypeParty  TargetType = 5 // 队伍成员
	TargetTypeEnemy  TargetType = 6 // 敌方目标
)

type EffectType int32 // 效果类型

const (
	EffectTypeDamage       EffectType = 1  // 直接伤害
	EffectTypeHeal         EffectType = 2  // 直接治疗
	EffectTypeBuff         EffectType = 3  // 属性增益
	EffectTypeDebuff       EffectType = 4  // 属性减益
	EffectTypeDot          EffectType = 5  // 持续伤害(Damage over Time)
	EffectTypeHot          EffectType = 6  // 持续治疗(Heal over Time)
	EffectTypeStun         EffectType = 7  // 眩晕控制
	EffectTypeSilence      EffectType = 8  // 沉默控制
	EffectTypeKnockback    EffectType = 9  // 击退效果
	EffectTypeSpeedBuff    EffectType = 10 // 移动速度加成
	EffectTypeDamageShield EffectType = 11 // 伤害护盾
)

type SkillConfig struct {
	SkillID     int64           // 技能唯一ID
	SkillName   string          // 技能名称
	SkillType   SkillType       // 技能类型（主动/被动/连招）
	SkillClass  SkillClass      // 技能分类（物理/魔法/治疗等）
	MaxLevel    int             // 最高等级
	Cooldown    int             // 冷却时间(毫秒)
	ManaCost    int             // 魔法消耗
	Range       float32         // 技能射程
	TargetType  TargetType      // 目标类型
	Description string          // 技能描述
	Effects     []*EffectConfig // 技能效果列表
}

type EffectConfig struct {
	EffectID     int64      // 效果唯一ID
	EffectType   EffectType // 效果类型
	Value        float32    // 效果数值
	Duration     int        // 持续时间(毫秒)，0表示瞬时
	TickInterval int        // 周期性效果的触发间隔(毫秒)
	StackCount   int        // 可堆叠层数
}

type PlayerSkill struct {
	SkillID     int64 // 技能ID
	Level       int   // 当前等级
	LastCast    int64 // 上次释放时间戳(毫秒)
	CooldownEnd int64 // 冷却结束时间戳(毫秒)
	IsUnlocked  bool  // 是否已解锁
}

type ComboSequence struct {
	SkillIDs    []int64       // 连招技能ID序列
	NextIndex   int           // 当前应释放的技能索引
	LastTime    int64         // 上一个技能释放时间戳
	MaxDelay    int           // 连招最大间隔时间(毫秒)
	BonusEffect *EffectConfig // 连招完成后的额外奖励效果
}

type SkillManager struct {
	skillConfigs   map[int64]*SkillConfig           // 技能配置表
	playerSkills   map[int64]map[int64]*PlayerSkill // 玩家技能状态（玩家ID -> 技能ID -> 技能状态）
	comboSequences map[int64]*ComboSequence         // 玩家连招序列（玩家ID -> 连招状态）
	timerManager   *timer.TimerManager              // 定时器管理器
	mu             sync.RWMutex                     // 并发访问互斥锁
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

// SkillCaster 技能释放者接口
type SkillCaster interface {
	GetID() int64                  // 获取释放者ID
	ConsumeMana(amount int32) bool // 消耗魔法值，返回是否成功
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
		m.ApplyEffectToTarget(casterID, targetID, effectResult, effectConfig, effectMultiplier)
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

func (m *SkillManager) ApplyEffectToTarget(casterID int64, targetID int64, effect *EffectResult, config *EffectConfig, multiplier float32) {
	switch effect.EffectType {
	case EffectTypeDamage:
		targetEntity := GetTargetEntity(targetID)
		if targetEntity != nil {
			damage := int32(effect.Value)
			defense := targetEntity.GetDefense()
			armorReduction := float32(defense) / (float32(defense) + 100)
			finalDamage := int32(float32(damage) * (1 - armorReduction))
			if finalDamage < 1 {
				finalDamage = 1
			}
			targetEntity.TakeDamage(finalDamage)
			TriggerSkillEffect(casterID, targetID, effect.EffectType, finalDamage)
		}
	case EffectTypeHeal:
		targetEntity := GetTargetEntity(targetID)
		if targetEntity != nil {
			healAmount := int32(effect.Value)
			targetEntity.Heal(healAmount)
			TriggerSkillEffect(casterID, targetID, effect.EffectType, healAmount)
		}
	case EffectTypeBuff:
		targetEntity := GetTargetEntity(targetID)
		if targetEntity != nil {
			buff := &BuffEffect{
				EffectID:   config.EffectID,
				EffectType: config.EffectType,
				Value:      config.Value * multiplier,
				Duration:   config.Duration,
				StackCount: 1,
				MaxStacks:  config.StackCount,
				StartTime:  time.Now().UnixNano() / 1e6,
				IsDebuff:   false,
			}
			targetEntity.AddBuffEffect(buff)
			TriggerSkillEffect(casterID, targetID, effect.EffectType, int32(config.Value))
		}
	case EffectTypeDebuff:
		targetEntity := GetTargetEntity(targetID)
		if targetEntity != nil {
			buff := &BuffEffect{
				EffectID:   config.EffectID,
				EffectType: config.EffectType,
				Value:      config.Value * multiplier,
				Duration:   config.Duration,
				StackCount: 1,
				MaxStacks:  config.StackCount,
				StartTime:  time.Now().UnixNano() / 1e6,
				IsDebuff:   true,
			}
			targetEntity.AddBuffEffect(buff)
			TriggerSkillEffect(casterID, targetID, effect.EffectType, int32(config.Value))
		}
	case EffectTypeDot:
		tickValue := config.Value * multiplier
		tickCount := config.Duration / config.TickInterval

		taskID := snowflake.GenerateID()
		tick := 0

		m.timerManager.AddTask(taskID, 0, time.Duration(config.TickInterval)*time.Millisecond, func() {
			tick++
			targetEntity := GetTargetEntity(targetID)
			if targetEntity != nil && !targetEntity.IsDead() {
				defense := targetEntity.GetDefense()
				armorReduction := float32(defense) / (float32(defense) + 100)
				finalDamage := int32(float32(tickValue) * (1 - armorReduction))
				if finalDamage < 1 {
					finalDamage = 1
				}
				targetEntity.TakeDamage(finalDamage)
				TriggerSkillEffect(casterID, targetID, effect.EffectType, finalDamage)
			}
			if tick >= tickCount {
				m.timerManager.RemoveTask(taskID)
			}
		})
	case EffectTypeHot:
		tickValue := config.Value * multiplier
		tickCount := config.Duration / config.TickInterval

		taskID := snowflake.GenerateID()
		tick := 0

		m.timerManager.AddTask(taskID, 0, time.Duration(config.TickInterval)*time.Millisecond, func() {
			tick++
			targetEntity := GetTargetEntity(targetID)
			if targetEntity != nil && !targetEntity.IsDead() {
				targetEntity.Heal(int32(tickValue))
				TriggerSkillEffect(casterID, targetID, effect.EffectType, int32(tickValue))
			}
			if tick >= tickCount {
				m.timerManager.RemoveTask(taskID)
			}
		})
	case EffectTypeStun:
		targetEntity := GetTargetEntity(targetID)
		if targetEntity != nil {
			buff := &BuffEffect{
				EffectID:   config.EffectID,
				EffectType: config.EffectType,
				Value:      config.Value,
				Duration:   config.Duration,
				StackCount: 1,
				MaxStacks:  1,
				StartTime:  time.Now().UnixNano() / 1e6,
				IsDebuff:   true,
			}
			targetEntity.AddBuffEffect(buff)
			TriggerSkillEffect(casterID, targetID, effect.EffectType, 0)
		}
	case EffectTypeSilence:
		targetEntity := GetTargetEntity(targetID)
		if targetEntity != nil {
			buff := &BuffEffect{
				EffectID:   config.EffectID,
				EffectType: config.EffectType,
				Value:      config.Value,
				Duration:   config.Duration,
				StackCount: 1,
				MaxStacks:  1,
				StartTime:  time.Now().UnixNano() / 1e6,
				IsDebuff:   true,
			}
			targetEntity.AddBuffEffect(buff)
			TriggerSkillEffect(casterID, targetID, effect.EffectType, 0)
		}
	case EffectTypeKnockback:
		TriggerSkillEffect(casterID, targetID, effect.EffectType, int32(config.Value))
	case EffectTypeSpeedBuff:
		targetEntity := GetTargetEntity(targetID)
		if targetEntity != nil {
			buff := &BuffEffect{
				EffectID:   config.EffectID,
				EffectType: config.EffectType,
				Value:      config.Value * multiplier,
				Duration:   config.Duration,
				StackCount: 1,
				MaxStacks:  config.StackCount,
				StartTime:  time.Now().UnixNano() / 1e6,
				IsDebuff:   false,
			}
			targetEntity.AddBuffEffect(buff)
			TriggerSkillEffect(casterID, targetID, effect.EffectType, int32(config.Value))
		}
	case EffectTypeDamageShield:
		targetEntity := GetTargetEntity(targetID)
		if targetEntity != nil {
			buff := &BuffEffect{
				EffectID:   config.EffectID,
				EffectType: config.EffectType,
				Value:      config.Value * multiplier,
				Duration:   config.Duration,
				StackCount: int(config.Value * multiplier),
				MaxStacks:  1,
				StartTime:  time.Now().UnixNano() / 1e6,
				IsDebuff:   false,
			}
			targetEntity.AddBuffEffect(buff)
			TriggerSkillEffect(casterID, targetID, effect.EffectType, int32(config.Value))
		}
	}
}

var effectTrigger func(casterID, targetID int64, effectType EffectType, value int32)

// SetEffectTrigger 设置技能效果触发回调（用于客户端显示特效）
func SetEffectTrigger(trigger func(casterID, targetID int64, effectType EffectType, value int32)) {
	effectTrigger = trigger
}

// TriggerSkillEffect 触发技能效果展示
func TriggerSkillEffect(casterID, targetID int64, effectType EffectType, value int32) {
	if effectTrigger != nil {
		effectTrigger(casterID, targetID, effectType, value)
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

// BuffEffect 增益/减益效果
type BuffEffect struct {
	EffectID   int64      // 效果ID
	EffectType EffectType // 效果类型
	Value      float32    // 效果数值
	Duration   int        // 持续时间(毫秒)
	StackCount int        // 当前堆叠层数
	MaxStacks  int        // 最大堆叠层数
	StartTime  int64      // 开始时间
	IsDebuff   bool       // 是否为减益效果
}

// EntityTarget 技能目标实体接口
type EntityTarget interface {
	TakeDamage(damage int32)                // 受到伤害
	Heal(amount int32)                      // 恢复生命值
	GetDefense() int32                      // 获取防御值
	GetMaxHealth() int32                    // 获取最大生命值
	IsDead() bool                           // 是否死亡
	AddBuffEffect(buff *BuffEffect)         // 添加增益效果
	RemoveBuffEffect(effectID int64)        // 移除增益效果
	HasBuffEffect(effectID int64) bool      // 是否有增益效果
	GetBuffEffectStacks(effectID int64) int // 获取增益堆叠层数
}
