package skill

// SkillResult 技能释放结果
type SkillResult struct {
	SkillID   int64           // 技能ID
	TargetID  int64           // 目标ID
	Effects   []*EffectResult // 效果结果列表
	IsCombo   bool            // 是否触发连招
	ComboName string          // 连招名称
}

// EffectResult 效果执行结果
type EffectResult struct {
	EffectID   int64      // 效果ID
	EffectType EffectType // 效果类型
	Value      float32    // 效果数值
	Duration   int        // 持续时间
}

// SkillInfo 技能信息
type SkillInfo struct {
	SkillID     int64      // 技能ID
	SkillName   string     // 技能名称
	SkillType   SkillType  // 技能类型
	SkillClass  SkillClass // 技能分类
	Level       int        // 当前等级
	MaxLevel    int        // 最高等级
	Cooldown    int        // 冷却时间
	CooldownEnd int64      // 冷却结束时间
	ManaCost    int        // 魔法消耗
	Range       float32    // 技能射程
	TargetType  TargetType // 目标类型
	Description string     // 技能描述
}

// PlayerSkillInfo 玩家技能信息
type PlayerSkillInfo struct {
	SkillID     int64  // 技能ID
	SkillName   string // 技能名称
	Level       int    // 当前等级
	MaxLevel    int    // 最高等级
	CooldownEnd int64  // 冷却结束时间
	IsUnlocked  bool   // 是否已解锁
}
